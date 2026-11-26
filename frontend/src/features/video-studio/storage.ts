import type { VideoAsset, VideoDraft, VideoPrompt, VideoRecord } from './types';
type Rows = {
    records: VideoRecord;
    drafts: VideoDraft;
    prompts: VideoPrompt;
    assets: VideoAsset;
};
type Store = keyof Rows;
export function createVideoStudioDB(userID: number) {
    if (!Number.isSafeInteger(userID) || userID <= 0)
        throw new Error('videoStudio.storageFailed');
    let pending: Promise<IDBDatabase> | undefined, closed = false;
    function open(): Promise<IDBDatabase> {
        if (closed)
            return Promise.reject(new Error('videoStudio.storageFailed'));
        if (!pending)
            pending = new Promise<IDBDatabase>((resolve, reject) => {
                if (typeof indexedDB === 'undefined') {
                    reject(new Error('videoStudio.storageFailed'));
                    return;
                }
                const request = indexedDB.open(`sub2api-video-studio-${userID}`, 1);
                request.onupgradeneeded = () => { for (const name of ['records', 'drafts', 'prompts', 'assets'])
                    if (!request.result.objectStoreNames.contains(name))
                        request.result.createObjectStore(name, { keyPath: 'id' }); };
                request.onerror = () => reject(new Error('videoStudio.storageFailed'));
                request.onblocked = () => reject(new Error('videoStudio.storageBlocked'));
                request.onsuccess = () => { if (closed) {
                    request.result.close();
                    reject(new Error('videoStudio.storageFailed'));
                    return;
                } request.result.onversionchange = () => { request.result.close(); pending = undefined; }; resolve(request.result); };
            }).catch(error => { pending = undefined; throw error; });
        return pending;
    }
    async function transaction<T>(names: Store[], mode: IDBTransactionMode, work: (tx: IDBTransaction, result: (value: T) => void) => void): Promise<T> {
        const db = await open();
        return new Promise((resolve, reject) => {
            const tx = db.transaction(names, mode);
            let result: T;
            tx.oncomplete = () => resolve(result);
            tx.onabort = () => reject(new Error(tx.error?.name === 'QuotaExceededError' ? 'videoStudio.storageQuota' : 'videoStudio.storageFailed'));
            tx.onerror = () => { };
            try {
                work(tx, value => { result = value; });
            }
            catch (error) {
                tx.abort();
                reject(error);
            }
        });
    }
    return {
        list: <K extends Store>(name: K) => transaction<Rows[K][]>([name], 'readonly', (tx, set) => { const req = tx.objectStore(name).getAll(); req.onsuccess = () => set(req.result); }),
        put: <K extends Store>(name: K, row: Rows[K]) => transaction<void>([name], 'readwrite', tx => { tx.objectStore(name).put(row); }),
        remove: (name: Store, id: string) => transaction<void>([name], 'readwrite', tx => { tx.objectStore(name).delete(id); }),
        updateRecord: (id: string, generationId: string, patch: Partial<VideoRecord>) => transaction<boolean>(['records'], 'readwrite', (tx, set) => {
            const store = tx.objectStore('records'), req = store.get(id);
            req.onsuccess = () => { const row = req.result as VideoRecord | undefined; if (!row || row.generationId !== generationId) {
                set(false);
                return;
            } store.put({ ...row, ...patch, id: row.id, generationId: row.generationId, sessionId: row.sessionId }); set(true); };
        }),
        deleteSession: (sessionId: string) => transaction<void>(['records'], 'readwrite', tx => {
            const req = tx.objectStore('records').openCursor();
            req.onsuccess = () => { const cursor = req.result; if (!cursor)
                return; if (cursor.value.sessionId === sessionId)
                cursor.delete(); cursor.continue(); };
        }),
        close: () => { closed = true; void pending?.then(db => db.close()).catch(() => { }); }
    };
}
export type VideoStudioDB = ReturnType<typeof createVideoStudioDB>;
