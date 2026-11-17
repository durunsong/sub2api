import type {
  ImageAsset, ImageCategory, ImageDraft, ImageHistory, ImagePrompt,
  ImageSnapshot, ImageStudioBackup, ImageTask
} from './types'

export const MAX_IMAGE_BYTES = 32 * 1024 * 1024
export const MAX_BACKUP_BYTES = 128 * 1024 * 1024
const stores = ['history', 'assets', 'categories', 'drafts', 'prompts', 'tasks'] as const
type Store = typeof stores[number]
type Row = ImageHistory | ImageAsset | ImageCategory | ImageDraft | ImagePrompt | ImageTask
type StorageErrorCode = 'invalid' | 'unavailable' | 'quota' | 'closed'

export class ImageStudioStorageError extends Error {
  constructor(public readonly code: StorageErrorCode, message: string) {
    super(message)
    this.name = 'ImageStudioStorageError'
  }
}

function invalid(message = 'Invalid image studio data'): never {
  throw new ImageStudioStorageError('invalid', message)
}

function storageError(error: unknown): ImageStudioStorageError {
  if (error instanceof ImageStudioStorageError) return error
  return new ImageStudioStorageError(
    error instanceof DOMException && error.name === 'QuotaExceededError' ? 'quota' : 'unavailable',
    'Local image storage could not complete the operation'
  )
}

function object(value: unknown, keys: string[]): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value) ||
    Object.keys(value).some(key => !keys.includes(key))) invalid()
  return value as Record<string, unknown>
}

function string(value: unknown, max = 256, empty = false): asserts value is string {
  if (typeof value !== 'string' || value.length > max || (!empty && !value.trim())) invalid()
}

function timestamp(value: unknown) {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) invalid()
}

function category(value: unknown) {
  if (value !== null) string(value)
}

function categories(value: unknown) {
  if (value === undefined) return
  if (!Array.isArray(value) || value.length > 100 || new Set(value).size !== value.length) invalid()
  for (const id of value) string(id)
}

function readBlob(blob: Blob, mode: 'text' | 'dataURL' | 'arrayBuffer'): Promise<string | ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string | ArrayBuffer)
    reader.onerror = () => reject(storageError(reader.error))
    if (mode === 'text') reader.readAsText(blob)
    else if (mode === 'dataURL') reader.readAsDataURL(blob)
    else reader.readAsArrayBuffer(blob)
  })
}

async function image(value: unknown): Promise<void> {
  if (!(value instanceof Blob) || !['image/png', 'image/jpeg', 'image/webp'].includes(value.type) ||
    value.size === 0 || value.size > MAX_IMAGE_BYTES) invalid('Invalid or oversized image')
  const bytes = new Uint8Array(await readBlob(value.slice(0, 12), 'arrayBuffer') as ArrayBuffer)
  const matches = value.type === 'image/png'
    ? [137, 80, 78, 71, 13, 10, 26, 10].every((byte, i) => bytes[i] === byte)
    : value.type === 'image/jpeg'
      ? bytes[0] === 255 && bytes[1] === 216 && bytes[2] === 255
      : String.fromCharCode(...bytes.slice(0, 4)) === 'RIFF' && String.fromCharCode(...bytes.slice(8, 12)) === 'WEBP'
  if (!matches) invalid('Image bytes do not match their MIME type')
  // Decode as well as checking the header: truncated files must never become history.
  try {
    if (typeof createImageBitmap === 'function') {
      const bitmap = await createImageBitmap(value)
      const valid = bitmap.width > 0 && bitmap.height > 0
      bitmap.close()
      if (!valid) invalid('Empty image')
    } else {
      const url = URL.createObjectURL(value)
      try {
        const decoded = new Image()
        decoded.src = url
        await decoded.decode()
        if (!decoded.naturalWidth || !decoded.naturalHeight) invalid('Empty image')
      } finally {
        URL.revokeObjectURL(url)
      }
    }
  } catch {
    invalid('Image cannot be decoded')
  }
}

async function images(value: unknown, required = false) {
  if (!Array.isArray(value) || value.length > 16 || (required && !value.length)) invalid()
  // Sequential decoding bounds memory when importing a collection of large images.
  for (const blob of value) await image(blob)
}

async function snapshot(value: unknown): Promise<void> {
  const row = object(value, ['groupID', 'model', 'modelID', 'prompt', 'size', 'quality', 'count', 'aspectRatio', 'resolution', 'modelLabel', 'mask'])
  if (!Number.isSafeInteger(row.groupID) || (row.groupID as number) < 1 ||
    !Number.isInteger(row.count) || (row.count as number) < 1 || (row.count as number) > 16) invalid()
  string(row.model)
  if (row.modelID !== undefined) string(row.modelID)
  string(row.prompt, 32000, true)
  string(row.size, 100, true)
  string(row.quality, 100, true)
  for (const key of ['aspectRatio', 'resolution', 'modelLabel']) if (row[key] !== undefined) string(row[key], 256, true)
  if (row.mask !== undefined) await image(row.mask)
}

export async function validateHistoryRecord(value: unknown): Promise<ImageHistory> {
  const row = object(value, ['id', 'createdAt', 'status', 'snapshot', 'images', 'references', 'favorite', 'categoryID', 'categoryIDs'])
  string(row.id)
  timestamp(row.createdAt)
  if (row.status !== 'success' || typeof row.favorite !== 'boolean') invalid()
  category(row.categoryID)
  categories(row.categoryIDs)
  await snapshot(row.snapshot)
  await images(row.images, true)
  await images(row.references)
  return row as unknown as ImageHistory
}

async function validate(store: Store, value: unknown): Promise<Row> {
  if (store === 'history') return validateHistoryRecord(value)
  const fields = {
    assets: ['id', 'name', 'createdAt', 'image', 'categoryID', 'categoryIDs'],
    categories: ['id', 'name'], drafts: ['id', 'snapshot', 'references', 'updatedAt', 'layout'],
    prompts: ['id', 'title', 'prompt', 'createdAt', 'updatedAt', 'category', 'tags', 'seriesPrompts'],
    tasks: ['id', 'ownerID', 'createdAt', 'status', 'snapshot', 'references', 'error']
  }
  const row = object(value, fields[store])
  string(row.id)
  if (store === 'categories' || store === 'assets') string(row.name)
  if (store === 'assets') {
    category(row.categoryID)
    categories(row.categoryIDs)
    await image(row.image)
  }
  if (store === 'drafts' || store === 'tasks') {
    await snapshot(row.snapshot)
    await images(row.references)
  }
  if (store === 'drafts') {
    if (row.id !== 'current') invalid()
    timestamp(row.updatedAt)
    if (row.layout !== undefined && row.layout !== 'side' && row.layout !== 'bottom') invalid()
  } else if (store !== 'categories') timestamp(row.createdAt)
  if (store === 'prompts') {
    string(row.title)
    string(row.prompt, 32000)
    if (row.updatedAt !== undefined) timestamp(row.updatedAt)
    if (row.category !== undefined) string(row.category, 256, true)
    for (const key of ['tags', 'seriesPrompts']) if (row[key] !== undefined) {
      const values = row[key]
      if (!Array.isArray(values) || values.length > 100) invalid()
      for (const item of values) string(item, key === 'tags' ? 256 : 32000)
    }
  }
  if (store === 'tasks') {
    if (!['queued', 'running', 'failed', 'cancelled', 'interrupted', 'success'].includes(row.status as string)) invalid()
    if (row.error !== undefined) string(row.error, 10000, true)
    if (row.ownerID !== undefined) string(row.ownerID, 128)
  }
  return row as unknown as Row
}

function decodeImage(value: unknown): Blob {
  const encoded = object(value, ['type', 'data'])
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(encoded.type as string) || typeof encoded.data !== 'string' ||
    !encoded.data.length || encoded.data.length > Math.ceil(MAX_IMAGE_BYTES / 3) * 4 ||
    encoded.data.length % 4 !== 0 || !/^[A-Za-z0-9+/]+={0,2}$/.test(encoded.data)) invalid('Invalid backup image')
  const binary = atob(encoded.data)
  return new Blob([Uint8Array.from(binary, char => char.charCodeAt(0))], { type: encoded.type as string })
}

export async function parseBackup(file: Blob): Promise<ImageStudioBackup> {
  if (!(file instanceof Blob) || !file.size || file.size > MAX_BACKUP_BYTES) invalid('Backup is empty or too large')
  try {
    const backup = object(JSON.parse(await readBlob(file, 'text') as string), ['version', ...stores])
    if (backup.version !== 1) invalid('Unsupported backup version')
    let entries = 0
    for (const store of stores) {
      const rows = backup[store]
      if (!Array.isArray(rows) || (entries += rows.length) > 10000 || (store === 'drafts' && rows.length > 1)) invalid()
      const ids = new Set<string>()
      for (const value of rows) {
        const row = object(value, ['id', 'name', 'title', 'prompt', 'createdAt', 'updatedAt', 'status', 'snapshot', 'images', 'image', 'references', 'favorite', 'categoryID', 'categoryIDs', 'error', 'category', 'tags', 'seriesPrompts', 'layout', 'ownerID'])
        if (Array.isArray(row.images)) row.images = row.images.map(decodeImage)
        if (Array.isArray(row.references)) row.references = row.references.map(decodeImage)
        if (row.image !== undefined) row.image = decodeImage(row.image)
        if (row.snapshot && typeof row.snapshot === 'object' && 'mask' in row.snapshot) {
          const params = row.snapshot as Record<string, unknown>
          params.mask = decodeImage(params.mask)
        }
        const validated = await validate(store, row)
        if (ids.has(validated.id)) invalid('Duplicate backup key')
        ids.add(validated.id)
      }
    }
    const result = backup as unknown as ImageStudioBackup
    const categories = new Set(result.categories.map(row => row.id))
    if ([...result.history, ...result.assets].some(row =>
      (row.categoryID !== null && !categories.has(row.categoryID)) || row.categoryIDs?.some(id => !categories.has(id))
    )) invalid('Missing backup category')
    return result
  } catch (error) {
    if (error instanceof ImageStudioStorageError) throw error
    invalid('Invalid backup file')
  }
}

async function encode(value: unknown): Promise<unknown> {
  if (value instanceof Blob) {
    const dataURL = await readBlob(value, 'dataURL') as string
    return { type: value.type, data: dataURL.slice(dataURL.indexOf(',') + 1) }
  }
  if (Array.isArray(value)) {
    const result: unknown[] = []
    for (const item of value) result.push(await encode(item))
    return result
  }
  if (value && typeof value === 'object') {
    const result: Record<string, unknown> = {}
    for (const [key, item] of Object.entries(value)) result[key] = await encode(item)
    return result
  }
  return value
}

export function createImageStudioDB(userID: number) {
  if (!Number.isSafeInteger(userID) || userID < 1) invalid('A signed-in user is required')
  let connection: Promise<IDBDatabase> | undefined
  let closed = false
  const transactions = new Set<IDBTransaction>()

  function open(): Promise<IDBDatabase> {
    if (closed) return Promise.reject(new ImageStudioStorageError('closed', 'Image storage is closed'))
    if (connection) return connection
    connection = new Promise<IDBDatabase>((resolve, reject) => {
      if (!globalThis.indexedDB) return reject(new ImageStudioStorageError('unavailable', 'IndexedDB is unavailable'))
      let settled = false
      const request = indexedDB.open(`sub2api-image-studio-${userID}`, 1)
      request.onupgradeneeded = () => {
        for (const store of stores) if (!request.result.objectStoreNames.contains(store)) request.result.createObjectStore(store, { keyPath: 'id' })
      }
      request.onerror = () => { settled = true; reject(storageError(request.error)) }
      request.onblocked = () => { settled = true; reject(new ImageStudioStorageError('unavailable', 'Image storage is blocked by another tab')) }
      request.onsuccess = () => {
        const db = request.result
        if (closed || settled) {
          db.close()
          reject(new ImageStudioStorageError('closed', 'Image storage is closed'))
          return
        }
        db.onversionchange = () => { db.close(); connection = undefined }
        db.onclose = () => { connection = undefined }
        resolve(db)
      }
    }).catch(error => { connection = undefined; throw storageError(error) })
    return connection
  }

  async function transaction<T>(names: Store[], mode: IDBTransactionMode, run: (tx: IDBTransaction, done: (value: T) => void) => void): Promise<T> {
    const db = await open()
    if (closed) throw new ImageStudioStorageError('closed', 'Image storage is closed')
    return new Promise<T>((resolve, reject) => {
      let tx: IDBTransaction
      let result: T
      try {
        tx = db.transaction(names, mode)
        transactions.add(tx)
        tx.oncomplete = () => { transactions.delete(tx); resolve(result) }
        tx.onabort = () => {
          transactions.delete(tx)
          reject(closed ? new ImageStudioStorageError('closed', 'Image storage is closed') : storageError(tx.error))
        }
        tx.onerror = () => { /* Request errors abort the transaction; onabort reports the final result. */ }
        run(tx, value => { result = value })
      } catch (error) {
        if (tx!) {
          transactions.delete(tx)
          try { tx.abort() } catch { /* Already completed. */ }
        }
        reject(storageError(error))
      }
    })
  }

  async function list<T extends Row>(store: Store): Promise<T[]> {
    const rows = await transaction<unknown[]>([store], 'readonly', (tx, done) => {
      const request = tx.objectStore(store).getAll()
      request.onsuccess = () => done(request.result)
    })
    const valid: T[] = []
    for (const row of rows) {
      try { valid.push(await validate(store, row) as T) } catch { /* Missing or corrupt local bytes are not history. */ }
    }
    if (closed) throw new ImageStudioStorageError('closed', 'Image storage is closed')
    return valid
  }

  async function save(store: Store, row: Row): Promise<void> {
    await validate(store, row)
    await transaction<void>([store], 'readwrite', tx => { tx.objectStore(store).put(row) })
  }

  async function remove(store: Store, id: string): Promise<void> {
    string(id)
    await transaction<void>([store], 'readwrite', tx => { tx.objectStore(store).delete(id) })
  }

  return {
    listHistory: async () => (await list<ImageHistory>('history')).sort((a, b) => b.createdAt - a.createdAt),
    saveHistory: async (row: ImageHistory): Promise<void> => {
      await validateHistoryRecord(row)
      // A committed result and its task must change together, including on tab close.
      await transaction<void>(['history', 'tasks'], 'readwrite', tx => {
        tx.objectStore('history').put(row)
        tx.objectStore('tasks').delete(row.id)
      })
    },
    deleteHistory: (id: string) => remove('history', id),
    toggleFavorite: async (id: string): Promise<void> => {
      string(id)
      await transaction<void>(['history'], 'readwrite', tx => {
        const store = tx.objectStore('history')
        const request = store.get(id)
        request.onsuccess = () => {
          const row = request.result as ImageHistory | undefined
          if (row) store.put({ ...row, favorite: !row.favorite })
        }
      })
    },
    assignCategories: async (ids: string[], categoryIDs: string[]): Promise<void> => {
      for (const id of ids) string(id)
      categories(categoryIDs)
      // Read and update together so another tab cannot revive deleted records or categories.
      await transaction<void>(['history', 'categories'], 'readwrite', tx => {
        const available = tx.objectStore('categories').getAllKeys()
        available.onsuccess = () => {
          const selected = categoryIDs.filter(id => available.result.includes(id))
          const store = tx.objectStore('history')
          for (const id of new Set(ids)) {
            const request = store.get(id)
            request.onsuccess = () => {
              const row = request.result as ImageHistory | undefined
              if (row) store.put({ ...row, categoryID: selected[0] ?? null, categoryIDs: selected })
            }
          }
        }
      })
    },
    clearHistory: () => transaction<void>(['history'], 'readwrite', tx => { tx.objectStore('history').clear() }),
    listAssets: () => list<ImageAsset>('assets'),
    saveAsset: (row: ImageAsset) => save('assets', row),
    deleteAsset: (id: string) => remove('assets', id),
    listCategories: () => list<ImageCategory>('categories'),
    saveCategory: (row: ImageCategory) => save('categories', row),
    deleteCategory: async (id: string): Promise<void> => {
      string(id)
      await transaction<void>(['categories', 'history', 'assets'], 'readwrite', tx => {
        tx.objectStore('categories').delete(id)
        for (const name of ['history', 'assets']) {
          const request = tx.objectStore(name).openCursor()
          request.onsuccess = () => {
            const cursor = request.result
            if (!cursor) return
            const value = cursor.value as ImageHistory | ImageAsset
            if (value.categoryID === id || value.categoryIDs?.includes(id)) cursor.update({
              ...value,
              categoryID: value.categoryID === id ? null : value.categoryID,
              ...(value.categoryIDs ? { categoryIDs: value.categoryIDs.filter(categoryID => categoryID !== id) } : {})
            })
            cursor.continue()
          }
        }
      })
    },
    getDraft: async (): Promise<ImageDraft | null> => (await list<ImageDraft>('drafts'))[0] ?? null,
    saveDraft: (row: ImageDraft | null) => row ? save('drafts', row) : remove('drafts', 'current'),
    listPrompts: () => list<ImagePrompt>('prompts'),
    savePrompt: (row: ImagePrompt) => save('prompts', row),
    deletePrompt: (id: string) => remove('prompts', id),
    listTasks: () => list<ImageTask>('tasks'),
    saveTask: (row: ImageTask) => save('tasks', row),
    deleteTask: (id: string) => remove('tasks', id),
    exportBackup: async (): Promise<Blob> => {
      const backup = await transaction<Record<string, unknown>>(Array.from(stores), 'readonly', (tx, done) => {
        const result: Record<string, unknown> = { version: 1 }
        for (const store of stores) {
          const request = tx.objectStore(store).getAll()
          request.onsuccess = () => { result[store] = request.result }
        }
        done(result)
      })
      // Do not export corrupt rows as legitimate recoverable history.
      for (const store of stores) {
        const valid: Row[] = []
        for (const row of backup[store] as unknown[]) {
          try { valid.push(await validate(store, row)) } catch { /* Skip corrupt local records. */ }
        }
        backup[store] = valid
      }
      const categoryIDs = new Set((backup.categories as ImageCategory[]).map(row => row.id))
      for (const row of [...backup.history as ImageHistory[], ...backup.assets as ImageAsset[]]) {
        if (row.categoryID !== null && !categoryIDs.has(row.categoryID)) row.categoryID = null
        if (row.categoryIDs) row.categoryIDs = row.categoryIDs.filter(id => categoryIDs.has(id))
      }
      const file = new Blob([JSON.stringify(await encode(backup))], { type: 'application/json' })
      if (file.size > MAX_BACKUP_BYTES) invalid('Backup exceeds the import size limit')
      if (closed) throw new ImageStudioStorageError('closed', 'Image storage is closed')
      return file
    },
    importBackup: async (file: Blob): Promise<ImageDraft | null> => {
      const backup = await parseBackup(file)
      await transaction<void>(Array.from(stores), 'readwrite', tx => {
        for (const store of stores) for (const row of backup[store]) {
          // Imported in-flight requests are never replayed or presented as still running.
          const value = store === 'tasks' && ['queued', 'running'].includes((row as ImageTask).status)
            ? { ...row, status: 'interrupted' } : row
          tx.objectStore(store).put(value)
        }
      })
      return backup.drafts[0] ?? null
    },
    close: () => {
      closed = true
      for (const tx of transactions) { try { tx.abort() } catch { /* Already committed. */ } }
      void connection?.then(db => db.close()).catch(() => undefined)
      connection = undefined
    }
  }
}

export type ImageStudioDB = ReturnType<typeof createImageStudioDB>
export type { ImageSnapshot }
