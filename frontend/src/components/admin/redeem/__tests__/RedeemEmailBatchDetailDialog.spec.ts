import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Detail from '../RedeemEmailBatchDetailDialog.vue'

const api = vi.hoisted(() => ({ get: vi.fn(), items: vi.fn(), retry: vi.fn() }))
vi.mock('@/api/admin', () => ({
  adminAPI: { redeem: { getEmailBatch: api.get, listEmailBatchItems: api.items, retryFailedEmailItems: api.retry } }
}))
const known = new Set(['admin.redeemEmail.deliveryErrors.DELIVERY_OUTCOME_UNKNOWN', 'admin.redeemEmail.deliveryErrors.RECIPIENT_UNSUBSCRIBED'])
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key),
    te: (key: string) => known.has(key)
  })
}))
vi.mock('@/utils/format', () => ({ formatDateTime: (value: string) => `formatted:${value}` }))

const batch = (extra: Record<string, unknown> = {}) => ({
  id: 4,
  status: 'partial_failed',
  segments: ['inactive_1_2m'],
  reference_time: '2026-10-15T00:00:00Z',
  coupon_value: 5,
  code_expires_at: '2026-11-14T00:00:00Z',
  total_count: 3,
  pending_count: 0,
  sending_count: 0,
  sent_count: 1,
  failed_count: 1,
  skipped_count: 1,
  created_by: 1,
  started_at: null,
  finished_at: null,
  created_at: '2026-10-15T00:00:00Z',
  updated_at: '2026-10-15T00:00:00Z',
  ...extra
})

const item = (id: number, status: string, errorCode: string | null, errorMessage: string | null = null) => ({
  id,
  batch_id: 4,
  position: id,
  user_id: id,
  recipient_name: `user${id}`,
  recipient_email: `u${id}@example.com`,
  redeem_code_id: id,
  redeem_code: `CODE${id}`,
  status,
  attempt_count: 1,
  sent_at: null,
  last_attempt_at: null,
  error_code: errorCode,
  error_message: errorMessage,
  created_at: '2026-10-15T00:00:00Z',
  updated_at: '2026-10-15T00:00:00Z'
})

const mountDetail = () =>
  mount(Detail, {
    props: { show: true, batchId: 4 },
    global: {
      stubs: {
        BaseDialog: { props: ['show', 'title'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
        Pagination: true
      }
    }
  })

describe('RedeemEmailBatchDetailDialog', () => {
  beforeEach(() => {
    api.get.mockReset().mockResolvedValue(batch())
    api.items.mockReset().mockResolvedValue({
      items: [
        item(1, 'sent', null),
        item(2, 'failed', 'DELIVERY_OUTCOME_UNKNOWN', 'raw'),
        item(3, 'failed', 'NEW_CODE', 'fallback msg'),
        item(4, 'skipped', 'RECIPIENT_UNSUBSCRIBED', 'recipient unsubscribed')
      ],
      total: 4
    })
    api.retry.mockReset().mockResolvedValue(batch({ status: 'pending', failed_count: 0, pending_count: 1 }))
  })

  it('shows items with localized delivery errors and the unknown-outcome warning', async () => {
    const wrapper = mountDetail()
    await flushPromises()

    expect(api.items).toHaveBeenCalledWith(4, 1, 50, undefined, expect.anything())
    expect(wrapper.text()).toContain('CODE1')
    expect(wrapper.get('[data-test="item-error-2"]').text()).toBe('admin.redeemEmail.deliveryErrors.DELIVERY_OUTCOME_UNKNOWN')
    expect(wrapper.get('[data-test="item-error-3"]').text()).toBe('fallback msg')
    expect(wrapper.get('[data-test="item-error-4"]').text()).toBe('admin.redeemEmail.deliveryErrors.RECIPIENT_UNSUBSCRIBED')
    expect(wrapper.text()).toContain('admin.redeemEmail.itemStatus.skipped')
    expect(wrapper.text()).toContain('admin.redeemEmail.history.skippedCount:{"n":1}')
    expect(wrapper.text()).toContain('admin.redeemEmail.detail.unknownOutcomeHint')
  })

  it('filters by status and retries failed items', async () => {
    const wrapper = mountDetail()
    await flushPromises()

    await wrapper.get('[data-test="item-status-filter"]').setValue('skipped')
    await flushPromises()
    expect(api.items).toHaveBeenLastCalledWith(4, 1, 50, 'skipped', expect.anything())
    await wrapper.get('[data-test="item-status-filter"]').setValue('failed')
    await flushPromises()
    expect(api.items).toHaveBeenLastCalledWith(4, 1, 50, 'failed', expect.anything())

    await wrapper.get('[data-test="retry-failed"]').trigger('click')
    await flushPromises()
    expect(api.retry).toHaveBeenCalledWith(4)
    expect(wrapper.text()).toContain('admin.redeemEmail.detail.retried')
  })
})
