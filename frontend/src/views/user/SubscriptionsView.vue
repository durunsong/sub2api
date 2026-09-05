<template>
  <AppLayout>
    <div class="space-y-6">
      <div
        v-if="!loading && resetCardTotal > 0"
        role="status"
        data-test="reset-card-notice"
        class="flex items-center gap-2 border-l-4 border-emerald-500 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-800 dark:bg-emerald-950/25 dark:text-emerald-200"
      >
        <Icon name="creditCard" size="md" class="shrink-0" />
        {{ t('userSubscriptions.resetCards.notice', { count: resetCardTotal }) }}
      </div>
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="subscriptions.length === 0" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
      </div>

      <!-- Subscriptions Grid -->
      <div v-else class="grid gap-6 lg:grid-cols-2">
        <div
          v-for="subscription in subscriptions"
          :key="subscription.id"
          class="overflow-hidden rounded-2xl border bg-white dark:bg-dark-800"
          :class="platformBorderClass(userFacingPlatform(subscription.group?.platform || ''))"
        >
          <!-- Header -->
          <div
            class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div :class="['h-1.5 w-1.5 shrink-0 rounded-full', platformAccentDotClass(userFacingPlatform(subscription.group?.platform || ''))]" />
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <h3 class="truncate font-semibold text-gray-900 dark:text-white">
                    {{ subscription.group?.name ? userFacingPlatformText(subscription.group.name) : `Group #${subscription.group_id}` }}
                  </h3>
                  <span :class="['shrink-0 whitespace-nowrap rounded-md border px-2 py-0.5 text-[11px] font-medium', platformBadgeClass(userFacingPlatform(subscription.group?.platform || ''))]">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="mt-0.5 break-words text-xs leading-relaxed text-gray-500 dark:text-dark-400">
                  {{ userFacingPlatformText(subscription.group.description) }}
                </p>
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <span
                :class="[
                  'whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium',
                  subscription.status === 'active'
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : subscription.status === 'expired'
                      ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
                ]"
              >
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active'"
                :class="['whitespace-nowrap rounded-lg px-3 py-1.5 text-xs font-semibold text-white transition-colors', platformButtonClass(userFacingPlatform(subscription.group?.platform || ''))]"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </div>

          <!-- Reset card benefits -->
          <div
            v-if="subscription.reset_cards && subscription.reset_cards.total > 0"
            data-test="reset-cards"
            class="flex flex-wrap items-center justify-between gap-3 border-b border-emerald-100 bg-emerald-50/70 px-4 py-3 dark:border-emerald-900/60 dark:bg-emerald-950/25"
          >
            <div class="flex items-center gap-2 text-xs">
              <span class="font-semibold text-emerald-800 dark:text-emerald-200">
                {{ t('userSubscriptions.resetCards.title', { count: subscription.reset_cards.total }) }}
              </span>
              <span class="text-emerald-700/75 dark:text-emerald-300/75">
                {{ t('userSubscriptions.resetCards.permanent') }}
              </span>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="card in subscription.reset_cards.groups"
                :key="card.validity_days"
                type="button"
                class="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-emerald-500 dark:hover:bg-emerald-400"
                :disabled="!canConsumeResetCard(subscription) || resettingId !== null"
                @click="handleConsumeResetCard(subscription, card.validity_days)"
              >
                {{
                  resettingId === subscription.id && resettingValidityDays === card.validity_days
                    ? t('userSubscriptions.resetCards.processing')
                    : t('userSubscriptions.resetCards.action', { days: card.validity_days, count: card.count })
                }}
              </button>
            </div>
          </div>

          <!-- Usage Progress -->
          <div class="space-y-4 p-4">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.daily_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <p class="text-xs tabular-nums text-gray-500 dark:text-dark-400">
                {{
                  t('userSubscriptions.tokenConsumed', {
                    tokens: formatTokenCount(subscription.daily_usage_tokens)
                  })
                }}
              </p>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.daily_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ formatDailyUsageWindow(subscription) }}
              </p>
              <p
                v-if="isOneTimeDailyQuota(subscription)"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ t('userSubscriptions.oneTimeDailyHint') }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.weekly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <p class="text-xs tabular-nums text-gray-500 dark:text-dark-400">
                {{
                  t('userSubscriptions.tokenConsumed', {
                    tokens: formatTokenCount(subscription.weekly_usage_tokens)
                  })
                }}
              </p>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.weekly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168)
                  })
                }}
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.monthly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <p class="text-xs tabular-nums text-gray-500 dark:text-dark-400">
                {{
                  t('userSubscriptions.tokenConsumed', {
                    tokens: formatTokenCount(subscription.monthly_usage_tokens)
                  })
                }}
              </p>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.monthly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720)
                  })
                }}
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !subscription.group?.daily_limit_usd &&
                !subscription.group?.weekly_limit_usd &&
                !subscription.group?.monthly_limit_usd
              "
              class="flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 py-6 dark:from-emerald-900/20 dark:to-teal-900/20"
            >
              <div class="flex items-center gap-3">
                <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                  <p class="mt-1 text-xs tabular-nums text-emerald-600/70 dark:text-emerald-400/70">
                    {{
                      t('userSubscriptions.tokenConsumed', {
                        tokens: formatTokenCount(
                          subscription.monthly_usage_tokens ||
                            subscription.weekly_usage_tokens ||
                            subscription.daily_usage_tokens
                        )
                      })
                    }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTimeToMinute } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  platformBorderClass,
  platformBadgeClass,
  platformButtonClass,
  platformLabel,
  userFacingPlatform,
  userFacingPlatformText,
} from '@/utils/platformColors'
import {
  getExpirationDateRelation,
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts,
} from '@/utils/subscriptionQuota'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const resetCardTotal = computed(() => subscriptions.value.reduce((total, subscription) => total + (subscription.reset_cards?.total ?? 0), 0))
const loading = ref(true)
const resettingId = ref<number | null>(null)
const resettingValidityDays = ref<number | null>(null)

async function loadSubscriptions(showLoadError = true, showLoading = true) {
  try {
    if (showLoading) loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    if (showLoadError) appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    if (showLoading) loading.value = false
  }
}

function isSubscriptionExpired(subscription: UserSubscription): boolean {
  if (subscription.status === 'expired') return true
  if (!subscription.expires_at) return false
  return new Date(subscription.expires_at).getTime() <= Date.now()
}

function canConsumeResetCard(subscription: UserSubscription): boolean {
  return subscription.status === 'active' || subscription.status === 'expired'
}

function createIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `reset-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

async function handleConsumeResetCard(subscription: UserSubscription, validityDays: number) {
  if (!canConsumeResetCard(subscription) || resettingId.value != null) return
  if (!window.confirm(t('userSubscriptions.resetCards.confirm', { days: validityDays }))) return

  resettingId.value = subscription.id
  resettingValidityDays.value = validityDays
  try {
    const updated = await subscriptionsAPI.consumeResetCard(subscription.id, validityDays, createIdempotencyKey())
    const index = subscriptions.value.findIndex((item) => item.id === subscription.id)
    if (index >= 0) subscriptions.value[index] = updated
    appStore.showSuccess(t('userSubscriptions.resetCards.success', { days: validityDays }))
  } catch (error) {
    console.error('Failed to consume reset card:', error)
    const message = extractApiErrorMessage(error, t('userSubscriptions.resetCards.failed'))
    await loadSubscriptions(false, false)
    appStore.showError(message)
  } finally {
    resettingId.value = null
    resettingValidityDays.value = null
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const relation = getExpirationDateRelation(expires, now)

  if (relation === null) return ''

  if (relation === 'expired') {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (relation === 'today') {
    return `${dateStr} (${t('common.today')})`
  }
  if (relation === 'tomorrow') {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (diff <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    if (isSubscriptionExpired(subscription)) {
      return t('userSubscriptions.quotaEnded')
    }
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.quotaEnded')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatTokenCount(tokens?: number | null): string {
  return (tokens || 0).toLocaleString()
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>
