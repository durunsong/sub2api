<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import BlobImage from './BlobImage.vue'
import type { MaskEditResult } from '../types'

const props = defineProps<{ show: boolean; blob: Blob; nativeMask?: boolean }>()
const emit = defineEmits<{ close: []; confirm: [payload: MaskEditResult] }>()
const { t } = useI18n()
const canvas = ref<HTMLCanvasElement>()
const source = ref<HTMLImageElement>()
const prompt = ref('')
const brushSize = ref(100)
const mode = ref<'paint' | 'erase'>('paint')
const error = ref('')
const busy = ref(false)
const cursor = ref<{ x: number; y: number } | null>(null)
let revision = 0
let drawing = false
let last: { x: number; y: number } | null = null

watch(() => [props.show, props.blob], () => {
  revision++
  prompt.value = ''
  error.value = ''
  brushSize.value = 100
  mode.value = 'paint'
  source.value = undefined
  busy.value = false
  cursor.value = null
  void nextTick(reset)
})
onBeforeUnmount(() => { revision++ })

function loaded(image: HTMLImageElement) {
  source.value = image
  if (canvas.value) {
    canvas.value.width = image.naturalWidth
    canvas.value.height = image.naturalHeight
  }
}
function reset() {
  const el = canvas.value
  if (el) el.getContext('2d')?.clearRect(0, 0, el.width, el.height)
  drawing = false
  last = null
  error.value = ''
}
function draw(event: PointerEvent) {
  const el = canvas.value
  const ctx = el?.getContext('2d')
  if (!el || !ctx) return
  const bounds = el.getBoundingClientRect()
  if (!bounds.width || !bounds.height) return
  cursor.value = event.pointerType === 'touch' ? null : { x: event.clientX - bounds.left, y: event.clientY - bounds.top }
  if (!drawing || busy.value) return
  const point = { x: (event.clientX - bounds.left) * el.width / bounds.width, y: (event.clientY - bounds.top) * el.height / bounds.height }
  ctx.globalCompositeOperation = mode.value === 'paint' ? 'source-over' : 'destination-out'
  ctx.strokeStyle = '#2563eb'
  ctx.fillStyle = '#2563eb'
  const width = brushSize.value * el.width / bounds.width
  ctx.lineWidth = width
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.beginPath()
  if (last) {
    ctx.moveTo(last.x, last.y)
    ctx.lineTo(point.x, point.y)
    ctx.stroke()
  } else {
    ctx.arc(point.x, point.y, width / 2, 0, Math.PI * 2)
    ctx.fill()
  }
  last = point
  error.value = ''
}
function start(event: PointerEvent) {
  if (!source.value || busy.value || event.button !== 0) return
  canvas.value?.setPointerCapture(event.pointerId)
  drawing = true
  last = null
  draw(event)
}
function stop() { drawing = false; last = null }

async function submit() {
  if (busy.value) return
  const el = canvas.value
  if (!prompt.value.trim() || !el || !source.value) {
    error.value = 'imageStudio.maskRequirements'
    return
  }
  const selection = el.getContext('2d')?.getImageData(0, 0, el.width, el.height)
  // Erasing an antialiased brush can leave faint edge pixels; they are not a selection.
  if (!selection?.data.some((value, index) => index % 4 === 3 && value >= 128)) {
    error.value = 'imageStudio.maskRequirements'
    return
  }
  busy.value = true
  const currentRevision = revision
  const original = props.blob, instructions = prompt.value.trim()
  try {
    const output = document.createElement('canvas')
    output.width = el.width
    output.height = el.height
    const context = output.getContext('2d')
    if (!context) throw new Error('canvas')
    for (let index = 3; index < selection.data.length; index += 4) {
      const selected = selection.data[index] >= 128
      selection.data[index] = props.nativeMask ? (selected ? 0 : 255) : (selected ? 255 : 0)
      if (props.nativeMask) selection.data[index - 3] = selection.data[index - 2] = selection.data[index - 1] = 255
    }
    context.putImageData(selection, 0, 0)
    if (!props.nativeMask) {
      context.globalCompositeOperation = 'destination-in'
      context.fillStyle = 'rgba(255,255,255,0.38)'
      context.fillRect(0, 0, output.width, output.height)
      context.globalCompositeOperation = 'destination-over'
      context.drawImage(source.value, 0, 0)
    }
    const blob = await new Promise<Blob>((resolve, reject) => output.toBlob((value) => value ? resolve(value) : reject(new Error('blob')), 'image/png'))
    if (currentRevision !== revision || !props.show) return
    if (blob.size > 10 * 1024 * 1024 || original.size > 10 * 1024 * 1024) { error.value = 'imageStudio.maskTooLarge'; return }
    emit('confirm', { original, ...(props.nativeMask ? { mask: blob } : { guide: blob }), prompt: instructions })
  } catch {
    if (currentRevision === revision) error.value = 'imageStudio.maskFailed'
  } finally {
    if (currentRevision === revision) busy.value = false
  }
}
</script>

<template>
  <BaseDialog :show="show" :title="t('imageStudio.maskTitle')" close-tooltip width="extra-wide" :close-on-escape="!busy" :show-close-button="!busy" @close="emit('close')">
    <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div class="flex min-h-64 items-center justify-center overflow-hidden rounded-xl bg-gray-100 dark:bg-dark-900">
        <div class="relative max-w-full">
          <BlobImage :blob="blob" :alt="t('imageStudio.originalImage')" eager class="block max-h-[65vh] max-w-full object-contain" @load="loaded" @error="error = 'imageStudio.maskFailed'" />
          <canvas ref="canvas" class="absolute inset-0 h-full w-full touch-none cursor-crosshair opacity-[0.38]" :aria-label="t('imageStudio.maskCanvas')" @pointerdown.prevent="start" @pointermove.prevent="draw" @pointerup="stop" @pointercancel="stop" @lostpointercapture="stop" @pointerleave="cursor = null" />
          <span v-if="cursor && !busy" data-brush-cursor aria-hidden="true" class="pointer-events-none absolute rounded-full border-2 border-white shadow-[0_0_0_1px_#0f172a]" :style="{ width: `${brushSize}px`, height: `${brushSize}px`, left: `${cursor.x - brushSize / 2}px`, top: `${cursor.y - brushSize / 2}px` }" />
        </div>
      </div>
      <div class="space-y-5">
        <div class="space-y-2"><p class="text-sm font-semibold">{{ t(nativeMask ? 'imageStudio.nativeMaskMode' : 'imageStudio.guidedEditMode') }}</p><p class="text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t(nativeMask ? 'imageStudio.maskHelp' : 'imageStudio.guidedEditHelp') }}</p></div>
        <div class="flex gap-2">
          <button class="btn flex-1" :disabled="busy" :class="mode === 'paint' ? 'btn-primary' : 'btn-secondary'" :aria-pressed="mode === 'paint'" @click="mode = 'paint'">{{ t('imageStudio.paint') }}</button>
          <button class="btn flex-1" :disabled="busy" :class="mode === 'erase' ? 'btn-primary' : 'btn-secondary'" :aria-pressed="mode === 'erase'" @click="mode = 'erase'">{{ t('imageStudio.erase') }}</button>
        </div>
        <p class="text-xs leading-5 text-gray-600 dark:text-dark-300">{{ t('imageStudio.eraseHelp') }}</p>
        <label class="block text-sm font-medium">
          {{ t('imageStudio.brushSize') }} <span class="float-right tabular-nums">{{ brushSize }}px</span>
          <input v-model.number="brushSize" :disabled="busy" type="range" min="8" max="160" step="2" class="mt-3 w-full accent-primary-600" />
        </label>
        <label class="block text-sm font-medium">{{ t('imageStudio.editInstructions') }}
          <textarea v-model="prompt" :disabled="busy" rows="6" maxlength="10000" class="input mt-2" :placeholder="t('imageStudio.editPlaceholder')" />
        </label>
        <p v-if="error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ t(error) }}</p>
        <button class="btn btn-secondary w-full" :disabled="busy" @click="reset">{{ t('imageStudio.resetMask') }}</button>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" :disabled="busy" @click="emit('close')">{{ t('imageStudio.cancel') }}</button>
        <button data-testid="mask-submit" class="btn btn-primary" :disabled="busy" @click="submit">{{ busy ? t('imageStudio.processing') : t('imageStudio.applyEdit') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>
