import { afterEach, describe, expect, it, vi } from 'vitest'
import { createImageStudioDB, MAX_BACKUP_BYTES, parseBackup, validateHistoryRecord } from './storage'

const snapshot = { groupID: 1, model: 'image-model', prompt: 'A tree', size: '1024x1024', quality: 'auto', count: 1 }
const record = (images: unknown[] = []) => ({
  id: 'result-1', createdAt: 1, status: 'success', snapshot,
  images, references: [], favorite: false, categoryID: null
})
const backup = (extra: Record<string, unknown> = {}) => new Blob([JSON.stringify({
  version: 1, history: [], assets: [], categories: [], drafts: [], prompts: [], tasks: [], ...extra
})], { type: 'application/json' })

afterEach(() => vi.unstubAllGlobals())

describe('image studio storage boundaries', () => {
  it('rejects a success record without locally stored image bytes', async () => {
    await expect(validateHistoryRecord(record())).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects remote URLs and SVG masquerading as stored images', async () => {
    await expect(validateHistoryRecord(record(['https://example.com/image.png']))).rejects.toMatchObject({ code: 'invalid' })
    await expect(validateHistoryRecord(record([new Blob(['<svg/>'], { type: 'image/svg+xml' })]))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects corrupt image bytes even when their MIME claims PNG', async () => {
    await expect(validateHistoryRecord(record([new Blob(['not an image'], { type: 'image/png' })]))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('reports unavailable storage instead of fabricating an empty history', async () => {
    vi.stubGlobal('indexedDB', undefined)
    await expect(createImageStudioDB(7).listHistory()).rejects.toMatchObject({ code: 'unavailable' })
  })

  it('refuses invalid user identities so they cannot share a database', () => {
    expect(() => createImageStudioDB(0)).toThrow()
    expect(() => createImageStudioDB(Number.NaN)).toThrow()
  })

  it('rejects backup versions and unknown fields before importing', async () => {
    await expect(parseBackup(backup({ version: 2 }))).rejects.toMatchObject({ code: 'invalid' })
    await expect(parseBackup(backup({ url: 'https://example.com/images' }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects remote backup image locations without fetching them', async () => {
    const entry = { ...record(), images: [{ type: 'image/png', data: 'https://example.com/image.png' }] }
    await expect(parseBackup(backup({ history: [entry] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects duplicate keys rather than silently dropping imported entries', async () => {
    await expect(parseBackup(backup({ categories: [{ id: 'a', name: 'First' }, { id: 'a', name: 'Second' }] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('accepts a complete empty backup and rejects unknown nested parameters', async () => {
    await expect(parseBackup(backup())).resolves.toMatchObject({ version: 1, history: [], tasks: [] })
    await expect(parseBackup(backup({ drafts: [{ id: 'current', updatedAt: 1, references: [], snapshot: { ...snapshot, endpoint: 'https://example.com' } }] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('preserves optional task ownership and rejects invalid owner identifiers', async () => {
    const task = { id: 'task', ownerID: 'tab-owner', createdAt: 1, status: 'running', snapshot, references: [] }
    await expect(parseBackup(backup({ tasks: [task] }))).resolves.toMatchObject({ tasks: [task] })
    await expect(parseBackup(backup({ tasks: [{ ...task, ownerID: {} }] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects an oversized backup before reading its contents', async () => {
    const oversized = backup()
    Object.defineProperty(oversized, 'size', { value: MAX_BACKUP_BYTES + 1 })
    await expect(parseBackup(oversized)).rejects.toMatchObject({ code: 'invalid' })
  })

  it('rejects decoder failures after a valid file signature', async () => {
    vi.stubGlobal('createImageBitmap', async () => { throw new Error('truncated PNG') })
    const truncated = new Blob([new Uint8Array([137, 80, 78, 71, 13, 10, 26, 10])], { type: 'image/png' })
    await expect(validateHistoryRecord(record([truncated]))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('preserves local template metadata and rejects malformed prompt sequences', async () => {
    const template = { id: 'p1', title: 'Portraits', prompt: '{{subject}} portrait', createdAt: 1,
      updatedAt: 2, category: 'Portrait', tags: ['Photo'], seriesPrompts: ['Daylight', 'Night'] }
    await expect(parseBackup(backup({ prompts: [template] }))).resolves.toMatchObject({ prompts: [template] })
    await expect(parseBackup(backup({ prompts: [{ ...template, seriesPrompts: [null] }] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('restores the draft layout and exact catalog model identifier', async () => {
    const draft = { id: 'current', updatedAt: 1, references: [], layout: 'bottom', snapshot: { ...snapshot, modelID: 'catalog-entry' } }
    await expect(parseBackup(backup({ drafts: [draft] }))).resolves.toMatchObject({ drafts: [draft] })
    await expect(parseBackup(backup({ drafts: [{ ...draft, layout: 'unknown' }] }))).rejects.toMatchObject({ code: 'invalid' })
  })

  it('preserves multiple categories and rejects missing category references', async () => {
    vi.stubGlobal('createImageBitmap', async () => ({ width: 1, height: 1, close() {} }))
    const image = { type: 'image/png', data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Wl6SAAAAABJRU5ErkJggg==' }
    const asset = { id: 'a', name: 'Reference', createdAt: 1, image, categoryID: null, categoryIDs: ['portrait', 'photo'] }
    await expect(parseBackup(backup({ assets: [asset], categories: [{ id: 'portrait', name: 'Portrait' }, { id: 'photo', name: 'Photo' }] }))).resolves.toMatchObject({ assets: [{ categoryIDs: ['portrait', 'photo'] }] })
    await expect(parseBackup(backup({ assets: [asset], categories: [{ id: 'portrait', name: 'Portrait' }] }))).rejects.toMatchObject({ code: 'invalid' })
  })
})

// jsdom has no IndexedDB. Drive only the native event boundary here; integration
// tests in a real browser cover persistent records, isolation and image decoding.
function controlledDatabase(data: Record<string, unknown[]> = {}, autoComplete = false) {
  let current: IDBTransaction | undefined
  const db = {
    close() {},
    transaction() {
      let reads = 0
      const tx = {
        error: null,
        objectStore: (name: string) => ({
          put: () => ({}),
          getAll() {
            reads++
            const request = { result: data[name] ?? [], onsuccess: undefined as (() => void) | undefined }
            queueMicrotask(() => {
              request.onsuccess?.()
              if (--reads === 0 && autoComplete) tx.oncomplete?.()
            })
            return request
          }
        }),
        abort() { tx.onabort?.() },
        oncomplete: undefined as (() => void) | undefined,
        onabort: undefined as (() => void) | undefined,
        onerror: undefined as (() => void) | undefined
      }
      current = tx as unknown as IDBTransaction
      return current
    }
  }
  vi.stubGlobal('indexedDB', { open() {
    const request = { result: db, onsuccess: undefined as (() => void) | undefined }
    queueMicrotask(() => request.onsuccess?.())
    return request
  } })
  return async () => {
    await vi.waitFor(() => expect(current).toBeDefined())
    return current!
  }
}

describe('transaction outcome', () => {
  it('does not report a successful save until the transaction commits', async () => {
    const pending = controlledDatabase()
    const db = createImageStudioDB(1)
    let saved = false
    const result = db.saveCategory({ id: 'a', name: 'Landscape' }).then(() => { saved = true })
    const tx = await pending()
    expect(saved).toBe(false)
    tx.oncomplete?.call(tx, new Event('complete'))
    await result
    expect(saved).toBe(true)
    db.close()
  })

  it('reports quota exhaustion when the transaction aborts after accepting a write', async () => {
    const pending = controlledDatabase()
    const db = createImageStudioDB(1)
    const result = db.saveCategory({ id: 'a', name: 'Landscape' })
    const rejected = expect(result).rejects.toMatchObject({ code: 'quota' })
    const tx = await pending()
    Object.defineProperty(tx, 'error', { value: new DOMException('Full', 'QuotaExceededError') })
    tx.onabort?.call(tx, new Event('abort'))
    await rejected
    db.close()
  })

  it('aborts pending writes and refuses reuse after the user leaves their account', async () => {
    const pending = controlledDatabase()
    const db = createImageStudioDB(1)
    const result = db.saveCategory({ id: 'a', name: 'Landscape' })
    const rejected = expect(result).rejects.toMatchObject({ code: 'closed' })
    await pending()
    db.close()
    await rejected
    await expect(db.listHistory()).rejects.toMatchObject({ code: 'closed' })
  })

  it('does not return a previous account read after its database is closed', async () => {
    const pending = controlledDatabase()
    const db = createImageStudioDB(1)
    const result = db.listCategories()
    const rejected = expect(result).rejects.toMatchObject({ code: 'closed' })
    const tx = await pending()
    tx.oncomplete?.call(tx, new Event('complete'))
    db.close()
    await rejected
  })

  it('exports recoverable images when their category was removed outside the app', async () => {
    vi.stubGlobal('createImageBitmap', async () => ({ width: 1, height: 1, close() {} }))
    const image = new Blob([new Uint8Array([137, 80, 78, 71, 13, 10, 26, 10])], { type: 'image/png' })
    controlledDatabase({ assets: [{ id: 'a', name: 'Reference', createdAt: 1, image, categoryID: 'deleted', categoryIDs: ['deleted', 'kept'] }], categories: [{ id: 'kept', name: 'Kept' }] }, true)
    const db = createImageStudioDB(1)
    const file = await db.exportBackup()
    await expect(parseBackup(file)).resolves.toMatchObject({ assets: [{ categoryID: null, categoryIDs: ['kept'] }] })
    db.close()
  })
})
