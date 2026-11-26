import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '@/api/client';
import { videoStudioAPI } from './api';
vi.mock('@/api/client', () => ({ apiClient: { post: vi.fn(), get: vi.fn(), put: vi.fn() } }));
beforeEach(() => vi.clearAllMocks());
describe('video request boundary', () => {
    it('never replays a billed create via the auth interceptor', async () => {
        vi.mocked(apiClient.post).mockResolvedValue({ data: { id: 'task-1', status: 'pending' } });
        const signal = new AbortController().signal;
        const result = await videoStudioAPI.generate({ model_id: '1:m', prompt: 'pan', seconds: 6, resolution: '720', size: 'auto', mode: 'frames', references: [] }, signal);
        expect(result.id).toBe('task-1');
        expect(apiClient.post).toHaveBeenCalledWith('/image-studio/videos/generate', expect.anything(), expect.objectContaining({ _retry: true, signal }));
    });
    it('rejects malformed status and non-video downloads', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { id: '../task', status: 'completed' } });
        await expect(videoStudioAPI.status(1, 'safe', new AbortController().signal)).rejects.toThrow();
        vi.mocked(apiClient.get).mockResolvedValue({ data: new Blob(['error'], { type: 'text/html' }) });
        await expect(videoStudioAPI.content(1, 'safe', new AbortController().signal)).rejects.toThrow();
    });
});
