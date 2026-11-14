import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/stores/app'
import type { PublicSettings } from '@/types'
import LoginView from '@/views/auth/LoginView.vue'
import RegisterView from '@/views/auth/RegisterView.vue'

const { getPublicSettingsMock, pushMock } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  pushMock: vi.fn()
}))

const publicSettings = {
  registration_enabled: true,
  turnstile_enabled: false,
  turnstile_site_key: '',
  tencent_captcha_enabled: false,
  tencent_captcha_app_id: '',
  aliyun_captcha_enabled: false,
  aliyun_captcha_scene_id: '',
  aliyun_captcha_prefix: '',
  linuxdo_oauth_enabled: false,
  dingtalk_oauth_enabled: false,
  wechat_oauth_enabled: false,
  backend_mode_enabled: false,
  oidc_oauth_enabled: false,
  oidc_oauth_provider_name: 'OIDC',
  github_oauth_enabled: false,
  google_oauth_enabled: false,
  password_reset_enabled: false,
  passkey_enabled: false,
  login_agreement_enabled: false,
  login_agreement_documents: []
}

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({
    push: pushMock,
    currentRoute: { value: { query: {} } }
  })
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      t: (key: string) => key
    }
  }),
  useI18n: () => ({
    locale: { value: 'zh' },
    t: (key: string) => key
  })
}))

vi.mock('@/stores', async () => ({
  useAuthStore: () => ({
    login: vi.fn(),
    loginWithPasskey: vi.fn(),
    login2FA: vi.fn()
  }),
  useAppStore: (await vi.importActual<typeof import('@/stores/app')>('@/stores/app')).useAppStore
}))

vi.mock('@/api/auth', () => ({
  buildOAuthLoginStartURL: vi.fn(),
  getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
  isTotp2FARequired: vi.fn(() => false),
  isWeChatWebOAuthEnabled: vi.fn(() => false),
  startOAuthLogin: vi.fn()
}))

const mounted: ReturnType<typeof mount>[] = []

function mountPage(page: typeof LoginView | typeof RegisterView) {
  const wrapper = mount(page, {
    attachTo: document.body,
    global: {
      stubs: {
        LocaleSwitcher: true,
        DingTalkOAuthSection: true,
        EmailOAuthButtons: true,
        Icon: true,
        LinuxDoOAuthSection: true,
        LoginAgreementPrompt: true,
        OidcOAuthSection: true,
        RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
        TotpLoginModal: true,
        TurnstileWidget: true,
        WechatOAuthSection: true,
        transition: false
      }
    }
  })
  mounted.push(wrapper)
  return wrapper
}

describe.each([
  ['login', LoginView, 'auth.loginSupportHint'],
  ['register', RegisterView, 'auth.registerSupportHint']
] as const)('%s support contact', (_name, page, hint) => {
  beforeEach(() => {
    setActivePinia(createPinia())
    delete window.__APP_CONFIG__
    getPublicSettingsMock.mockReset().mockResolvedValue({
      ...publicSettings,
      contact_info: '微信: support-test',
      contact_qr_code: '/support-qr.png'
    })
  })

  afterEach(() => {
    mounted.splice(0).forEach(wrapper => wrapper.unmount())
    delete window.__APP_CONFIG__
    vi.restoreAllMocks()
  })

  it('opens the configured QR and contact from the page-specific help row, then toggles closed', async () => {
    const wrapper = mountPage(page)
    await flushPromises()

    expect(wrapper.text()).toContain(hint)
    const button = wrapper.get('button[aria-label="nav.supportWechat"]')
    expect(button.attributes('type')).toBe('button')
    expect(button.attributes('aria-expanded')).toBe('false')
    await button.trigger('click')
    expect(button.attributes('aria-expanded')).toBe('true')
    const panel = wrapper.get('[role="region"][aria-label="nav.supportWechat"]')
    expect(panel.get('img').attributes('src')).toBe('/support-qr.png')
    expect(panel.text()).toContain('微信: support-test')
    await button.trigger('click')
    expect(wrapper.find('[role="region"]').exists()).toBe(false)
  })

  it('closes on Escape and outside click', async () => {
    const wrapper = mountPage(page)
    await flushPromises()
    const button = wrapper.get('button[aria-label="nav.supportWechat"]')
    await button.trigger('click')
    await button.trigger('keydown', { key: 'Escape' })
    expect(button.attributes('aria-expanded')).toBe('false')
    await button.trigger('click')
    document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(button.attributes('aria-expanded')).toBe('false')
  })

  it('keeps a clicked mouse preview open after pointer leave, and closes when focus moves outside', async () => {
    const wrapper = mountPage(page)
    await flushPromises()
    const button = wrapper.get('button[aria-label="nav.supportWechat"]')
    const trigger = button.element.parentElement!
    trigger.dispatchEvent(Object.assign(new Event('pointerenter'), { pointerType: 'mouse' }))
    await flushPromises()
    expect(button.attributes('aria-expanded')).toBe('true')
    await button.trigger('click')
    trigger.dispatchEvent(new Event('pointerleave'))
    await flushPromises()
    expect(button.attributes('aria-expanded')).toBe('true')
    await button.trigger('focusout', { relatedTarget: wrapper.get('#email').element })
    expect(button.attributes('aria-expanded')).toBe('false')
  })

  it('keeps the panel available when clicking non-focusable contact text', async () => {
    const wrapper = mountPage(page)
    await flushPromises()
    const button = wrapper.get('button[aria-label="nav.supportWechat"]')
    await button.trigger('click')
    await button.trigger('focusout', { relatedTarget: null })
    expect(button.attributes('aria-expanded')).toBe('true')
    await wrapper.get('[role="region"] p').trigger('click')
    expect(wrapper.get('[role="region"]').text()).toContain('微信: support-test')
    document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()
    expect(button.attributes('aria-expanded')).toBe('false')
  })

  it('uses injected support settings even when the public API fails', async () => {
    window.__APP_CONFIG__ = { ...publicSettings, contact_info: 'cached-support', contact_qr_code: '/cached-qr.png' } as PublicSettings
    useAppStore().initFromInjectedConfig()
    getPublicSettingsMock.mockRejectedValue(new Error('Network unavailable'))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const wrapper = mountPage(page)
    await flushPromises()
    await wrapper.get('button[aria-label="nav.supportWechat"]').trigger('click')
    expect(wrapper.get('[role="region"]').text()).toContain('cached-support')
    expect(wrapper.get('[role="region"] img').attributes('src')).toBe('/cached-qr.png')
  })

  it('hides the help row when no safe support contact is configured', async () => {
    getPublicSettingsMock.mockResolvedValue({ ...publicSettings, contact_info: '  ', contact_qr_code: 'javascript:alert(1)' })
    const wrapper = mountPage(page)
    await flushPromises()
    expect(wrapper.find('button[aria-label="nav.supportWechat"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain(hint)
  })

  it.each([
    ['support-only', '', false],
    ['', '/qr-only.png', true]
  ])('supports partial configuration (%s, %s)', async (contact, qr, hasImage) => {
    getPublicSettingsMock.mockResolvedValue({ ...publicSettings, contact_info: contact, contact_qr_code: qr })
    const wrapper = mountPage(page)
    await flushPromises()
    await wrapper.get('button[aria-label="nav.supportWechat"]').trigger('click')
    const panel = wrapper.get('[role="region"]')
    expect(panel.find('img').exists()).toBe(hasImage)
    if (contact) expect(panel.text()).toContain(contact)
  })
})
