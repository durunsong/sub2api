import type { RedeemEmailBatchStatus, RedeemEmailItemStatus } from '@/types'

const COUPON_SCALE = 1e8

/** 与后端 FormatRedeemEmailCouponValue 一致：至少 2 位、至多 8 位小数 */
export function formatCouponUSD(value: number): string {
  let s = value.toFixed(8).replace(/0+$/, '')
  const decimals = s.length - s.indexOf('.') - 1
  if (decimals < 2) s += '0'.repeat(2 - decimals)
  return `$${s}`
}

/** 以 1e-8 为单位做整数乘法，避免 0.1 × 3 这类浮点误差 */
export function totalCouponValue(value: number, count: number): number {
  return (Math.round(value * COUPON_SCALE) * count) / COUPON_SCALE
}

export function isValidCouponValue(value: unknown): value is number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0 || value > 1_000_000) return false
  const scaled = value * COUPON_SCALE
  return Math.abs(scaled - Math.round(scaled)) < 1e-6
}

export const batchStatusClass: Record<RedeemEmailBatchStatus, string> = {
  pending: 'badge badge-gray',
  running: 'badge badge-primary',
  succeeded: 'badge badge-success',
  partial_failed: 'badge badge-warning',
  failed: 'badge badge-danger'
}

export const itemStatusClass: Record<RedeemEmailItemStatus, string> = {
  pending: 'badge badge-gray',
  sending: 'badge badge-primary',
  sent: 'badge badge-success',
  failed: 'badge badge-danger',
  skipped: 'badge badge-warning'
}

export const isActiveBatchStatus = (status: RedeemEmailBatchStatus) => status === 'pending' || status === 'running'

export function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 14)}`
}
