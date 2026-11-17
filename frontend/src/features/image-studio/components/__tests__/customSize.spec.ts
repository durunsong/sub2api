import { describe, expect, it } from 'vitest'
import { normalizeDimensions, isValidCustomSize } from '../customSize'

describe('custom image dimensions', () => {
  it('rounds upward to the declared step only when alignment is enabled', () => {
    expect(normalizeDimensions(1025, 1023, true, 16)).toBe('1040x1024')
    expect(normalizeDimensions(1025, 1023, false, 16)).toBe('1025x1023')
  })
  it('enforces total pixels, maximum edge, aspect ratio and alignment together', () => {
    expect(isValidCustomSize('1024x1024', 16)).toBe(true)
    expect(isValidCustomSize('3840x2160', 16)).toBe(true)
    for (const size of ['1025x1024', '3840x3840', '4096x1024', '512x512', '3072x800', '0x1024', 'NaNx1024']) {
      expect(isValidCustomSize(size, 16), size).toBe(false)
    }
  })
})
