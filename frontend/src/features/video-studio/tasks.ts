import type { VideoRecord, VideoTaskStatus } from './types';
import { readVideoMetadata, type VideoMetadata } from './media';
interface TaskDependencies {
    signal: AbortSignal;
    status: () => Promise<VideoTaskStatus>;
    content: () => Promise<Blob>;
    update: (patch: Partial<VideoRecord>) => Promise<boolean>;
    onChange: () => void;
}
function delay(signal: AbortSignal): Promise<void> {
    return new Promise((resolve, reject) => {
        const abort = () => { clearTimeout(timer); reject(new DOMException('Aborted', 'AbortError')); };
        const timer = setTimeout(() => { signal.removeEventListener('abort', abort); resolve(); }, 5000);
        signal.addEventListener('abort', abort, { once: true });
        if (signal.aborted)
            abort();
    });
}
export async function followVideo(record: VideoRecord, deps: TaskDependencies): Promise<VideoMetadata & {
    video: Blob;
    durationMs: number;
    unsaved: boolean;
} | undefined> {
    for (let i = 0; i < 120; i++) {
        deps.signal.throwIfAborted();
        const task = await deps.status();
        deps.signal.throwIfAborted();
        if (!await deps.update({ progress: task.progress, error: undefined }))
            return;
        deps.onChange();
        if (task.status === 'failed') {
            await deps.update({ status: 'failed', error: 'videoStudio.generationFailed' });
            deps.onChange();
            return;
        }
        if (task.status === 'completed') {
            const video = await deps.content();
            deps.signal.throwIfAborted();
            const metadata = await readVideoMetadata(video, deps.signal);
            deps.signal.throwIfAborted();
            const durationMs = Date.now() - record.createdAt;
            try {
                if (!await deps.update({ status: 'success', video, durationMs, ...metadata, error: undefined }))
                    return;
            }
            catch {
                return { video, durationMs, ...metadata, unsaved: true };
            }
            deps.onChange();
            return { video, durationMs, ...metadata, unsaved: false };
        }
        await delay(deps.signal);
    }
    await deps.update({ error: 'videoStudio.pollTimeout' });
    deps.onChange();
}
