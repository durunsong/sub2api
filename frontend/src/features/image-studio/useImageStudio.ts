import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import { generateImages, imageStudioAPI, imageToDataURL, type StudioModel } from './api'
import { createImageStudioDB } from './storage'
import { createTaskOwner } from './taskOwner'
import type { ImageAsset, ImageCategory, ImageEdit, ImageHistory, ImagePrompt, ImageSnapshot, ImageTask } from './types'

export interface StudioTask extends ImageTask {
  images?: Blob[]
  saved?: boolean
  completedAt?: number
  categoryID?: string | null
}

export interface BatchCompletion { id: string; success: number; failed: number; cancelled: number; unsaved: number }

export function useImageStudio() {
  const auth = useAuthStore()
  const models = shallowRef<StudioModel[]>([])
  const enabled = ref(true)
  const loading = ref(true)
  const catalogLoading = ref(false)
  const error = ref('')
  const storageError = ref('')
  const modelID = ref('')
  const prompt = ref('')
  const size = ref('auto')
  const quality = ref('auto')
  const count = ref(1)
  const aspectRatio = ref('1:1')
  const resolution = ref('1K')
  const references = shallowRef<Blob[]>([])
  const layout = ref<'side' | 'bottom'>('side')
  const history = shallowRef<ImageHistory[]>([])
  const tasks = ref<StudioTask[]>([])
  const completedBatch = shallowRef<BatchCompletion | null>(null)
  const pendingBatches = new Map<string, { pending: Set<string>; unsaved: Map<string, number>; result: BatchCompletion }>()
  const externalTasksCount = ref(0)
  const assets = shallowRef<ImageAsset[]>([])
  const categories = shallowRef<ImageCategory[]>([])
  const prompts = shallowRef<ImagePrompt[]>([])
  const categoryID = ref<string | null>(null)
  const selectedModel = computed(() => models.value.find(model => model.id === modelID.value))
  const activeCount = computed(() => tasks.value.filter(task => ['queued', 'running'].includes(task.status)).length)
  const canGenerate = computed(() => enabled.value && !!selectedModel.value && !!prompt.value.trim() && activeCount.value < 30)
  let db: ReturnType<typeof createImageStudioDB> | undefined
  let owner: ReturnType<typeof createTaskOwner> | undefined
  let epoch = 0
  let refreshSequence = 0
  let draftTimer: ReturnType<typeof setTimeout> | undefined
  let draftWrite = Promise.resolve()
  let draftImportEpoch: number | undefined
  let broadcast: BroadcastChannel | undefined
  let catalogAbort: AbortController | undefined
  const controllers = new Map<string, AbortController>()
  const dismissedTasks = new Set<string>()

  function message(err: unknown) { return extractApiErrorMessage(err, 'imageStudio.errors.request') }
  function localFailure() { storageError.value = 'imageStudio.errors.storage' }
  function changed() { broadcast?.postMessage('changed') }
  function snapshot(text = prompt.value): ImageSnapshot {
    const model = selectedModel.value!
    return { groupID: model.group_id, modelID: model.id, model: model.model, modelLabel: model.label, prompt: text, size: size.value, quality: quality.value, count: 1, aspectRatio: aspectRatio.value, resolution: resolution.value }
  }
  function restore(saved: ImageSnapshot, images: Blob[]) {
    modelID.value = models.value.find(model => model.id === saved.modelID || (model.group_id === saved.groupID && model.model === saved.model))?.id ?? ''
    prompt.value = saved.prompt
    size.value = saved.size
    quality.value = saved.quality
    count.value = saved.count
    aspectRatio.value = saved.aspectRatio || '1:1'
    resolution.value = saved.resolution || '1K'
    references.value = [...images]
  }
  async function readLocal(currentEpoch = epoch) {
    const source = db
    if (!source) return
    const sequence = ++refreshSequence
    try {
      const [logs, items, groups, templates, storedTasks, liveOwners] = await Promise.all([source.listHistory(), source.listAssets(), source.listCategories(), source.listPrompts(), source.listTasks(), owner?.activeOwners() ?? Promise.resolve(null)])
      if (currentEpoch !== epoch || sequence !== refreshSequence) return
      const storedIDs = new Set(storedTasks.map(task => task.id))
      const savedIDs = new Set(logs.map(log => log.id))
      const ownedIDs = new Set(tasks.value.filter(task => task.ownerID === owner?.id).map(task => task.id))
      const recovered: StudioTask[] = []
      let external = 0
      for (const stored of storedTasks) {
        if (stored.status === 'success' || savedIDs.has(stored.id) || ownedIDs.has(stored.id)) continue
        let task = stored
        if (stored.status === 'running' || stored.status === 'queued') {
          if (stored.ownerID && liveOwners?.has(stored.ownerID)) { external++; continue }
          task = { ...stored, status: 'interrupted' }
          if (liveOwners === null) task.error = 'imageStudio.untrackedTaskHint'
          else await source.saveTask(task)
        }
        recovered.push(task)
      }
      if (currentEpoch !== epoch || sequence !== refreshSequence) return
      const ownTasks = tasks.value.filter(task =>
        (task.images?.length && !task.saved) || (task.ownerID === owner?.id &&
          (['queued', 'running'].includes(task.status) || (storedIDs.has(task.id) && !savedIDs.has(task.id))))
      )
      const retainedIDs = new Set(ownTasks.map(task => task.id))
      tasks.value = [...ownTasks, ...recovered.filter(task => !retainedIDs.has(task.id))]
      externalTasksCount.value = external
      history.value = logs
      assets.value = items
      categories.value = groups
      prompts.value = templates
      if (categoryID.value && !groups.some(group => group.id === categoryID.value)) categoryID.value = null
      storageError.value = ''
    } catch {
      if (currentEpoch !== epoch || sequence !== refreshSequence) return
      // A failed read is not permission to present stale in-memory records as local history.
      history.value = []; assets.value = []; categories.value = []; prompts.value = []
      localFailure()
    }
  }
  async function readCatalog(currentEpoch = epoch) {
    catalogAbort?.abort()
    const controller = new AbortController()
    catalogAbort = controller
    catalogLoading.value = true
    try {
      const catalog = await imageStudioAPI.catalog(controller.signal)
      if (currentEpoch !== epoch || controller.signal.aborted) return
      enabled.value = catalog.enabled
      models.value = catalog.models
      if (!catalog.models.some(model => model.id === modelID.value)) modelID.value = catalog.models[0]?.id || ''
    } catch (err) {
      if (currentEpoch !== epoch || controller.signal.aborted) return
      models.value = []; modelID.value = ''; error.value = message(err)
    } finally { if (currentEpoch === epoch && !controller.signal.aborted) catalogLoading.value = false }
  }
  async function refresh() { error.value = ''; await Promise.all([readLocal(), readCatalog()]) }

  async function initialize() {
    const currentEpoch = ++epoch
    clearTimeout(draftTimer)
    for (const controller of controllers.values()) controller.abort()
    controllers.clear()
    dismissedTasks.clear()
    pendingBatches.clear(); completedBatch.value = null
    owner?.release(); owner = undefined
    externalTasksCount.value = 0
    db?.close(); db = undefined
    broadcast?.close(); broadcast = undefined
    catalogAbort?.abort()
    history.value = []; tasks.value = []; assets.value = []; categories.value = []; prompts.value = []
    prompt.value = ''; references.value = []; modelID.value = ''; categoryID.value = null
    models.value = []; error.value = ''; storageError.value = ''; loading.value = true
    const userID = auth.user?.id
    if (!userID) { loading.value = false; return }
    const currentOwner = createTaskOwner(userID)
    owner = currentOwner
    await currentOwner.ready
    if (currentEpoch !== epoch) { currentOwner.release(); return }
    try {
      db = createImageStudioDB(userID)
      if (typeof BroadcastChannel !== 'undefined') {
        broadcast = new BroadcastChannel(`sub2api-image-studio-${userID}`)
        broadcast.onmessage = () => { void readLocal() }
      }
    } catch { localFailure() }
    await Promise.all([readCatalog(currentEpoch), readLocal(currentEpoch)])
    const source = db
    if (currentEpoch !== epoch) return
    if (!source) { loading.value = false; return }
    try {
      const draft = await source.getDraft()
      if (currentEpoch !== epoch) return
      if (draft) { restore(draft.snapshot, draft.references); layout.value = draft.layout || 'side' }
    } catch { if (currentEpoch === epoch) localFailure() }
    finally { if (currentEpoch === epoch) loading.value = false }
  }

  async function storeTask(task: StudioTask, source = db) {
    if (!source) return
    const { id, ownerID, createdAt, status, snapshot: saved, references: refs, error: detail, batch } = task
    try {
      await source.saveTask({ id, ...(batch ? { batch: { ...batch } } : {}), ...(ownerID ? { ownerID } : {}), createdAt, status, snapshot: { ...saved }, references: [...refs], ...(detail ? { error: detail } : {}) })
      if (dismissedTasks.has(id)) await source.deleteTask(id)
      if (source === db) changed()
    }
    catch { if (source === db) localFailure() }
  }
  function record(task: StudioTask): ImageHistory {
    return { id: task.id, ...(task.batch ? { batch: { ...task.batch } } : {}), createdAt: task.createdAt, status: 'success', snapshot: { ...task.snapshot }, images: [...task.images!], references: [...task.references], favorite: false, categoryID: task.categoryID || null }
  }
  function settleBatch(task: StudioTask) {
    const batch = task.batch && pendingBatches.get(task.batch.id)
    if (!batch || !batch.pending.delete(task.id)) return
    if (task.status === 'success') {
      batch.result.success += task.images?.length || 0
      if (!task.saved) batch.unsaved.set(task.id, task.images?.length || 0)
    } else if (task.status === 'cancelled') batch.result.cancelled++
    else batch.result.failed++
    if (!batch.pending.size) {
      completedBatch.value = { ...batch.result, unsaved: [...batch.unsaved.values()].reduce((total, count) => total + count, 0) }
      pendingBatches.delete(batch.result.id)
    }
  }
  async function run(task: StudioTask, currentEpoch: number) {
    const source = db
    const controller = new AbortController()
    controllers.set(task.id, controller)
    task.status = 'running'
    await storeTask(task, source)
    try {
      if (controller.signal.aborted || currentEpoch !== epoch) return
      const model = models.value.find(model => model.id === task.snapshot.modelID && model.group_id === task.snapshot.groupID && model.model === task.snapshot.model)
      if (!model || !enabled.value) throw new Error('imageStudio.errors.modelUnavailable')
      if (task.snapshot.mask && (!model.supports_mask || !task.references.length)) throw new Error('imageStudio.maskUnsupported')
      const refs = await Promise.all(task.references.map(imageToDataURL))
      const mask = task.snapshot.mask ? await imageToDataURL(task.snapshot.mask) : undefined
      if (controller.signal.aborted || currentEpoch !== epoch) return
      const result = await generateImages({ model_id: model.id, prompt: task.snapshot.prompt, n: 1, size: task.snapshot.size, ...(model.qualities.length ? { quality: task.snapshot.quality } : {}), ...(model.protocol === 'gemini' ? { aspect_ratio: task.snapshot.aspectRatio, resolution: task.snapshot.size } : {}), references: refs, ...(mask ? { mask } : {}) }, controller.signal)
      if (controller.signal.aborted || currentEpoch !== epoch) return
      task.status = 'success'; task.images = result.images; task.saved = false
      try {
        if (!source) throw new Error('storage unavailable')
        await source.saveHistory(record(task))
        if (currentEpoch !== epoch) return
        task.saved = true
        await readLocal(currentEpoch)
        changed()
      } catch { if (currentEpoch === epoch) localFailure() }
    } catch (err) {
      if (currentEpoch !== epoch) return
      task.status = controller.signal.aborted ? 'cancelled' : 'failed'
      task.error = controller.signal.aborted ? 'imageStudio.cancelledHint' : message(err)
      await storeTask(task, source)
    } finally {
      if (currentEpoch === epoch) {
        task.completedAt = Date.now()
        settleBatch(task)
        controllers.delete(task.id)
        if (task.saved) tasks.value = tasks.value.filter(item => item.id !== task.id)
        pump()
      }
    }
  }
  function pump() {
    if (!auth.user?.id) return
    const waiting = tasks.value.filter(task => task.status === 'queued')
    const available = Math.max(0, 3 - controllers.size)
    for (const task of waiting.slice(0, available)) void run(task, epoch)
  }
  async function submit(series?: string[], edit?: ImageEdit) {
    error.value = ''
    const model = selectedModel.value
    const texts = series || Array.from({ length: Math.min(10, Math.max(1, Number(count.value) || 1)) }, () => prompt.value)
    if (!enabled.value || !model) { error.value = 'imageStudio.errors.modelUnavailable'; return }
    if (edit && (edit.modelID !== model.id || !edit.references.length || (edit.mask && (!model.supports_mask || edit.mask.type !== 'image/png')))) { error.value = 'imageStudio.maskUnsupported'; return }
    if (!texts.length || texts.some(text => !text.trim() || Array.from(text).length > 16000) || texts.length + activeCount.value > 30) { error.value = 'imageStudio.errors.invalidPrompt'; return }
    const input = [...(edit?.references || references.value)]
    if (input.length > model.max_references) { error.value = 'imageStudio.errors.tooManyReferences'; return }
    if (edit && [...input, ...(edit.mask ? [edit.mask] : [])].some(image => !image.size || image.size > 10 * 1024 * 1024)) { error.value = 'imageStudio.errors.invalidImage'; return }
    const currentEpoch = epoch, source = db
    const batchID = crypto.randomUUID()
    const newTasks: StudioTask[] = texts.map((text, index) => ({ id: crypto.randomUUID(), batch: { id: batchID, index, total: texts.length }, ownerID: owner?.id, createdAt: Date.now(), status: 'queued', snapshot: { ...snapshot(text.trim()), ...(edit?.mask ? { mask: edit.mask } : {}) }, references: input, categoryID: categoryID.value }))
    pendingBatches.set(batchID, { pending: new Set(newTasks.map(task => task.id)), unsaved: new Map(), result: { id: batchID, success: 0, failed: 0, cancelled: 0, unsaved: 0 } })
    tasks.value = [...newTasks, ...tasks.value]
    await Promise.all(newTasks.map(task => storeTask(task, source)))
    if (currentEpoch === epoch) pump()
  }
  function cancel(id: string) {
    const task = tasks.value.find(item => item.id === id)
    if (!task || !['queued', 'running'].includes(task.status)) return
    const queued = task.status === 'queued'
    controllers.get(id)?.abort()
    task.status = 'cancelled'; task.completedAt = Date.now(); task.error = 'imageStudio.cancelledHint'
    if (queued) settleBatch(task)
    void storeTask(task)
  }
  async function retry(task: ImageTask) {
    restore(task.snapshot, task.references)
    if (task.snapshot.mask) await submit([task.snapshot.prompt], { modelID: modelID.value, references: task.references, mask: task.snapshot.mask })
    else await submit([task.snapshot.prompt])
  }
  async function dismissTask(id: string) {
    const task = tasks.value.find(item => item.id === id)
    if (!task || !['failed', 'cancelled', 'interrupted'].includes(task.status)) return
    const currentEpoch = epoch
    await localAction(async source => {
      dismissedTasks.add(id)
      await source.deleteTask(id)
      if (currentEpoch === epoch) tasks.value = tasks.value.filter(item => item.id !== id)
    })
  }
  function loadHistory(item: ImageHistory) { restore(item.snapshot, item.references) }
  function newDraft() { prompt.value = ''; references.value = []; count.value = 1 }
  async function localAction(action: (source: NonNullable<typeof db>) => Promise<unknown>) {
    const source = db, currentEpoch = epoch
    if (!source) { localFailure(); return }
    try {
      await action(source)
      if (currentEpoch !== epoch) return
      await readLocal(currentEpoch); changed()
    } catch (err) { if (currentEpoch === epoch) { localFailure(); error.value = message(err) } }
  }
  async function addReferences(images: Blob[]) {
    const currentEpoch = epoch
    const limit = selectedModel.value?.max_references || 0
    if (!limit || images.length + references.value.length > limit) { error.value = 'imageStudio.errors.tooManyReferences'; return }
    try {
      for (const image of images) {
        if (image.size > 10 * 1024 * 1024) throw new Error('imageStudio.errors.invalidImage')
        await imageToDataURL(image)
      }
      if (currentEpoch !== epoch) return
      const currentLimit = selectedModel.value?.max_references || 0
      if (!currentLimit || images.length + references.value.length > currentLimit) { error.value = 'imageStudio.errors.tooManyReferences'; return }
      references.value = [...references.value, ...images]
    } catch (err) { if (currentEpoch === epoch) error.value = message(err) }
  }
  function removeReference(index: number) { references.value = references.value.filter((_, i) => i !== index) }
  async function deleteHistory(ids: string[]) { await localAction(async source => { for (const id of ids) await source.deleteHistory(id) }) }
  async function favorite(item: ImageHistory) { await localAction(source => source.toggleFavorite(item.id)) }
  async function categorize(ids: string[], category: string | null | string[]) {
    await localAction(source => source.assignCategories(ids, Array.isArray(category) ? category : category ? [category] : []))
  }
  async function saveCategory(name: string, id: string = crypto.randomUUID()) { await localAction(source => source.saveCategory({ id, name: name.trim() })) }
  async function deleteCategory(id: string) { await localAction(source => source.deleteCategory(id)) }
  async function saveAsset(image: Blob, name = 'Image') { await localAction(source => source.saveAsset({ id: crypto.randomUUID(), name, image, createdAt: Date.now(), categoryID: categoryID.value })) }
  async function deleteAsset(id: string) { await localAction(source => source.deleteAsset(id)) }
  async function savePrompt(item: ImagePrompt) { await localAction(source => source.savePrompt(item)) }
  async function deletePrompt(id: string) { await localAction(source => source.deletePrompt(id)) }
  async function exportBackup() {
    if (!db) { localFailure(); return }
    try { downloadImage(await db.exportBackup(), 'image-studio-backup.json') } catch { localFailure() }
  }
  async function importBackup(file: Blob) {
    const source = db, currentEpoch = epoch
    if (!source) { localFailure(); return }
    clearTimeout(draftTimer)
    draftImportEpoch = currentEpoch
    let restored = false
    try {
      // An older asynchronous Blob validation must finish before the backup replaces its draft.
      await draftWrite
      if (currentEpoch !== epoch) return
      const draft = await source.importBackup(file)
      if (currentEpoch !== epoch) return
      if (draft) { restore(draft.snapshot, draft.references); layout.value = draft.layout || 'side'; restored = true }
      await readLocal(currentEpoch)
      if (currentEpoch === epoch) changed()
    } catch (err) {
      if (currentEpoch === epoch) { localFailure(); error.value = message(err) }
    } finally {
      if (currentEpoch === epoch) {
        draftImportEpoch = undefined
        if (!restored) scheduleDraft()
      }
    }
  }
  async function saveUnsaved(task: StudioTask) {
    const currentEpoch = epoch
    await localAction(async source => {
      await source.saveHistory(record(task))
      if (currentEpoch === epoch) {
        task.saved = true
        if (task.batch) pendingBatches.get(task.batch.id)?.unsaved.delete(task.id)
        tasks.value = tasks.value.filter(item => item.id !== task.id)
      }
    })
  }
  function onFocus() { if (document.visibilityState !== 'hidden') void refresh() }
  function beforeUnload(event: BeforeUnloadEvent) { if (activeCount.value || tasks.value.some(task => task.images?.length && !task.saved)) { event.preventDefault(); event.returnValue = '' } }
  watch(selectedModel, model => {
    if (!model) return
    if (!model.sizes.includes(size.value) && !(model.supports_custom_size && /^\d+x\d+$/.test(size.value))) size.value = model.sizes[0] || 'auto'
    if (!model.qualities.includes(quality.value)) quality.value = model.qualities[0] || 'auto'
  })
  function scheduleDraft() {
    clearTimeout(draftTimer)
    if (loading.value || draftImportEpoch === epoch || !selectedModel.value || !db) return
    const source = db, currentEpoch = epoch
    const draft = { id: 'current' as const, snapshot: { ...snapshot(), count: count.value }, references: [...references.value], layout: layout.value, updatedAt: Date.now() }
    draftTimer = setTimeout(() => {
      draftWrite = draftWrite.then(async () => {
        if (currentEpoch === epoch) await source.saveDraft(draft)
      }).catch(() => { if (currentEpoch === epoch) localFailure() })
    }, 400)
  }
  watch([prompt, modelID, size, quality, count, aspectRatio, resolution, references, layout], scheduleDraft)
  watch(() => auth.user?.id, () => { void initialize() })
  onMounted(() => { void initialize(); window.addEventListener('focus', onFocus); document.addEventListener('visibilitychange', onFocus); window.addEventListener('beforeunload', beforeUnload) })
  onBeforeUnmount(() => {
    epoch++; clearTimeout(draftTimer); catalogAbort?.abort()
    for (const controller of controllers.values()) controller.abort()
    controllers.clear(); owner?.release(); db?.close(); broadcast?.close()
    window.removeEventListener('focus', onFocus); document.removeEventListener('visibilitychange', onFocus); window.removeEventListener('beforeunload', beforeUnload)
  })
  return { models, enabled, catalogLoading, loading, error, storageError, modelID, prompt, size, quality, count, aspectRatio, resolution, references, layout, history, tasks, completedBatch, externalTasksCount, assets, categories, prompts, categoryID, selectedModel, activeCount, canGenerate, refresh, submit, cancel, retry, dismissTask, loadHistory, newDraft, addReferences, removeReference, deleteHistory, favorite, categorize, saveCategory, deleteCategory, saveAsset, deleteAsset, savePrompt, deletePrompt, exportBackup, importBackup, saveUnsaved }
}

export function downloadImage(image: Blob, name = `image-${Date.now()}.${image.type === 'image/jpeg' ? 'jpg' : image.type === 'image/webp' ? 'webp' : 'png'}`) {
  const url = URL.createObjectURL(image)
  const link = document.createElement('a')
  link.href = url; link.download = name
  document.body.appendChild(link); link.click(); link.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
