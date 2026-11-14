import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { SubscriptionResetHistory } from '@/types'
import SubscriptionResetHistoryDialog from '../SubscriptionResetHistoryDialog.vue'

const { getUserHistory, getAdminHistory } = vi.hoisted(() => ({ getUserHistory: vi.fn(), getAdminHistory: vi.fn() }))
vi.mock('@/api/subscriptions', () => ({ default: { getResetHistory: getUserHistory } }))
vi.mock('@/api/admin/subscriptions', () => ({ default: { getResetHistory: getAdminHistory } }))
vi.mock('vue-i18n', async () => ({ ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'), useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => key + (params ? ':' + Object.values(params).join(',') : '') }) }))

const card: SubscriptionResetHistory = { id: 1, type: 'reset_card', reset_at: '2026-11-14T08:00:00Z', daily: true, weekly: true, monthly: true, validity_days: 7 }
const admin: SubscriptionResetHistory = { id: 2, type: 'admin', reset_at: '2026-11-14T09:00:00Z', daily: false, weekly: true, monthly: false }
const result = (items: SubscriptionResetHistory[] = [card], total = items.length) => ({ items, total, page: 1, page_size: 20, pages: Math.ceil(total / 20) })
const mountDialog = (props = {}) => mount(SubscriptionResetHistoryDialog, {
  props: { subscriptionId: 42, subscriptionName: 'Plan', ...props },
  global: { stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' }, Pagination: true } }
})

describe('SubscriptionResetHistoryDialog', () => {
  beforeEach(() => { vi.clearAllMocks(); getUserHistory.mockResolvedValue(result()); getAdminHistory.mockResolvedValue(result()) })

  it('loads only when opened and renders card term and selected reset windows', async () => {
    const wrapper = mountDialog({ subscriptionId: null })
    expect(getUserHistory).not.toHaveBeenCalled()
    getUserHistory.mockResolvedValue(result([card, admin]))
    await wrapper.setProps({ subscriptionId: 42 })
    await flushPromises()
    expect(getUserHistory).toHaveBeenCalledWith(42, 1, 20, expect.any(AbortSignal))
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('userSubscriptions.resetHistory.cardTerm:7')
    expect(rows[1].text()).toContain('userSubscriptions.weekly')
    expect(rows[1].text()).not.toContain('userSubscriptions.daily')
    expect(wrapper.text()).toContain('userSubscriptions.resetHistory.coverageHint')
    wrapper.unmount()
  })

  it('uses the admin endpoint and fetches the selected page', async () => {
    getAdminHistory.mockResolvedValue(result([card], 21))
    const wrapper = mountDialog({ admin: true })
    await flushPromises()
    wrapper.getComponent({ name: 'Pagination' }).vm.$emit('update:page', 2)
    await flushPromises()
    expect(getAdminHistory).toHaveBeenLastCalledWith(42, 2, 20, expect.any(AbortSignal))
    expect(getUserHistory).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('distinguishes an error from an empty history and supports retry', async () => {
    getUserHistory.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce(result([]))
    const wrapper = mountDialog()
    await flushPromises()
    expect(wrapper.get('[role=alert]').text()).toContain('userSubscriptions.resetHistory.failedToLoad')
    expect(wrapper.text()).not.toContain('userSubscriptions.resetHistory.empty')
    await wrapper.get('[data-test=retry-history]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role=alert]').exists()).toBe(false)
    expect(wrapper.text()).toContain('userSubscriptions.resetHistory.empty')
    wrapper.unmount()
  })

  it('cancels old requests and ignores stale results when switching or closing', async () => {
    let resolveOld!: (value: ReturnType<typeof result>) => void
    getUserHistory.mockReturnValueOnce(new Promise(resolve => { resolveOld = resolve }))
    const wrapper = mountDialog()
    const firstSignal = getUserHistory.mock.calls[0][3] as AbortSignal
    getUserHistory.mockResolvedValueOnce(result([]))
    await wrapper.setProps({ subscriptionId: 99 })
    await flushPromises()
    expect(firstSignal.aborted).toBe(true)
    resolveOld(result())
    await flushPromises()
    expect(wrapper.findAll('tbody tr')).toHaveLength(0)
    expect(wrapper.text()).toContain('userSubscriptions.resetHistory.empty')
    const secondSignal = getUserHistory.mock.calls[1][3] as AbortSignal
    await wrapper.setProps({ subscriptionId: null })
    expect(secondSignal.aborted).toBe(true)
    await wrapper.setProps({ subscriptionId: 42 })
    await flushPromises()
    expect(getUserHistory).toHaveBeenLastCalledWith(42, 1, 20, expect.any(AbortSignal))
    wrapper.unmount()
    expect((getUserHistory.mock.calls[2][3] as AbortSignal).aborted).toBe(true)
  })
})
