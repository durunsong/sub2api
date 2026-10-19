<template>
  <BaseDialog :show="show" :title="t('admin.redeemEmail.title')" width="extra-wide" @close="close">
    <div class="space-y-5">
      <ol class="flex flex-wrap items-center gap-2 text-sm" data-test="campaign-steps">
        <li
          v-for="(key, index) in stepKeys"
          :key="key"
          :class="[
            'rounded-full px-3 py-1',
            index === step
              ? 'bg-primary-600 text-white'
              : index < step
                ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
          ]"
        >
          {{ index + 1 }}. {{ t(`admin.redeemEmail.steps.${key}`) }}
        </li>
      </ol>

      <div
        v-if="defaults && !defaults.smtp_configured"
        data-test="smtp-warning"
        class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800 dark:bg-amber-900/30 dark:text-amber-200"
      >
        {{ t('admin.redeemEmail.smtpNotConfigured') }}
      </div>
      <div
        v-if="error"
        data-test="campaign-error"
        class="rounded-lg bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/30 dark:text-red-200"
      >
        {{ error }}
      </div>

      <!-- Step 1: coupon -->
      <section v-if="step === 0" class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="input-label" for="redeem-email-coupon-value">{{ t('admin.redeemEmail.coupon.value') }}</label>
          <input
            id="redeem-email-coupon-value"
            v-model.number="couponValue"
            data-test="coupon-value"
            type="number"
            min="0.00000001"
            step="0.01"
            class="input"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.coupon.valueHint') }}</p>
        </div>
        <div>
          <span class="input-label">{{ t('admin.redeemEmail.coupon.expiry') }}</span>
          <label class="mt-2 flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="neverExpires" data-test="never-expires" type="checkbox" class="rounded" />
            {{ t('admin.redeemEmail.coupon.neverExpires') }}
          </label>
          <div v-if="!neverExpires" class="mt-2">
            <label class="sr-only" for="redeem-email-expiry-days">{{ t('admin.redeemEmail.coupon.expiresInDays') }}</label>
            <div class="flex items-center gap-2">
              <input
                id="redeem-email-expiry-days"
                v-model.number="expiresInDays"
                data-test="expires-in-days"
                type="number"
                min="1"
                :max="maxExpiresInDays"
                step="1"
                class="input w-32"
              />
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.coupon.expiresInDays') }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Step 2: recipients -->
      <section v-else-if="step === 1" class="space-y-4">
        <div>
          <span id="redeem-email-segments-label" class="input-label">{{ t('admin.redeemEmail.recipients.segments') }}</span>
          <div class="mt-1 flex flex-wrap gap-2" role="group" aria-labelledby="redeem-email-segments-label">
            <button
              v-for="segment in segmentOptions"
              :key="segment"
              type="button"
              :data-test="`segment-${segment}`"
              :aria-pressed="segments.includes(segment)"
              :title="segments.length === 1 && segments.includes(segment) ? t('admin.redeemEmail.recipients.segmentKeepOne') : undefined"
              :class="[
                'rounded-full border px-3 py-1.5 text-sm transition-colors',
                segments.includes(segment)
                  ? 'border-primary-600 bg-primary-600 text-white'
                  : 'border-gray-300 bg-white text-gray-700 hover:border-primary-400 dark:border-dark-500 dark:bg-dark-800 dark:text-gray-300'
              ]"
              @click="toggleSegment(segment)"
            >
              {{ t(`admin.redeemEmail.recipients.segment.${segment}`) }}
            </button>
          </div>
        </div>
        <div class="flex flex-wrap items-end gap-3">
          <div class="min-w-48 flex-1">
            <label class="sr-only" for="redeem-email-search">{{ t('admin.redeemEmail.recipients.search') }}</label>
            <input
              id="redeem-email-search"
              v-model="search"
              type="text"
              maxlength="100"
              class="input"
              :placeholder="t('admin.redeemEmail.recipients.search')"
              @keyup.enter="searchRecipients"
            />
          </div>
          <button type="button" data-test="search-recipients" class="btn btn-secondary" :disabled="searching" @click="searchRecipients">
            {{ searching ? t('admin.redeemEmail.recipients.filtering') : t('admin.redeemEmail.recipients.filter') }}
          </button>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.recipients.hint') }}</p>

        <div v-if="recipients === null" class="py-10 text-center text-sm text-gray-500">
          {{ t('admin.redeemEmail.recipients.notSearched') }}
        </div>
        <div v-else-if="!recipients.length" class="py-10 text-center text-sm text-gray-500">
          {{ t('admin.redeemEmail.recipients.empty') }}
        </div>
        <template v-else>
          <div class="flex flex-wrap items-center justify-between gap-3 text-sm">
            <span data-test="selected-count" class="text-gray-600 dark:text-gray-300">
              {{ t('admin.redeemEmail.recipients.selectedCount', { selected: selectedIds.size, total: recipients.length }) }}
            </span>
            <div class="flex gap-2">
              <button type="button" data-test="select-all" class="btn btn-secondary btn-sm" @click="selectAll">
                {{ t('admin.redeemEmail.recipients.selectAll') }}
              </button>
              <button type="button" data-test="deselect-all" class="btn btn-secondary btn-sm" @click="selectedIds = new Set()">
                {{ t('admin.redeemEmail.recipients.deselectAll') }}
              </button>
            </div>
          </div>
          <p v-if="hasUnsubscribed" data-test="unsubscribed-hint" class="text-xs text-amber-700 dark:text-amber-300">
            {{ t('admin.redeemEmail.recipients.unsubscribedHint') }}
          </p>
          <div class="max-h-80 overflow-auto rounded-lg border dark:border-dark-600">
            <table class="w-full text-sm">
              <thead class="sticky top-0 bg-gray-50 dark:bg-dark-700">
                <tr class="text-left text-gray-500 dark:text-gray-400">
                  <th class="w-10 px-3 py-2"></th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.recipients.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.recipients.username') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.recipients.segmentColumn') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.recipients.lastUsedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in recipients" :key="r.user_id" class="border-t dark:border-dark-700">
                  <td class="px-3 py-2">
                    <input
                      type="checkbox"
                      class="rounded"
                      :data-test="`recipient-${r.user_id}`"
                      :checked="selectedIds.has(r.user_id)"
                      :aria-label="r.email"
                      @change="toggle(r.user_id)"
                    />
                  </td>
                  <td class="px-3 py-2 text-gray-900 dark:text-gray-100">
                    {{ r.email }}
                    <span
                      v-if="r.unsubscribed"
                      :data-test="`unsubscribed-${r.user_id}`"
                      class="ml-1 rounded bg-amber-100 px-1.5 py-0.5 text-xs text-amber-800 dark:bg-amber-900/40 dark:text-amber-200"
                    >
                      {{ t('admin.redeemEmail.recipients.unsubscribed') }}
                    </span>
                  </td>
                  <td class="px-3 py-2 text-gray-600 dark:text-gray-300">{{ r.username || '-' }}</td>
                  <td class="whitespace-nowrap px-3 py-2 text-gray-600 dark:text-gray-300">
                    {{ t(`admin.redeemEmail.recipients.segment.${r.segment}`) }}
                  </td>
                  <td class="whitespace-nowrap px-3 py-2 text-gray-500 dark:text-gray-400">
                    {{ r.last_used_at ? formatDateTime(r.last_used_at) : t('admin.redeemEmail.recipients.neverUsed') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </section>

      <!-- Step 3: template preview (template is maintained in /admin/settings) -->
      <section v-else-if="step === 2" class="space-y-3">
        <p class="rounded-lg bg-gray-50 p-3 text-sm text-gray-700 dark:bg-dark-700 dark:text-gray-300" data-test="template-source">
          {{ t('admin.redeemEmail.template.source') }}
        </p>
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.template.previewLocale') }}</span>
          <button
            v-for="loc in previewLocales"
            :key="loc"
            type="button"
            :data-test="`preview-locale-${loc}`"
            :aria-pressed="previewLocale === loc"
            :class="['btn btn-sm', previewLocale === loc ? 'btn-primary' : 'btn-secondary']"
            @click="switchPreviewLocale(loc)"
          >
            {{ t(`admin.redeemEmail.template.locale.${loc}`) }}
          </button>
          <span class="flex-1"></span>
          <a
            :href="settingsTemplateHref"
            target="_blank"
            rel="noopener"
            data-test="edit-template-link"
            class="btn btn-secondary btn-sm"
          >
            {{ t('admin.redeemEmail.template.editInSettings') }}
          </a>
          <button type="button" data-test="preview-email" class="btn btn-secondary btn-sm" :disabled="previewing" @click="loadPreview">
            {{ previewing ? t('admin.redeemEmail.template.previewing') : t('admin.redeemEmail.template.refresh') }}
          </button>
        </div>
        <EmailPreview :preview="preview" />
      </section>

      <!-- Step 4: confirm -->
      <section v-else class="grid gap-4 lg:grid-cols-2">
        <dl class="grid grid-cols-[auto,1fr] content-start gap-x-4 gap-y-2 text-sm" data-test="confirm-summary">
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.recipients') }}</dt>
          <dd data-test="confirm-count" class="font-medium text-gray-900 dark:text-gray-100">{{ selectedIds.size }}</dd>
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.perCode') }}</dt>
          <dd class="font-medium text-gray-900 dark:text-gray-100">{{ formatCouponUSD(couponValue || 0) }}</dd>
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.total') }}</dt>
          <dd data-test="confirm-total" class="font-semibold text-primary-600 dark:text-primary-400">
            {{ formatCouponUSD(totalCouponValue(couponValue || 0, selectedIds.size)) }}
          </dd>
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.expiry') }}</dt>
          <dd class="text-gray-900 dark:text-gray-100">
            {{ neverExpires ? t('admin.redeemEmail.coupon.neverExpires') : t('admin.redeemEmail.confirm.expiryDays', { days: expiresInDays }) }}
          </dd>
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.filter') }}</dt>
          <dd data-test="confirm-segments" class="text-gray-900 dark:text-gray-100">{{ segmentLabels }}</dd>
          <dt class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.subject') }}</dt>
          <dd class="break-all text-gray-900 dark:text-gray-100">{{ preview?.subject ?? '-' }}</dd>
          <dd v-if="selectedUnsubscribedCount" data-test="confirm-unsubscribed" class="col-span-2 mt-2 text-xs text-amber-700 dark:text-amber-300">
            {{ t('admin.redeemEmail.confirm.unsubscribeHint') }}
          </dd>
          <dd class="col-span-2 mt-2 text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.redeemEmail.confirm.rateHint', { rate: defaults?.rate_per_second ?? 1 }) }}
          </dd>
        </dl>
        <EmailPreview :preview="preview" />
      </section>
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="close">{{ t('common.cancel') }}</button>
      <button v-if="step > 0" type="button" data-test="prev-step" class="btn btn-secondary" :disabled="submitting" @click="goPrev">
        {{ t('admin.redeemEmail.prev') }}
      </button>
      <button v-if="step < stepKeys.length - 1" type="button" data-test="next-step" class="btn btn-primary" :disabled="loadingNext" @click="goNext">
        {{ t('admin.redeemEmail.next') }}
      </button>
      <button
        v-else
        type="button"
        data-test="submit-campaign"
        class="btn btn-primary"
        :disabled="submitting || !defaults?.smtp_configured"
        @click="submit"
      >
        {{ submitting ? t('admin.redeemEmail.confirm.sending') : t('admin.redeemEmail.confirm.send') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import { extractI18nErrorMessage } from '@/utils/apiError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { formatCouponUSD, isValidCouponValue, newIdempotencyKey, totalCouponValue } from './redeemEmailFormat'
import type {
  CreateRedeemEmailBatchRequest,
  RedeemEmailBatch,
  RedeemEmailDefaults,
  RedeemEmailLocale,
  RedeemEmailPreview,
  RedeemEmailRecipient,
  RedeemEmailSegment
} from '@/types'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: []; created: [batch: RedeemEmailBatch] }>()
const { t, locale } = useI18n()

// 邮件 HTML 来自管理员可编辑的模板，放进无权限的 sandbox iframe 渲染，与设置页预览一致。
const EmailPreview = defineComponent({
  props: { preview: { type: Object as PropType<RedeemEmailPreview | null>, default: null } },
  setup(p) {
    return () =>
      h('div', { class: 'space-y-2' }, [
        h('div', { class: 'input-label' }, t('admin.redeemEmail.template.previewTitle')),
        p.preview
          ? h('div', { class: 'overflow-hidden rounded-lg border dark:border-dark-600', 'data-test': 'email-preview' }, [
              h('div', { class: 'border-b px-3 py-2 text-sm font-semibold text-gray-900 dark:border-dark-600 dark:text-gray-100' }, p.preview.subject),
              h('iframe', {
                class: 'h-[28rem] w-full bg-white',
                sandbox: '',
                srcdoc: p.preview.html,
                title: t('admin.redeemEmail.template.previewTitle')
              })
            ])
          : h('p', { class: 'rounded-lg border border-dashed p-6 text-center text-sm text-gray-500 dark:border-dark-600' }, t('admin.redeemEmail.template.previewHint'))
      ])
  }
})

const stepKeys = ['coupon', 'recipients', 'template', 'confirm'] as const
const fallbackSegments: RedeemEmailSegment[] = ['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used', 'zero_balance']
const previewLocales: RedeemEmailLocale[] = ['zh', 'en']
const step = ref(0)
const defaults = ref<RedeemEmailDefaults | null>(null)
const error = ref('')

const couponValue = ref<number>(5)
const neverExpires = ref(false)
const expiresInDays = ref<number>(30)
const segments = ref<RedeemEmailSegment[]>(fallbackSegments.slice(0, 3))
const search = ref('')
const recipients = ref<RedeemEmailRecipient[] | null>(null)
const selectedIds = ref<Set<number>>(new Set())
const searching = ref(false)
const previewLocale = ref<RedeemEmailLocale>('zh')
const preview = ref<RedeemEmailPreview | null>(null)
const previewing = ref(false)
const loadingNext = ref(false)
const submitting = ref(false)

let idempotencyKey = ''
let idempotencyPayload = ''

const segmentOptions = computed<RedeemEmailSegment[]>(() => defaults.value?.segment_options ?? fallbackSegments)
const maxExpiresInDays = computed(() => defaults.value?.max_expires_in_days ?? 3650)
const maxRecipients = computed(() => defaults.value?.max_recipients ?? 1000)
const hasUnsubscribed = computed(() => (recipients.value ?? []).some((r) => r.unsubscribed))
const selectedUnsubscribedCount = computed(
  () => (recipients.value ?? []).filter((r) => r.unsubscribed && selectedIds.value.has(r.user_id)).length
)
const segmentLabels = computed(() =>
  segmentOptions.value
    .filter((s) => segments.value.includes(s))
    .map((s) => t(`admin.redeemEmail.recipients.segment.${s}`))
    .join(locale.value.startsWith('zh') ? '、' : ', ')
)
const settingsTemplateHref = computed(() => {
  const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')
  const event = encodeURIComponent(defaults.value?.template_event ?? 'redeem.inactive_coupon')
  return `${base}/admin/settings?tab=email&email_event=${event}`
})

const fail = (value: unknown) => {
  error.value = extractI18nErrorMessage(value, t, 'admin.redeemEmail.errors', t('admin.redeemEmail.errors.generic'))
}
const isCanceled = (value: unknown) => (value as { code?: string } | null)?.code === 'ERR_CANCELED'

function couponPayload() {
  return neverExpires.value
    ? { coupon_value: couponValue.value, never_expires: true }
    : { coupon_value: couponValue.value, code_expires_in_days: expiresInDays.value }
}

function validateCoupon(): boolean {
  if (!isValidCouponValue(couponValue.value)) {
    error.value = t('admin.redeemEmail.coupon.valueInvalid')
    return false
  }
  const days = expiresInDays.value
  if (!neverExpires.value && (!Number.isInteger(days) || days < 1 || days > maxExpiresInDays.value)) {
    error.value = t('admin.redeemEmail.coupon.expiryInvalid', { max: maxExpiresInDays.value })
    return false
  }
  return true
}

function toggleSegment(segment: RedeemEmailSegment) {
  // 余额为 0 与未活跃分段互斥：选中后只保留这一项，避免和“未用过 API”取并集。
  if (segment === 'zero_balance') {
    if (segments.value.length === 1 && segments.value[0] === 'zero_balance') return
    segments.value = ['zero_balance']
    return
  }
  const current = segments.value.filter((s) => s !== 'zero_balance')
  if (current.includes(segment)) {
    if (current.length === 1) return
    segments.value = current.filter((s) => s !== segment)
    return
  }
  segments.value = segmentOptions.value.filter((s) => s !== 'zero_balance' && (s === segment || current.includes(s)))
}

// 改变分段后旧结果不再对应提交时的筛选快照，必须重新筛选。
watch(segments, () => {
  recipients.value = null
  selectedIds.value = new Set()
})

async function searchRecipients() {
  if (searching.value) return
  searching.value = true
  error.value = ''
  try {
    const result = await adminAPI.redeem.searchEmailRecipients({
      segments: segments.value,
      search: search.value.trim() || undefined
    })
    recipients.value = result.items
    selectedIds.value = new Set(result.items.filter((r) => !r.unsubscribed).map((r) => r.user_id))
  } catch (value) {
    recipients.value = null
    selectedIds.value = new Set()
    if ((value as { reason?: string })?.reason === 'REDEEM_EMAIL_RECIPIENT_LIMIT_EXCEEDED') {
      error.value = t('admin.redeemEmail.recipients.limitExceeded', { max: maxRecipients.value })
    } else {
      fail(value)
    }
  } finally {
    searching.value = false
  }
}

function toggle(id: number) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

function selectAll() {
  selectedIds.value = new Set((recipients.value ?? []).map((r) => r.user_id))
}

async function loadPreview(): Promise<boolean> {
  previewing.value = true
  error.value = ''
  try {
    preview.value = await adminAPI.redeem.previewEmail({ ...couponPayload(), locale: previewLocale.value })
    return true
  } catch (value) {
    preview.value = null
    fail(value)
    return false
  } finally {
    previewing.value = false
  }
}

function switchPreviewLocale(loc: RedeemEmailLocale) {
  if (previewLocale.value === loc) return
  previewLocale.value = loc
  loadPreview()
}

async function goNext() {
  error.value = ''
  if (step.value === 0 && !validateCoupon()) return
  if (step.value === 1) {
    if (selectedIds.value.size === 0) {
      error.value = t('admin.redeemEmail.recipients.selectAtLeastOne')
      return
    }
    if (selectedIds.value.size > maxRecipients.value) {
      error.value = t('admin.redeemEmail.recipients.limitExceeded', { max: maxRecipients.value })
      return
    }
  }
  if (step.value === 1 || step.value === 2) {
    loadingNext.value = true
    const ok = await loadPreview()
    loadingNext.value = false
    if (!ok) return
  }
  step.value += 1
}

function goPrev() {
  error.value = ''
  step.value = Math.max(0, step.value - 1)
}

async function submit() {
  if (submitting.value || !validateCoupon()) return
  const payload: CreateRedeemEmailBatchRequest = {
    ...couponPayload(),
    segments: [...segments.value],
    user_ids: [...selectedIds.value].sort((a, b) => a - b)
  }
  // 同一请求体复用同一个 key：网络重试会命中服务端幂等回放，而修改内容后重新提交会使用新 key。
  const signature = JSON.stringify(payload)
  if (signature !== idempotencyPayload) {
    idempotencyKey = newIdempotencyKey()
    idempotencyPayload = signature
  }
  submitting.value = true
  error.value = ''
  try {
    const batch = await adminAPI.redeem.createEmailBatch(payload, idempotencyKey)
    emit('created', batch)
  } catch (value) {
    fail(value)
  } finally {
    submitting.value = false
  }
}

function resetState() {
  step.value = 0
  error.value = ''
  couponValue.value = 5
  neverExpires.value = false
  segments.value = fallbackSegments.slice(0, 3)
  search.value = ''
  recipients.value = null
  selectedIds.value = new Set()
  previewLocale.value = locale.value.startsWith('zh') ? 'zh' : 'en'
  preview.value = null
  idempotencyKey = ''
  idempotencyPayload = ''
}

async function loadDefaults() {
  try {
    defaults.value = await adminAPI.redeem.getEmailDefaults()
    expiresInDays.value = defaults.value.default_expires_in_days || 30
    if (defaults.value.default_segments?.length) segments.value = [...defaults.value.default_segments]
  } catch (value) {
    if (!isCanceled(value)) fail(value)
  }
}

function close() {
  if (submitting.value) return
  emit('close')
}

watch(
  () => props.show,
  (show) => {
    if (!show) return
    resetState()
    loadDefaults()
  },
  { immediate: true }
)
</script>
