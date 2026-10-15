import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import History from '../RedeemEmailBatchHistoryDialog.vue'

const api = vi.hoisted(() => ({ list: vi.fn(), delete: vi.fn() }))
vi.mock('@/api/admin', () => ({
  adminAPI: { redeem: { listEmailBatches: api.list, deleteEmailBatch: api.delete } }
}))
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key)
  })
}))
vi.mock('@/utils/format', () => ({ formatDateTime: (value: string) => `formatted:${value}` }))

const batch = (id: number, status: string, extra: Record<string, unknown> = {}) => ({
  id,
  status,
  segments: ['inactive_2_3m', 'never_used'],
  reference_time: '2026-10-15T10:00:00Z',
  coupon_value: 1.5,
  code_expires_at: null,
  total_count: 2,
  pending_count: 0,
  sending_count: 0,
  sent_count: 2,
  failed_count: 0,
  skipped_count: 0,
  created_by: 1,
  started_at: null,
  finished_at: null,
  created_at: '2026-10-15T10:00:00Z',
  updated_at: '2026-10-15T10:00:00Z',
  ...extra
})

const mountHistory = () =>
  mount(History, {
    props: { show: true },
    global: {
      stubs: {
        BaseDialog: { props: ['show', 'title'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
        Pagination: true,
        Icon: true,
        RedeemEmailBatchDetailDialog: { props: ['show', 'batchId'], template: '<div v-if="show" data-test="detail-stub">{{ batchId }}</div>' },
        ConfirmDialog: {
          props: ['show'],
          template: '<button v-if="show" data-test="confirm-delete" @click="$emit(\'confirm\')" />'
        }
      }
    }
  })

describe('RedeemEmailBatchHistoryDialog', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    api.list.mockReset()
    api.delete.mockReset().mockResolvedValue({ message: 'ok' })
  })
  afterEach(() => vi.useRealTimers())

  it('lists batches with value, segments, progress and hides delete for active batches', async () => {
    api.list.mockResolvedValue({
      items: [batch(9, 'succeeded', { sent_count: 1, skipped_count: 1 }), batch(10, 'running', { sent_count: 0, pending_count: 2 })],
      total: 2
    })
    const wrapper = mountHistory()
    await flushPromises()

    expect(wrapper.text()).toContain('$1.50')
    expect(wrapper.get('[data-test="batch-segments-9"]').text()).toBe(
      'admin.redeemEmail.recipients.segment.inactive_2_3m / admin.redeemEmail.recipients.segment.never_used'
    )
    expect(wrapper.text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('admin.redeemEmail.history.skippedCount:{"n":1}')
    expect(wrapper.get('[data-test="batch-status-10"]').text()).toBe('admin.redeemEmail.status.running')
    expect(wrapper.find('[data-test="delete-batch-9"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="delete-batch-10"]').exists()).toBe(false)

    await wrapper.get('[data-test="view-batch-9"]').trigger('click')
    expect(wrapper.get('[data-test="detail-stub"]').text()).toBe('9')
    wrapper.unmount()
  })

  it('polls while a batch is active and stops once all are terminal', async () => {
    api.list
      .mockResolvedValueOnce({ items: [batch(10, 'running')], total: 1 })
      .mockResolvedValue({ items: [batch(10, 'succeeded')], total: 1 })
    const wrapper = mountHistory()
    await flushPromises()
    expect(api.list).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(5000)
    await flushPromises()
    expect(api.list).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-test="batch-status-10"]').text()).toBe('admin.redeemEmail.status.succeeded')

    await vi.advanceTimersByTimeAsync(15000)
    expect(api.list).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('deletes a terminal batch after confirmation and shows mapped errors', async () => {
    api.list.mockResolvedValue({ items: [batch(9, 'failed')], total: 1 })
    const wrapper = mountHistory()
    await flushPromises()

    await wrapper.get('[data-test="delete-batch-9"]').trigger('click')
    await wrapper.get('[data-test="confirm-delete"]').trigger('click')
    await flushPromises()
    expect(api.delete).toHaveBeenCalledWith(9)
    expect(api.list).toHaveBeenCalledTimes(2)

    api.delete.mockRejectedValueOnce({ reason: 'REDEEM_EMAIL_BATCH_NOT_DELETABLE', message: 'x' })
    await wrapper.get('[data-test="delete-batch-9"]').trigger('click')
    await wrapper.get('[data-test="confirm-delete"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-test="history-error"]').text()).toContain('admin.redeemEmail.errors.REDEEM_EMAIL_BATCH_NOT_DELETABLE')
    wrapper.unmount()
  })
})
