<template>
  <BaseDialog :show="show" :title="t('admin.redeemEmail.history.title')" width="extra-wide" @close="close">
    <div class="space-y-4">
      <div class="flex justify-end">
        <button type="button" class="btn btn-secondary" :title="t('common.refresh')" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>
      <div v-if="notice" class="rounded-lg bg-green-50 p-3 text-sm text-green-700 dark:bg-green-900/30 dark:text-green-200">
        {{ notice }}
      </div>
      <div v-if="error" data-test="history-error" class="rounded-lg bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/30 dark:text-red-200">
        {{ error }}
      </div>

      <div v-if="loading && !rows.length" class="py-12 text-center text-gray-500">{{ t('common.loading') }}</div>
      <div v-else-if="!rows.length" class="py-12 text-center text-gray-500">{{ t('admin.redeemEmail.history.empty') }}</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[880px] text-sm">
          <thead>
            <tr class="border-b text-left text-gray-500 dark:border-dark-600 dark:text-gray-400">
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.batch') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.coupon') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.filter') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.status') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.progress') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.redeemEmail.history.createdAt') }}</th>
              <th class="px-3 py-2 text-right font-medium">{{ t('admin.redeemEmail.history.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="batch in rows" :key="batch.id" class="border-b dark:border-dark-700">
              <td class="px-3 py-3 tabular-nums text-gray-500">#{{ batch.id }}</td>
              <td class="px-3 py-3 tabular-nums">{{ formatCouponUSD(batch.coupon_value) }}</td>
              <td :data-test="`batch-segments-${batch.id}`" class="px-3 py-3 text-gray-600 dark:text-gray-300">
                {{ batch.segments.map((s) => t(`admin.redeemEmail.recipients.segment.${s}`)).join(' / ') }}
              </td>
              <td class="px-3 py-3">
                <span :data-test="`batch-status-${batch.id}`" :class="batchStatusClass[batch.status]">
                  {{ t(`admin.redeemEmail.status.${batch.status}`) }}
                </span>
              </td>
              <td class="whitespace-nowrap px-3 py-3 tabular-nums text-gray-600 dark:text-gray-300">
                {{ batch.sent_count }} / {{ batch.total_count }}
                <span v-if="batch.failed_count" class="ml-2 text-red-600 dark:text-red-400">
                  {{ t('admin.redeemEmail.history.failedCount', { n: batch.failed_count }) }}
                </span>
                <span v-if="batch.skipped_count" class="ml-2 text-amber-600 dark:text-amber-400">
                  {{ t('admin.redeemEmail.history.skippedCount', { n: batch.skipped_count }) }}
                </span>
              </td>
              <td class="whitespace-nowrap px-3 py-3 text-gray-500 dark:text-gray-400">{{ formatDateTime(batch.created_at) }}</td>
              <td class="px-3 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button
                    type="button"
                    :data-test="`view-batch-${batch.id}`"
                    class="rounded p-2 text-gray-500 hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700"
                    :title="t('admin.redeemEmail.history.view')"
                    :aria-label="t('admin.redeemEmail.history.view')"
                    @click="detailId = batch.id"
                  >
                    <Icon name="eye" size="sm" />
                  </button>
                  <button
                    v-if="!isActiveBatchStatus(batch.status)"
                    type="button"
                    :data-test="`delete-batch-${batch.id}`"
                    class="rounded p-2 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                    :title="t('admin.redeemEmail.history.delete')"
                    :aria-label="t('admin.redeemEmail.history.delete')"
                    :disabled="deleting"
                    @click="deleteTarget = batch"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </td>
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

  <RedeemEmailBatchDetailDialog :show="detailId !== null" :batch-id="detailId" @close="onDetailClose" />

  <ConfirmDialog
    :show="deleteTarget !== null"
    :title="t('admin.redeemEmail.history.delete')"
    :message="t('admin.redeemEmail.history.deleteConfirm')"
    :confirm-text="deleting ? t('admin.redeemEmail.history.deleting') : t('common.delete')"
    :cancel-text="t('common.cancel')"
    danger
    @confirm="confirmDelete"
    @cancel="deleteTarget = null"
  />
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import { extractI18nErrorMessage } from '@/utils/apiError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import RedeemEmailBatchDetailDialog from './RedeemEmailBatchDetailDialog.vue'
import { batchStatusClass, formatCouponUSD, isActiveBatchStatus } from './redeemEmailFormat'
import type { RedeemEmailBatch } from '@/types'

const POLL_INTERVAL_MS = 5000

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()

const rows = ref<RedeemEmailBatch[]>([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const loading = ref(false)
const deleting = ref(false)
const error = ref('')
const notice = ref('')
const deleteTarget = ref<RedeemEmailBatch | null>(null)
const detailId = ref<number | null>(null)
let controller: AbortController | null = null
let pollTimer: ReturnType<typeof setTimeout> | null = null

const fail = (value: unknown) => {
  error.value = extractI18nErrorMessage(value, t, 'admin.redeemEmail.errors', t('admin.redeemEmail.errors.generic'))
}

function stopPolling() {
  if (pollTimer) clearTimeout(pollTimer)
  pollTimer = null
}

function schedulePoll() {
  stopPolling()
  if (props.show && rows.value.some((b) => isActiveBatchStatus(b.status))) {
    pollTimer = setTimeout(load, POLL_INTERVAL_MS)
  }
}

async function load() {
  controller?.abort()
  const requestController = new AbortController()
  controller = requestController
  loading.value = true
  error.value = ''
  try {
    const result = await adminAPI.redeem.listEmailBatches(page.value, pageSize, { signal: requestController.signal })
    rows.value = result?.items ?? []
    total.value = result?.total ?? 0
  } catch (value) {
    if ((value as { code?: string })?.code !== 'ERR_CANCELED') fail(value)
  } finally {
    if (controller === requestController) {
      loading.value = false
      schedulePoll()
    }
  }
}

function changePage(value: number) {
  page.value = value
  load()
}

function onDetailClose() {
  detailId.value = null
  load()
}

async function confirmDelete() {
  if (!deleteTarget.value || deleting.value) return
  deleting.value = true
  error.value = ''
  notice.value = ''
  try {
    await adminAPI.redeem.deleteEmailBatch(deleteTarget.value.id)
    deleteTarget.value = null
    notice.value = t('admin.redeemEmail.history.deleted')
    if (rows.value.length === 1 && page.value > 1) page.value -= 1
    await load()
  } catch (value) {
    deleteTarget.value = null
    fail(value)
  } finally {
    deleting.value = false
  }
}

function close() {
  controller?.abort()
  stopPolling()
  deleteTarget.value = null
  emit('close')
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      page.value = 1
      notice.value = ''
      load()
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
