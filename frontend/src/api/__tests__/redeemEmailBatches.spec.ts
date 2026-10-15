import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createEmailBatch,
  listEmailBatchItems,
  previewEmail,
  retryFailedEmailItems,
  searchEmailRecipients
} from '../admin/redeem'
import { formatCouponUSD, isValidCouponValue, totalCouponValue } from '@/components/admin/redeem/redeemEmailFormat'

const { get, post } = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }))
vi.mock('../client', () => ({ apiClient: { get, post } }))

describe('redeem email batch API', () => {
  beforeEach(() => vi.clearAllMocks())

  it('sends the Idempotency-Key header when creating a batch', async () => {
    post.mockResolvedValue({ data: { id: 1 } })
    const payload = {
      segments: ['inactive_1_2m' as const, 'never_used' as const],
      user_ids: [1],
      coupon_value: 5,
      code_expires_in_days: 30
    }
    await createEmailBatch(payload, 'key-123456')
    expect(post).toHaveBeenCalledWith('/admin/redeem-codes/email-batches', payload, {
      headers: { 'Idempotency-Key': 'key-123456' }
    })
  })

  it('posts the recipient filter and omits an empty item status', async () => {
    post.mockResolvedValue({ data: { items: [], total: 0, max_recipients: 1000 } })
    get.mockResolvedValue({ data: { items: [], total: 0 } })
    await searchEmailRecipients({ segments: ['inactive_3m_plus'] })
    expect(post).toHaveBeenCalledWith(
      '/admin/redeem-codes/email-batches/recipients/search',
      { segments: ['inactive_3m_plus'] },
      undefined
    )
    await previewEmail({ coupon_value: 5, never_expires: true, locale: 'en' })
    expect(post).toHaveBeenLastCalledWith(
      '/admin/redeem-codes/email-batches/preview',
      { coupon_value: 5, never_expires: true, locale: 'en' },
      undefined
    )
    await listEmailBatchItems(9, 2, 50)
    expect(get).toHaveBeenCalledWith('/admin/redeem-codes/email-batches/9/items', {
      params: { page: 2, page_size: 50, status: undefined }
    })
    await retryFailedEmailItems(9)
    expect(post).toHaveBeenLastCalledWith('/admin/redeem-codes/email-batches/9/retry-failed', undefined, undefined)
  })
})

describe('redeem email formatting', () => {
  it('matches backend coupon formatting and avoids float drift', () => {
    expect(formatCouponUSD(5)).toBe('$5.00')
    expect(formatCouponUSD(0.125)).toBe('$0.125')
    expect(formatCouponUSD(0.12345678)).toBe('$0.12345678')
    expect(totalCouponValue(0.1, 3)).toBe(0.3)
    expect(isValidCouponValue(0.12345678)).toBe(true)
    expect(isValidCouponValue(0.123456789)).toBe(false)
    expect(isValidCouponValue(0)).toBe(false)
    expect(isValidCouponValue(1_000_001)).toBe(false)
  })
})
