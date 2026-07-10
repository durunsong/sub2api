import { describe, expect, it } from 'vitest'
import zh from '../locales/zh'
import en from '../locales/en'

function get(obj: unknown, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (!acc || typeof acc !== 'object') return undefined
    return (acc as Record<string, unknown>)[key]
  }, obj)
}

describe('fork overlay i18n keys', () => {
  const keys = [
    'payment.methods.xorpay',
    'payment.meteredTitle',
    'payment.meteredDesc',
    'payment.rechargeNow',
    'payment.mySubscriptions',
    'home.providers.kiro',
    'admin.accounts.oauth.kiro.authModeTitle',
    'admin.accounts.oauth.kiro.importAndUpdate',
    'admin.settings.payment.providerXorpay',
    'admin.ops.runtime.metricThresholds'
  ]

  for (const key of keys) {
    it(`zh has ${key}`, () => {
      const v = get(zh, key)
      expect(typeof v).toBe('string')
      expect(v).not.toBe(key)
    })
    it(`en has ${key}`, () => {
      const v = get(en, key)
      expect(typeof v).toBe('string')
      expect(v).not.toBe(key)
    })
  }
})
