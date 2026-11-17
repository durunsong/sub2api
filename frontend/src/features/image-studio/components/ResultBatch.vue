<script setup lang="ts">
import { vTooltip } from '@/components/common/tooltip'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import BlobImage from './BlobImage.vue'
import GenerationPending from './GenerationPending.vue'
import { downloadImage } from '../useImageStudio'
import { prepareImageDownload } from '../download'
import type { ResultBatch } from '../batches'

const props = defineProps<{ group: ResultBatch; elapsed?: string }>()
const emit = defineEmits<{ preview: [images: Blob[], captions: string[], index: number]; notice: [message: string]; dialog: [open: boolean] }>()
const { t } = useI18n()
const open = ref(false), choosing = ref(false), downloading = ref(false)
const selected = ref<string[]>([])
const trigger = ref<HTMLButtonElement>()
watch(choosing, value => emit('dialog', value))
let alive = true
onBeforeUnmount(() => { alive = false; emit('dialog', false) })
const items = computed(() => props.group.entries.map(entry => (entry.history || entry.task)!))
const images = computed(() => props.group.entries.flatMap(entry => {
  const item = (entry.history || entry.task)!
  return (item.images || []).map((blob, index) => ({ key: `${entry.id}:${index}`, blob, caption: item.snapshot.prompt }))
}))
const running = computed(() => items.value.filter(item => ['running', 'queued'].includes(item.status)).length)
const failed = computed(() => items.value.filter(item => ['failed', 'cancelled', 'interrupted'].includes(item.status)).length)
const unsaved = computed(() => props.group.entries.some(entry => entry.task?.images?.length && !entry.task.saved))
const title = computed(() => items.value[0]?.snapshot.prompt || '')
const fan = computed(() => images.value.slice(0, 5))
watch(images, value => {
  selected.value = selected.value.filter(key => value.some(image => image.key === key))
  if (!value.length) choosing.value = false
})
function choose() { selected.value = images.value.map(image => image.key); choosing.value = true }
async function download() {
  if (downloading.value) return
  const selection = images.value.filter(image => selected.value.includes(image.key))
  if (!selection.length) return
  downloading.value = true
  try {
    const result = await prepareImageDownload(selection.map(image => image.blob))
    if (alive) { downloadImage(result.blob, result.name); choosing.value = false }
  } catch { if (alive) emit('notice', t('imageStudio.downloadFailed')) }
  finally { if (alive) downloading.value = false }
}
async function collapse() { open.value = false; await nextTick(); trigger.value?.focus() }
function inspect(index: number) { emit('preview', images.value.map(item => item.blob), images.value.map(item => item.caption), index) }
</script>

<template>
  <slot v-if="!group.grouped" :entries="group.entries" />
  <section v-else class="result-batch" :class="{ 'batch-expanded': open }" data-result-batch :aria-label="t('imageStudio.batchCurrentResults')">
    <div class="batch-overview">
      <button v-tooltip ref="trigger" type="button" class="batch-fan" :class="{ 'fan-open': open }" :aria-label="t(open ? 'imageStudio.collapseBatch' : 'imageStudio.expandBatch')" :aria-expanded="open" @click="open = !open">
        <template v-if="fan.length">
          <span v-if="fan.length === 1" class="fan-sheet fan-placeholder" aria-hidden="true" />
          <span v-for="(image, index) in fan" :key="image.key" class="fan-sheet" :style="{ '--fan-offset': index - (fan.length - 1) / 2, '--fan-z': index + 1 }"><BlobImage :blob="image.blob" alt="" class="h-full w-full object-cover" :eager="index === fan.length - 1" /></span>
        </template>
        <GenerationPending v-else-if="running" class="batch-pending" :status="items.some(item => item.status === 'running') ? 'running' : 'queued'" :elapsed="elapsed || ''" :title="t('imageStudio.batchPending', { count: running })" :subtitle="t('imageStudio.batchOpenHint')" :count-label="t('imageStudio.elapsedTime')" />
        <span v-else class="batch-empty"><Icon name="exclamationCircle" size="lg" />{{ t('imageStudio.batchFailed', { count: failed }) }}</span>
        <span v-if="images.length" class="fan-count"><Icon name="grid" size="xs" />{{ t('imageStudio.batchImages', { count: images.length }) }}</span>
      </button>
      <div class="batch-info">
        <p class="line-clamp-2 whitespace-pre-wrap break-words text-sm font-medium leading-6">{{ title }}</p>
        <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-xs" aria-live="polite">
          <span v-if="running" class="text-primary-700 dark:text-primary-300">{{ t('imageStudio.batchPending', { count: running }) }}</span>
          <span v-if="failed" class="text-red-700 dark:text-red-300">{{ t('imageStudio.batchFailed', { count: failed }) }}</span>
          <span v-if="unsaved" class="text-amber-800 dark:text-amber-200">{{ t('imageStudio.notSaved') }}</span>
          <span v-if="!running && !failed && !unsaved" class="text-gray-500 dark:text-dark-300">{{ items[0]?.snapshot.modelLabel || items[0]?.snapshot.model }}</span>
        </div>
        <div class="mt-4 flex flex-wrap gap-2">
          <button type="button" class="btn btn-primary btn-sm" :aria-expanded="open" @click="open = !open"><Icon :name="open ? 'chevronUp' : 'grid'" size="sm" />{{ t(open ? 'imageStudio.collapseBatch' : 'imageStudio.expandBatch') }}</button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="!images.length" @click="choose"><Icon name="download" size="sm" />{{ t('imageStudio.batchDownload') }}</button>
        </div>
      </div>
    </div>
    <div v-if="open" class="batch-entries" data-batch-entries><slot :entries="group.entries" /></div>
    <button v-if="open" type="button" class="batch-collapse text-sm font-medium text-primary-700 dark:text-primary-300" @click="collapse">{{ t('imageStudio.collapseBatch') }}<Icon name="chevronUp" size="sm" /></button>
    <BaseDialog close-tooltip :show="choosing" :title="t('imageStudio.batchDownload')" width="wide" @close="choosing = false">
      <div class="mb-4 flex flex-wrap items-center gap-3 text-sm">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="downloading" @click="selected = images.map(image => image.key)">{{ t('imageStudio.selectVisible') }}</button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="downloading" @click="selected = []">{{ t('imageStudio.clearSelection') }}</button>
        <span>{{ t('imageStudio.selectedCount', { count: selected.length }) }}</span>
      </div>
      <div class="download-grid">
        <div v-for="(image, index) in images" :key="image.key" class="relative">
          <button v-tooltip type="button" class="download-preview" :aria-label="t('imageStudio.previewReference', { number: index + 1 })" @click="inspect(index)"><BlobImage :blob="image.blob" :alt="t('imageStudio.batchSequence', { number: index + 1 })" class="aspect-square w-full rounded-lg object-cover" /></button>
          <label class="download-choice"><input v-model="selected" data-download-choice type="checkbox" :value="image.key" :disabled="downloading" :aria-label="t('imageStudio.batchSequence', { number: index + 1 })" class="h-5 w-5 accent-primary-700" /><span>{{ index + 1 }}</span></label>
        </div>
      </div>
      <p class="mt-4 text-sm text-gray-600 dark:text-dark-300">{{ t('imageStudio.downloadArchiveHint') }}</p>
      <template #footer><button type="button" data-download-selected class="btn btn-primary" :disabled="!selected.length || downloading" @click="download">{{ downloading ? t('imageStudio.downloadPreparing') : t('imageStudio.downloadSelection', { count: selected.length }) }}</button></template>
    </BaseDialog>
  </section>
</template>

<style scoped>
.result-batch { @apply min-w-0 rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900; }
.batch-expanded { grid-column: 1 / -1; }
.batch-overview { @apply min-w-0; }
.batch-fan { position: relative; display: block; width: 100%; height: 250px; overflow: hidden; border-radius: 12px 12px 0 0; background: #f3f6f5; isolation: isolate; }
.dark .batch-fan { background: #171e24; }
.fan-sheet { --fan-offset: 0; --fan-z: 1; position: absolute; left: 20%; top: 15%; width: 60%; height: 66%; overflow: hidden; border: 4px solid white; border-radius: 9px; box-shadow: 0 8px 18px rgb(20 40 35 / 16%); transform-origin: 50% 105%; transform: translateX(calc(var(--fan-offset) * 15px)) rotate(calc(var(--fan-offset) * 6deg)); z-index: var(--fan-z); transition: transform 400ms cubic-bezier(.16,1,.3,1), box-shadow 400ms; }
.fan-placeholder { background: #d7e5df; transform: translateX(-10px) rotate(-7deg); }
.batch-fan:hover .fan-sheet, .batch-fan:focus-visible .fan-sheet, .fan-open .fan-sheet { transform: translateX(calc(var(--fan-offset) * 28px)) rotate(calc(var(--fan-offset) * 11deg)); box-shadow: 0 12px 24px rgb(20 40 35 / 18%); }
.batch-fan:focus-visible { outline: 2px solid var(--color-primary-500, #0d9488); outline-offset: -3px; }
.fan-count { @apply absolute bottom-3 left-1/2 z-10 flex -translate-x-1/2 items-center gap-1.5 whitespace-nowrap rounded-full bg-gray-900 px-3 py-1.5 text-xs font-medium text-white; }
.batch-info { @apply min-w-0 p-4; }
.batch-empty { @apply flex h-full flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-dark-300; }
.batch-pending { width: 100%; height: 100%; }
.batch-entries { @apply grid min-w-0 gap-4 border-t border-gray-100 p-4 dark:border-dark-700; grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr)); }
.batch-collapse { @apply flex min-h-11 w-full items-center justify-center gap-2 border-t border-gray-100 dark:border-dark-700; }
.download-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 12px; }
.download-preview { @apply block w-full cursor-zoom-in; }
.download-choice { @apply absolute left-2 top-2 flex min-h-9 items-center gap-2 rounded-md bg-white px-2 text-xs font-semibold text-gray-900 shadow-sm dark:bg-dark-900 dark:text-white; }
@media (min-width: 768px) { .batch-expanded .batch-overview { display: grid; grid-template-columns: minmax(200px, 300px) 1fr; align-items: center; } .batch-expanded .batch-fan { height: 210px; } }
@media (prefers-reduced-motion: reduce) { .fan-sheet { transition: none; } }
</style>
