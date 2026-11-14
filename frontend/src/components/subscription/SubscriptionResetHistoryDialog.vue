<template>
  <BaseDialog
    :show="subscriptionId !== null"
    :title="t('userSubscriptions.resetHistory.title')"
    width="wide"
    @close="emit('close')"
  >
    <p class="mb-2 break-words font-medium text-gray-900 dark:text-white">{{ subscriptionName }}</p>
    <p class="mb-5 text-sm text-gray-500 dark:text-dark-400">
      {{ t('userSubscriptions.resetHistory.coverageHint') }}
    </p>
    <div v-if="loading" role="status" class="py-10 text-center text-gray-500 dark:text-dark-400">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="error" role="alert" class="space-y-3 py-8 text-center">
      <p class="text-red-600 dark:text-red-400">{{ t('userSubscriptions.resetHistory.failedToLoad') }}</p>
      <p v-if="error !== t('userSubscriptions.resetHistory.failedToLoad')" class="break-words text-sm text-gray-500 dark:text-dark-400">{{ error }}</p>
      <button type="button" class="btn btn-secondary" data-test="retry-history" @click="loadHistory(page)">
        {{ t('userSubscriptions.resetHistory.retry') }}
      </button>
    </div>
    <p v-else-if="items.length === 0" role="status" class="py-10 text-center text-gray-500 dark:text-dark-400">
      {{ t('userSubscriptions.resetHistory.empty') }}
    </p>
    <div v-else class="overflow-x-auto">
      <table class="w-full text-left text-sm">
        <thead class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
          <tr>
            <th scope="col" class="px-3 py-3 font-medium">{{ t('userSubscriptions.resetHistory.time') }}</th>
            <th scope="col" class="px-3 py-3 font-medium">{{ t('userSubscriptions.resetHistory.method') }}</th>
            <th scope="col" class="px-3 py-3 font-medium">{{ t('userSubscriptions.resetHistory.scope') }}</th>
            <th scope="col" class="px-3 py-3 font-medium">{{ t('userSubscriptions.resetHistory.term') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 text-gray-900 dark:divide-dark-700 dark:text-dark-100">
          <tr v-for="item in items" :key="`${item.type}-${item.id}`">
            <td class="px-3 py-3 tabular-nums"><time :datetime="item.reset_at">{{ formatDateTime(item.reset_at) }}</time></td>
            <td class="px-3 py-3">{{ t(`userSubscriptions.resetHistory.${item.type}`) }}</td>
            <td class="px-3 py-3">{{ resetScope(item) }}</td>
            <td class="px-3 py-3">
              {{ item.type === 'reset_card'
                ? t('userSubscriptions.resetHistory.cardTerm', { days: item.validity_days })
                : t('userSubscriptions.resetHistory.unchangedTerm') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <template #footer>
      <Pagination
        v-if="!loading && !error && total > 0"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :show-page-size-selector="false"
        @update:page="loadHistory"
      />
      <div class="mt-3 flex justify-end">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.close') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import subscriptionsAPI from '@/api/subscriptions'
import adminSubscriptionsAPI from '@/api/admin/subscriptions'
import type { SubscriptionResetHistory } from '@/types'
import { formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = withDefaults(defineProps<{
  subscriptionId: number | null
  subscriptionName?: string
  admin?: boolean
}>(), { subscriptionName: '', admin: false })
const emit = defineEmits<{ (event: 'close'): void }>()
const { t } = useI18n()
const items = ref<SubscriptionResetHistory[]>([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const loading = ref(false)
const error = ref('')
let controller: AbortController | undefined

function resetScope(item: SubscriptionResetHistory): string {
  return (['daily', 'weekly', 'monthly'] as const)
    .filter(window => item[window])
    .map(window => t(`userSubscriptions.${window}`))
    .join(' / ')
}

async function loadHistory(nextPage = 1) {
  controller?.abort()
  items.value = []
  total.value = 0
  error.value = ''
  page.value = nextPage
  const id = props.subscriptionId
  if (id === null) {
    loading.value = false
    return
  }
  const request = new AbortController()
  controller = request
  loading.value = true
  try {
    const api = props.admin ? adminSubscriptionsAPI : subscriptionsAPI
    const data = await api.getResetHistory(id, nextPage, pageSize, request.signal)
    if (request.signal.aborted) return
    items.value = data.items
    total.value = data.total
  } catch (cause) {
    if (request.signal.aborted) return
    error.value = extractApiErrorMessage(cause, t('userSubscriptions.resetHistory.failedToLoad'))
  } finally {
    if (!request.signal.aborted) loading.value = false
  }
}

watch(() => [props.subscriptionId, props.admin], () => { void loadHistory(1) }, { immediate: true })
onUnmounted(() => controller?.abort())
</script>
