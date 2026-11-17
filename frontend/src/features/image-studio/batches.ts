import type { ImageHistory } from './types'
import type { StudioTask } from './useImageStudio'

export type ResultEntry = { id: string; history?: ImageHistory; task?: StudioTask }
export interface ResultBatch { id: string; grouped: boolean; entries: ResultEntry[]; createdAt: number }

// Only persisted batch IDs relate records; identical prompts or timestamps do not.
export function groupResults(history: ImageHistory[], tasks: StudioTask[]): ResultBatch[] {
  const groups = new Map<string, ResultBatch>()
  const saved = new Set(history.map(item => item.id))
  const entries: ResultEntry[] = [
    ...tasks.filter(task => !saved.has(task.id)).map(task => ({ id: task.id, task })),
    ...history.map(item => ({ id: item.id, history: item }))
  ]
  for (const entry of entries) {
    const item = (entry.history || entry.task)!
    const id = item.batch ? `batch:${item.batch.id}` : `record:${item.id}`
    const group = groups.get(id) || { id, grouped: Boolean(item.batch && item.batch.total > 1), entries: [], createdAt: item.createdAt }
    group.entries.push(entry)
    group.createdAt = Math.max(group.createdAt, item.createdAt)
    groups.set(id, group)
  }
  for (const group of groups.values()) group.entries.sort((a, b) => ((a.history || a.task)!.batch?.index || 0) - ((b.history || b.task)!.batch?.index || 0))
  return [...groups.values()].sort((a, b) => b.createdAt - a.createdAt)
}
