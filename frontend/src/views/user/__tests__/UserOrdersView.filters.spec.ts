import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import UserOrdersView from '../UserOrdersView.vue'
import Pagination from '@/components/common/Pagination.vue'

const api = vi.hoisted(() => ({ getMyOrders: vi.fn(), getRefundEligibleProviders: vi.fn() }))
vi.mock('@/api/payment', () => ({ paymentAPI: api }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key })
}))
enableAutoUnmount(afterEach)

beforeEach(() => {
  vi.clearAllMocks()
  api.getMyOrders.mockResolvedValue({ data: { items: [], total: 100 } })
  api.getRefundEligibleProviders.mockResolvedValue({ data: { provider_instance_ids: [] } })
})

async function openOrders() {
  const wrapper = mount(UserOrdersView, {
    global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      OrderTable: true, BaseDialog: true, Icon: true, Pagination: true, teleport: true
    } }
  })
  await flushPromises()
  wrapper.getComponent(Pagination).vm.$emit('update:page', 4)
  await flushPromises()
  return wrapper
}

describe('order status filtering', () => {
  it('shows all five status buttons with All selected initially', async () => {
    const wrapper = await openOrders()
    const buttons = wrapper.get('[role="group"]').findAll('button')
    expect(buttons.map(button => button.text())).toEqual([
      'common.all', 'payment.status.pending', 'payment.status.completed',
      'payment.status.failed', 'payment.status.refunded'
    ])
    expect(buttons.map(button => button.attributes('aria-pressed'))).toEqual([
      'true', 'false', 'false', 'false', 'false'
    ])
  })

  it.each([
    ['payment.status.pending', 'PENDING'],
    ['payment.status.completed', 'COMPLETED'],
    ['payment.status.failed', 'FAILED'],
    ['payment.status.refunded', 'REFUNDED'],
  ])('loads the first page when selecting %s', async (label, status) => {
    const wrapper = await openOrders()
    const button = wrapper.get('[role="group"]').findAll('button').find(button => button.text() === label)!
    await button.trigger('click')
    await flushPromises()
    expect(api.getMyOrders).toHaveBeenLastCalledWith({ page: 1, page_size: 20, status })
    expect(wrapper.getComponent(Pagination).props('page')).toBe(1)
    expect(api.getMyOrders).toHaveBeenCalledTimes(3)
    expect(button.attributes('aria-pressed')).toBe('true')
    expect(wrapper.findAll('[aria-pressed="true"]')).toHaveLength(1)
  })

  it('clears the status and returns to the first page when selecting All', async () => {
    const wrapper = await openOrders()
    const buttons = wrapper.get('[role="group"]').findAll('button')
    await buttons[1].trigger('click')
    await flushPromises()
    wrapper.getComponent(Pagination).vm.$emit('update:page', 3)
    await flushPromises()
    await buttons[0].trigger('click')
    await flushPromises()
    expect(api.getMyOrders).toHaveBeenLastCalledWith({ page: 1, page_size: 20, status: undefined })
    expect(buttons[0].attributes('aria-pressed')).toBe('true')
  })

  it('does not reload or reset pagination when selecting the active status', async () => {
    const wrapper = await openOrders()
    await wrapper.get('[aria-pressed="true"]').trigger('click')
    await flushPromises()
    expect(api.getMyOrders).toHaveBeenCalledTimes(2)
    expect(wrapper.getComponent(Pagination).props('page')).toBe(4)
  })

  it('disables status changes while loading and enables them after completion', async () => {
    const wrapper = await openOrders()
    let complete!: (value: { data: { items: never[]; total: number } }) => void
    api.getMyOrders.mockImplementationOnce(() => new Promise(resolve => { complete = resolve }))
    await wrapper.get('[title="common.refresh"]').trigger('click')
    const buttons = wrapper.get('[role="group"]').findAll('button')
    expect(buttons.every(button => button.attributes('disabled') !== undefined)).toBe(true)
    await buttons[1].trigger('click')
    expect(api.getMyOrders).toHaveBeenCalledTimes(3)
    complete({ data: { items: [], total: 100 } })
    await flushPromises()
    expect(buttons.every(button => button.attributes('disabled') === undefined)).toBe(true)
  })

  it('keeps the current page on manual refresh', async () => {
    const wrapper = await openOrders()
    await wrapper.get('[title="common.refresh"]').trigger('click')
    await flushPromises()
    expect(api.getMyOrders).toHaveBeenLastCalledWith({ page: 4, page_size: 20, status: undefined })
  })
})
