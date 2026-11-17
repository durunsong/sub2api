import { flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createTaskOwner } from './taskOwner'

describe('image task owner locks', () => {
  it('tracks live owners per account and releases ownership when a page leaves', async () => {
    const held = new Set<string>()
    const locks = {
      request: vi.fn(async (name: string, callback: () => Promise<void>) => {
        held.add(name)
        try { await callback() } finally { held.delete(name) }
      }),
      query: vi.fn(async () => ({ held: [...held].map(name => ({ name })) }))
    } as unknown as LockManager
    const first = createTaskOwner(7, locks), second = createTaskOwner(7, locks), otherAccount = createTaskOwner(8, locks)
    await Promise.all([first.ready, second.ready, otherAccount.ready])
    expect(await first.activeOwners()).toEqual(new Set([first.id, second.id]))
    second.release()
    await flushPromises()
    expect(await first.activeOwners()).toEqual(new Set([first.id]))
    first.release(); otherAccount.release()
    await flushPromises()
    expect(held.size).toBe(0)
  })

  it('reports unknown ownership when lock access fails', async () => {
    const locks = { request: vi.fn().mockRejectedValue(new Error('denied')), query: vi.fn() } as unknown as LockManager
    const owner = createTaskOwner(7, locks)
    await owner.ready
    expect(await owner.activeOwners()).toBeNull()
    owner.release()
  })
})
