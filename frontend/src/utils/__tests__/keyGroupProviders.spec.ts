import { describe, expect, it } from 'vitest'
import { getKeyGroupProvider } from '../keyGroupProviders'

describe('Fork key group providers', () => {
  it('includes Kiro groups in the user-facing Claude provider', () => {
    expect(getKeyGroupProvider('kiro')).toBe('anthropic')
  })
})
