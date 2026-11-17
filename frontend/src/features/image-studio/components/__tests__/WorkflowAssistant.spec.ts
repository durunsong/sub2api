import { mount, flushPromises } from '@vue/test-utils'
import { reactive } from 'vue'
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import WorkflowAssistant from '../WorkflowAssistant.vue'

const mocks = vi.hoisted(() => ({ user: { user: { id: 7 } as { id: number } | null }, get: vi.fn(), post: vi.fn() }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.user }))
vi.mock('@/api/client', () => ({ apiClient: { get: mocks.get, post: mocks.post } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const textModel = { id: 'text', group_id: 7, group_name: 'Text', model: 'gpt-4', label: 'GPT', supports_references: true }
let wrapper: ReturnType<typeof mount>
beforeEach(() => {
  mocks.user = reactive({ user: { id: 7 } })
  mocks.get.mockReset().mockResolvedValue({ data: { enabled: true, models: [textModel] } })
  mocks.post.mockReset()
})
afterEach(() => wrapper?.unmount())
async function render(mode: 'workflow' | 'series' = 'series', references: Blob[] = []) {
  wrapper = mount(WorkflowAssistant, { props: { mode, initialPrompt: 'initial brief', references }, global: { stubs: { Icon: true, Select: true } } })
  await flushPromises()
}
function button(key: string) { return wrapper.findAll('button').find(item => item.text() === `imageStudio.${key}`)! }

describe('workflow assistant lifecycle', () => {
  it('discards an old series response when its template variables change', async () => {
    let finish!: (data: unknown) => void
    mocks.post.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    await render()
    await button('aiGenerateSeries').trigger('click')
    await flushPromises()
    const signal = mocks.post.mock.calls[0][2].signal
    await wrapper.setProps({ initialPrompt: 'different product' })
    expect(signal.aborted).toBe(true)
    finish({ data: { prompts: ['one', 'two', 'three', 'four'] } })
    await flushPromises()
    expect(wrapper.emitted('series')).toBeUndefined()
    expect(wrapper.get('textarea').element.value).toBe('different product')
  })
  it('cancels and clears private input on sign-out without fetching another catalog', async () => {
    let finish!: (data: unknown) => void
    mocks.post.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    await render()
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    mocks.user.user = null
    await flushPromises()
    expect(mocks.post.mock.calls[0][2].signal.aborted).toBe(true)
    expect(wrapper.get('textarea').element.value).toBe('')
    expect(mocks.get).toHaveBeenCalledTimes(1)
    finish({ data: { prompts: ['one', 'two', 'three', 'four'] } }); await flushPromises()
    expect(wrapper.emitted('series')).toBeUndefined()
  })
  it('stops waiting without emitting a late result and permits a new manual attempt', async () => {
    let finish!: (data: unknown) => void
    mocks.post.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    await render()
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    expect(button('aiGenerateSeries').attributes('disabled')).toBeDefined()
    await button('cancel').trigger('click')
    expect(mocks.post.mock.calls[0][2].signal.aborted).toBe(true)
    finish({ data: { prompts: ['one', 'two', 'three', 'four'] } }); await flushPromises()
    expect(wrapper.emitted('series')).toBeUndefined()
    expect(button('aiGenerateSeries').attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).toContain('imageStudio.aiCancelled')
  })
  it('aborts on unmount and never retries upstream errors automatically', async () => {
    mocks.post.mockRejectedValueOnce(new Error('provider failed'))
    await render()
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('provider failed')
    expect(mocks.post).toHaveBeenCalledTimes(1)
    mocks.post.mockImplementationOnce(() => new Promise(() => {}))
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    wrapper.unmount()
    expect(mocks.post.mock.calls[1][2].signal.aborted).toBe(true)
  })
  it('sends local references only to models that declare vision support', async () => {
    mocks.post.mockResolvedValue({ data: { prompts: ['one', 'two', 'three', 'four'] } })
    await render('series', [new Blob(['image'], { type: 'image/png' })])
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    await vi.waitFor(() => expect(mocks.post).toHaveBeenCalledTimes(1))
    expect(mocks.post.mock.calls[0][1].references).toEqual(['data:image/png;base64,aW1hZ2U='])
    wrapper.unmount()
    mocks.get.mockResolvedValue({ data: { enabled: true, models: [{ ...textModel, supports_references: false }] } })
    await render('series', [new Blob(['image'], { type: 'image/png' })])
    await button('aiGenerateSeries').trigger('click'); await flushPromises()
    expect(mocks.post.mock.calls[1][1].references).toEqual([])
  })
  it('disables generation when no models are available', async () => {
    mocks.get.mockResolvedValue({ data: { enabled: true, models: [] } })
    await render()
    expect(wrapper.text()).toContain('imageStudio.aiNoModels')
    expect(button('aiGenerateSeries').attributes('disabled')).toBeDefined()
    expect(mocks.post).not.toHaveBeenCalled()
  })
})
