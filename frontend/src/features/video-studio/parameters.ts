// Adapted from infinite-canvas media-size.ts (MIT); see docs/VIDEO_STUDIO_UPSTREAM.md.
import type { VideoParameters } from './types';
export const videoRatios = ['1:1', '3:4', '4:3', '16:9', '9:16', '21:9', 'auto'];
export function computeVideoSize(resolution: string, ratio: string): string {
    if (ratio === 'auto')
        return 'auto';
    const [w, h] = ratio.split(':').map(Number);
    if (!w || !h)
        return 'auto';
    const p = Number(resolution.replace(/p$/, '')) || 720;
    return w >= h ? `${Math.round(p * w / h)}x${p}` : `${p}x${Math.round(p * h / w)}`;
}
export function inferVideoRatio(size: string): string {
    if (size === 'auto')
        return 'auto';
    if (videoRatios.includes(size))
        return size;
    const [w, h] = size.split('x').map(Number);
    if (!w || !h)
        return '16:9';
    return videoRatios.slice(0, -1).reduce((best, ratio) => {
        const [a, b] = ratio.split(':').map(Number), [c, d] = best.split(':').map(Number);
        return Math.abs(a / b - w / h) < Math.abs(c / d - w / h) ? ratio : best;
    }, '16:9');
}
export function normalizeParameters(input: {
    seconds?: number;
    resolution?: string;
    size?: string;
    mode?: string;
} = {}): VideoParameters {
    const p = String(input.resolution || '720').replace(/p$/, '');
    const resolution = /^\d+$/.test(p) && Number(p) > 0 && Number(p) <= 4320 ? p : '720';
    const dimensions = /^(\d+)x(\d+)$/.exec(input.size || '');
    const size = dimensions && dimensions.slice(1).every(v => Number(v) > 0 && Number(v) <= 8192) ? input.size! : 'auto';
    return { seconds: Math.max(4, Math.min(30, Math.floor(input.seconds || 6))), resolution, size, mode: input.mode === 'reference' ? 'reference' : 'frames' };
}
export function normalizeModelParameters(input: VideoParameters, protocol?: string): VideoParameters {
    if (protocol !== 'muse') return input;
    const resolution = ['480', '720', '1080'].includes(input.resolution) ? input.resolution : '720';
    const ratio = inferVideoRatio(input.size);
    const size = ratio === 'auto' ? 'auto' : computeVideoSize(resolution, ratio === '9:16' ? '9:16' : '16:9');
    return { ...input, seconds: Math.min(15, input.seconds), mode: 'reference', resolution, size };
}
export function groupSessions(records: {
    id: string;
    sessionId: string;
    prompt: string;
    createdAt: number;
}[]) {
    const sessions = new Map<string, {
        id: string;
        title: string;
        count: number;
        updatedAt: number;
    }>();
    for (const row of [...records].sort((a, b) => a.createdAt - b.createdAt)) {
        const session = sessions.get(row.sessionId) || { id: row.sessionId, title: row.prompt.slice(0, 60), count: 0, updatedAt: 0 };
        session.count++;
        session.updatedAt = Math.max(session.updatedAt, row.createdAt);
        sessions.set(session.id, session);
    }
    return [...sessions.values()].sort((a, b) => b.updatedAt - a.updatedAt);
}
