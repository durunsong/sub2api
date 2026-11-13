<template>
  <BaseDialog :show="show" :title="t('keys.viewAvailableModels')" width="wide" @close="close">
    <div class="space-y-4">
      <h4 class="break-words text-base font-medium text-gray-800 dark:text-dark-100">
        {{ userFacingPlatformText(apiKey?.group?.name) || t('keys.noGroup') }}
      </h4>
      <section class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900/50" :aria-busy="loading">
        <div class="mb-4 flex flex-wrap items-baseline gap-x-3 gap-y-1">
          <h5 class="text-sm font-semibold text-gray-700 dark:text-dark-200">
            {{ t('keys.availableModels.modelName') }}
          </h5>
          <p class="text-sm text-gray-500 dark:text-dark-400">
            {{ t('keys.availableModels.hint') }}
          </p>
        </div>
        <p v-if="apiKey?.group_id == null" role="status" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('keys.availableModels.noGroup') }}
        </p>
        <p v-else-if="loading" role="status" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('common.loading') }}
        </p>
        <div v-else-if="failed" role="alert" class="space-y-3 py-4 text-center">
          <p class="text-sm text-red-600 dark:text-red-400">{{ t('keys.availableModels.failed') }}</p>
          <button type="button" class="btn btn-secondary min-h-11" data-testid="retry-models" @click="loadModels">
            {{ t('keys.availableModels.retry') }}
          </button>
        </div>
        <p v-else-if="models.length === 0" role="status" class="py-6 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('keys.availableModels.empty') }}
        </p>
        <ul v-else class="flex flex-wrap gap-2">
          <li v-for="model in models" :key="model" class="max-w-full">
            <button
              type="button"
              data-testid="model-name"
              class="flex min-h-11 max-w-full items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-left text-sm text-gray-700 transition-colors hover:border-primary-300 hover:bg-primary-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-200 dark:hover:border-primary-500 dark:hover:bg-dark-700"
              :aria-label="t('keys.availableModels.copy', { model })"
              @click="copyToClipboard(model, t('keys.availableModels.copied'))"
            >
              <span class="min-w-0 font-mono [overflow-wrap:anywhere]">{{ model }}</span>
              <Icon name="copy" size="sm" class="shrink-0" aria-hidden="true" />
            </button>
          </li>
        </ul>
      </section>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import { listKeyModels } from '@/api/keyModels'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { userFacingPlatformText } from '@/utils/platformColors'

const props = defineProps<{ show: boolean; apiKey: ApiKey | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const models = ref<string[]>([])
const loading = ref(false)
const failed = ref(false)
let controller: AbortController | undefined

async function loadModels() {
  controller?.abort()
  models.value = []
  failed.value = false
  loading.value = false
  const key = props.apiKey
  if (!props.show || !key || key.group_id == null) return

  const request = new AbortController()
  controller = request
  loading.value = true
  try {
    const result = await listKeyModels(key.key, request.signal)
    if (!request.signal.aborted) models.value = result
  } catch {
    if (!request.signal.aborted) failed.value = true
  } finally {
    if (!request.signal.aborted) loading.value = false
  }
}

function close() {
  controller?.abort()
  emit('close')
}

watch(
  () => [props.show, props.apiKey?.key, props.apiKey?.group_id, props.apiKey?.group?.platform],
  loadModels,
  { immediate: true }
)
onBeforeUnmount(() => controller?.abort())
</script>
