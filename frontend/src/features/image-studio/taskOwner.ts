export function createTaskOwner(userID: number, locks = globalThis.navigator?.locks) {
  const id = crypto.randomUUID()
  const prefix = `sub2api-image-studio-owner:${userID}:`
  let release!: () => void
  const lifetime = new Promise<void>(resolve => { release = resolve })
  let acquired!: () => void
  const ready = new Promise<void>(resolve => { acquired = resolve })
  let available = Boolean(locks)
  if (locks) {
    void locks.request(prefix + id, async () => { acquired(); await lifetime }).catch(() => { available = false; acquired() })
  } else acquired()
  return {
    id, ready, release,
    async activeOwners(): Promise<Set<string> | null> {
      await ready
      if (!available || !locks) return null
      try {
        const snapshot = await locks.query()
        return new Set((snapshot.held || []).flatMap(lock => lock.name?.startsWith(prefix) ? [lock.name.slice(prefix.length)] : []))
      } catch { return null }
    }
  }
}
