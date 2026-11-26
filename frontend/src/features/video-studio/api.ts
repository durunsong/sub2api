import { apiClient } from '@/api/client';
import type { AxiosRequestConfig } from 'axios';
import type { VideoConfig, VideoModel, VideoTaskStatus } from './types';
export interface VideoRequest {
    model_id: string;
    prompt: string;
    seconds: number;
    resolution: string;
    size: string;
    mode: 'frames' | 'reference';
    references: string[];
}
function task(value: unknown): VideoTaskStatus {
    const v = value as VideoTaskStatus;
    if (!v || !/^[A-Za-z0-9][A-Za-z0-9_-]{0,199}$/.test(v.id) || !['pending', 'completed', 'failed'].includes(v.status))
        throw new Error('videoStudio.invalidResponse');
    return { id: v.id, status: v.status, progress: typeof v.progress === 'number' && v.progress >= 0 && v.progress <= 100 ? v.progress : undefined };
}
export const videoStudioAPI = {
    catalog: async (signal?: AbortSignal): Promise<{
        enabled: boolean;
        models: VideoModel[];
    }> => (await apiClient.get('/image-studio/videos/catalog', { signal })).data,
    adminCatalog: async (): Promise<{
        enabled: boolean;
        models: VideoModel[];
    }> => (await apiClient.get('/admin/image-studio/videos/catalog')).data,
    config: async (): Promise<VideoConfig> => (await apiClient.get('/admin/image-studio/videos/config')).data,
    saveConfig: async (value: VideoConfig): Promise<VideoConfig> => (await apiClient.put('/admin/image-studio/videos/config', value)).data,
    generate: async (value: VideoRequest, signal: AbortSignal): Promise<VideoTaskStatus> => {
        const options: AxiosRequestConfig & {
            _retry: boolean;
        } = { signal, timeout: 180000, _retry: true };
        return task((await apiClient.post('/image-studio/videos/generate', value, options)).data);
    },
    status: async (group: number, id: string, signal: AbortSignal): Promise<VideoTaskStatus> => task((await apiClient.get('/image-studio/videos/status', { signal, timeout: 60000, params: { group_id: group, task_id: id } })).data),
    content: async (group: number, id: string, signal: AbortSignal): Promise<Blob> => {
        const blob = (await apiClient.get('/image-studio/videos/content', { signal, timeout: 180000, responseType: 'blob', params: { group_id: group, task_id: id } })).data;
        if (!(blob instanceof Blob) || !['video/mp4', 'video/webm'].includes(blob.type) || !blob.size || blob.size > 256 * 1024 * 1024)
            throw new Error('videoStudio.invalidVideo');
        return blob;
    }
};
