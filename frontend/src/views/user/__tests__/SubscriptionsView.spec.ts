import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'
import type { UserSubscription } from '@/types'

const getMySubscriptions = vi.hoisted(() => vi.fn())
const resetDailyQuota = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

vi.mock('@/api/subscriptions', () => ({ default: { getMySubscriptions, resetDailyQuota } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string, params?: { count?: number }) => params?.count == null ? key : `${key}:${params.count}` }) }
})

const now = Date.now()
function subscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 1, user_id: 7, group_id: 3, status: 'active',
    starts_at: new Date(now - 60 * 60 * 1000).toISOString(),
    expires_at: new Date(now + 30 * 24 * 60 * 60 * 1000).toISOString(),
    daily_usage_usd: 5, weekly_usage_usd: 5, monthly_usage_usd: 5,
    daily_usage_tokens: 50, weekly_usage_tokens: 50, monthly_usage_tokens: 50,
    manual_reset_credits: 1, daily_window_start: null, weekly_window_start: null, monthly_window_start: null,
    created_at: new Date(now).toISOString(), updated_at: new Date(now).toISOString(),
    group: { id: 3, name: 'Plan', description: '', platform: 'openai', daily_limit_usd: 0 } as UserSubscription['group'],
    ...overrides,
  }
}

async function mountView(items: UserSubscription[]) {
  getMySubscriptions.mockResolvedValueOnce(items)
  const wrapper = mount(SubscriptionsView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } } })
  await flushPromises()
  return wrapper
}
function resetButton(wrapper: Awaited<ReturnType<typeof mountView>>) {
  return wrapper.findAll('button').find(button => button.text().includes('userSubscriptions.manualReset.button'))
}

describe('SubscriptionsView manual daily reset', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getMySubscriptions.mockReset()
    resetDailyQuota.mockReset()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('shows credits when daily limit is zero, including zero credits as disabled', async () => {
    const wrapper = await mountView([subscription({ manual_reset_credits: 0 })])
    expect(wrapper.text()).toContain('userSubscriptions.manualReset.remaining:0')
    expect(resetButton(wrapper)?.attributes('disabled')).toBeDefined()
  })

  it('shows null credits as zero and disables reset', async () => {
    const wrapper = await mountView([
      subscription({ manual_reset_credits: null as unknown as number }),
    ])
    expect(wrapper.text()).toContain('userSubscriptions.manualReset.remaining:0')
    expect(resetButton(wrapper)?.attributes('disabled')).toBeDefined()
  })

  it('hides reset controls only when manual_reset_credits is absent', async () => {
    const item = subscription(); delete item.manual_reset_credits
    const wrapper = await mountView([item])
    expect(wrapper.text()).not.toContain('userSubscriptions.manualReset.remaining')
  })

  it.each([
    ['active monthly', subscription(), false],
    ['expired one-time day card', subscription({ status: 'expired', starts_at: new Date(now - 48 * 60 * 60 * 1000).toISOString(), expires_at: new Date(now - 24 * 60 * 60 * 1000).toISOString() }), false],
    ['expired monthly card', subscription({ status: 'expired', starts_at: new Date(now - 31 * 24 * 60 * 60 * 1000).toISOString(), expires_at: new Date(now - 60 * 60 * 1000).toISOString() }), true],
    ['revoked card', subscription({ status: 'revoked' }), true],
    ['suspended card', subscription({ status: 'suspended' }), true],
  ])('matches backend reset eligibility for %s', async (_name, item, disabled) => {
    const wrapper = await mountView([item])
    expect(resetButton(wrapper)?.attributes('disabled') !== undefined).toBe(disabled)
  })

  it('replaces with complete server response and prevents double click', async () => {
    const updated = subscription({
      manual_reset_credits: 0,
      daily_usage_usd: 0,
      daily_window_start: '2026-08-14T00:00:00Z',
      monthly_usage_usd: 37.25,
      monthly_usage_tokens: 98765,
      group: undefined,
    })
    let resolveReset!: (value: UserSubscription) => void
    resetDailyQuota.mockReturnValue(new Promise<UserSubscription>(resolve => { resolveReset = resolve }))
    const wrapper = await mountView([subscription()]); const button = resetButton(wrapper)!
    await button.trigger('click'); await button.trigger('click')
    expect(resetDailyQuota).toHaveBeenCalledTimes(1); expect(button.attributes('disabled')).toBeDefined()
    resolveReset(updated); await flushPromises()
    expect(wrapper.text()).toContain('userSubscriptions.manualReset.remaining:0')
    expect(wrapper.text()).not.toContain('Plan')
    const rendered = (wrapper.vm as unknown as { subscriptions: UserSubscription[] }).subscriptions[0]
    expect(rendered.daily_window_start).toBe('2026-08-14T00:00:00Z')
    expect(rendered.monthly_usage_usd).toBe(37.25)
    expect(rendered.monthly_usage_tokens).toBe(98765)
    expect(rendered.monthly_usage_usd).not.toBe(0)
    expect(rendered.monthly_usage_tokens).not.toBe(0)
  })

  it('keeps API error message and silently refreshes after failure', async () => {
    resetDailyQuota.mockRejectedValueOnce({ message: 'server reset rejected' })
    const wrapper = await mountView([subscription()])
    getMySubscriptions.mockResolvedValueOnce([subscription({ manual_reset_credits: 0 })])
    await resetButton(wrapper)!.trigger('click'); await flushPromises()
    expect(getMySubscriptions).toHaveBeenCalledTimes(2); expect(showError).toHaveBeenCalledWith('server reset rejected')
    expect(wrapper.text()).toContain('userSubscriptions.manualReset.remaining:0')
  })

  it('does not overwrite original API error when silent refresh fails', async () => {
    resetDailyQuota.mockRejectedValueOnce({ message: 'original reset error' })
    const wrapper = await mountView([subscription()])
    getMySubscriptions.mockRejectedValueOnce(new Error('refresh failed'))
    await resetButton(wrapper)!.trigger('click'); await flushPromises()
    expect(showError).toHaveBeenCalledTimes(1); expect(showError).toHaveBeenCalledWith('original reset error')
  })
})
