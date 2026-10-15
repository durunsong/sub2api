<template>
  <BaseDialog
    :show="show"
    :title="t('admin.redeemEmail.detail.title', { id: batchId ?? '' })"
    width="extra-wide"
    :z-index="60"
    @close="close"
  >
    <div class="space-y-4">
      <div v-if="error" data-test="detail-error" class="rounded-lg bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/30 dark:text-red-200">
        {{ error }}
      </div>
      <div v-if="notice" class="rounded-lg bg-green-50 p-3 text-sm text-green-700 dark:bg-green-900/30 dark:text-green-200">
        {{ notice }}
      </div>

      <div v-if="batch" class="grid gap-3 text-sm sm:grid-cols-4">
        <div>
          <div class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.history.status') }}</div>
          <span data-test="detail-status" :class="batchStatusClass[batch.status]">{{ t(`admin.redeemEmail.status.${batch.status}`) }}</span>
        </div>
        <div>
          <div class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.history.progress') }}</div>
          <div class="tabular-nums text-gray-900 dark:text-gray-100">
            {{ batch.sent_count }} / {{ batch.total_count }}
            <span v-if="batch.failed_count" class="ml-1 text-red-600">{{ t('admin.redeemEmail.history.failedCount', { n: batch.failed_count }) }}</span>
            <span v-if="batch.skipped_count" class="ml-1 text-amber-600">{{ t('admin.redeemEmail.history.skippedCount', { n: batch.skipped_count }) }}</span>
          </div>
        </div>
        <div>
          <div class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.history.coupon') }}</div>
          <div class="text-gray-900 dark:text-gray-100">{{ formatCouponUSD(batch.coupon_value) }}</div>
        </div>
        <div>
          <div class="text-gray-500 dark:text-gray-400">{{ t('admin.redeemEmail.confirm.expiry') }}</div>
          <div class="text-gray-900 dark:text-gray-100">
            {{ batch.code_expires_at ? formatDateTime(batch.code_expires_at) : t('admin.redeemEmail.coupon.neverExpires') }}
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <select v-model="statusFilter" data-test="item-status-filter" class="input w-40" :aria-label="t('admin.redeemEmail.detail.status')">
          <option value="">{{ t('admin.redeemEmail.detail.allStatus') }}</option>
          <option v-for="s in itemStatuses" :key="s" :value="s">{{ t(`admin.redeemEmail.itemStatus.${s}`) }}</option>
        </select>
        <button
          v-if="batch && batch.failed_count > 0"
          type="button"
          data-test="retry-failed"
          class="btn btn-primary"
          :disabled="retrying"
          @click="retryFailed"
        >
          {{ retrying ? t('admin.redeemEmail.detail.retrying') : t('admin.redeemEmail.detail.retryFailed') }}
        </button>
      </div>
      <p v-if="hasUnknownOutcome" class="rounded-lg bg-amber-50 p-3 text-xs text-amber-800 dark:bg-amber-900/30 dark:text-amber-200">
        {{ t('admin.redeemEmail.detail.unknownOutcomeHint') }}
      </p>

      <div v-if="!items.length && !loading" class="py-10 text-center text-sm text-gray-500">{{ t('admin.redeemEmail.detail.empty') }}</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[820px] text-sm">
          <thead>
            <tr class="border-b text-left text-gray-500 dark:border-dark-600 dark:text-gray-400">
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.recipient') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.code') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.status') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.attempts') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.sentAt') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.detail.error') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id" class="border-b dark:border-dark-700">
              <td class="px-3 py-2">
                <div class="text-gray-900 dark:text-gray-100">{{ item.recipient_email }}</div>
                <div class="text-xs text-gray-500">{{ item.recipient_name }}</div>
              </td>
              <td class="px-3 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{{ item.redeem_code }}</td>
              <td class="px-3 py-2"><span :class="itemStatusClass[item.status]">{{ t(`admin.redeemEmail.itemStatus.${item.status}`) }}</span></td>
              <td class="px-3 py-2 tabular-nums">{{ item.attempt_count }}</td>
              <td class="whitespace-nowrap px-3 py-2 text-gray-500">{{ item.sent_at ? formatDateTime(item.sent_at) : '-' }}</td>
              <td
                :class="['px-3 py-2', item.status === 'skipped' ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400']"
                :data-test="`item-error-${item.id}`"
              >{{ deliveryError(item.error_code, item.error_message) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <Pagination
        v-if="total > pageSize"
        :page="page"
        :total="total"
        :page-size="pageSize"
        :show-page-size-selector="false"
        @update:page="changePage"
      />
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="close">{{ t('common.close') }}</button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import { extractI18nErrorMessage } from '@/utils/apiError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import { batchStatusClass, formatCouponUSD, isActiveBatchStatus, itemStatusClass } from './redeemEmailFormat'
import type { RedeemEmailBatch, RedeemEmailItem, RedeemEmailItemStatus } from '@/types'

const POLL_INTERVAL_MS = 5000
const itemStatuses: RedeemEmailItemStatus[] = ['pending', 'sending', 'sent', 'failed', 'skipped']

const props = defineProps<{ show: boolean; batchId: number | null }>()
const emit = defineEmits<{ close: [] }>()
const { t, te } = useI18n()

const batch = ref<RedeemEmailBatch | null>(null)
const items = ref<RedeemEmailItem[]>([])
const statusFilter = ref<RedeemEmailItemStatus | ''>('')
const page = ref(1)
const pageSize = 50
const total = ref(0)
const loading = ref(false)
const retrying = ref(false)
const error = ref('')
const notice = ref('')
let controller: AbortController | null = null
let pollTimer: ReturnType<typeof setTimeout> | null = null

const hasUnknownOutcome = computed(() => items.value.some((i) => i.error_code === 'DELIVERY_OUTCOME_UNKNOWN'))

function deliveryError(code: string | null, message: string | null): string {
  if (!code) return message || '-'
  const key = `admin.redeemEmail.deliveryErrors.${code}`
  return te(key) ? t(key) : message || code
}

const fail = (value: unknown) => {
  error.value = extractI18nErrorMessage(value, t, 'admin.redeemEmail.errors', t('admin.redeemEmail.errors.generic'))
}

function stopPolling() {
  if (pollTimer) clearTimeout(pollTimer)
  pollTimer = null
}

async function load() {
  if (props.batchId === null) return
  controller?.abort()
  const requestController = new AbortController()
  controller = requestController
  const id = props.batchId
  loading.value = true
  error.value = ''
  stopPolling()
  try {
    const [b, list] = await Promise.all([
      adminAPI.redeem.getEmailBatch(id, { signal: requestController.signal }),
      adminAPI.redeem.listEmailBatchItems(id, page.value, pageSize, statusFilter.value || undefined, { signal: requestController.signal })
    ])
    batch.value = b
    items.value = list?.items ?? []
    total.value = list?.total ?? 0
  } catch (value) {
    if ((value as { code?: string })?.code !== 'ERR_CANCELED') fail(value)
  } finally {
    if (controller === requestController) {
      loading.value = false
      if (props.show && batch.value && isActiveBatchStatus(batch.value.status)) {
        pollTimer = setTimeout(load, POLL_INTERVAL_MS)
      }
    }
  }
}

function changePage(value: number) {
  page.value = value
  load()
}

async function retryFailed() {
  if (props.batchId === null || retrying.value) return
  retrying.value = true
  error.value = ''
  notice.value = ''
  try {
    batch.value = await adminAPI.redeem.retryFailedEmailItems(props.batchId)
    notice.value = t('admin.redeemEmail.detail.retried')
    await load()
  } catch (value) {
    fail(value)
  } finally {
    retrying.value = false
  }
}

function close() {
  controller?.abort()
  stopPolling()
  emit('close')
}

watch(statusFilter, () => {
  page.value = 1
  load()
})

watch(
  () => [props.show, props.batchId] as const,
  ([show]) => {
    if (show) {
      batch.value = null
      items.value = []
      notice.value = ''
      page.value = 1
      if (statusFilter.value) statusFilter.value = ''
      else load()
    } else {
      controller?.abort()
      stopPolling()
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  controller?.abort()
  stopPolling()
})
</script>
