import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import MaskEditor from '../MaskEditor.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { nextTick, reactive } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ImageStudioView from '@/views/user/ImageStudioView.vue'

const fixture = vi.hoisted(() => ({ state: {} as Record<string, any>, auth: { user: { id: 1 } } }))
vi.mock('vue-i18n', async (original) => ({ ...(await original<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key, te: () => true, locale: { value: 'en' } }) }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ sidebarCollapsed: false, mobileOpen: false }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => fixture.auth }))
vi.mock('@/features/image-studio/useImageStudio', () => ({ useImageStudio: () => fixture.state, downloadImage: vi.fn() }))

enableAutoUnmount(afterEach)
let observeSubmit: (visible: boolean) => void = () => {}
beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn()
  vi.stubGlobal('visualViewport', undefined)
  vi.stubGlobal('IntersectionObserver', class {
    constructor(callback: IntersectionObserverCallback) {
      observeSubmit = visible => callback([{ isIntersecting: visible, intersectionRatio: visible ? 1 : 0 } as IntersectionObserverEntry], this as unknown as IntersectionObserver)
    }
    observe() {}
    disconnect() {}
  })
  fixture.auth = reactive({ user: { id: 1 } })
  fixture.state = {
    models: [], enabled: true, catalogLoading: false, loading: false, error: '', storageError: '',
    modelID: '', prompt: '', size: 'auto', quality: 'auto', count: 1, aspectRatio: '1:1', resolution: '1K',
    references: [], layout: 'side', assets: [], prompts: [], categoryID: null, selectedModel: undefined, activeCount: 0, canGenerate: false,
    categories: [{ id: 'a', name: 'Design' }, { id: 'b', name: 'Ideas' }],
    history: [{ id: 'one', createdAt: 1, status: 'success', snapshot: { model: 'image-model', prompt: 'A sailboat', size: 'auto', quality: 'auto' }, images: [new Blob(['x'])], references: [], categoryID: 'a', categoryIDs: ['a'], favorite: false }],
    tasks: [{ id: 'one', status: 'success', saved: true }],
    refresh: vi.fn(), submit: vi.fn(), cancel: vi.fn(), dismissTask: vi.fn(), retry: vi.fn(), loadHistory: vi.fn(), newDraft: vi.fn(), addReferences: vi.fn(), removeReference: vi.fn(), deleteHistory: vi.fn(), favorite: vi.fn(), categorize: vi.fn(), saveCategory: vi.fn(), deleteCategory: vi.fn(), saveAsset: vi.fn(), deleteAsset: vi.fn(), savePrompt: vi.fn(), deletePrompt: vi.fn(), exportBackup: vi.fn(), importBackup: vi.fn(), saveUnsaved: vi.fn()
  }
})

function render() {
  return mount(ImageStudioView, { attachTo: document.body, global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
    GenerationPending: true, ConfirmDialog: true, BlobImage: true, Select: true, Icon: true, MaskEditor: true, WorkflowLibrary: true
  } } })
}

describe('ImageStudioView', () => {
  it('requires confirmation for another generation from either entry and submits once', async () => {
    fixture.state.activeCount = 1
    fixture.state.canGenerate = true
    const wrapper = render()
    await wrapper.get('form').trigger('submit')
    expect(fixture.state.submit).not.toHaveBeenCalled()
    expect(wrapper.getComponent(ConfirmDialog).props('show')).toBe(true)
    wrapper.getComponent(ConfirmDialog).vm.$emit('cancel')
    await nextTick()
    expect(fixture.state.submit).not.toHaveBeenCalled()
    observeSubmit(false)
    await nextTick()
    await wrapper.get('[data-generate]').trigger('click')
    wrapper.getComponent(ConfirmDialog).vm.$emit('confirm')
    wrapper.getComponent(ConfirmDialog).vm.$emit('confirm')
    await flushPromises()
    expect(fixture.state.submit).toHaveBeenCalledTimes(1)
  })

  it('makes favorites readable and dismisses completion feedback after five seconds', async () => {
    vi.useFakeTimers()
    fixture.state = reactive(fixture.state)
    fixture.state.history[0].favorite = true
    const wrapper = render()
    const favorite = wrapper.get('[aria-label="imageStudio.unfavorite"]')
    expect(favorite.text()).toContain('imageStudio.favorited')
    expect(favorite.find('icon-stub').attributes('name')).toBe('check')
    fixture.state.completedBatch = { id: 'new-batch', success: 2, failed: 0, cancelled: 0, unsaved: 0 }
    await nextTick()
    expect(wrapper.find('[data-studio-notice]').text()).toContain('imageStudio.batchCompleted')
    await vi.advanceTimersByTimeAsync(5000)
    expect(wrapper.find('[data-studio-notice]').exists()).toBe(false)
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('offers a dock when the original submit is clipped, with the same validation and one submission', async () => {
    fixture.state = reactive(fixture.state)
    const wrapper = render()
    observeSubmit(false)
    await nextTick()
    const dock = wrapper.get('[data-generation-dock]')
    expect(dock.isVisible()).toBe(true)
    expect(dock.get('[data-generate]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('.composer-submit button[type="submit"]').isVisible()).toBe(false)
    fixture.state.canGenerate = true
    await nextTick()
    await dock.get('[data-generate]').trigger('click')
    expect(fixture.state.submit).toHaveBeenCalledTimes(1)
    observeSubmit(true)
    await nextTick()
    expect(wrapper.find('[data-generation-dock]').exists()).toBe(false)
    expect(wrapper.get('.composer-submit button[type="submit"]').isVisible()).toBe(true)
    wrapper.unmount()
  })

  it('keeps the action above the visual viewport when a keyboard covers the original button', async () => {
    const viewport = Object.assign(new EventTarget(), { height: window.innerHeight, offsetTop: 0 })
    vi.stubGlobal('visualViewport', viewport)
    const wrapper = render()
    const anchor = wrapper.get('.composer-submit > div').element
    vi.spyOn(anchor, 'getBoundingClientRect').mockReturnValue({ bottom: 500 } as DOMRect)
    observeSubmit(true)
    await nextTick()
    expect(wrapper.find('[data-generation-dock]').exists()).toBe(false)
    viewport.height = 300
    viewport.dispatchEvent(new Event('resize'))
    await nextTick()
    expect(wrapper.get('[data-generation-dock]').attributes('style')).toContain(`bottom: ${window.innerHeight - 300}px`)
    viewport.height = window.innerHeight
    viewport.dispatchEvent(new Event('resize'))
    await nextTick()
    expect(wrapper.find('[data-generation-dock]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('hides the dock while a dialog is open and rejects invalid custom dimensions', async () => {
    fixture.state.canGenerate = true
    fixture.state.size = '17x17'
    fixture.state.selectedModel = { sizes: ['1024x1024'], qualities: ['auto'], supports_custom_size: true, pricing: { mode: 'token' } }
    const wrapper = render()
    observeSubmit(false)
    await nextTick()
    expect(wrapper.get('[data-generation-dock] [data-generate]').attributes('disabled')).toBeDefined()
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.workflows')!.trigger('click')
    expect(wrapper.find('[data-generation-dock]').exists()).toBe(false)
    expect(fixture.state.submit).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('reveals newly completed local work once without animating loaded history', async () => {
    fixture.state = reactive(fixture.state)
    const wrapper = render()
    expect(wrapper.find('[data-new-result="true"]').exists()).toBe(false)
    const task = { id: 'new', status: 'running', createdAt: Date.now(), snapshot: { prompt: 'New image', model: 'image-model', size: 'auto', quality: 'auto' } }
    fixture.state.tasks = [task]
    await nextTick()
    expect(wrapper.find('generation-pending-stub').attributes('status')).toBe('running')
    fixture.state.history = [{ ...fixture.state.history[0], id: 'new' }, ...fixture.state.history]
    fixture.state.tasks = []
    await nextTick()
    const result = wrapper.get('[data-new-result="true"]')
    await result.trigger('animationend')
    expect(wrapper.find('[data-new-result="true"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('pastes clipboard image files directly without reading clipboard permissions', async () => {
    const wrapper = render()
    const image = new File(['image'], 'paste.png', { type: 'image/png' })
    const event = new Event('paste', { bubbles: true, cancelable: true })
    Object.defineProperty(event, 'clipboardData', { value: { items: [{ kind: 'file', type: 'image/png', getAsFile: () => image }], files: [image] } })
    document.dispatchEvent(event)
    await flushPromises()
    expect(event.defaultPrevented).toBe(true)
    expect(fixture.state.addReferences).toHaveBeenCalledTimes(1)
    expect(fixture.state.addReferences).toHaveBeenCalledWith([image])
    wrapper.unmount()
    document.dispatchEvent(event)
    expect(fixture.state.addReferences).toHaveBeenCalledTimes(1)
  })

  it('preserves plain text paste and ignores image paste in open dialogs', async () => {
    const wrapper = render()
    const text = new Event('paste', { bubbles: true, cancelable: true })
    Object.defineProperty(text, 'clipboardData', { value: { items: [], files: [] } })
    document.dispatchEvent(text)
    expect(text.defaultPrevented).toBe(false)
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.workflows')!.trigger('click')
    const event = new Event('paste', { bubbles: true, cancelable: true })
    Object.defineProperty(event, 'clipboardData', { value: { files: [new File(['image'], 'paste.png', { type: 'image/png' })] } })
    document.dispatchEvent(event)
    expect(event.defaultPrevented).toBe(false)
    expect(fixture.state.addReferences).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('highlights file drag within the reference zone and adds dropped images once', async () => {
    const wrapper = render()
    const image = new File(['image'], 'drop.png', { type: 'image/png' })
    const zone = wrapper.get('[data-reference-dropzone]')
    const dataTransfer = { types: ['Files'], files: [image], dropEffect: 'none' }
    await zone.trigger('dragenter', { dataTransfer })
    expect(zone.attributes('data-drag-active')).toBe('true')
    await zone.trigger('dragenter', { dataTransfer })
    await zone.trigger('dragleave', { dataTransfer })
    expect(zone.attributes('data-drag-active')).toBe('true')
    await zone.trigger('drop', { dataTransfer })
    expect(zone.attributes('data-drag-active')).toBe('false')
    expect(fixture.state.addReferences).toHaveBeenCalledTimes(1)
    expect(fixture.state.addReferences).toHaveBeenCalledWith([image])
    wrapper.unmount()
  })

  it('previews an uploaded asset without adding it as a reference or generating', async () => {
    fixture.state.assets = [{ id: 'asset', name: 'Uploaded landscape', image: new Blob(['asset'], { type: 'image/png' }) }]
    const wrapper = render()
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.assets')!.trigger('click')
    const previewButton = wrapper.find('[data-preview-asset]')
    expect(previewButton.exists()).toBe(true)
    await previewButton.trigger('click')
    expect(wrapper.find('[data-image-preview]').exists()).toBe(true)
    expect(fixture.state.addReferences).not.toHaveBeenCalled()
    expect(fixture.state.submit).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('allows inspecting the references saved with a generated result', async () => {
    fixture.state.history[0].references = [new Blob(['reference'], { type: 'image/png' })]
    const wrapper = render()
    const reference = wrapper.find('.result-card [data-preview-reference]')
    expect(reference.exists()).toBe(true)
    await reference.trigger('click')
    expect(wrapper.find('[data-image-preview]').exists()).toBe(true)
    expect(fixture.state.submit).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('closes image previews when the account changes', async () => {
    const wrapper = render()
    await wrapper.get('[aria-label="imageStudio.previewImage"]').trigger('click')
    expect(wrapper.find('input[type="range"]').exists()).toBe(true)
    fixture.auth.user.id = 2
    await nextTick()
    expect(wrapper.find('input[type="range"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows saved success only in history and loads its original snapshot', async () => {
    const wrapper = render()
    expect(wrapper.findAll('.result-card')).toHaveLength(1)
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.load')!.trigger('click')
    expect(fixture.state.loadHistory).toHaveBeenCalledWith(fixture.state.history[0])
    wrapper.unmount()
  })

  it('submits exactly one mask edit when the normal batch count is larger', async () => {
    fixture.state.count = 3
    fixture.state.selectedModel = { id: 'native', supports_mask: true, max_references: 3, pricing: { mode: 'token' } }
    const original = new Blob(['original'], { type: 'image/png' }), mask = new Blob(['mask'], { type: 'image/png' })
    const wrapper = render()
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.maskEdit')!.trigger('click')
    wrapper.getComponent(MaskEditor).vm.$emit('confirm', { original, mask, prompt: 'blue ceramic' })
    await flushPromises()
    expect(fixture.state.submit).toHaveBeenCalledWith(['imageStudio.nativeMaskPromptPrefixblue ceramic'], { modelID: 'native', references: [original], mask })
    expect(fixture.state.references).toEqual([])
    expect(fixture.state.prompt).toBe('')
    wrapper.unmount()
  })

  it('keeps guided edit inputs separate and confirms before submitting during an active task', async () => {
    fixture.state.selectedModel = { id: 'guided', supports_mask: false, max_references: 3, pricing: { mode: 'token' } }
    fixture.state.activeCount = 1
    const original = new Blob(['original'], { type: 'image/png' }), guide = new Blob(['guide'], { type: 'image/png' })
    const wrapper = render()
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.maskEdit')!.trigger('click')
    wrapper.getComponent(MaskEditor).vm.$emit('confirm', { original, guide, prompt: 'remove object' })
    await flushPromises()
    expect(fixture.state.submit).not.toHaveBeenCalled()
    wrapper.getComponent(ConfirmDialog).vm.$emit('confirm')
    wrapper.getComponent(ConfirmDialog).vm.$emit('confirm')
    await flushPromises()
    expect(fixture.state.submit).toHaveBeenCalledTimes(1)
    expect(fixture.state.submit).toHaveBeenCalledWith(['imageStudio.maskPromptPrefixremove object'], { modelID: 'guided', references: [original, guide] })
  })

  it('confirms before dismissing a failed task', async () => {
    fixture.state.tasks = [{ id: 'failed', status: 'failed', createdAt: Date.now(), snapshot: { prompt: 'Failed image', model: 'image-model', size: 'auto', quality: 'auto' } }]
    const wrapper = render()
    await wrapper.findAll('.result-card')[0].findAll('button').find(button => button.text() === 'imageStudio.delete')!.trigger('click')
    expect(fixture.state.dismissTask).not.toHaveBeenCalled()
    wrapper.getComponent(ConfirmDialog).vm.$emit('confirm')
    await flushPromises()
    expect(fixture.state.dismissTask).toHaveBeenCalledWith('failed')
    wrapper.unmount()
  })

  it('preserves existing categories when assigning a second category', async () => {
    const wrapper = render()
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.categorize' && button.attributes('disabled') === undefined)!.trigger('click')
    const categories = wrapper.findAll('[data-category-choice]')
    expect(categories).toHaveLength(2)
    expect((categories[0].element as HTMLInputElement).checked).toBe(true)
    await categories[1].setValue(true)
    await wrapper.findAll('button').find(button => button.text() === 'imageStudio.save')!.trigger('click')
    expect(fixture.state.categorize).toHaveBeenCalledWith(['one'], ['a', 'b'])
    wrapper.unmount()
  })
})
