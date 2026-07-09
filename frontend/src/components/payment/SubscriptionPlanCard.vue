<template>
  <div
    :class="[
      'group relative flex flex-col overflow-hidden rounded-2xl border transition-all',
      'hover:shadow-xl hover:-translate-y-0.5',
      borderClass,
      'bg-white dark:bg-dark-800',
    ]"
  >
    <!-- Colored top accent bar -->
    <div :class="['h-1', accentClass]" />

    <div class="flex flex-1 flex-col p-6">
      <!-- Header: badge + readable name + validity -->
      <div class="mb-4">
        <div class="mb-3 flex min-w-0 items-start gap-2">
          <span :class="['mt-0.5 shrink-0 rounded-md border px-2 py-0.5 text-xs font-medium', badgeLightClass, borderClass]">
            {{ pLabel }}
          </span>
          <h3 class="min-w-0 flex-1 text-xl font-bold leading-tight text-gray-900 line-clamp-2 dark:text-white">
            {{ plan.name }}
          </h3>
          <span :class="['mt-0.5 shrink-0 rounded-full px-2 py-0.5 text-xs font-medium', badgeLightClass]">
            {{ validitySuffix }}
          </span>
        </div>
        <p v-if="plan.description" class="text-sm leading-relaxed text-gray-500 line-clamp-2 dark:text-dark-400">
          {{ plan.description }}
        </p>
      </div>

      <!-- Price -->
      <div class="mb-5 flex items-end gap-2">
        <span v-if="plan.original_price" class="pb-1 text-sm text-gray-400 line-through dark:text-dark-500">${{ plan.original_price }}</span>
        <div class="flex items-baseline gap-1">
          <span class="text-xl font-semibold text-gray-400 dark:text-dark-500">$</span>
          <span :class="['text-4xl font-extrabold leading-none tracking-tight', textClass]">{{ plan.price }}</span>
        </div>
        <span class="pb-1 text-sm text-gray-500 dark:text-dark-400">/ {{ validitySuffix }}</span>
        <span v-if="plan.original_price" :class="['mb-1 rounded px-1.5 py-0.5 text-[11px] font-semibold', discountClass]">{{ discountText }}</span>
      </div>

      <!-- Group quota info (compact) -->
      <div class="mb-5 grid grid-cols-2 gap-x-8 gap-y-4 text-sm">
        <div v-if="plan.daily_limit_usd != null">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <div class="mt-1 text-lg font-bold text-gray-900 dark:text-gray-100">${{ plan.daily_limit_usd }}</div>
        </div>
        <div v-if="plan.weekly_limit_usd != null">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <div class="mt-1 text-lg font-bold text-gray-900 dark:text-gray-100">${{ plan.weekly_limit_usd }}</div>
        </div>
        <div v-if="plan.monthly_limit_usd != null">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <div class="mt-1 text-lg font-bold text-gray-900 dark:text-gray-100">${{ plan.monthly_limit_usd }}</div>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.quota') }}</span>
          <div class="mt-1 text-lg font-bold text-gray-900 dark:text-gray-100">{{ t('payment.planCard.unlimited') }}</div>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="col-span-2">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.models') }}</span>
          <div class="mt-1 flex flex-wrap gap-1">
            <span v-for="scope in modelScopeLabels" :key="scope"
              class="rounded bg-gray-200/80 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">
              {{ scope }}
            </span>
          </div>
        </div>
      </div>

      <!-- Features list (compact) -->
      <div v-if="plan.features.length > 0" class="mb-4 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <svg :class="['mt-0.5 h-3.5 w-3.5 flex-shrink-0', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <!-- Subscribe Button -->
      <button
        type="button"
        :class="['inline-flex w-full items-center justify-center gap-1.5 rounded-xl py-3 text-sm font-semibold transition-all active:scale-[0.98]', btnClass]"
        @click="emit('select', plan)"
      >
        <Icon name="bolt" size="sm" :stroke-width="2" />
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBorderClass,
  platformTextClass,
  platformIconClass,
  platformButtonClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'
import { formatPlanValiditySuffix } from '@/utils/subscriptionPlanValidity'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const accentClass = computed(() => platformAccentBarClass(platform.value))
const borderClass = computed(() => platformBorderClass(platform.value))
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const btnClass = computed(() => platformButtonClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})
const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  if (platform.value !== 'antigravity') return []
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() =>
  formatPlanValiditySuffix(props.plan.validity_days, props.plan.validity_unit, t),
)
</script>
