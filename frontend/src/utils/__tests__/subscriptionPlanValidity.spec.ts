import { describe, expect, it } from 'vitest'
import { formatPlanValiditySuffix } from '../subscriptionPlanValidity'

const t = (key: string) => {
  const labels: Record<string, string> = {
    'payment.admin.days': '天',
    'payment.admin.weeks': '周',
    'payment.admin.months': '月',
    'payment.years': '年',
  }
  return labels[key] ?? key
}

describe('formatPlanValiditySuffix', () => {
  it('formats day units', () => {
    expect(formatPlanValiditySuffix(1, 'days', t)).toBe('1天')
    expect(formatPlanValiditySuffix(30, 'day', t)).toBe('30天')
  })

  it('formats week units', () => {
    expect(formatPlanValiditySuffix(1, 'weeks', t)).toBe('1周')
    expect(formatPlanValiditySuffix(2, 'week', t)).toBe('2周')
  })

  it('formats month units', () => {
    expect(formatPlanValiditySuffix(1, 'months', t)).toBe('1月')
    expect(formatPlanValiditySuffix(3, 'month', t)).toBe('3月')
  })

  it('formats year units', () => {
    expect(formatPlanValiditySuffix(1, 'years', t)).toBe('1年')
  })

  it('falls back to days for unknown units', () => {
    expect(formatPlanValiditySuffix(7, 'unknown', t)).toBe('7天')
    expect(formatPlanValiditySuffix(7, null, t)).toBe('7天')
  })
})
