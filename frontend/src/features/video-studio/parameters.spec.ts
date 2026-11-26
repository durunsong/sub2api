import { describe, expect, it } from 'vitest';
import { computeVideoSize, inferVideoRatio, normalizeParameters, groupSessions } from './parameters';
describe('video parameters and sessions', () => {
    it('matches upstream short-edge sizing and automatic ratio', () => {
        expect(computeVideoSize('720', '16:9')).toBe('1280x720');
        expect(computeVideoSize('1080', '9:16')).toBe('1080x1920');
        expect(computeVideoSize('480', 'auto')).toBe('auto');
        expect(inferVideoRatio('1080x1920')).toBe('9:16');
    });
    it('bounds input and preserves independent runtime options', () => {
        expect(normalizeParameters({ seconds: 100, resolution: '1080p', size: 'auto', mode: 'reference' })).toEqual({ seconds: 30, resolution: '1080', size: 'auto', mode: 'reference' });
        expect(normalizeParameters({ seconds: 0, resolution: 'bad', size: 'bad', mode: 'bad' })).toEqual({ seconds: 6, resolution: '720', size: 'auto', mode: 'frames' });
    });
    it('keeps separate conversations and sorts by latest generation', () => {
        expect(groupSessions([{ id: 'a', sessionId: 's1', prompt: 'First', createdAt: 1 }, { id: 'b', sessionId: 's2', prompt: 'Second', createdAt: 2 }, { id: 'c', sessionId: 's1', prompt: 'Follow-up', createdAt: 3 }])).toEqual([{ id: 's1', title: 'First', count: 2, updatedAt: 3 }, { id: 's2', title: 'Second', count: 1, updatedAt: 2 }]);
    });
});
