export interface VideoModel {
    id: string;
    group_id: number;
    group_name: string;
    model: string;
    label: string;
    protocol: 'grok';
}
export interface VideoModelSetting {
    group_id: number;
    model: string;
    label: string;
    enabled: boolean;
    protocol?: 'grok';
}
export interface VideoConfig {
    enabled: boolean;
    models: VideoModelSetting[];
}
export interface VideoParameters {
    seconds: number;
    resolution: string;
    size: string;
    mode: 'frames' | 'reference';
}
export interface VideoReference {
    id: string;
    name: string;
    blob: Blob;
}
export interface VideoDraft {
    id: 'current';
    sessionId: string;
    modelID: string;
    prompt: string;
    parameters: VideoParameters;
    references: VideoReference[];
}
export interface VideoRecord {
    id: string;
    generationId: string;
    sessionId: string;
    createdAt: number;
    prompt: string;
    model: VideoModel;
    parameters: VideoParameters;
    references: VideoReference[];
    status: 'pending' | 'success' | 'failed' | 'interrupted';
    taskID?: string;
    progress?: number;
    error?: string;
    video?: Blob;
    durationMs?: number;
    width?: number;
    height?: number;
    videoDurationMs?: number;
}
export interface VideoPrompt {
    id: string;
    title: string;
    prompt: string;
}
export interface VideoAsset {
    id: string;
    name: string;
    kind: 'image' | 'video';
    blob: Blob;
    createdAt: number;
}
export interface VideoTaskStatus {
    id: string;
    status: 'pending' | 'completed' | 'failed';
    progress?: number;
}
