import type { StudioModel, StudioModelSetting, WorkflowModel } from './api'
const key = (value: StudioModelSetting | StudioModel) => `${value.group_id}:${value.model}`
export function mergeModelSettings(models: (StudioModel | WorkflowModel)[], settings: StudioModelSetting[]): StudioModelSetting[] {
  const entries = new Map<string, StudioModelSetting>(models.map(item => [`${item.group_id}:${item.model}`, { group_id: item.group_id, model: item.model, label: item.model, ...('protocol' in item ? { protocol: item.protocol } : {}), enabled: true }]))
  for (const setting of settings) entries.set(key(setting), { ...setting })
  return [...entries.values()]
}
export function updateModelSetting(settings: StudioModelSetting[], value: StudioModelSetting): StudioModelSetting[] {
  const entries = new Map(settings.map(setting => [key(setting), setting]))
  entries.set(key(value), { ...value })
  return [...entries.values()]
}
