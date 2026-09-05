import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'
import type { UserSubscription } from '@/types'

const getMySubscriptions = vi.hoisted(() => vi.fn())
const consumeResetCard = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

vi.mock('@/api/subscriptions', () => ({ default: { getMySubscriptions, consumeResetCard } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string, params?: Record<string, number>) => {
    if (!params) return key
    return key + ':' + Object.values(params).join(',')
  } }) }
})

const now = Date.now()
function subscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 1, user_id: 7, group_id: 3, status: 'active',
    starts_at: new Date(now - 60 * 60 * 1000).toISOString(),
    expires_at: new Date(now + 30 * 24 * 60 * 60 * 1000).toISOString(),
    daily_usage_usd: 5, weekly_usage_usd: 5, monthly_usage_usd: 5,
    daily_usage_tokens: 50, weekly_usage_tokens: 50, monthly_usage_tokens: 50,
    manual_reset_credits: 9, reset_cards: { total: 3, groups: [{ validity_days: 1, count: 2 }, { validity_days: 30, count: 1 }] },
    daily_window_start: null, weekly_window_start: null, monthly_window_start: null,
    created_at: new Date(now).toISOString(), updated_at: new Date(now).toISOString(),
    group: { id: 3, name: 'Plan', description: '', platform: 'openai', daily_limit_usd: 10, weekly_limit_usd: 20, monthly_limit_usd: 30 } as UserSubscription['group'],
    ...overrides,
  }
}

async function mountView(items: UserSubscription[]) {
  getMySubscriptions.mockResolvedValueOnce(items)
  const wrapper = mount(SubscriptionsView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } } })
  await flushPromises()
  return wrapper
}

function cardButtons(wrapper: Awaited<ReturnType<typeof mountView>>) {
  return wrapper.findAll('button').filter(button => button.text().includes('userSubscriptions.resetCards.action'))
}

describe('SubscriptionsView reset cards', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getMySubscriptions.mockReset()
    consumeResetCard.mockReset()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('shows the benefit bar after the header and hides it when total is zero', async () => {
    const wrapper = await mountView([subscription()])
    expect(wrapper.text()).toContain('userSubscriptions.resetCards.title:3')
    expect(wrapper.text()).toContain('userSubscriptions.resetCards.permanent')
    expect(wrapper.get('[data-test=reset-card-notice]').text()).toContain('userSubscriptions.resetCards.notice:3')
    expect(cardButtons(wrapper).map(button => button.text())).toEqual([
      'userSubscriptions.resetCards.action:1,2',
      'userSubscriptions.resetCards.action:30,1',
    ])
    expect(wrapper.find('[data-test=reset-cards]').element.previousElementSibling?.textContent).toContain('Plan')
    expect(wrapper.find('[data-test=reset-cards]').element.nextElementSibling?.textContent).toContain('userSubscriptions.expires')

    const hidden = await mountView([subscription({ reset_cards: { total: 0, groups: [] } })])
    expect(hidden.find('[data-test=reset-cards]').exists()).toBe(false)
    expect(hidden.find('[data-test=reset-card-notice]').exists()).toBe(false)
    expect(hidden.text()).not.toContain('userSubscriptions.manualReset')
  })

  it.each([
    ['active', false], ['expired', false], ['suspended', true], ['revoked', true],
  ] as const)('sets all card buttons disabled for %s subscriptions', async (status, disabled) => {
    const wrapper = await mountView([subscription({ status })])
    expect(cardButtons(wrapper).every(button => (button.attributes('disabled') !== undefined) === disabled)).toBe(true)
  })

  it('confirms the full reset impact, disables the subscription, and replaces the complete response', async () => {
    const updated = subscription({
      starts_at: '2026-08-16T00:00:00Z', expires_at: '2026-09-15T00:00:00Z',
      daily_usage_usd: 0, weekly_usage_usd: 0, monthly_usage_usd: 0,
      daily_usage_tokens: 0, weekly_usage_tokens: 0, monthly_usage_tokens: 0,
      reset_cards: { total: 2, groups: [{ validity_days: 1, count: 2 }] }, group: undefined,
    })
    let resolveConsume!: (value: UserSubscription) => void
    consumeResetCard.mockReturnValue(new Promise<UserSubscription>(resolve => { resolveConsume = resolve }))
    const wrapper = await mountView([subscription()])
    const buttons = cardButtons(wrapper)
    await buttons[1].trigger('click')
    await buttons[0].trigger('click')

    expect(window.confirm).toHaveBeenCalledWith('userSubscriptions.resetCards.confirm:30')
    expect(consumeResetCard).toHaveBeenCalledTimes(1)
    expect(consumeResetCard).toHaveBeenCalledWith(1, 30, expect.stringMatching(/^[A-Za-z0-9._:-]{8,128}$/))
    expect(cardButtons(wrapper).every(button => button.attributes('disabled') !== undefined)).toBe(true)
    expect(buttons[1].text()).toBe('userSubscriptions.resetCards.processing')

    resolveConsume(updated)
    await flushPromises()
    const rendered = (wrapper.vm as unknown as { subscriptions: UserSubscription[] }).subscriptions[0]
    expect(rendered).toEqual(updated)
    expect(wrapper.get('[data-test=reset-card-notice]').text()).toContain('userSubscriptions.resetCards.notice:2')
    expect(wrapper.text()).not.toContain('Plan')
    expect(showSuccess).toHaveBeenCalledWith('userSubscriptions.resetCards.success:30')
  })

  it('counts cards on expired subscriptions and does not use the legacy counter', async () => {
    const wrapper = await mountView([
      subscription(),
      subscription({ id: 2, status: 'expired', reset_cards: { total: 2, groups: [{ validity_days: 7, count: 2 }] } }),
      subscription({ id: 3, manual_reset_credits: 99, reset_cards: undefined }),
    ])
    expect(wrapper.get('[data-test=reset-card-notice]').text()).toContain('userSubscriptions.resetCards.notice:5')
  })

  it('keeps the backend error and silently refreshes after failure', async () => {
    consumeResetCard.mockRejectedValueOnce({ message: 'server consume rejected' })
    const wrapper = await mountView([subscription()])
    let resolveRefresh!: (value: UserSubscription[]) => void
    getMySubscriptions.mockReturnValueOnce(new Promise<UserSubscription[]>(resolve => { resolveRefresh = resolve }))
    await cardButtons(wrapper)[0].trigger('click')
    await flushPromises()

    expect(getMySubscriptions).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Plan')
    expect(wrapper.find('[data-test=reset-cards]').exists()).toBe(true)

    resolveRefresh([subscription({ reset_cards: { total: 0, groups: [] } })])
    await flushPromises()
    expect(showError).toHaveBeenCalledWith('server consume rejected')
    expect(wrapper.find('[data-test=reset-cards]').exists()).toBe(false)
  })

  it('does not overwrite the original error when silent refresh fails', async () => {
    consumeResetCard.mockRejectedValueOnce({ message: 'original consume error' })
    const wrapper = await mountView([subscription()])
    getMySubscriptions.mockRejectedValueOnce(new Error('refresh failed'))
    await cardButtons(wrapper)[0].trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenCalledTimes(1)
    expect(showError).toHaveBeenCalledWith('original consume error')
  })
})
