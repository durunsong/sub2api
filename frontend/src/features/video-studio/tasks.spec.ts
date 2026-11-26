import { describe, expect, it, vi } from 'vitest';
import { followVideo } from './tasks';
import type { VideoRecord } from './types';
vi.mock('./media', () => ({ readVideoMetadata: async () => ({ width: 320, height: 180, videoDurationMs: 1000 }) }));
const record: VideoRecord = { id: 'r', generationId: 'g', sessionId: 's', createdAt: 1, prompt: 'pan', model: { id: '1:m', group_id: 1, group_name: 'A', model: 'm', label: 'M', protocol: 'grok' }, parameters: { seconds: 6, resolution: '720', size: 'auto', mode: 'frames' }, references: [], status: 'pending', taskID: 'task' };
describe('video task recovery', () => {
    it('resumes status and saves output without another create', async () => {
        const blob = new Blob(['video'], { type: 'video/mp4' }), updates: Partial<VideoRecord>[] = [];
        const result = await followVideo(record, { signal: new AbortController().signal, status: async () => ({ id: 'task', status: 'completed' }), content: async () => blob, update: async (patch) => { updates.push(patch); return true; }, onChange: () => { } });
        expect(result?.video).toBe(blob);
        expect(updates.at(-1)).toMatchObject({ status: 'success', video: blob, width: 320, height: 180, videoDurationMs: 1000 });
    });
    it('does not fetch output or resurrect a deleted record', async () => {
        const content = vi.fn();
        const result = await followVideo(record, { signal: new AbortController().signal, status: async () => ({ id: 'task', status: 'completed' }), content, update: async () => false, onChange: () => { } });
        expect(result).toBeUndefined();
        expect(content).not.toHaveBeenCalled();
    });
    it('returns unsaved output when local quota fails after download', async () => {
        const blob = new Blob(['video'], { type: 'video/mp4' });
        const result = await followVideo(record, { signal: new AbortController().signal, status: async () => ({ id: 'task', status: 'completed' }), content: async () => blob, update: async (patch) => { if (patch.video)
                throw new Error('quota'); return true; }, onChange: () => { } });
        expect(result).toMatchObject({ video: blob, unsaved: true });
    });
});
