import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/stores/app'
import type { PublicSettings } from '@/types'
import LoginView from '@/views/auth/LoginView.vue'

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

function mountLogin() {
  const wrapper = mount(LoginView, {
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

describe('LoginView registration entry', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    delete window.__APP_CONFIG__
    getPublicSettingsMock.mockReset()
    pushMock.mockReset()
    getPublicSettingsMock.mockResolvedValue(publicSettings)
  })

  afterEach(() => {
    mounted.splice(0).forEach(wrapper => wrapper.unmount())
    delete window.__APP_CONFIG__
    vi.restoreAllMocks()
  })

  it('shows the registration entry when registration is enabled', async () => {
    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.text()).toContain('auth.signUp')
  })

  it('hides the registration entry when registration is disabled', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...publicSettings,
      registration_enabled: false
    })

    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.text()).not.toContain('auth.signUp')
  })

  it('keeps registration and password reset available from injected settings when the API is unavailable', async () => {
    window.__APP_CONFIG__ = { ...publicSettings, password_reset_enabled: true } as PublicSettings
    useAppStore().initFromInjectedConfig()
    getPublicSettingsMock.mockRejectedValue(new Error('Network unavailable'))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.get('.auth-footer a').attributes('href')).toBe('/register')
    expect(wrapper.find('a[href="/forgot-password"]').exists()).toBe(true)
    expect(getPublicSettingsMock).not.toHaveBeenCalled()
  })

  it('uses cached settings when navigating to login without another API request', async () => {
    await useAppStore().fetchPublicSettings()
    delete window.__APP_CONFIG__
    getPublicSettingsMock.mockClear().mockRejectedValue(new Error('Network unavailable'))
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.get('.auth-footer a').attributes('href')).toBe('/register')
    expect(getPublicSettingsMock).not.toHaveBeenCalled()
  })

  it('shares the pending settings request with the layout and waits before showing registration', async () => {
    let resolveSettings!: (settings: typeof publicSettings) => void
    getPublicSettingsMock.mockReturnValue(new Promise(resolve => { resolveSettings = resolve }))
    const wrapper = mountLogin()

    expect(wrapper.find('a[href="/register"]').exists()).toBe(false)
    resolveSettings(publicSettings)
    await flushPromises()

    expect(wrapper.get('.auth-footer a').attributes('href')).toBe('/register')
    expect(getPublicSettingsMock).toHaveBeenCalledTimes(1)
  })

  it.each([
    { registration_enabled: false, backend_mode_enabled: false },
    { registration_enabled: true, backend_mode_enabled: true }
  ])('honors injected registration restrictions: %j', async (settings) => {
    window.__APP_CONFIG__ = { ...publicSettings, ...settings } as PublicSettings
    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.find('a[href="/register"]').exists()).toBe(false)
    expect(getPublicSettingsMock).not.toHaveBeenCalled()
  })

  it('does not enable registration when no settings are available', async () => {
    getPublicSettingsMock.mockRejectedValue(new Error('Network unavailable'))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const wrapper = mountLogin()
    await flushPromises()

    expect(wrapper.find('a[href="/register"]').exists()).toBe(false)
    expect(getPublicSettingsMock).toHaveBeenCalledTimes(1)
  })
})
