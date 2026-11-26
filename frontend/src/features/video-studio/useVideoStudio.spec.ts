import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, reactive } from 'vue';
import { beforeEach, expect, it, vi } from 'vitest';
import { useVideoStudio } from './useVideoStudio';
import type { VideoRecord } from './types';
const mocks = vi.hoisted(() => ({
    auth: { user: { id: 7 } },
    catalog: vi.fn(), generate: vi.fn(), put: vi.fn(), updateRecord: vi.fn(),
    list: vi.fn(), close: vi.fn()
}));
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.auth }));
vi.mock('./api', () => ({ videoStudioAPI: { catalog: mocks.catalog, generate: mocks.generate } }));
vi.mock('./storage', () => ({ createVideoStudioDB: () => mocks }));
vi.mock('@/features/image-studio/api', () => ({ imageToDataURL: vi.fn() }));
beforeEach(() => {
    vi.resetAllMocks();
    mocks.auth = reactive({ user: { id: 7 } });
    vi.stubGlobal('BroadcastChannel', undefined);
    mocks.catalog.mockResolvedValue({ enabled: true, models: [{ id: '7:video', group_id: 7, model: 'video' }] });
    mocks.list.mockResolvedValue([]);
    mocks.put.mockResolvedValue(undefined);
});
function start() {
    let studio!: ReturnType<typeof useVideoStudio>;
    const wrapper = mount(defineComponent({ setup() { studio = useVideoStudio(); return () => null; } }));
    return { studio, wrapper };
}
it('keeps local history and drafts available when the model catalog is offline', async () => {
    mocks.catalog.mockRejectedValue(new Error('offline'));
    mocks.list.mockImplementation(async (name: string) => name === 'drafts'
        ? [{ id: 'current', sessionId: 'local', prompt: 'my local prompt', parameters: {}, references: [] }]
        : name === 'records' ? [{ id: 'record', sessionId: 'local', status: 'success', createdAt: 1 }] : []);
    const { studio, wrapper } = start();
    await flushPromises();
    expect(studio.prompt.value).toBe('my local prompt');
    expect(studio.records.value).toHaveLength(1);
    expect(studio.error.value).toBe('videoStudio.loadFailed');
    expect(studio.storageError.value).toBe('');
    wrapper.unmount();
});
it('never submits paid work if the initial local request cannot be saved', async () => {
    const { studio, wrapper } = start();
    await flushPromises();
    studio.prompt.value = 'a landscape';
    mocks.put.mockRejectedValue(new Error('QuotaExceededError'));
    await studio.submit();
    expect(mocks.generate).not.toHaveBeenCalled();
    expect(studio.storageError.value).toBe('videoStudio.storageFailed');
    expect(studio.error.value).not.toBe('videoStudio.submitUncertain');
    wrapper.unmount();
});

it.each(['task', 'blob'])('keeps unsaved %s recovery data through a same-user refresh', async kind => {
    const stored = { id: 'r', generationId: 'g', sessionId: 's', status: 'pending', prompt: 'local', createdAt: 1, references: [], parameters: {}, model: {} } as VideoRecord;
    mocks.list.mockImplementation(async (name: string) => name === 'records' ? [stored] : []);
    const { studio, wrapper } = start();
    await flushPromises();
    const pending = { ...stored, taskID: 'paid-task', ...(kind === 'blob' ? { status: 'success' as const, video: new Blob(['output']) } : {}) };
    studio.unsaved.value = { r: pending };
    await studio.load();
    await flushPromises();
    expect(studio.unsaved.value.r?.taskID).toBe('paid-task');
    expect(studio.records.value[0].taskID).toBe('paid-task');
    if (kind === 'blob') expect(studio.records.value[0].video).toBe(pending.video);
    expect(mocks.generate).not.toHaveBeenCalled();
    wrapper.unmount();
});

it('clears unsaved private results when the signed-in account changes', async () => {
    const { studio, wrapper } = start();
    await flushPromises();
    studio.unsaved.value = { r: { id: 'r', taskID: 'private-task', video: new Blob(['private']) } as VideoRecord };
    mocks.auth.user.id = 8;
    await flushPromises();
    expect(studio.unsaved.value).toEqual({});
    expect(studio.records.value).toEqual([]);
    wrapper.unmount();
});

it('does not mutate the new account when an old unsaved result finishes saving', async () => {
    const { studio, wrapper } = start();
    await flushPromises();
    let resolveSave!: (value: boolean) => void;
    mocks.updateRecord.mockImplementationOnce(() => new Promise<boolean>(resolve => { resolveSave = resolve; }));
    const row = { id: 'old', generationId: 'g', taskID: 'old-task', video: new Blob(['old']) } as VideoRecord;
    studio.unsaved.value = { old: row };
    const saving = studio.saveUnsaved(row);
    mocks.auth.user.id = 8;
    await flushPromises();
    studio.storageError.value = 'new-account-error';
    resolveSave(true);
    await saving;
    expect(studio.storageError.value).toBe('new-account-error');
    expect(studio.unsaved.value).toEqual({});
    wrapper.unmount();
});
