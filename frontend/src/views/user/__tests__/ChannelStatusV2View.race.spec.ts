import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const api = vi.hoisted(() => ({
  getDimensions: vi.fn(),
  getSnapshot: vi.fn(),
  getMatrix: vi.fn(),
  getModels: vi.fn(),
  getErrors: vi.fn(),
  getUsers: vi.fn(),
}))

vi.mock('@/api/channelMonitorV2', () => api)
vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({ replace: vi.fn() }),
}))
vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key, te: () => false, locale: { value: 'zh-CN' } }),
  }
})
vi.mock('@/utils/featureFlags', () => ({ isChannelMonitorThroughputHidden: () => false }))

import { useAppStore } from '@/stores/app'
import ChannelStatusV2View from '../ChannelStatusV2View.vue'

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((next, fail) => {
    resolve = next
    reject = fail
  })
  return { promise, resolve, reject }
}


const metric = {
  success_requests: 0,
  error_requests: 0,
  request_count: 0,
  token_count: 0,
  rpm: 0,
  tpm: 0,
  error_rate: 0,
  cache_rate: 0,
  cache_rate_numerator: 0,
  cache_rate_denominator: 0,
  ttft: { sample_count: 0, p50_ms: null, p95_ms: null, avg_ms: null },
  duration: { sample_count: 0, p50_ms: null, p95_ms: null, avg_ms: null },
}
const coverage = {
  requested_start: '',
  coverage_start: '',
  data_through: '',
  computed_at: '',
  aggregation_lag_seconds: 0,
  coverage_complete: true,
  bucket_seconds: 300,
}
const snapshot = {
  config: { refresh_interval_seconds: 300 },
  coverage,
  metrics: metric,
  health: { overall: 'unknown', error_rate: 'unknown', ttft: 'unknown', minimum_sample: 1 },
  trend: [],
}

const Stub = defineComponent({ setup: () => () => h('div') })

function mountView() {
  return mount(ChannelStatusV2View, {
    global: {
      plugins: [createPinia()],
      stubs: {
        AppLayout: defineComponent({ setup: (_, { slots }) => () => h('div', slots.default?.()) }),
        Icon: Stub,
        LoadingSpinner: Stub,
        Select: Stub,
        FilterMultiSelect: Stub,
        MetricCell: Stub,
        MonitorRankBadge: Stub,
        MonitorTrendChart: Stub,
        RelayPulseMatrix: Stub,
      },
    },
  })
}

describe('ChannelStatusV2View tab requests', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.getDimensions.mockResolvedValue({ platforms: [], groups: [], models: [] })
    api.getSnapshot.mockReturnValue(new Promise(() => {}))
    api.getMatrix.mockReturnValue(new Promise(() => {}))
    api.getModels.mockResolvedValue({ items: [] })
  })

  it('does not let the main reload clear a newer tab request loading state', async () => {
    const models = deferred<{ items: [] }>()
    const users = deferred<{ items: [] }>()
    api.getSnapshot.mockResolvedValueOnce(snapshot)
    api.getMatrix.mockResolvedValueOnce({ coverage, group_by: 'platform_group', items: [] })
    api.getModels.mockReturnValueOnce(models.promise)
    api.getUsers.mockReturnValueOnce(users.promise)
    const wrapper = mountView()
    await flushPromises()

    const mainSignal = api.getModels.mock.calls[0]?.[2] as AbortSignal
    await wrapper.findAll('[role="tab"]')[2].trigger('click')
    const tabSignal = api.getUsers.mock.calls[0]?.[2] as AbortSignal
    expect(mainSignal.aborted).toBe(false)
    expect(tabSignal.aborted).toBe(false)

    models.resolve({ items: [] })
    await flushPromises()
    expect(wrapper.find('.empty-state').text()).toBe('common.loading')

    users.resolve({ items: [] })
    await flushPromises()
    expect(wrapper.find('.empty-state').text()).not.toBe('common.loading')
  })

  it('aborts the previous independent tab request without aborting the main reload signal', async () => {
    const models = deferred<{ items: [] }>()
    const errors = deferred<{ items: [] }>()
    const users = deferred<{ items: [] }>()
    api.getSnapshot.mockResolvedValueOnce(snapshot)
    api.getMatrix.mockResolvedValueOnce({ coverage, group_by: 'platform_group', items: [] })
    api.getModels.mockReturnValueOnce(models.promise)
    api.getErrors.mockReturnValueOnce(errors.promise)
    api.getUsers.mockReturnValueOnce(users.promise)
    const wrapper = mountView()
    await flushPromises()

    const mainSignal = api.getModels.mock.calls[0]?.[2] as AbortSignal
    const tabs = wrapper.findAll('[role="tab"]')
    await tabs[1].trigger('click')
    const errorsSignal = api.getErrors.mock.calls[0]?.[2] as AbortSignal
    await tabs[2].trigger('click')

    expect(errorsSignal.aborted).toBe(true)
    expect(mainSignal.aborted).toBe(false)

    models.resolve({ items: [] })
    users.resolve({ items: [] })
    await flushPromises()
  })

  it('ignores a stale errors failure after the users tab has loaded', async () => {
    const errors = deferred<{ items: [] }>()
    const users = deferred<{ items: [] }>()
    api.getErrors.mockReturnValueOnce(errors.promise)
    api.getUsers.mockReturnValueOnce(users.promise)
    const wrapper = mountView()

    const tabs = wrapper.findAll('[role="tab"]')
    await tabs[1].trigger('click')
    await tabs[2].trigger('click')
    users.resolve({ items: [] })
    await flushPromises()
    expect(wrapper.find('.empty-state').text()).not.toBe('common.loading')

    errors.reject(new Error('stale errors request'))
    await flushPromises()

    expect(useAppStore().toasts).toHaveLength(0)
  })
  it('keeps the active tab loading when an older tab request finishes first', async () => {
    const errors = deferred<{ items: [] }>()
    const users = deferred<{ items: [] }>()
    api.getErrors.mockReturnValueOnce(errors.promise)
    api.getUsers.mockReturnValueOnce(users.promise)
    const wrapper = mountView()

    const tabs = wrapper.findAll('[role="tab"]')
    await tabs[1].trigger('click')
    await tabs[2].trigger('click')
    expect(api.getErrors).toHaveBeenCalledOnce()
    expect(api.getUsers).toHaveBeenCalledOnce()

    errors.resolve({ items: [] })
    await flushPromises()

    expect(wrapper.find('.empty-state').text()).toBe('common.loading')
    users.resolve({ items: [] })
    await flushPromises()
    expect(wrapper.find('.empty-state').text()).not.toBe('common.loading')
  })
})
