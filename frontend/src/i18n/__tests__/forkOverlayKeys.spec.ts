import { describe, expect, it } from 'vitest'
import zh from '../locales/zh'
import en from '../locales/en'

function get(obj: unknown, path: string): unknown {
  return path.split('.').reduce<unknown>((acc, key) => {
    if (!acc || typeof acc !== 'object') return undefined
    return (acc as Record<string, unknown>)[key]
  }, obj)
}

function collectStringPaths(value: unknown, prefix = ''): string[] {
  if (!value || typeof value !== 'object') return []

  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) => {
    const path = prefix ? `${prefix}.${key}` : key
    return typeof child === 'string' ? [path] : collectStringPaths(child, path)
  })
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

  for (const namespace of ['home.showcase', 'auth.showcase']) {
    it(`${namespace} has matching non-empty zh and en messages`, () => {
      const zhMessages = get(zh, namespace)
      const enMessages = get(en, namespace)
      const zhKeys = collectStringPaths(zhMessages).sort()
      const enKeys = collectStringPaths(enMessages).sort()

      expect(zhKeys).toEqual(enKeys)
      expect(zhKeys.length).toBeGreaterThan(0)

      for (const key of zhKeys) {
        expect(get(zhMessages, key)).toBeTruthy()
        expect(get(enMessages, key)).toBeTruthy()
      }
    })
  }
})
