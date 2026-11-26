import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { imageToDataURL } from '@/features/image-studio/api';
import { videoStudioAPI } from './api';
import { createVideoStudioDB, type VideoStudioDB } from './storage';
import { groupSessions, normalizeParameters } from './parameters';
import { followVideo } from './tasks';
import type { VideoAsset, VideoDraft, VideoModel, VideoParameters, VideoPrompt, VideoRecord, VideoReference } from './types';
export function useVideoStudio() {
    const auth = useAuthStore(), models = ref<VideoModel[]>([]), enabled = ref(true), loading = ref(true), error = ref(''), storageError = ref(''), notice = ref('');
    const records = shallowRef<VideoRecord[]>([]), assets = shallowRef<VideoAsset[]>([]), prompts = shallowRef<VideoPrompt[]>([]);
    const sessionID = ref<string>(crypto.randomUUID()), modelID = ref(''), prompt = ref(''), parameters = ref<VideoParameters>(normalizeParameters()), references = shallowRef<VideoReference[]>([]);
    const active = ref<string[]>([]), unsaved = shallowRef<Record<string, VideoRecord>>({});
    let db: VideoStudioDB | undefined, epoch = 0, ready = false, closed = false, saveTimer: ReturnType<typeof setTimeout> | undefined, noticeTimer: ReturnType<typeof setTimeout> | undefined;
    let loadedUserID: number | undefined;
    let refreshVersion = 0;
    let channel: BroadcastChannel | undefined, loadAbort = new AbortController();
    let saveQueue: Promise<unknown> = Promise.resolve();
    const controllers = new Map<string, AbortController>();
    const selectedModel = computed(() => models.value.find(m => m.id === modelID.value));
    const entries = computed(() => records.value.filter(row => row.sessionId === sessionID.value).sort((a, b) => a.createdAt - b.createdAt));
    const sessions = computed(() => groupSessions(records.value));
    const currentRunning = computed(() => records.value.some(r => r.sessionId === sessionID.value && active.value.includes(r.id)));
    const referenceLimit = computed(() => parameters.value.mode === 'frames' ? 2 : 9);
    const canGenerate = computed(() => ready && !loading.value && enabled.value && !!selectedModel.value && !!prompt.value.trim() && !currentRunning.value && references.value.length <= referenceLimit.value);
    function notify(message: string) { notice.value = message; clearTimeout(noticeTimer); noticeTimer = setTimeout(() => { notice.value = ''; }, 5000); }
    function announce() { channel?.postMessage('refresh'); }
    async function refresh() {
        const local = db, version = epoch, request = ++refreshVersion;
        if (!local)
            return;
        try {
            const [rows, a, p] = await Promise.all([local.list('records'), local.list('assets'), local.list('prompts')]);
            if (version !== epoch || request !== refreshVersion)
                return;
            records.value = rows.map(row => unsaved.value[row.id] || row);
            assets.value = a;
            prompts.value = p;
        }
        catch {
            if (version === epoch)
                storageError.value = 'videoStudio.storageFailed';
        }
    }
    function rawReferences() { return references.value.map(r => ({ ...r, blob: r.blob })); }
    function saveDraft() {
        if (!ready || !db)
            return;
        const local = db, version = epoch;
        const draft: VideoDraft = { id: 'current', sessionId: sessionID.value, modelID: modelID.value, prompt: prompt.value, parameters: { ...parameters.value }, references: rawReferences() };
        saveQueue = saveQueue.catch(() => { }).then(() => local.put('drafts', draft)).catch(() => { if (version === epoch)
            storageError.value = 'videoStudio.storageFailed'; });
    }
    watch([sessionID, modelID, prompt, parameters, references], () => { clearTimeout(saveTimer); if (ready)
        saveTimer = setTimeout(saveDraft, 250); }, { deep: true });
    function stop(id: string) { controllers.get(id)?.abort(); }
    async function resume(row: VideoRecord) {
        if (!db || !row.taskID || controllers.has(row.id))
            return;
        const local = db, version = epoch, abort = new AbortController();
        controllers.set(row.id, abort);
        active.value = [...controllers.keys()];
        try {
            const result = await followVideo(row, { signal: abort.signal, status: () => videoStudioAPI.status(row.model.group_id, row.taskID!, abort.signal), content: () => videoStudioAPI.content(row.model.group_id, row.taskID!, abort.signal), update: patch => local.updateRecord(row.id, row.generationId, patch), onChange: () => { if (version === epoch) {
                    void refresh();
                    announce();
                } } });
            if (version !== epoch || abort.signal.aborted)
                return;
            if (result?.unsaved) {
                unsaved.value = { ...unsaved.value, [row.id]: { ...row, status: 'success', video: result.video, durationMs: result.durationMs, width: result.width, height: result.height, videoDurationMs: result.videoDurationMs } };
                storageError.value = 'videoStudio.unsaved';
                await refresh();
            }
            else if (result)
                notify('videoStudio.completed');
        }
        catch (err) {
            if (version !== epoch)
                return;
            const message = abort.signal.aborted ? 'videoStudio.paused' : 'videoStudio.lookupFailed';
            try {
                await local.updateRecord(row.id, row.generationId, { error: message });
            }
            catch {
                storageError.value = 'videoStudio.storageFailed';
            }
        }
        finally {
            if (version === epoch) {
                await refresh();
                if (version === epoch) {
                    controllers.delete(row.id);
                    active.value = [...controllers.keys()];
                    announce();
                }
            }
        }
    }
    async function submit() {
        if (!canGenerate.value || !db || !selectedModel.value)
            return;
        const local = db, version = epoch, abort = new AbortController();
        const row: VideoRecord = { id: crypto.randomUUID(), generationId: crypto.randomUUID(), sessionId: sessionID.value, createdAt: Date.now(), prompt: prompt.value, model: { ...selectedModel.value }, parameters: { ...parameters.value }, references: rawReferences(), status: 'pending' };
        controllers.set(row.id, abort);
        active.value = [...controllers.keys()];
        refreshVersion++;
        records.value = [...records.value, row];
        error.value = '';
        let created = false, submitted = false, saved = false;
        try {
            // Saving the immutable request first is required for safe refresh/recovery.
            await local.put('records', row);
            saved = true;
            announce();
            const refs = await Promise.all(row.references.map(r => imageToDataURL(r.blob)));
            abort.signal.throwIfAborted();
            submitted = true;
            const task = await videoStudioAPI.generate({ model_id: row.model.id, prompt: row.prompt, ...row.parameters, references: refs }, abort.signal);
            created = true;
            row.taskID = task.id;
            if (version !== epoch)
                return;
            if (!await local.updateRecord(row.id, row.generationId, { taskID: task.id })) {
                return;
            }
            await refresh();
            if (version !== epoch || abort.signal.aborted)
                return;
            announce();
            // Hand off synchronously: never render a paused state between submit and polling.
            controllers.delete(row.id);
            void resume(row);
        }
        catch {
            if (version !== epoch)
                return;
            if (created) {
                unsaved.value = { ...unsaved.value, [row.id]: row };
                storageError.value = 'videoStudio.unsavedTask';
                records.value = records.value.map(r => r.id === row.id ? row : r);
            }
            else if (!saved) {
                storageError.value = 'videoStudio.storageFailed';
            }
            else {
                const message = submitted ? 'videoStudio.submitUncertain' : 'videoStudio.invalidReference';
                try {
                    await local.updateRecord(row.id, row.generationId, { status: 'interrupted', error: message });
                }
                catch {
                    storageError.value = 'videoStudio.storageFailed';
                }
                error.value = message;
            }
        }
        finally {
            if (version === epoch && controllers.get(row.id) === abort) {
                await refresh();
                if (version === epoch) {
                    controllers.delete(row.id);
                    active.value = [...controllers.keys()];
                    announce();
                }
            }
        }
    }
    async function saveUnsaved(row: VideoRecord) {
        if (!db)
            return;
        const local = db, version = epoch;
        try {
            const saved = await local.updateRecord(row.id, row.generationId, { video: row.video, status: row.status, taskID: row.taskID, durationMs: row.durationMs, width: row.width, height: row.height, videoDurationMs: row.videoDurationMs, error: undefined });
            if (version !== epoch)
                return;
            if (saved) {
                const next = { ...unsaved.value };
                delete next[row.id];
                unsaved.value = next;
                storageError.value = '';
                await refresh();
                announce();
                if (version === epoch && row.taskID && !row.video)
                    void resume(row);
            }
            else {
                const next = { ...unsaved.value };
                delete next[row.id];
                unsaved.value = next;
                await refresh();
            }
        }
        catch {
            if (version !== epoch)
                return;
            storageError.value = 'videoStudio.storageFailed';
        }
    }
    async function addReferences(files: File[]) {
        const version = epoch;
        try {
            const prepared: VideoReference[] = [];
            for (const file of files) {
                if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type) || !file.size || file.size > 10 * 1024 * 1024)
                    throw new Error('invalid');
                const bitmap = await createImageBitmap(file);
                if (!bitmap.width || !bitmap.height) {
                    bitmap.close();
                    throw new Error('invalid');
                }
                bitmap.close();
                prepared.push({ id: crypto.randomUUID(), name: file.name, blob: file });
            }
            if (version !== epoch)
                return;
            if (references.value.length + prepared.length > referenceLimit.value || [...references.value, ...prepared].reduce((sum, r) => sum + r.blob.size, 0) > 30 * 1024 * 1024) {
                error.value = 'videoStudio.referenceLimit';
                return;
            }
            references.value = [...references.value, ...prepared];
            error.value = '';
        }
        catch {
            if (version === epoch)
                error.value = 'videoStudio.invalidReference';
        }
    }
    function removeReference(id: string) { references.value = references.value.filter(r => r.id !== id); }
    function newSession() { sessionID.value = crypto.randomUUID(); prompt.value = ''; references.value = []; error.value = ''; }
    function openSession(id: string) { sessionID.value = id; error.value = ''; }
    function reuse(row: VideoRecord) { prompt.value = row.prompt; modelID.value = row.model.id; parameters.value = { ...row.parameters }; references.value = row.references.map(r => ({ ...r })); }
    async function deleteSession(id: string) {
        const local = db, version = epoch;
        if (!local)
            return;
        try {
            await local.deleteSession(id);
            if (version !== epoch)
                return;
            for (const row of records.value.filter(r => r.sessionId === id)) {
                stop(row.id);
                const next = { ...unsaved.value };
                delete next[row.id];
                unsaved.value = next;
            }
            await refresh();
            announce();
            if (sessionID.value === id)
                newSession();
        }
        catch {
            if (version === epoch)
                storageError.value = 'videoStudio.storageFailed';
        }
    }
    async function savePrompt(title: string, text: string) { if (!db || !title.trim() || !text.trim())
        return; const local = db, version = epoch; try {
        await local.put('prompts', { id: crypto.randomUUID(), title: title.trim().slice(0, 120), prompt: text.slice(0, 32000) });
        if (version !== epoch)
            return;
        await refresh();
        announce();
    }
    catch {
        if (version === epoch)
            storageError.value = 'videoStudio.storageFailed';
    } }
    async function saveAsset(row: VideoRecord) { if (!db || !row.video)
        return; const local = db, version = epoch; try {
        await local.put('assets', { id: row.id, name: row.prompt.slice(0, 100), kind: 'video', blob: row.video, createdAt: row.createdAt });
        if (version !== epoch)
            return;
        await refresh();
        announce();
        notify('videoStudio.assetSaved');
    }
    catch {
        if (version === epoch)
            storageError.value = 'videoStudio.storageFailed';
    } }
    async function uploadAssets(files: File[]) {
        const local = db, version = epoch;
        if (!local)
            return;
        try {
            for (const file of files) {
                if (!['image/png', 'image/jpeg', 'image/webp', 'video/mp4', 'video/webm'].includes(file.type) || !file.size || file.size > (file.type.startsWith('image/') ? 10 : 256) * 1024 * 1024)
                    throw new Error('invalid');
                await local.put('assets', { id: crypto.randomUUID(), name: file.name, kind: file.type.startsWith('image/') ? 'image' : 'video', blob: file, createdAt: Date.now() });
            }
            if (version === epoch) {
                await refresh();
                announce();
            }
        }
        catch {
            if (version === epoch)
                storageError.value = 'videoStudio.storageFailed';
        }
    }
    async function removeLibrary(kind: 'assets' | 'prompts', id: string) { if (!db)
        return; const local = db, version = epoch; try {
        await local.remove(kind, id);
        if (version !== epoch)
            return;
        await refresh();
        announce();
    }
    catch {
        if (version === epoch)
            storageError.value = 'videoStudio.storageFailed';
    } }
    async function load() {
        const userID = auth.user?.id;
        const preserveUnsaved = loadedUserID === userID && !!userID;
        loadedUserID = userID;
        clearTimeout(saveTimer);
        saveDraft();
        const previous = db;
        const queued = saveQueue;
        ready = false;
        epoch++;
        const version = epoch;
        for (const c of controllers.values())
            c.abort();
        controllers.clear();
        active.value = [];
        loadAbort.abort();
        loadAbort = new AbortController();
        channel?.close();
        void queued.finally(() => previous?.close());
        db = undefined;
        records.value = [];
        assets.value = [];
        prompts.value = [];
        if (!preserveUnsaved) unsaved.value = {};
        models.value = [];
        modelID.value = '';
        prompt.value = '';
        references.value = [];
        parameters.value = normalizeParameters();
        sessionID.value = crypto.randomUUID();
        error.value = '';
        storageError.value = Object.values(unsaved.value).some(row => row.video) ? 'videoStudio.unsaved' : Object.keys(unsaved.value).length ? 'videoStudio.unsavedTask' : '';
        loading.value = true;
        if (!userID) {
            loading.value = false;
            return;
        }
        db = createVideoStudioDB(userID);
        const local = db;
        if (typeof BroadcastChannel !== 'undefined') {
            channel = new BroadcastChannel(`sub2api-video-studio-${userID}`);
            channel.onmessage = () => { void refresh(); };
        }
        const catalogRequest = videoStudioAPI.catalog(loadAbort.signal).then(catalog => {
            if (version !== epoch)
                return;
            models.value = catalog.models;
            enabled.value = catalog.enabled;
            if (!modelID.value)
                modelID.value = models.value[0]?.id || '';
        }).catch(() => { if (version === epoch)
            error.value = 'videoStudio.loadFailed'; });
        try {
            await queued;
            if (version !== epoch)
                return;
            const drafts = await local.list('drafts');
            if (version !== epoch)
                return;
            const draft = drafts.find(d => d.id === 'current');
            if (draft) {
                sessionID.value = draft.sessionId;
                modelID.value = draft.modelID;
                prompt.value = draft.prompt;
                parameters.value = normalizeParameters(draft.parameters);
                references.value = draft.references;
            }
            await refresh();
            if (version !== epoch)
                return;
            ready = true;
            for (const row of records.value) {
                if (row.status === 'pending' && row.taskID && !unsaved.value[row.id])
                    void resume(row);
            }
        }
        catch {
            if (version === epoch)
                storageError.value = 'videoStudio.storageFailed';
        }
        finally {
            if (version === epoch)
                loading.value = false;
        }
        await catalogRequest;
        if (version === epoch && !modelID.value)
            modelID.value = models.value[0]?.id || '';
    }
    const focus = () => { if (!closed)
        void refresh(); };
    window.addEventListener('focus', focus);
    watch(() => auth.user?.id, () => { void load(); }, { immediate: true });
    onBeforeUnmount(() => { saveDraft(); closed = true; epoch++; ready = false; clearTimeout(saveTimer); clearTimeout(noticeTimer); loadAbort.abort(); for (const c of controllers.values())
        c.abort(); channel?.close(); window.removeEventListener('focus', focus); const local = db; void saveQueue.finally(() => local?.close()); });
    return { models, enabled, loading, error, storageError, notice, records, assets, prompts, sessionID, modelID, prompt, parameters, references, active, unsaved, entries, sessions, selectedModel, currentRunning, canGenerate, referenceLimit, load, submit, resume, stop, saveUnsaved, addReferences, removeReference, newSession, openSession, reuse, deleteSession, savePrompt, saveAsset, uploadAssets, removeLibrary, refresh };
}
