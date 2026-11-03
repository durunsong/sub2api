import { describe, expect, it } from 'vitest'

import { getExpirationDateRelation, getRemainingExpiryDuration, nextSubscriptionReset } from '../subscriptionQuota'

describe('subscription expiry timing', () => {
  it('uses local calendar dates for today and tomorrow', () => {
    const now = new Date(2026, 2, 7, 23, 30)

    expect(getExpirationDateRelation(new Date(2026, 2, 7, 23, 45), now)).toBe('today')
    expect(getExpirationDateRelation(new Date(2026, 2, 8, 3, 30), now)).toBe('tomorrow')
  })

  it('treats the exact expiry instant and elapsed expiries as expired', () => {
    const now = new Date(2026, 6, 30, 9, 0)

    expect(getExpirationDateRelation(now, now)).toBe('expired')
    expect(getRemainingExpiryDuration(now, now)).toBeNull()
    expect(getExpirationDateRelation(new Date(2026, 6, 30, 8, 59), now)).toBe('expired')
    expect(getRemainingExpiryDuration(new Date(2026, 6, 30, 8, 59), now)).toBeNull()
  })

  it('rejects invalid target and current dates', () => {
    const invalid = new Date('invalid')
    const valid = new Date(2026, 6, 30, 9, 0)

    expect(getExpirationDateRelation(invalid, valid)).toBeNull()
    expect(getExpirationDateRelation(valid, invalid)).toBeNull()
    expect(getRemainingExpiryDuration(invalid, valid)).toBeNull()
    expect(getRemainingExpiryDuration(valid, invalid)).toBeNull()
  })

  it('returns rounded-up hours and minutes for an expiry under 24 hours away', () => {
    const now = new Date(2026, 6, 30, 9, 0)

    expect(getRemainingExpiryDuration(new Date(2026, 6, 31, 8, 30), now)).toEqual({
      unit: 'hoursMinutes',
      hours: 23,
      minutes: 30
    })
    expect(getRemainingExpiryDuration(new Date(now.getTime() + 1), now)).toEqual({
      unit: 'hoursMinutes',
      hours: 0,
      minutes: 1
    })
    expect(getRemainingExpiryDuration(new Date(now.getTime() + 23 * 60 * 60 * 1000 + 1), now)).toEqual({
      unit: 'hoursMinutes',
      hours: 23,
      minutes: 1
    })
  })

  it('preserves rounded-up day display from 24 hours onward', () => {
    const now = new Date(2026, 6, 30, 9, 0)

    expect(getRemainingExpiryDuration(new Date(now.getTime() + 24 * 60 * 60 * 1000), now)).toEqual({
      unit: 'days',
      days: 1
    })
    expect(getRemainingExpiryDuration(new Date(now.getTime() + 24 * 60 * 60 * 1000 + 1), now)).toEqual({
      unit: 'days',
      days: 2
    })
  })
})

describe('subscription reset countdown', () => {
  const day = 24 * 60 * 60 * 1000

  it('uses the purchase instant instead of midnight for a 90-day monthly window', () => {
    const startsAt = new Date(2026, 6, 1, 15, 5, 0)
    const windowStart = new Date(2026, 6, 1, 0, 0, 0)
    const now = new Date(startsAt.getTime() + 2 * 60 * 60 * 1000)

    const reset = nextSubscriptionReset(
      {
        starts_at: startsAt.toISOString(),
        expires_at: new Date(startsAt.getTime() + 90 * day).toISOString(),
      },
      windowStart.toISOString(),
      'monthly',
      now
    )

    expect(reset).toEqual({
      at: new Date(startsAt.getTime() + 30 * day),
      endsWithSubscription: false,
    })
  })

  it('rolls a 90-day plan to the current 30-day window', () => {
    const startsAt = new Date(2026, 6, 1, 15, 5, 0)
    const now = new Date(startsAt.getTime() + 40 * day)

    const reset = nextSubscriptionReset(
      {
        starts_at: startsAt.toISOString(),
        expires_at: new Date(startsAt.getTime() + 90 * day).toISOString(),
      },
      startsAt.toISOString(),
      'monthly',
      now
    )

    expect(reset?.endsWithSubscription).toBe(false)
    expect(reset?.at.getTime()).toBe(startsAt.getTime() + 60 * day)
  })

  it('ends the last slice of a 90-day plan with the subscription', () => {
    const startsAt = new Date(2026, 6, 1, 15, 5, 0)
    const expiresAt = new Date(startsAt.getTime() + 90 * day)
    const now = new Date(startsAt.getTime() + 70 * day)

    const reset = nextSubscriptionReset(
      { starts_at: startsAt.toISOString(), expires_at: expiresAt.toISOString() },
      new Date(startsAt.getTime() + 60 * day).toISOString(),
      'monthly',
      now
    )

    expect(reset).toEqual({ at: expiresAt, endsWithSubscription: true })
  })

  it('resets a daily window at the next local midnight', () => {
    const now = new Date(2026, 7, 6, 18, 0, 0)
    const windowStart = new Date(2026, 7, 6, 16, 49, 0)

    const reset = nextSubscriptionReset(
      { starts_at: windowStart.toISOString(), expires_at: new Date(2026, 8, 6, 18, 0, 0).toISOString() },
      windowStart.toISOString(),
      'daily',
      now
    )

    expect(reset?.endsWithSubscription).toBe(false)
    expect(reset?.at.getTime()).toBe(new Date(2026, 7, 7, 0, 0, 0).getTime())
  })
})
