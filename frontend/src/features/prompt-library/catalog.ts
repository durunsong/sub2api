// Source list and normalization ported from basketikun/infinite-canvas.
// Upstream commit and MIT notice: docs/VIDEO_STUDIO_UPSTREAM.md.
export const sources = [
  ['banana-prompt-quicker', 'Banana Prompt Quicker'],
  ['davidwu-gpt-image2-prompts', 'DavidWu GPT Image 2'],
  ['freestylefly-gpt-image-2', 'Freestylefly GPT Image 2'],
  ['awesome-gpt-image', 'Awesome GPT Image'],
  ['awesome-gpt4o-image-prompts', 'Awesome GPT-4o'],
  ['youmind-gpt-image-2', 'YouMind GPT Image 2'],
  ['youmind-nano-banana-pro', 'YouMind Nano Banana Pro'],
].map(([id, name]) => ({ id, name, url: `https://raw.githubusercontent.com/yukkcat/image-prompts/main/dist/sources/${id}.json` }))
export type PromptSource = typeof sources[number]
export interface LibraryPrompt {
  id: string; sourceId: string; category: string; title: string; prompt: string;
  description: string; coverUrl: string; sourceUrl: string; tags: string[]; author: string;
}
export function safeURL(value: unknown, base: string): string {
  if (typeof value !== 'string' || !value.trim()) return ''
  try { const url = new URL(value, base); return url.protocol === 'https:' ? url.href : '' } catch { return '' }
}
export function normalizePrompts(data: unknown, source: PromptSource): LibraryPrompt[] {
  if (!Array.isArray(data)) throw new Error('Invalid catalog')
  const seen = new Set<string>()
  return data.flatMap((value, index) => {
    if (!value || typeof value !== 'object') return []
    const text = (key: string) => typeof value[key] === 'string' ? value[key].trim() : ''
    const title = text('title'), prompt = text('prompt'), id = text('id') || `${source.id}-${index}`
    if (!title || !prompt || seen.has(id)) return []
    seen.add(id)
    return [{ id, title, prompt, sourceId: source.id, category: source.name,
      description: text('description'), author: text('author'),
      coverUrl: safeURL(value.coverUrl || value.referenceImageUrls?.[0], source.url),
      sourceUrl: safeURL(value.sourceUrl, source.url),
      tags: Array.isArray(value.tags) ? value.tags.filter((v: unknown): v is string => typeof v === 'string' && !!v.trim()).map((v: string) => v.trim()) : [] }]
  })
}
export function filterPrompts(items: LibraryPrompt[], category: string, keyword: string, tags: string[]) {
  const query = keyword.trim().toLowerCase()
  return items.filter(item => (!category || item.sourceId === category)
    && (!tags.length || tags.some(tag => item.tags.includes(tag)))
    && (!query || [item.title, item.prompt, item.description, item.category, ...item.tags].join(' ').toLowerCase().includes(query)))
}
interface Cache { id: string; fetchedAt: number; items: LibraryPrompt[] }
const memory = new Map<string, Cache>()
const requests = new Map<string, Promise<{ items: LibraryPrompt[]; stale: boolean }>>()
async function cacheStore<T>(mode: IDBTransactionMode, action: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  const db = await new Promise<IDBDatabase>((resolve, reject) => {
    const request = indexedDB.open('sub2api-public-prompt-library', 1)
    let blocked = false
    request.onupgradeneeded = () => request.result.createObjectStore('sources', { keyPath: 'id' })
    request.onsuccess = () => { if (blocked) request.result.close(); else resolve(request.result) }
    request.onerror = () => reject(request.error)
    request.onblocked = () => { blocked = true; reject(new Error('Cache blocked')) }
  })
  try {
    return await new Promise<T>((resolve, reject) => {
      const transaction = db.transaction('sources', mode)
      const request = action(transaction.objectStore('sources'))
      transaction.oncomplete = () => resolve(request.result)
      transaction.onerror = transaction.onabort = () => reject(transaction.error)
    })
  } finally { db.close() }
}
export function loadSource(source: PromptSource, force = false) {
  const existing = requests.get(source.id)
  if (existing) return existing
  const request = (async () => {
    const cached = memory.get(source.id) || await cacheStore<Cache | undefined>('readonly', store => store.get(source.id)).catch(() => undefined)
    if (cached && !force && Date.now() - cached.fetchedAt < 3600000) {
      memory.set(source.id, cached)
      return { items: cached.items, stale: false }
    }
    const abort = new AbortController(), timer = setTimeout(() => abort.abort(), 30000)
    try {
      // Public sources only: never attach panel JWT or user-authenticated API interceptors.
      const response = await fetch(source.url, { signal: abort.signal, credentials: 'omit', referrerPolicy: 'no-referrer' })
      if (!response.ok) throw new Error('Source unavailable')
      const items = normalizePrompts(await response.json(), source)
      if (!items.length) throw new Error('Empty source')
      const next = { id: source.id, fetchedAt: Date.now(), items }
      memory.set(source.id, next)
      await cacheStore('readwrite', store => store.put(next)).catch(() => undefined)
      return { items, stale: false }
    } catch (error) {
      if (cached?.items.length) return { items: cached.items, stale: true }
      throw error
    } finally { clearTimeout(timer) }
  })().finally(() => requests.delete(source.id))
  requests.set(source.id, request)
  return request
}
