import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import { buildGatewayUrl } from '@/api/url'
import zh from '@/i18n/locales/zh/dashboard'
import AvailableModelsModal from '../AvailableModelsModal.vue'

const { listKeyModels, copyToClipboard } = vi.hoisted(() => ({
  listKeyModels: vi.fn(), copyToClipboard: vi.fn()
}))
vi.mock('@/api/keyModels', () => ({ listKeyModels }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard }) }))

const key = {
  key: 'sk-selected', group_id: 1, group: { name: 'ChatGPT', platform: 'openai' }
} as ApiKey
const wrappers: VueWrapper[] = []
function mountModal(apiKey: ApiKey | null = key, show = true) {
  const wrapper = mount(AvailableModelsModal, {
    props: { show, apiKey },
    global: {
      plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: { ...zh, common: { loading: '加载中...' } } } })],
      stubs: { teleport: true, transition: false }
    }
  })
  wrappers.push(wrapper)
  return wrapper
}
function deferred() {
  let resolve!: (models: string[]) => void
  let reject!: (error: Error) => void
  const promise = new Promise<string[]>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

describe('AvailableModelsModal', () => {
  beforeEach(() => {
    listKeyModels.mockReset().mockResolvedValue(['allowed-model', 'another-model'])
    copyToClipboard.mockReset().mockResolvedValue(true)
  })
  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
    vi.unstubAllGlobals()
  })

  it('shows the selected group and only returned model IDs, and copies a clicked ID', async () => {
    const wrapper = mountModal()
    await flushPromises()
    expect(listKeyModels).toHaveBeenCalledWith('sk-selected', expect.any(AbortSignal))
    expect(wrapper.get('[role=dialog]').text()).toContain('查看可用模型')
    expect(wrapper.text()).toContain('ChatGPT')
    expect(wrapper.text()).toContain('仅显示所选分组授权的模型')
    const models = wrapper.findAll('[data-testid=model-name]')
    expect(models.map((model) => model.text())).toEqual(['allowed-model', 'another-model'])
    await models[0].trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('allowed-model', '模型名称已复制')
    expect(wrapper.text()).not.toContain('sk-selected')
  })

  it('shows Antigravity custom aliases from the group-aware gateway catalog', async () => {
    const actual = await vi.importActual<typeof import('@/api/keyModels')>('@/api/keyModels')
    listKeyModels.mockImplementation(actual.listKeyModels)
    const fetchMock = vi.fn(async (url: string) => new Response(JSON.stringify({
      data: url === buildGatewayUrl('/v1/models') ? [{ id: 'my-model' }] : []
    })))
    vi.stubGlobal('fetch', fetchMock)
    const wrapper = mountModal({ ...key, group: { ...key.group!, name: 'Antigravity', platform: 'antigravity' } })
    await flushPromises()
    expect(wrapper.findAll('[data-testid=model-name]').map((model) => model.text())).toEqual(['my-model'])
    expect(fetchMock).toHaveBeenCalledWith(buildGatewayUrl('/v1/models'), expect.objectContaining({
      headers: { Accept: 'application/json', Authorization: 'Bearer sk-selected' }
    }))
  })

  it('does not fetch until opened', async () => {
    const wrapper = mountModal(key, false)
    expect(listKeyModels).not.toHaveBeenCalled()
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(listKeyModels).toHaveBeenCalledTimes(1)
  })

  it('shows loading, then an explicit empty state', async () => {
    const request = deferred()
    listKeyModels.mockReturnValue(request.promise)
    const wrapper = mountModal()
    expect(wrapper.get('[role=status]').text()).toContain('加载中')
    request.resolve([])
    await flushPromises()
    expect(wrapper.get('[role=status]').text()).toContain('当前分组暂无可用模型')
  })

  it('explains a missing group without sending the key', async () => {
    const wrapper = mountModal({ ...key, group_id: null, group: undefined })
    await flushPromises()
    expect(listKeyModels).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('请先为此密钥选择分组')
  })

  it('preserves the user-facing Claude group alias', async () => {
    const wrapper = mountModal({ ...key, group: { ...key.group!, name: 'Kiro Pro', platform: 'kiro' } })
    await flushPromises()
    expect(wrapper.text()).toContain('Claude Pro')
    expect(wrapper.text()).not.toContain('Kiro')
    expect(listKeyModels).toHaveBeenCalledWith('sk-selected', expect.any(AbortSignal))
  })

  it('shows a safe error and lets the user retry', async () => {
    listKeyModels.mockRejectedValueOnce(new Error('private upstream details'))
    const wrapper = mountModal()
    await flushPromises()
    expect(wrapper.get('[role=alert]').text()).toContain('获取模型失败')
    expect(wrapper.text()).not.toContain('private upstream details')
    await wrapper.get('[data-testid=retry-models]').trigger('click')
    await flushPromises()
    expect(listKeyModels).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.findAll('[data-testid=model-name]')).toHaveLength(2)
  })

  it.each(['resolve', 'reject'] as const)('ignores an old request that later %ss after switching keys', async (finish) => {
    const first = deferred()
    listKeyModels.mockReturnValueOnce(first.promise).mockResolvedValueOnce(['second-group-model'])
    const wrapper = mountModal()
    const firstSignal = listKeyModels.mock.calls[0][1] as AbortSignal
    await wrapper.setProps({ apiKey: { ...key, key: 'sk-second', group_id: 2, group: { ...key.group!, name: 'Second group' } } })
    await flushPromises()
    expect(firstSignal.aborted).toBe(true)
    if (finish === 'resolve') first.resolve(['old-group-model'])
    else first.reject(new Error('old failure'))
    await flushPromises()
    expect(wrapper.text()).toContain('Second group')
    expect(wrapper.findAll('[data-testid=model-name]').map((model) => model.text())).toEqual(['second-group-model'])
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
  })

  it('clears the prior group models while a new group is loading', async () => {
    const wrapper = mountModal()
    await flushPromises()
    listKeyModels.mockReturnValueOnce(deferred().promise)
    await wrapper.setProps({ apiKey: { ...key, group_id: 2 } })
    expect(wrapper.findAll('[data-testid=model-name]')).toHaveLength(0)
    expect(wrapper.get('[role=status]').text()).toContain('加载中')
  })

  it('aborts on close and fetches fresh data when reopened', async () => {
    const request = deferred()
    listKeyModels.mockReturnValueOnce(request.promise)
    const wrapper = mountModal()
    const signal = listKeyModels.mock.calls[0][1] as AbortSignal
    await wrapper.get('[aria-label="Close modal"]').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
    expect(signal.aborted).toBe(true)
    await wrapper.setProps({ show: false })
    request.resolve(['old-model'])
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(listKeyModels).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).not.toContain('old-model')
  })

  it('aborts a pending request on unmount', () => {
    listKeyModels.mockReturnValueOnce(deferred().promise)
    const wrapper = mountModal()
    const signal = listKeyModels.mock.calls[0][1] as AbortSignal
    wrapper.unmount()
    expect(signal.aborted).toBe(true)
  })
})
