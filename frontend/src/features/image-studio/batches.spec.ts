import { describe, expect, it } from 'vitest'
import { groupResults } from './batches'
import type { ImageHistory } from './types'
import type { StudioTask } from './useImageStudio'
const history = (id: string, index?: number): ImageHistory => ({ id, createdAt: 10, status: 'success', snapshot: { groupID: 1, model: 'image', prompt: 'prompt', size: 'auto', quality: 'auto', count: 1 }, images: [new Blob(['image'])], references: [], favorite: false, categoryID: null, ...(index === undefined ? {} : { batch: { id: 'batch', index, total: 3 } }) })
describe('local result batches', () => {
  it('groups persisted results with unfinished tasks in submission order, without duplicate saved tasks', () => {
    const first = history('a', 0), third = history('c', 2)
    const task: StudioTask = { ...history('b', 1), status: 'failed' }
    const groups = groupResults([third, first, history('legacy')], [task, { ...first, saved: true }])
    expect(groups).toHaveLength(2)
    expect(groups[0].entries.map(entry => entry.id)).toEqual(['a', 'b', 'c'])
    expect(groups[0].grouped).toBe(true)
    expect(groups[1].grouped).toBe(false)
  })
  it('does not invent removed images or merge old records by prompt or timestamp', () => {
    const groups = groupResults([history('a', 0), history('legacy1'), history('legacy2')], [])
    expect(groups).toHaveLength(3)
    expect(groups[0].entries).toHaveLength(1)
    expect(groups[0].grouped).toBe(true)
  })
})
