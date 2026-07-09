import { describe, expect, it } from 'vitest'
import enFork from '../en/fork'
import zhFork from '../zh/fork'

describe('fork ipBans i18n messages', () => {
  const cases = [
    { name: 'zh', messages: zhFork.admin.ipBans },
    { name: 'en', messages: enFork.admin.ipBans }
  ] as const

  for (const locale of cases) {
    it(`${locale.name}: email helper texts escape @ for vue-i18n linked messages`, () => {
      // Runtime vue-i18n treats bare "@foo" as linked-message syntax and can throw
      // when IpBansView switches to email_suffix / email_regex and renders these keys.
      expect(locale.messages.emailSuffixPlaceholder).toContain("{'@'}")
      expect(locale.messages.emailSuffixHelp).toContain("{'@'}")
      expect(locale.messages.emailRegexPlaceholder).toContain("{'@'}")
      expect(locale.messages.emailSuffixPlaceholder).not.toMatch(/(^|[^'{])@\w/)
      expect(locale.messages.emailSuffixHelp).not.toMatch(/(^|[^'{])@\w/)
    })
  }
})
