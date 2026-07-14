import { describe, expect, it } from 'vitest'
import {
  platformBadgeClass,
  platformGradientClass,
  platformLabel,
  platformSearchText,
  platformTextClass,
  userFacingPlatform,
  userFacingPlatformText
} from '../platformColors'

describe('platformColors', () => {
  it('Kiro 平台使用独立紫色主题，不复用 Anthropic 橙色', () => {
    expect(platformBadgeClass('kiro')).toContain('violet')
    expect(platformTextClass('kiro')).toContain('violet')
    expect(platformGradientClass('kiro')).toContain('from-violet-500')
    expect(platformGradientClass('kiro')).toContain('to-fuchsia-500')
    expect(platformBadgeClass('kiro')).not.toContain('orange')
  })

  it('用户侧将 Kiro 显示和搜索为 Claude', () => {
    expect(platformLabel('kiro')).toBe('Claude')
    expect(userFacingPlatform('kiro')).toBe('anthropic')
    expect(userFacingPlatformText('Kiro Pro / kiro')).toBe('Claude Pro / Claude')
    expect(platformSearchText('kiro')).toContain('claude')
  })
})
