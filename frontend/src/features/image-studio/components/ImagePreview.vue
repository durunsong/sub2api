<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import BlobImage from './BlobImage.vue'

const props = withDefaults(defineProps<{
  show: boolean
  images: Blob[]
  captions?: string[]
  initialIndex?: number
}>(), { captions: () => [], initialIndex: 0 })
const emit = defineEmits<{ close: []; download: [blob: Blob] }>()
const { t } = useI18n()
const root = ref<HTMLElement>()
const stage = ref<HTMLElement>()
const index = ref(0)
const zoom = ref(1)
const natural = reactive({ width: 0, height: 0 })
const viewport = reactive({ width: 0, height: 0 })
const pan = reactive({ x: 0, y: 0 })
const loading = ref(true)
const failed = ref(false)
const retryKey = ref(0)
const pointers = new Map<number, { x: number; y: number }>()
let observer: ResizeObserver | undefined
const current = computed(() => props.images[index.value])
const caption = computed(() => props.captions[index.value] || '')
const fitScale = computed(() => natural.width && viewport.width ? Math.min(viewport.width / natural.width, viewport.height / natural.height, 1) : 1)
const ready = computed(() => !loading.value && !failed.value && Boolean(current.value))
const imageStyle = computed(() => ({
  width: `${natural.width}px`, height: `${natural.height}px`,
  transform: `translate(-50%, -50%) translate(${pan.x}px, ${pan.y}px) scale(${zoom.value})`,
  visibility: ready.value ? 'visible' as const : 'hidden' as const
}))
const format = computed(() => current.value?.type.split('/')[1]?.toUpperCase() || '')
const fileSize = computed(() => !current.value ? '' : current.value.size >= 1024 * 1024 ? `${(current.value.size / 1024 / 1024).toFixed(1)} MB` : `${Math.max(1, Math.round(current.value.size / 1024))} KB`)

function clampPan() {
  const x = Math.max(0, (natural.width * zoom.value - viewport.width) / 2)
  const y = Math.max(0, (natural.height * zoom.value - viewport.height) / 2)
  pan.x = Math.max(-x, Math.min(x, pan.x))
  pan.y = Math.max(-y, Math.min(y, pan.y))
}
function fit() { zoom.value = fitScale.value; pan.x = 0; pan.y = 0 }
function measure() {
  const wasFitted = Math.abs(zoom.value - fitScale.value) < 0.001
  const rect = stage.value?.getBoundingClientRect()
  if (!rect) return
  viewport.width = stage.value?.clientWidth || rect.width
  viewport.height = stage.value?.clientHeight || rect.height
  if (wasFitted) fit()
  else clampPan()
}
function resetImage() {
  loading.value = true; failed.value = false
  natural.width = 0; natural.height = 0
  zoom.value = 1; pan.x = 0; pan.y = 0
  pointers.clear()
}
function loaded(image: HTMLImageElement) {
  // Ignore an old load event after switching to another local image.
  if (!stage.value?.contains(image)) return
  natural.width = image.naturalWidth; natural.height = image.naturalHeight
  loading.value = false
  measure(); fit()
}
function loadFailed() { loading.value = false; failed.value = true }
function retry() { resetImage(); retryKey.value++ }
function select(next: number) {
  if (next < 0 || next >= props.images.length || next === index.value) return
  index.value = next
}
function zoomTo(value: number, point?: { x: number; y: number }) {
  if (!ready.value) return
  const next = Math.max(Math.min(0.1, fitScale.value), Math.min(8, value))
  if (point && stage.value) {
    const rect = stage.value.getBoundingClientRect()
    const x = point.x - rect.left - rect.width / 2
    const y = point.y - rect.top - rect.height / 2
    pan.x = x - (x - pan.x) * next / zoom.value
    pan.y = y - (y - pan.y) * next / zoom.value
  }
  zoom.value = next
  clampPan()
}
function wheel(event: WheelEvent) { zoomTo(zoom.value * Math.exp(-event.deltaY * 0.002), { x: event.clientX, y: event.clientY }) }
function pointerDown(event: PointerEvent) {
  if (!ready.value || (event.pointerType === 'mouse' && event.button !== 0)) return
  event.preventDefault()
  stage.value?.focus({ preventScroll: true })
  pointers.set(event.pointerId, { x: event.clientX, y: event.clientY })
  stage.value?.setPointerCapture(event.pointerId)
}
function pointerMove(event: PointerEvent) {
  const before = pointers.get(event.pointerId)
  if (!before || !ready.value) return
  const oldPoints = [...pointers.values()]
  const after = { x: event.clientX, y: event.clientY }
  pointers.set(event.pointerId, after)
  const newPoints = [...pointers.values()]
  if (oldPoints.length === 2) {
    const distance = (points: typeof oldPoints) => Math.hypot(points[0].x - points[1].x, points[0].y - points[1].y)
    const previousDistance = distance(oldPoints)
    if (previousDistance > 0 && stage.value) {
      const rect = stage.value.getBoundingClientRect()
      const oldX = (oldPoints[0].x + oldPoints[1].x) / 2 - rect.left - rect.width / 2
      const oldY = (oldPoints[0].y + oldPoints[1].y) / 2 - rect.top - rect.height / 2
      const imageX = (oldX - pan.x) / zoom.value
      const imageY = (oldY - pan.y) / zoom.value
      zoomTo(zoom.value * distance(newPoints) / previousDistance)
      pan.x = (newPoints[0].x + newPoints[1].x) / 2 - rect.left - rect.width / 2 - imageX * zoom.value
      pan.y = (newPoints[0].y + newPoints[1].y) / 2 - rect.top - rect.height / 2 - imageY * zoom.value
      clampPan()
    }
  } else if (oldPoints.length === 1) {
    pan.x += after.x - before.x; pan.y += after.y - before.y
    clampPan()
  }
}
function pointerEnd(event: PointerEvent) { pointers.delete(event.pointerId) }
function keydown(event: KeyboardEvent) {
  if (!props.show || event.isComposing) return
  if (event.key === 'Tab') {
    const panel = root.value?.closest('.modal-content') || root.value
    const elements = [...(panel?.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), summary, [tabindex="0"]') || [])].filter(element => {
      for (let ancestor: HTMLElement | null = element; ancestor && ancestor !== panel; ancestor = ancestor.parentElement) {
        const style = getComputedStyle(ancestor)
        if (style.display === 'none' || style.visibility === 'hidden') return false
      }
      return true
    })
    const first = elements[0], last = elements[elements.length - 1]
    if (first && (event.shiftKey ? document.activeElement === first : document.activeElement === last)) {
      event.preventDefault(); (event.shiftKey ? last : first).focus()
    }
    return
  }
  const target = event.target instanceof Element ? event.target : null
  if (target?.closest('input, textarea, [contenteditable="true"]') || event.ctrlKey || event.metaKey || event.altKey) return
  if (event.shiftKey && event.key.startsWith('Arrow') && ready.value) {
    event.preventDefault()
    pan.x += event.key === 'ArrowLeft' ? 40 : event.key === 'ArrowRight' ? -40 : 0
    pan.y += event.key === 'ArrowUp' ? 40 : event.key === 'ArrowDown' ? -40 : 0
    clampPan()
    return
  }
  const actions: Record<string, () => void> = {
    ArrowLeft: () => select(index.value - 1), ArrowRight: () => select(index.value + 1),
    '+': () => zoomTo(zoom.value * 1.25), '=': () => zoomTo(zoom.value * 1.25),
    '-': () => zoomTo(zoom.value / 1.25), '0': fit, '1': () => zoomTo(1)
  }
  if (actions[event.key]) { event.preventDefault(); actions[event.key]() }
}
watch(current, resetImage, { flush: 'sync' })
watch(() => [props.show, props.images, props.initialIndex] as const, async () => {
  observer?.disconnect()
  if (!props.show) { pointers.clear(); return }
  index.value = Math.min(Math.max(0, props.initialIndex), Math.max(0, props.images.length - 1))
  resetImage()
  await nextTick()
  if (!props.show || !stage.value) return
  measure()
  if (typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver(measure)
    observer.observe(stage.value)
  }
}, { immediate: true })
onMounted(() => {
  document.addEventListener('keydown', keydown, true)
  window.addEventListener('resize', measure)
})
onBeforeUnmount(() => {
  observer?.disconnect()
  document.removeEventListener('keydown', keydown, true)
  window.removeEventListener('resize', measure)
})
</script>

<template>
  <BaseDialog :show="show" :title="t('imageStudio.previewImage')" width="full" :z-index="70" :show-close-button="false" close-on-click-outside @close="emit('close')">
    <div v-if="show" ref="root" data-image-preview class="image-preview">
      <div class="preview-toolbar">
        <button type="button" class="preview-control order-last" :aria-label="t('imageStudio.closePreview')" @click="emit('close')"><Icon name="x" size="md" /></button>
        <div class="flex items-center gap-1">
          <button type="button" class="preview-control" :disabled="index === 0" :aria-label="t('imageStudio.previousImage')" @click="select(index - 1)"><Icon name="chevronLeft" size="md" /></button>
          <span class="min-w-12 text-center text-xs tabular-nums" role="status" aria-live="polite">{{ images.length ? index + 1 : 0 }} / {{ images.length }}</span>
          <button type="button" class="preview-control" :disabled="index >= images.length - 1" :aria-label="t('imageStudio.nextImage')" @click="select(index + 1)"><Icon name="chevronRight" size="md" /></button>
        </div>
        <button type="button" class="preview-control ml-auto gap-2 px-3" :disabled="!current" @click="current && emit('download', current)"><Icon name="download" size="sm" /><span>{{ t('imageStudio.download') }}</span></button>
      </div>
      <div class="preview-zoom">
        <button type="button" class="preview-control" :disabled="!ready || zoom <= Math.min(0.1, fitScale)" :aria-label="t('imageStudio.zoomOut')" @click="zoomTo(zoom / 1.25)"><span aria-hidden="true" class="text-lg">−</span></button>
        <input class="hidden min-w-16 flex-1 accent-primary-600 sm:block" type="range" :aria-label="t('imageStudio.zoom')" :min="Math.min(0.1, fitScale)" max="8" step="0.01" :value="zoom" :disabled="!ready" @input="zoomTo(Number(($event.target as HTMLInputElement).value))" />
        <output class="w-12 text-center text-xs tabular-nums">{{ Math.round(zoom * 100) }}%</output>
        <button type="button" class="preview-control" :disabled="!ready || zoom >= 8" :aria-label="t('imageStudio.zoomIn')" @click="zoomTo(zoom * 1.25)"><Icon name="plus" size="sm" /></button>
        <button type="button" class="preview-control ml-auto px-3" :disabled="!ready" @click="fit">{{ t('imageStudio.fitImage') }}</button>
        <button type="button" class="preview-control px-3" :disabled="!ready" @click="zoomTo(1)">{{ t('imageStudio.actualSize') }}</button>
      </div>
      <div ref="stage" data-preview-stage class="preview-stage" :class="{ 'is-zoomed': ready && zoom > fitScale + 0.001 }" tabindex="0" role="group" :aria-label="t('imageStudio.previewCanvas')" :aria-busy="loading" :data-zoom="zoom" :data-pan-x="pan.x" :data-pan-y="pan.y"
        @wheel.prevent="wheel" @dblclick="zoom > fitScale + 0.001 ? fit() : zoomTo(1)" @pointerdown="pointerDown" @pointermove="pointerMove" @pointerup="pointerEnd" @pointercancel="pointerEnd" @lostpointercapture="pointerEnd">
        <BlobImage v-if="current" :key="`${index}:${retryKey}`" :blob="current" :alt="caption || t('imageStudio.previewImage')" eager class="preview-image" :style="imageStyle" @load="loaded" @error="loadFailed" />
        <p v-if="loading" role="status" class="preview-message">{{ t('imageStudio.loadingImage') }}</p>
        <div v-else-if="failed" class="preview-message" role="alert"><p>{{ t('imageStudio.previewFailed') }}</p><button type="button" class="btn btn-secondary mt-3" @click="retry">{{ t('imageStudio.retryPreview') }}</button></div>
      </div>
      <div class="preview-info"><span class="tabular-nums">{{ ready ? `${natural.width} × ${natural.height} · ` : '' }}{{ format }} · {{ fileSize }}</span><span class="hidden md:inline">{{ t('imageStudio.previewKeyboardHint') }}</span><span class="md:hidden">{{ t('imageStudio.previewTouchHint') }}</span></div>
      <div v-if="images.length > 1" class="preview-thumbnails" :aria-label="t('imageStudio.previewGallery')">
        <button v-for="(blob, itemIndex) in images" :key="itemIndex" type="button" :aria-label="t('imageStudio.previewNumber', { number: itemIndex + 1 })" :aria-pressed="index === itemIndex" class="preview-thumbnail" :class="{ active: index === itemIndex }" @click="select(itemIndex)"><BlobImage :blob="blob" :alt="''" class="h-full w-full object-contain" /></button>
      </div>
      <details v-if="caption" class="preview-caption"><summary>{{ t('imageStudio.imageDescription') }}</summary><p>{{ caption }}</p></details>
    </div>
  </BaseDialog>
</template>

<style scoped>
.image-preview { display: flex; flex-direction: column; gap: 8px; height: min(72dvh, 800px); min-height: 220px; }
.preview-toolbar, .preview-zoom { @apply flex shrink-0 items-center gap-1 text-gray-700 dark:text-dark-100; }
.preview-control { @apply inline-flex shrink-0 items-center justify-center rounded-lg text-xs hover:bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-dark-700; min-height: 44px; min-width: 44px; }
.preview-stage { @apply relative min-h-20 flex-1 overflow-hidden rounded-lg border border-gray-200 bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 dark:border-dark-700 dark:bg-dark-950; touch-action: none; overscroll-behavior: contain; }
.preview-stage.is-zoomed { cursor: grab; }
.preview-stage.is-zoomed:active { cursor: grabbing; }
.preview-image { position: absolute; top: 50%; left: 50%; max-width: none; max-height: none; transform-origin: center; user-select: none; pointer-events: none; }
.preview-message { @apply absolute inset-0 flex flex-col items-center justify-center p-5 text-center text-sm leading-6 text-gray-600 dark:text-dark-300; }
.preview-info { @apply flex shrink-0 flex-wrap justify-between gap-x-3 gap-y-1 text-[11px] leading-5 text-gray-500 dark:text-dark-300; }
.preview-thumbnails { @apply flex shrink-0 gap-2 overflow-x-auto p-1; }
.preview-thumbnail { @apply h-12 w-12 shrink-0 overflow-hidden rounded-md border-2 border-transparent bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 dark:bg-dark-950; }
.preview-thumbnail.active { @apply border-primary-500; }
.preview-caption { @apply shrink-0 text-xs leading-5 text-gray-600 dark:text-dark-300; }
.preview-caption summary { @apply cursor-pointer rounded py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500; }
.preview-caption p { @apply mt-2 max-h-24 overflow-auto whitespace-pre-wrap break-words; }
@media (max-height: 500px) {
  .preview-thumbnails { display: none; }
}
@media (min-width: 768px) and (max-height: 500px) {
  .image-preview { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr); grid-template-rows: 44px minmax(80px, 1fr) auto; height: calc(100dvh - 130px); min-height: 180px; column-gap: 16px; }
  .preview-zoom input { display: none; }
  .preview-stage, .preview-info, .preview-caption { grid-column: 1 / -1; }
}
</style>
