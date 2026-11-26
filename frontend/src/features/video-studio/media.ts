export interface VideoMetadata {
    width: number;
    height: number;
    videoDurationMs: number;
}
/** Read the actual local result, rather than treating requested size or generation time as output metadata. */
export function readVideoMetadata(blob: Blob, signal?: AbortSignal): Promise<VideoMetadata> {
    return new Promise((resolve, reject) => {
        const video = document.createElement('video');
        const url = URL.createObjectURL(blob);
        const cleanup = () => {
            clearTimeout(timer);
            signal?.removeEventListener('abort', abort);
            video.onloadedmetadata = video.onerror = null;
            video.removeAttribute('src');
            video.load();
            URL.revokeObjectURL(url);
        };
        const fail = () => { cleanup(); reject(new Error('videoStudio.invalidVideo')); };
        const abort = () => { cleanup(); reject(new DOMException('Aborted', 'AbortError')); };
        const timer = setTimeout(fail, 15000);
        video.onloadedmetadata = () => {
            const result = { width: video.videoWidth, height: video.videoHeight, videoDurationMs: Math.round(video.duration * 1000) };
            if (!result.width || !result.height || !Number.isFinite(result.videoDurationMs) || result.videoDurationMs <= 0) {
                fail();
                return;
            }
            cleanup();
            resolve(result);
        };
        video.onerror = fail;
        signal?.addEventListener('abort', abort, { once: true });
        if (signal?.aborted) {
            abort();
            return;
        }
        video.preload = 'metadata';
        video.src = url;
    });
}

export function videoFilename(id: string, blob: Blob): string {
  return `video-${id}.${blob.type === 'video/webm' ? 'webm' : 'mp4'}`
}
