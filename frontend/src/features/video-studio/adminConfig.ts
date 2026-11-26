import type { VideoModel, VideoModelSetting } from './types';
export function mergeVideoModels(models: VideoModel[], settings: VideoModelSetting[]): VideoModelSetting[] {
    const rows = new Map(models.map(m => [`${m.group_id}:${m.model}`, { group_id: m.group_id, model: m.model, label: m.label, protocol: m.protocol, enabled: true } as VideoModelSetting]));
    for (const item of settings)
        rows.set(`${item.group_id}:${item.model}`, { ...item });
    return [...rows.values()];
}
export function setVideoVisibility(settings: VideoModelSetting[], filtered: VideoModelSetting[], enabled: boolean, validGroups: number[]) {
    const next = settings.map(s => ({ ...s }));
    for (const row of filtered) {
        if (!validGroups.includes(row.group_id))
            continue;
        const index = next.findIndex(s => s.group_id === row.group_id && s.model === row.model);
        const item = { ...row, enabled };
        if (index < 0)
            next.push(item);
        else
            next[index] = item;
    }
    return next;
}
