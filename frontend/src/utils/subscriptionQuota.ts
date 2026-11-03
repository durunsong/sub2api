import type { UserSubscription } from '@/types'

const ONE_DAY_MS = 24 * 60 * 60 * 1000

export type ExpirationDateRelation = 'expired' | 'today' | 'tomorrow' | 'later'

export type RemainingExpiryDuration =
  | { unit: 'days'; days: number }
  | { unit: 'hoursMinutes'; hours: number; minutes: number }

export interface RemainingDurationParts {
  days: number
  hours: number
  minutes: number
}

export function isOneTimeDailyQuota(
  subscription: Pick<UserSubscription, 'starts_at' | 'expires_at'>
): boolean {
  if (!subscription.starts_at || !subscription.expires_at) return false

  const startsAt = new Date(subscription.starts_at).getTime()
  const expiresAt = new Date(subscription.expires_at).getTime()

  if (!Number.isFinite(startsAt) || !Number.isFinite(expiresAt)) return false

  return expiresAt <= startsAt + ONE_DAY_MS
}

export function getRemainingDurationParts(
  targetAt: Date | string,
  now: Date = new Date()
): RemainingDurationParts | null {
  const targetTime = targetAt instanceof Date ? targetAt.getTime() : new Date(targetAt).getTime()
  const nowTime = now.getTime()

  if (!Number.isFinite(targetTime) || !Number.isFinite(nowTime)) return null

  const diffMs = targetTime - nowTime
  if (diffMs <= 0) return null

  const totalMinutes = Math.floor(diffMs / (1000 * 60))
  const days = Math.floor(totalMinutes / (24 * 60))
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60)
  const minutes = totalMinutes % 60

  return { days, hours, minutes }
}

export function getExpirationDateRelation(
  targetAt: Date | string,
  now: Date = new Date()
): ExpirationDateRelation | null {
  const target = targetAt instanceof Date ? targetAt : new Date(targetAt)
  const targetTime = target.getTime()
  const nowTime = now.getTime()

  if (!Number.isFinite(targetTime) || !Number.isFinite(nowTime)) return null
  if (targetTime <= nowTime) return 'expired'

  const targetDay = Date.UTC(target.getFullYear(), target.getMonth(), target.getDate())
  const currentDay = Date.UTC(now.getFullYear(), now.getMonth(), now.getDate())
  const calendarDays = Math.round((targetDay - currentDay) / ONE_DAY_MS)

  if (calendarDays === 0) return 'today'
  if (calendarDays === 1) return 'tomorrow'
  return 'later'
}

const WINDOW_PERIOD_MS = {
  weekly: 7 * ONE_DAY_MS,
  monthly: 30 * ONE_DAY_MS,
} as const

export type SubscriptionResetWindow = 'daily' | 'weekly' | 'monthly'

export interface SubscriptionResetInput {
  starts_at?: string | null
  expires_at?: string | null
}

export interface SubscriptionResetAt {
  at: Date
  /** 下一次滚动重置不早于到期，额度随订阅结束，不再重置。 */
  endsWithSubscription: boolean
}

function parseInstant(value: Date | string | null | undefined): Date | null {
  if (!value) return null
  const date = value instanceof Date ? value : new Date(value)
  return Number.isFinite(date.getTime()) ? date : null
}

function startOfLocalDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

function nextLocalMidnight(day: Date): Date {
  const next = startOfLocalDay(day)
  next.setDate(next.getDate() + 1)
  return next
}

// 早期订阅把首个周/月窗口记在开通日零点。真正滚动锚点是 StartsAt，
// 与后端 windowResetAnchor 相同；否则 90 天套餐会把重置时间提前若干小时。
function windowResetAnchor(startsAt: Date | null, windowStart: Date): Date {
  if (!startsAt) return windowStart
  const legacy = startOfLocalDay(startsAt)
  if (legacy.getTime() < startsAt.getTime() && windowStart.getTime() === legacy.getTime()) {
    return startsAt
  }
  return windowStart
}

export function nextSubscriptionReset(
  subscription: SubscriptionResetInput,
  windowStart: string | null | undefined,
  window: SubscriptionResetWindow,
  now: Date = new Date()
): SubscriptionResetAt | null {
  const start = parseInstant(windowStart)
  const nowTime = now.getTime()
  if (!start || !Number.isFinite(nowTime)) return null

  const expires = parseInstant(subscription.expires_at)
  const expiresTime = expires?.getTime()
  let reset = 0

  if (window === 'daily') {
    reset = nextLocalMidnight(start).getTime()
    if (reset <= nowTime) reset = nextLocalMidnight(now).getTime()
  } else {
    const anchor = windowResetAnchor(parseInstant(subscription.starts_at), start).getTime()
    const period = WINDOW_PERIOD_MS[window]
    reset = anchor + period
    const limit = expiresTime ?? Number.POSITIVE_INFINITY
    while (reset <= nowTime && reset < limit) reset += period
  }

  if (!Number.isFinite(reset) || reset <= nowTime) return null
  if (expiresTime != null && reset >= expiresTime) {
    if (expiresTime <= nowTime) return null
    return { at: new Date(expiresTime), endsWithSubscription: true }
  }
  return { at: new Date(reset), endsWithSubscription: false }
}

export function getRemainingExpiryDuration(
  targetAt: Date | string,
  now: Date = new Date()
): RemainingExpiryDuration | null {
  const targetTime = targetAt instanceof Date ? targetAt.getTime() : new Date(targetAt).getTime()
  const nowTime = now.getTime()

  if (!Number.isFinite(targetTime) || !Number.isFinite(nowTime)) return null

  const diffMs = targetTime - nowTime
  if (diffMs <= 0) return null
  if (diffMs >= ONE_DAY_MS) {
    return { unit: 'days', days: Math.ceil(diffMs / ONE_DAY_MS) }
  }

  const totalMinutes = Math.ceil(diffMs / (60 * 1000))
  return {
    unit: 'hoursMinutes',
    hours: Math.floor(totalMinutes / 60),
    minutes: totalMinutes % 60
  }
}
