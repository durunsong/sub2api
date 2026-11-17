import { defineComponent, nextTick, reactive } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useImageStudio } from './useImageStudio'

const mocks = vi.hoisted(() => ({ generate: vi.fn(), catalog: vi.fn(), activeOwners: vi.fn(), releaseOwner: vi.fn(), db: {} as Record<string, ReturnType<typeof vi.fn>>, user: { user: { id: 7 } } }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.user }))
vi.mock('./api', () => ({ imageStudioAPI: { catalog: mocks.catalog }, generateImages: mocks.generate, imageToDataURL: async () => 'data:image/png;base64,YQ==' }))
vi.mock('./taskOwner', () => ({ createTaskOwner: () => ({ id: 'current-owner', ready: Promise.resolve(), release: mocks.releaseOwner, activeOwners: mocks.activeOwners }) }))
vi.mock('./storage', () => ({ createImageStudioDB: () => mocks.db }))
let studio: ReturnType<typeof useImageStudio>
let wrapper: ReturnType<typeof mount>
const model = { id: '7:gpt-image-2', group_id: 7, group_name: 'Images', model: 'gpt-image-2', label: 'GPT Image', protocol: 'images', sizes: ['1024x1024'], qualities: ['auto'], max_images: 10, max_references: 4, supports_mask: false, pricing: { currency: 'USD', mode: 'image', rate_multiplier: 1, image_price_1k: 0.03, image_price_2k: null, image_price_4k: null } }
async function setup() {
  wrapper = mount(defineComponent({ setup() { studio = useImageStudio(); return () => null } }))
  await flushPromises()
}
beforeEach(() => {
  vi.clearAllMocks()
  mocks.activeOwners.mockResolvedValue(new Set(['current-owner']))
  mocks.user = reactive({ user: { id: 7 } })
  mocks.catalog.mockResolvedValue({ enabled: true, models: [model] })
  for (const name of ['listHistory','listAssets','listCategories','listTasks','listPrompts']) mocks.db[name] = vi.fn().mockResolvedValue([])
  mocks.db.getDraft = vi.fn().mockResolvedValue(null)
  for (const name of ['saveHistory','saveTask','deleteTask','saveDraft','close']) mocks.db[name] = vi.fn().mockResolvedValue(undefined)
  mocks.generate.mockResolvedValue({ images: [new Blob(['png'], { type: 'image/png' })] })
})
afterEach(() => { wrapper?.unmount(); vi.useRealTimers() })

describe('image studio task lifecycle', () => {
  it('sends and persists an edit mask with the original image, retries it, and isolates ordinary generation', async () => {
    mocks.catalog.mockResolvedValue({ enabled: true, models: [{ ...model, supports_mask: true }] })
    await setup()
    const original = new Blob(['original'], { type: 'image/png' })
    const mask = new Blob(['mask'], { type: 'image/png' })
    studio.references.value = [new Blob(['unrelated'], { type: 'image/png' })]
    await studio.submit(['Edit only here'], { modelID: model.id, references: [original], mask })
    await flushPromises()
    expect(mocks.generate).toHaveBeenLastCalledWith(expect.objectContaining({ mask: expect.any(String), references: [expect.any(String)] }), expect.any(AbortSignal))
    const saved = mocks.db.saveHistory.mock.calls[0][0]
    expect(saved.snapshot.mask).toBe(mask)
    expect(saved.references).toEqual([original])
    await studio.retry(saved)
    await flushPromises()
    expect(mocks.generate.mock.calls[1][0].mask).toEqual(expect.any(String))
    studio.prompt.value = 'ordinary'
    await studio.submit()
    await flushPromises()
    expect(mocks.generate.mock.calls[2][0].mask).toBeUndefined()
  })

  it('rejects stale models and unsupported masks before generating', async () => {
    await setup()
    const original = new Blob(['original'], { type: 'image/png' })
    const mask = new Blob(['mask'], { type: 'image/png' })
    await studio.submit(['edit'], { modelID: model.id, references: [original], mask })
    await studio.submit(['edit'], { modelID: 'stale-model', references: [original] })
    await flushPromises()
    expect(mocks.generate).not.toHaveBeenCalled()
    expect(mocks.db.saveTask).not.toHaveBeenCalled()
  })

  it('does not report a successfully re-saved image as unsaved when its batch finishes', async () => {
    await setup()
    studio.prompt.value = 'Re-save before batch finishes'
    studio.count.value = 2
    const resolvers: ((result: unknown) => void)[] = []
    mocks.generate.mockImplementation(() => new Promise(resolve => resolvers.push(resolve)))
    mocks.db.saveHistory.mockRejectedValueOnce(new Error('quota'))
    await studio.submit()
    await flushPromises()
    const image = { images: [new Blob(['png'], { type: 'image/png' })] }
    resolvers[0](image)
    await flushPromises()
    const unsaved = studio.tasks.value.find(task => task.status === 'success')!
    expect(unsaved.saved).toBe(false)
    await studio.saveUnsaved(unsaved)
    await flushPromises()
    resolvers[1](image)
    await flushPromises()
    expect(studio.completedBatch.value).toMatchObject({ success: 2, unsaved: 0 })
  })

  it('keeps one persistent batch identity for a submission and reports completion once', async () => {
    await setup()
    studio.prompt.value = 'Batch'
    studio.count.value = 2
    const resolvers: ((result: unknown) => void)[] = []
    mocks.generate.mockImplementation(() => new Promise(resolve => resolvers.push(resolve)))
    await studio.submit()
    await flushPromises()
    const tasks = [...studio.tasks.value]
    expect(tasks[0].batch).toEqual({ id: expect.any(String), index: 0, total: 2 })
    expect(tasks[1].batch).toEqual({ ...tasks[0].batch, index: 1 })
    expect(mocks.db.saveTask).toHaveBeenCalledWith(expect.objectContaining({ batch: tasks[0].batch }))
    resolvers[0]({ images: [new Blob(['png'], { type: 'image/png' })] })
    await flushPromises()
    expect(studio.completedBatch.value).toBeNull()
    expect(mocks.db.saveHistory).toHaveBeenCalledWith(expect.objectContaining({ batch: tasks[0].batch }))
    resolvers[1]({ images: [new Blob(['png'], { type: 'image/png' })] })
    await flushPromises()
    expect(studio.completedBatch.value).toMatchObject({ id: tasks[0].batch!.id, success: 2, failed: 0, cancelled: 0, unsaved: 0 })
    const completed = studio.completedBatch.value
    await studio.refresh()
    expect(studio.completedBatch.value).toBe(completed)
  })

  it('completes a batch after queued cancellation and partial failure without replay', async () => {
    await setup()
    studio.prompt.value = 'Mixed batch'
    studio.count.value = 4
    const rejectors: ((reason: Error) => void)[] = []
    mocks.generate.mockImplementation(() => new Promise((_, reject) => rejectors.push(reject)))
    await studio.submit()
    await flushPromises()
    studio.cancel(studio.tasks.value[3].id)
    rejectors.forEach(reject => reject(new Error('failed')))
    await flushPromises()
    expect(studio.completedBatch.value).toMatchObject({ success: 0, failed: 3, cancelled: 1 })
    expect(mocks.generate).toHaveBeenCalledTimes(3)
  })

  it('restores imported draft contents and cancels the pending old autosave', async () => {
    await setup()
    vi.useFakeTimers()
    studio.prompt.value = 'old prompt'
    await nextTick()
    const image = new Blob(['reference'], { type: 'image/png' })
    const imported = { id: 'current', snapshot: { modelID: model.id, model: model.model, groupID: 7, prompt: 'imported prompt', size: '1024x1024', quality: 'auto', count: 3 }, references: [image], layout: 'bottom', updatedAt: 1 }
    mocks.db.importBackup = vi.fn().mockResolvedValue(imported)
    await studio.importBackup(new Blob(['backup']))
    await vi.advanceTimersByTimeAsync(500)
    expect(studio.prompt.value).toBe('imported prompt')
    expect(studio.references.value).toEqual([image])
    expect(studio.count.value).toBe(3)
    expect(studio.layout.value).toBe('bottom')
    expect(mocks.db.saveDraft).not.toHaveBeenCalled()
    studio.count.value = 2
    await nextTick()
    await vi.advanceTimersByTimeAsync(500)
    expect(mocks.db.saveDraft).toHaveBeenCalledWith(expect.objectContaining({ snapshot: expect.objectContaining({ prompt: 'imported prompt', count: 2 }), references: [image], layout: 'bottom' }))
  })
  it('finishes an already started draft write before importing its replacement', async () => {
    await setup()
    vi.useFakeTimers()
    let finishSave!: () => void
    let storedPrompt = ''
    mocks.db.saveDraft.mockImplementationOnce((draft) => new Promise<void>(resolve => {
      finishSave = () => { storedPrompt = draft.snapshot.prompt; resolve() }
    }))
    studio.prompt.value = 'old prompt'
    await nextTick()
    await vi.advanceTimersByTimeAsync(400)
    mocks.db.importBackup = vi.fn(async () => { storedPrompt = 'imported prompt' })
    const importing = studio.importBackup(new Blob(['backup']))
    await nextTick()
    finishSave()
    await importing
    expect(storedPrompt).toBe('imported prompt')
  })
  it('does not restore an imported draft into a different account', async () => {
    await setup()
    let finishImport!: (draft: unknown) => void
    mocks.db.importBackup = vi.fn(() => new Promise(resolve => { finishImport = resolve }))
    const importing = studio.importBackup(new Blob(['backup']))
    await flushPromises()
    expect(finishImport).toBeTypeOf('function')
    mocks.user.user.id = 8
    await flushPromises()
    studio.prompt.value = 'new account'
    finishImport({ snapshot: { modelID: model.id, model: model.model, groupID: 7, prompt: 'old account secret', size: '1024x1024', quality: 'auto', count: 1 }, references: [], layout: 'bottom' })
    await importing
    expect(studio.prompt.value).toBe('new account')
  })
  it('retains the current draft and resumes autosave after an invalid import', async () => {
    await setup()
    vi.useFakeTimers()
    studio.prompt.value = 'keep this draft'
    await nextTick()
    mocks.db.importBackup = vi.fn().mockRejectedValue(new Error('invalid backup'))
    await studio.importBackup(new Blob(['invalid']))
    await vi.advanceTimersByTimeAsync(500)
    expect(studio.prompt.value).toBe('keep this draft')
    expect(mocks.db.saveDraft).toHaveBeenCalledWith(expect.objectContaining({ snapshot: expect.objectContaining({ prompt: 'keep this draft' }) }))
    expect(studio.storageError.value).toBeTruthy()
  })
  it('preserves unsaved edits when the backup contains no draft', async () => {
    await setup()
    vi.useFakeTimers()
    studio.prompt.value = 'unsaved edits'
    await nextTick()
    mocks.db.importBackup = vi.fn().mockResolvedValue(null)
    mocks.db.getDraft.mockResolvedValue({ snapshot: { modelID: model.id, model: model.model, groupID: 7, prompt: 'stale disk version', size: '1024x1024', quality: 'auto', count: 1 }, references: [], layout: 'side' })
    await studio.importBackup(new Blob(['backup']))
    await vi.advanceTimersByTimeAsync(500)
    expect(studio.prompt.value).toBe('unsaved edits')
    expect(mocks.db.saveDraft).toHaveBeenCalledWith(expect.objectContaining({ snapshot: expect.objectContaining({ prompt: 'unsaved edits' }) }))
  })
  it('ignores an old account autosave failure after switching accounts', async () => {
    await setup()
    vi.useFakeTimers()
    let failSave!: (error: Error) => void
    mocks.db.saveDraft.mockImplementationOnce(() => new Promise((_resolve, reject) => { failSave = reject }))
    studio.prompt.value = 'old account draft'
    await nextTick()
    await vi.advanceTimersByTimeAsync(400)
    mocks.user.user.id = 8
    await flushPromises()
    failSave(new Error('old database closed'))
    await flushPromises()
    expect(studio.storageError.value).toBe('')
  })
  it('enforces the reference limit when paste and drop arrive together', async () => {
    await setup()
    const image = new Blob(['png'], { type: 'image/png' })
    studio.references.value = [image, image, image]
    await Promise.all([studio.addReferences([image]), studio.addReferences([image])])
    expect(studio.references.value).toHaveLength(4)
    expect(studio.error.value).toBe('imageStudio.errors.tooManyReferences')
  })
  it('selects an authorized model automatically and splits counts into independent requests', async () => {
    await setup()
    expect(studio.modelID.value).toBe(model.id)
    studio.prompt.value = 'mountains'
    studio.count.value = 3
    await studio.submit()
    await flushPromises()
    expect(mocks.generate).toHaveBeenCalledTimes(3)
    expect(mocks.generate.mock.calls.every(([request]) => request.n === 1 && request.prompt === 'mountains')).toBe(true)
    expect(mocks.db.saveHistory).toHaveBeenCalledTimes(3)
  })
  it('keeps a failed local save out of history and retains a downloadable session result', async () => {
    mocks.db.saveHistory.mockRejectedValue(new Error('quota'))
    await setup()
    studio.prompt.value = 'mountains'
    await studio.submit()
    await flushPromises()
    expect(studio.history.value).toEqual([])
    expect(studio.tasks.value.find(task => task.images?.length)?.saved).toBe(false)
    expect(studio.storageError.value).toBeTruthy()
  })
  it('marks old running work interrupted without resubmitting after refresh', async () => {
    mocks.db.listTasks.mockResolvedValue([{ id: 'old', createdAt: 1, status: 'running', snapshot: { modelID: model.id, model: model.model, groupID: 7, prompt: 'old', size: '1024x1024', quality: 'auto', count: 1 }, references: [] }])
    await setup()
    expect(studio.tasks.value[0]?.status).toBe('interrupted')
    expect(mocks.generate).not.toHaveBeenCalled()
  })
  it('does not interrupt or offer retry for another live tab task', async () => {
    mocks.activeOwners.mockResolvedValue(new Set(['current-owner', 'other-tab']))
    const task = { id: 'other', ownerID: 'other-tab', createdAt: 1, status: 'running', snapshot: { modelID: model.id, model: model.model, groupID: 7, prompt: 'other', size: '1024x1024', quality: 'auto', count: 1 }, references: [] }
    mocks.db.listTasks.mockResolvedValue([task])
    await setup()
    expect(studio.tasks.value).toEqual([])
    expect(studio.externalTasksCount.value).toBe(1)
    expect(mocks.db.saveTask).not.toHaveBeenCalled()
    expect(mocks.generate).not.toHaveBeenCalled()
    mocks.activeOwners.mockResolvedValue(new Set(['current-owner']))
    await studio.refresh()
    expect(studio.tasks.value[0]?.status).toBe('interrupted')
    expect(studio.externalTasksCount.value).toBe(0)
    expect(mocks.generate).not.toHaveBeenCalled()
  })
  it('allows explicit recovery without rewriting unknown running tasks when Web Locks are unavailable', async () => {
    mocks.activeOwners.mockResolvedValue(null)
    mocks.db.listTasks.mockResolvedValue([{ id: 'unknown', status: 'running' }])
    await setup()
    expect(studio.tasks.value[0]?.status).toBe('interrupted')
    expect(studio.tasks.value[0]?.error).toBe('imageStudio.untrackedTaskHint')
    expect((await mocks.db.listTasks())[0].status).toBe('running')
    expect(mocks.db.saveTask).not.toHaveBeenCalled()
    expect(mocks.generate).not.toHaveBeenCalled()
  })
  it('does not recover a legacy running task whose result is already saved', async () => {
    mocks.db.listHistory.mockResolvedValue([{ id: 'saved' }])
    mocks.db.listTasks.mockResolvedValue([{ id: 'saved', status: 'running' }])
    await setup()
    expect(studio.tasks.value).toEqual([])
    expect(mocks.generate).not.toHaveBeenCalled()
  })
  it('ignores an older local-read failure after a newer refresh succeeds', async () => {
    await setup()
    let rejectOld!: (error: Error) => void
    mocks.db.listHistory.mockImplementationOnce(() => new Promise((_resolve, reject) => { rejectOld = reject }))
    const oldRefresh = studio.refresh()
    mocks.db.listHistory.mockResolvedValueOnce([{ id: 'new' }])
    await studio.refresh()
    rejectOld(new Error('old read failed'))
    await oldRefresh
    expect(studio.history.value).toEqual([{ id: 'new' }])
    expect(studio.storageError.value).toBe('')
  })
  it('removes externally deleted history on refresh rather than merging stale memory', async () => {
    mocks.db.listHistory.mockResolvedValueOnce([{ id: 'old' }])
    await setup()
    expect(studio.history.value).toHaveLength(1)
    await studio.refresh()
    expect(studio.history.value).toEqual([])
  })
  it('does not send requests when the model disappears from the catalog', async () => {
    await setup()
    mocks.catalog.mockResolvedValue({ enabled: true, models: [] })
    await studio.refresh()
    studio.prompt.value = 'mountains'
    await nextTick()
    await studio.submit()
    expect(mocks.generate).not.toHaveBeenCalled()
  })
  it('preserves a supported custom size across catalog refreshes', async () => {
    mocks.catalog.mockImplementation(async () => ({ enabled: true, models: [{ ...model, supports_custom_size: true, size_multiple: 16 }] }))
    await setup()
    studio.size.value = '1600x1024'
    await studio.refresh()
    await nextTick()
    expect(studio.size.value).toBe('1600x1024')
  })
  it('releases saved session images once IndexedDB owns the history', async () => {
    await setup()
    studio.prompt.value = 'mountains'
    await studio.submit()
    await flushPromises()
    expect(mocks.db.saveHistory).toHaveBeenCalledTimes(1)
    expect(studio.tasks.value).toEqual([])
  })
  it('keeps submitted snapshots independent when the draft changes during a request', async () => {
    let finish!: (value: { images: Blob[] }) => void
    mocks.generate.mockImplementation(() => new Promise(resolve => { finish = resolve }))
    await setup()
    studio.prompt.value = 'original'
    await studio.submit()
    await flushPromises()
    studio.prompt.value = 'changed'
    finish({ images: [new Blob(['png'], { type: 'image/png' })] })
    await flushPromises()
    expect(mocks.db.saveHistory.mock.calls[0]?.[0].snapshot.prompt).toBe('original')
  })
  it('discards a completed response after an account switch', async () => {
    let finish!: (value: { images: Blob[] }) => void
    mocks.generate.mockImplementation(() => new Promise(resolve => { finish = resolve }))
    await setup()
    studio.prompt.value = 'private'
    await studio.submit()
    await flushPromises()
    mocks.user.user.id = 8
    await flushPromises()
    finish({ images: [new Blob(['png'], { type: 'image/png' })] })
    await flushPromises()
    expect(mocks.db.saveHistory).not.toHaveBeenCalled()
    expect(studio.tasks.value).toEqual([])
  })
  it('cancels a queued task without a billable request and leaves other tasks running', async () => {
    mocks.generate.mockImplementation(() => new Promise(() => {}))
    await setup()
    studio.prompt.value = 'queued'; studio.count.value = 4
    await studio.submit()
    await flushPromises()
    const pending = studio.tasks.value.find(task => task.status === 'queued')!
    studio.cancel(pending.id)
    expect(pending.status).toBe('cancelled')
    expect(mocks.generate).toHaveBeenCalledTimes(3)
    expect(studio.activeCount.value).toBe(3)
  })
  it('persists queued snapshots before dispatch so refresh can recover interrupted work', async () => {
    mocks.generate.mockImplementation(() => new Promise(() => {}))
    await setup()
    studio.prompt.value = 'queued'; studio.count.value = 4
    await studio.submit()
    await flushPromises()
    const pending = studio.tasks.value.find(task => task.status === 'queued')!
    expect(mocks.db.saveTask.mock.calls.some(([saved]) => saved.id === pending.id && saved.status === 'queued')).toBe(true)
  })
  it('removes its own finished task after another tab deletes the stored record', async () => {
    mocks.generate.mockRejectedValueOnce(new Error('failed'))
    await setup()
    studio.prompt.value = 'failure'
    await studio.submit()
    await flushPromises()
    expect(studio.tasks.value[0]?.status).toBe('failed')
    await studio.refresh()
    expect(studio.tasks.value).toEqual([])
  })
  it('removes a dismissed failure from local storage and the task tray', async () => {
    mocks.generate.mockRejectedValue(new Error('provider failed'))
    await setup()
    studio.prompt.value = 'mountains'
    await studio.submit()
    await flushPromises()
    const failed = studio.tasks.value[0]!
    await studio.dismissTask(failed.id)
    expect(mocks.db.deleteTask).toHaveBeenCalledWith(failed.id)
    expect(studio.tasks.value).toEqual([])
  })

})
