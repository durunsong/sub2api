<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import ImagePreview from '@/features/image-studio/components/ImagePreview.vue'
import GenerationPending from '@/features/image-studio/components/GenerationPending.vue'
import BlobImage from '@/features/image-studio/components/BlobImage.vue'
import MaskEditor from '@/features/image-studio/components/MaskEditor.vue'
import WorkflowLibrary from '@/features/image-studio/components/WorkflowLibrary.vue'
import { downloadImage, useImageStudio, type StudioTask } from '@/features/image-studio/useImageStudio'
import { normalizeDimensions, isValidCustomSize } from '@/features/image-studio/components/customSize'
import type { ImageCategory, ImageHistory } from '@/features/image-studio/types'

const { t, te, locale } = useI18n()
const studio = reactive(useImageStudio())
const auth = useAuthStore()
const app = useAppStore()
const uploadInput = ref<HTMLInputElement>()
const importInput = ref<HTMLInputElement>()
const assetUploadInput = ref<HTMLInputElement>()
const libraryOpen = ref(false)
const assetsOpen = ref(false)
const assetSearch = ref('')
const search = ref('')
const favoritesOnly = ref(false)
const view = ref<'all' | 'category'>('all')
const selected = ref<string[]>([])
const categoryDialog = ref(false)
const categoryName = ref('')
const editingCategory = ref<ImageCategory | null>(null)
const assignmentOpen = ref(false)
const assignmentIDs = ref<string[]>([])
const assignmentCategories = ref<string[]>([])
const preview = ref<{ images: Blob[]; index: number; captions: string[] } | null>(null)
const maskImage = ref<Blob | null>(null)
const detail = ref('')
const notice = ref('')
const expanded = ref<string[]>([])
const confirmation = ref<{ title: string; message: string; danger: boolean; action: () => void | Promise<void> } | null>(null)
const customWidth = ref(1024)
const customHeight = ref(1024)
const alignDimensions = ref(true)
const customSelected = ref(false)
const referenceDragDepth = ref(0)
const submitAnchor = ref<HTMLElement>()
const originalSubmit = ref<HTMLButtonElement>()
const generationDock = ref<HTMLElement>()
const resultsElement = ref<HTMLElement>()
const submitVisible = ref(true)
const keyboardClipsSubmit = ref(false)
const keyboardInset = ref(0)
const modalOpen = computed(() => Boolean(libraryOpen.value || assetsOpen.value || assignmentOpen.value || categoryDialog.value || preview.value || maskImage.value || confirmation.value || detail.value))
const showGenerationDock = computed(() => (!submitVisible.value || keyboardClipsSubmit.value) && !modalOpen.value && !app.mobileOpen)
const generateLabel = computed(() => studio.activeCount ? t('imageStudio.continueGenerate', { count: studio.activeCount }) : t('imageStudio.generate'))
let submitObserver: IntersectionObserver | undefined
function updateVisualViewport() {
  const viewport = window.visualViewport
  keyboardInset.value = viewport ? Math.max(0, window.innerHeight - viewport.height - viewport.offsetTop) : 0
  const rect = submitAnchor.value?.getBoundingClientRect()
  keyboardClipsSubmit.value = Boolean(viewport && rect && rect.bottom > viewport.height + viewport.offsetTop)
}
onMounted(() => {
  if (typeof IntersectionObserver !== 'undefined' && submitAnchor.value) {
    submitObserver = new IntersectionObserver(([entry]) => {
      submitVisible.value = entry.isIntersecting && entry.intersectionRatio >= 0.99
      updateVisualViewport()
    }, { threshold: [0, 0.99, 1], rootMargin: '-80px 0px -12px 0px' })
    submitObserver.observe(submitAnchor.value)
  }
  updateVisualViewport()
  window.visualViewport?.addEventListener('resize', updateVisualViewport)
  window.visualViewport?.addEventListener('scroll', updateVisualViewport)
  window.addEventListener('scroll', updateVisualViewport, true)
})
onBeforeUnmount(() => {
  submitObserver?.disconnect()
  window.visualViewport?.removeEventListener('resize', updateVisualViewport)
  window.visualViewport?.removeEventListener('scroll', updateVisualViewport)
  window.removeEventListener('scroll', updateVisualViewport, true)
})
watch(showGenerationDock, async visible => {
  // Preserve keyboard focus when scrolling brings the original action back.
  if (!visible && !modalOpen.value && generationDock.value?.contains(document.activeElement)) {
    await nextTick()
    originalSubmit.value?.focus({ preventScroll: true })
  }
})
function editParameters() {
  document.getElementById('studio-prompt')?.focus({ preventScroll: true })
  document.getElementById('studio-prompt')?.scrollIntoView({ block: 'center', behavior: 'auto' })
}
async function generateFromDock() {
  if (!studio.canGenerate || customSizeError.value) return
  generate()
  await nextTick()
  resultsElement.value?.scrollIntoView({ block: 'start', behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' })
}
const dimensionStep = computed(() => studio.selectedModel?.size_multiple || 16)
const customSizeError = computed(() => studio.selectedModel?.supports_custom_size && /^\d+x\d+$/.test(studio.size) && !isValidCustomSize(studio.size, dimensionStep.value))
const sizeChoice = computed({
  get: () => customSelected.value || (studio.selectedModel?.supports_custom_size && !studio.selectedModel.sizes.includes(studio.size)) ? '__custom' : studio.size,
  set: (value: string | number | boolean | null) => {
    customSelected.value = value === '__custom'
    if (customSelected.value) commitDimensions()
    else studio.size = String(value || 'auto')
  }
})
const now = ref(Date.now())
const timer = window.setInterval(() => { now.value = Date.now() }, 1000)
onBeforeUnmount(() => window.clearInterval(timer))

const options = computed(() => studio.models.map(model => ({ value: model.id, label: `${model.label || model.model} · ${model.group_name}` })))
const sizes = computed(() => [...(studio.selectedModel?.sizes || []).map(value => ({ value, label: value === 'auto' ? t('imageStudio.auto') : value })), ...(studio.selectedModel?.supports_custom_size ? [{ value: '__custom', label: t('imageStudio.customSize') }] : [])])
const qualities = computed(() => (studio.selectedModel?.qualities || []).map(value => ({ value, label: te(`imageStudio.qualityLabels.${value}`) ? t(`imageStudio.qualityLabels.${value}`) : value })))
const filteredHistory = computed(() => studio.history.filter(item => {
  const matchesCategory = view.value !== 'category' || (studio.categoryID ? (item.categoryIDs || [item.categoryID]).includes(studio.categoryID) : !(item.categoryIDs?.length || item.categoryID))
  return matchesCategory && (!favoritesOnly.value || item.favorite) && `${item.snapshot.prompt} ${item.snapshot.model}`.toLowerCase().includes(search.value.trim().toLowerCase())
}))
const sessionTaskIDs = new Set<string>()
const newResultIDs = ref(new Set<string>())
watch(() => studio.tasks.map(task => ({ id: task.id, status: task.status })), tasks => {
  for (const task of tasks) {
    if (task.status === 'queued' || task.status === 'running') sessionTaskIDs.add(task.id)
    else if (['failed', 'cancelled', 'interrupted'].includes(task.status)) sessionTaskIDs.delete(task.id)
  }
})
watch(() => studio.history.map(item => item.id), ids => {
  newResultIDs.value = new Set([...newResultIDs.value].filter(id => ids.includes(id)))
  for (const id of ids) if (sessionTaskIDs.delete(id)) newResultIDs.value.add(id)
})
const runningCount = computed(() => studio.tasks.filter(task => task.status === 'running').length)
const queuedCount = computed(() => studio.tasks.filter(task => task.status === 'queued').length)
const visibleTasks = computed(() => studio.tasks.filter(task => task.status !== 'success' || !task.saved))
const filteredAssets = computed(() => studio.assets.filter(item => item.name.toLowerCase().includes(assetSearch.value.trim().toLowerCase())))
const allSelected = computed(() => filteredHistory.value.length > 0 && filteredHistory.value.every(item => selected.value.includes(item.id)))
const activeCategory = computed(() => studio.categories.find(item => item.id === studio.categoryID))
const formatError = (value: string) => te(value) ? t(value) : value
const formatDate = (value: number) => new Date(value).toLocaleString(locale.value, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
const duration = (task: StudioTask) => `${Math.max(0, Math.round(((task.completedAt || now.value) - task.createdAt) / 1000))}s`
const bytes = (blob: Blob) => blob.size > 1024 * 1024 ? `${(blob.size / 1024 / 1024).toFixed(1)} MB` : `${Math.round(blob.size / 1024)} KB`
const categoryImages = (id: string) => studio.history.filter(item => (item.categoryIDs || [item.categoryID]).includes(id)).flatMap(item => item.images).slice(0, 4)
const categoryCount = (id: string) => studio.history.filter(item => (item.categoryIDs || [item.categoryID]).includes(id)).length

watch(() => studio.history, () => { selected.value = selected.value.filter(id => studio.history.some(item => item.id === id)) })
watch(() => studio.categoryID, () => { selected.value = [] })
watch(() => auth.user?.id, () => {
  referenceDragDepth.value = 0
  sessionTaskIDs.clear()
  newResultIDs.value.clear()
  preview.value = null
  maskImage.value = null
  libraryOpen.value = false
  assetsOpen.value = false
  assignmentOpen.value = false
  categoryDialog.value = false
  confirmation.value = null
  selected.value = []
  detail.value = ''
  notice.value = ''
})
watch(() => studio.modelID, () => { customSelected.value = false })
watch(() => studio.size, (size) => {
  const match = /^(\d+)x(\d+)$/.exec(size)
  if (match) { customWidth.value = Number(match[1]); customHeight.value = Number(match[2]) }
}, { immediate: true })
function commitDimensions() {
  studio.size = normalizeDimensions(customWidth.value, customHeight.value, alignDimensions.value, dimensionStep.value)
  const [width, height] = studio.size.split('x').map(Number)
  customWidth.value = width
  customHeight.value = height
}
function generate(prompts?: string[]) { if (!customSizeError.value) void studio.submit(prompts) }

function confirm(title: string, message: string, action: () => void | Promise<void>, danger = false) {
  confirmation.value = { title: t(title), message, action, danger }
}
async function acceptConfirmation() {
  const action = confirmation.value?.action
  confirmation.value = null
  await action?.()
}
async function copyPrompt(value: string) {
  try { await navigator.clipboard.writeText(value); notice.value = t('imageStudio.copied') }
  catch { notice.value = t('imageStudio.clipboardFailed') }
}
async function pastePrompt() {
  const userID = auth.user?.id
  try { const value = await navigator.clipboard.readText(); if (auth.user?.id === userID) studio.prompt = value }
  catch { notice.value = t('imageStudio.clipboardFailed') }
}
async function pasteImages() {
  const userID = auth.user?.id
  try {
    const items = await navigator.clipboard.read()
    const blobs = await Promise.all(items.flatMap(item => item.types.filter(type => type.startsWith('image/')).map(type => item.getType(type))))
    if (auth.user?.id !== userID) return
    if (!blobs.length) { notice.value = t('imageStudio.noClipboardImage'); return }
    await studio.addReferences(blobs)
  } catch { notice.value = t('imageStudio.clipboardFailed') }
}
function transferImages(data: DataTransfer | null): File[] {
  if (!data) return []
  const images = Array.from(data.items || [])
    .filter(item => item.kind === 'file' && item.type.startsWith('image/'))
    .map(item => item.getAsFile()).filter((file): file is File => file !== null)
  return images.length ? images : Array.from(data.files || []).filter(file => file.type.startsWith('image/'))
}
function pasteReference(event: ClipboardEvent) {
  if (libraryOpen.value || assetsOpen.value || assignmentOpen.value || categoryDialog.value || preview.value || maskImage.value || confirmation.value) return
  const target = event.target instanceof Element ? event.target : null
  if (target?.closest('[role="dialog"], input, [contenteditable="true"]') || (target?.closest('textarea') && target.id !== 'studio-prompt')) return
  const images = transferImages(event.clipboardData)
  if (!images.length) return
  event.preventDefault()
  void studio.addReferences(images)
}
onMounted(() => document.addEventListener('paste', pasteReference))
onBeforeUnmount(() => document.removeEventListener('paste', pasteReference))
function dragReferences(event: DragEvent, entering = false) {
  if (!Array.from(event.dataTransfer?.types || []).includes('Files')) return
  event.preventDefault()
  if (event.dataTransfer) event.dataTransfer.dropEffect = studio.selectedModel?.max_references ? 'copy' : 'none'
  if (entering) referenceDragDepth.value++
}
async function upload(event: Event, toAssets = false) {
  const input = event.target as HTMLInputElement
  const files = [...(input.files || [])]
  const userID = auth.user?.id
  if (toAssets) { for (const file of files) { if (auth.user?.id !== userID) break; await studio.saveAsset(file, file.name) } }
  else await studio.addReferences(files)
  input.value = ''
}
async function dropped(event: DragEvent) {
  event.preventDefault()
  referenceDragDepth.value = 0
  const files = transferImages(event.dataTransfer)
  if (files.length) await studio.addReferences(files)
  else if (event.dataTransfer?.files.length) studio.error = 'imageStudio.errors.invalidImage'
}
function toggle(id: string) { selected.value = selected.value.includes(id) ? selected.value.filter(value => value !== id) : [...selected.value, id] }
function toggleAll() { selected.value = allSelected.value ? [] : filteredHistory.value.map(item => item.id) }
function removeHistory(ids: string[]) {
  confirm('imageStudio.deleteHistory', t('imageStudio.deleteHistoryConfirm', { count: ids.length }), () => studio.deleteHistory(ids), true)
}
function retry(task: StudioTask | ImageHistory) {
  confirm('imageStudio.retry', t('imageStudio.retryConfirm'), () => studio.retry(task))
}
function assign(ids: string[]) {
  assignmentIDs.value = [...ids]
  const records = studio.history.filter(item => ids.includes(item.id))
  assignmentCategories.value = studio.categories.filter(category => records.every(item => (item.categoryIDs || [item.categoryID]).includes(category.id))).map(category => category.id)
  assignmentOpen.value = true
}
function openCategory(item?: ImageCategory) { editingCategory.value = item || null; categoryName.value = item?.name || ''; categoryDialog.value = true }
async function saveCategory() {
  if (!categoryName.value.trim()) return
  await studio.saveCategory(categoryName.value.trim(), editingCategory.value?.id)
  categoryDialog.value = false
}
function downloadSelected() {
  studio.history.filter(item => selected.value.includes(item.id)).forEach(item => item.images.forEach(blob => downloadImage(blob)))
}
async function importBackup(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  confirm('imageStudio.importBackup', t('imageStudio.importConfirm'), () => studio.importBackup(file))
}
async function submitEdit(payload: { blob: Blob; prompt: string }) {
  const userID = auth.user?.id
  studio.references = []
  await studio.addReferences([payload.blob])
  if (auth.user?.id !== userID) return
  studio.prompt = t('imageStudio.maskPromptPrefix') + payload.prompt
  maskImage.value = null
  if (studio.references.length) generate([studio.prompt])
}
function enterPrompt(event: KeyboardEvent) {
  if (studio.layout === 'bottom' && event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
    event.preventDefault()
    if (studio.canGenerate) generate()
  }
}
function setPreview(images: Blob[], caption: string | string[], index = 0) {
  if (!images.length) return
  preview.value = { images: [...images], captions: typeof caption === 'string' ? images.map(() => caption) : [...caption], index }
}
function previewHistory(item: ImageHistory) {
  const images = filteredHistory.value.flatMap(record => record.images)
  const captions = filteredHistory.value.flatMap(record => record.images.map(() => record.snapshot.prompt))
  const preceding = filteredHistory.value.slice(0, filteredHistory.value.findIndex(record => record.id === item.id))
  setPreview(images, captions, preceding.reduce((count, record) => count + record.images.length, 0))
}
</script>

<template>
  <AppLayout>
    <div class="image-studio pb-28 text-gray-900 dark:text-gray-100">
      <header class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div><h1 class="text-2xl font-semibold tracking-tight">{{ t('imageStudio.title') }}</h1><p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ t('imageStudio.localNotice') }}</p></div>
        <div class="flex flex-wrap items-center gap-2">
          <button class="btn btn-secondary btn-sm" @click="libraryOpen = true"><Icon name="book" size="sm" />{{ t('imageStudio.workflows') }}</button>
          <button class="btn btn-secondary btn-sm" @click="studio.exportBackup()"><Icon name="download" size="sm" />{{ t('imageStudio.exportBackup') }}</button>
          <button class="btn btn-secondary btn-sm" @click="importInput?.click()"><Icon name="upload" size="sm" />{{ t('imageStudio.importBackup') }}</button>
          <div class="flex rounded-lg border border-gray-200 p-0.5 dark:border-dark-700" :aria-label="t('imageStudio.layout')">
            <button class="layout-button" :class="{ active: studio.layout === 'side' }" :aria-pressed="studio.layout === 'side'" @click="studio.layout = 'side'">{{ t('imageStudio.side') }}</button>
            <button class="layout-button" :class="{ active: studio.layout === 'bottom' }" :aria-pressed="studio.layout === 'bottom'" @click="studio.layout = 'bottom'">{{ t('imageStudio.bottom') }}</button>
          </div>
        </div>
      </header>
      <div v-if="studio.error || studio.storageError" role="alert" class="mb-4 flex items-start justify-between gap-3 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/30 dark:text-red-300">
        <div><p v-if="studio.error">{{ formatError(studio.error) }}</p><p v-if="studio.storageError" class="mt-1">{{ formatError(studio.storageError) }}</p></div>
        <button class="shrink-0 underline" @click="studio.refresh()">{{ t('imageStudio.refresh') }}</button>
      </div>
      <p v-if="studio.externalTasksCount" role="status" class="mb-4 text-sm text-gray-600 dark:text-dark-300">{{ t('imageStudio.externalTasksHint', { count: studio.externalTasksCount }) }}</p>
      <div v-if="notice" role="status" class="mb-3 flex items-center justify-between gap-3 text-sm text-primary-700 dark:text-primary-300"><p>{{ notice }}</p><button :aria-label="t('imageStudio.dismiss')" @click="notice = ''"><Icon name="x" size="sm" /></button></div>
      <div class="studio-layout" :class="`layout-${studio.layout}`">
        <form class="studio-composer" @submit.prevent="generate()">
          <div class="composer-main space-y-4">
            <div>
              <div class="mb-2 flex items-center justify-between gap-2"><label for="studio-prompt" class="text-sm font-semibold">{{ t('imageStudio.prompt') }}</label><button type="button" class="text-xs text-gray-500 hover:text-primary-600 dark:text-dark-300" @click="studio.prompt = ''">{{ t('imageStudio.clear') }}</button></div>
              <textarea id="studio-prompt" v-model="studio.prompt" rows="6" maxlength="32000" class="input resize-y leading-6" :placeholder="t('imageStudio.promptPlaceholder')" @keydown="enterPrompt" />
              <div class="mt-2 flex flex-wrap gap-2">
                <button type="button" class="text-button" @click="pastePrompt"><Icon name="clipboard" size="sm" />{{ t('imageStudio.paste') }}</button>
                <button type="button" class="text-button" @click="libraryOpen = true"><Icon name="book" size="sm" />{{ t('imageStudio.promptLibrary') }}</button>
                <button type="button" class="text-button" @click="assetsOpen = true"><Icon name="grid" size="sm" />{{ t('imageStudio.assets') }}</button>
              </div>
            </div>
            <section>
              <div class="mb-2 flex items-center justify-between gap-2"><h2 class="text-sm font-semibold">{{ t('imageStudio.references') }} <span class="ml-1 font-normal text-gray-500">{{ studio.references.length }}/{{ studio.selectedModel?.max_references || 0 }}</span></h2><div class="flex gap-2"><button type="button" class="text-button" @click="pasteImages">{{ t('imageStudio.paste') }}</button><button type="button" class="text-button" :disabled="!studio.selectedModel?.max_references" @click="uploadInput?.click()">{{ t('imageStudio.upload') }}</button></div></div>
              <div data-reference-dropzone :data-drag-active="referenceDragDepth > 0" tabindex="0" role="group" :aria-label="t('imageStudio.references')" aria-describedby="studio-reference-help"
                class="relative flex min-h-20 gap-2 overflow-x-auto rounded-lg border border-dashed p-2 transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500"
                :class="referenceDragDepth ? 'border-primary-500 bg-primary-50 dark:bg-primary-950' : 'border-gray-300 dark:border-dark-600'"
                @dragenter="dragReferences($event, true)" @dragover="dragReferences($event)" @dragleave="referenceDragDepth = Math.max(0, referenceDragDepth - 1)" @drop.stop="dropped">
                <div v-if="referenceDragDepth" class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-primary-50/95 p-2 text-center text-xs font-medium text-primary-700 dark:bg-primary-950/95 dark:text-primary-300">{{ t(studio.selectedModel?.max_references ? 'imageStudio.dropReferences' : 'imageStudio.noReferencesSupported') }}</div>
                <div v-for="(blob, index) in studio.references" :key="index" class="group relative h-16 w-16 shrink-0">
                  <button type="button" class="h-full w-full cursor-zoom-in" :aria-label="t('imageStudio.previewReference', { number: index + 1 })" @click="setPreview(studio.references, '', index)"><BlobImage :blob="blob" :alt="t('imageStudio.referenceNumber', { number: index + 1 })" class="h-full w-full rounded-md object-cover" /></button>
                  <button type="button" class="absolute -right-1 -top-1 rounded-full bg-gray-900 p-1 text-white" :aria-label="t('imageStudio.removeReference', { number: index + 1 })" @click="studio.removeReference(index)"><Icon name="x" size="xs" /></button>
                </div>
                <button v-if="!studio.references.length" type="button" class="w-full py-3 text-xs leading-5 text-gray-500 dark:text-dark-300" :disabled="!studio.selectedModel?.max_references" @click="uploadInput?.click()">{{ studio.selectedModel?.max_references ? t('imageStudio.referenceHint') : t('imageStudio.noReferencesSupported') }}</button>
              </div>
              <p id="studio-reference-help" class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('imageStudio.referencePasteHint') }}</p>
            </section>
          </div>
          <div class="composer-settings space-y-4">
            <div><label for="studio-model" class="mb-2 block text-sm font-semibold">{{ t('imageStudio.model') }}</label><Select id="studio-model" v-model="studio.modelID" :options="options" searchable :loading="studio.catalogLoading" :placeholder="studio.catalogLoading ? t('imageStudio.loadingModels') : t('imageStudio.selectModel')" :empty-text="t('imageStudio.noModels')" :aria-label="t('imageStudio.model')" /></div>
            <p v-if="!studio.catalogLoading && (!studio.enabled || !studio.models.length)" class="text-xs leading-5 text-amber-700 dark:text-amber-300">{{ !studio.enabled ? t('imageStudio.disabled') : t('imageStudio.noModelsHelp') }}</p>
            <div v-if="studio.selectedModel" class="space-y-1 text-xs leading-5 text-gray-500 dark:text-dark-300">
              <p v-if="studio.selectedModel.pricing.mode === 'token'">{{ t('imageStudio.tokenPricing') }}</p>
              <div v-else class="flex flex-wrap gap-x-3"><span v-for="tier in (['1k', '2k', '4k'] as const)" :key="tier">{{ tier.toUpperCase() }}: {{ studio.selectedModel.pricing[`image_price_${tier}`] == null ? t('imageStudio.priceUnavailable') : `$${studio.selectedModel.pricing[`image_price_${tier}`]} ${t('imageStudio.perImage')}` }}</span></div>
              <p>{{ t('imageStudio.billingNotice') }}</p>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div><label class="mb-2 block text-xs font-medium">{{ studio.selectedModel?.protocol === 'gemini' ? t('imageStudio.resolution') : t('imageStudio.size') }}</label><Select v-model="sizeChoice" :options="sizes" :aria-label="t('imageStudio.size')" :disabled="!sizes.length" /></div>
              <div v-if="qualities.length"><label class="mb-2 block text-xs font-medium">{{ t('imageStudio.quality') }}</label><Select v-model="studio.quality" :options="qualities" :aria-label="t('imageStudio.quality')" :disabled="!qualities.length" /></div>
            </div>
            <div v-if="studio.selectedModel?.supports_custom_size && sizeChoice === '__custom'" class="space-y-2">
              <div class="grid grid-cols-2 gap-3">
                <label class="text-xs font-medium">{{ t('imageStudio.width') }}<input v-model.number="customWidth" type="number" min="16" max="3840" class="input mt-2" :aria-invalid="Boolean(customSizeError)" @blur="commitDimensions" @keydown.enter.prevent="commitDimensions" /></label>
                <label class="text-xs font-medium">{{ t('imageStudio.height') }}<input v-model.number="customHeight" type="number" min="16" max="3840" class="input mt-2" :aria-invalid="Boolean(customSizeError)" @blur="commitDimensions" @keydown.enter.prevent="commitDimensions" /></label>
              </div>
              <label class="flex items-center gap-2 text-xs"><input v-model="alignDimensions" type="checkbox" class="accent-primary-600" @change="commitDimensions" />{{ t('imageStudio.alignDimensions', { step: dimensionStep }) }}</label>
              <p v-if="customSizeError" role="alert" class="text-xs leading-5 text-red-600 dark:text-red-400">{{ t('imageStudio.customSizeInvalid', { step: dimensionStep }) }}</p>
            </div>
            <div v-if="studio.selectedModel?.protocol === 'gemini'" class="grid grid-cols-2 gap-3">
              <div><label class="mb-2 block text-xs font-medium">{{ t('imageStudio.aspectRatio') }}</label><Select v-model="studio.aspectRatio" :options="['1:1', '2:3', '3:2', '3:4', '4:3', '9:16', '16:9', '21:9'].map(value => ({ value, label: value }))" :aria-label="t('imageStudio.aspectRatio')" /></div>
            </div>
            <div><label for="studio-count" class="mb-2 block text-xs font-medium">{{ t('imageStudio.count') }}</label><div class="flex gap-2"><button v-for="number in [1, 2, 4]" :key="number" type="button" class="count-button" :class="{ active: studio.count === number }" :aria-pressed="studio.count === number"  @click="studio.count = number">{{ number }}</button><input id="studio-count" v-model.number="studio.count" type="number" min="1" max="10" class="input min-w-0 flex-1 text-center" /></div></div>
          </div>
          <div class="composer-submit">
            <div ref="submitAnchor">
              <button ref="originalSubmit" type="submit" class="btn btn-primary w-full py-3" :style="{ visibility: showGenerationDock ? 'hidden' : undefined }" :disabled="!studio.canGenerate || Boolean(customSizeError)"><Icon name="sparkles" size="sm" />{{ generateLabel }}</button>
            </div>
            <p class="mt-2 text-center text-[11px] leading-5 text-gray-500 dark:text-dark-300">{{ t('imageStudio.queueHint') }}</p>
          </div>
        </form>

        <section ref="resultsElement" class="studio-results" :aria-label="t('imageStudio.results')">
          <div class="results-toolbar">
            <div class="flex flex-wrap items-center justify-between gap-3"><h2 class="text-lg font-semibold">{{ activeCategory?.name || t('imageStudio.results') }} <span class="ml-1 text-sm font-normal text-gray-500">{{ filteredHistory.length }}</span></h2><div class="flex gap-2"><button class="btn btn-secondary btn-sm" :disabled="studio.loading" @click="studio.refresh()"><Icon name="refresh" size="sm" /><span class="sr-only">{{ t('imageStudio.refresh') }}</span></button><button class="btn btn-secondary btn-sm" @click="studio.newDraft()"><Icon name="plus" size="sm" />{{ t('imageStudio.newDraft') }}</button></div></div>
            <div class="mt-3 flex flex-wrap items-center gap-2">
              <div class="flex rounded-lg border border-gray-200 p-0.5 dark:border-dark-700"><button class="layout-button" :class="{ active: view === 'all' }" @click="view = 'all'; studio.categoryID = null">{{ t('imageStudio.allResults') }}</button><button class="layout-button" :class="{ active: view === 'category' }" @click="view = 'category'">{{ t('imageStudio.byCategory') }}</button></div>
              <button v-if="view === 'category' && studio.categoryID" class="text-button" @click="studio.categoryID = null">{{ t('imageStudio.backCategories') }}</button>
              <button v-if="view === 'category'" class="text-button" @click="openCategory()"><Icon name="plus" size="sm" />{{ t('imageStudio.newCategory') }}</button>
              <label class="ml-auto flex items-center gap-2 text-xs"><input v-model="favoritesOnly" type="checkbox" class="rounded accent-primary-600" />{{ t('imageStudio.favoritesOnly') }}</label>
              <input v-model="search" class="input !w-44 !py-1.5 !text-xs" :aria-label="t('imageStudio.searchHistory')" :placeholder="t('imageStudio.searchHistory')" />
            </div>
            <div v-if="filteredHistory.length" class="mt-3 flex flex-wrap items-center gap-3 border-t border-gray-100 pt-3 text-xs dark:border-dark-800"><label class="flex items-center gap-2"><input type="checkbox" :checked="allSelected" class="rounded accent-primary-600" @change="toggleAll" />{{ t('imageStudio.selectAll') }}</label><span v-if="selected.length">{{ t('imageStudio.selectedCount', { count: selected.length }) }}</span><button class="text-button" :disabled="!selected.length" @click="downloadSelected">{{ t('imageStudio.download') }}</button><button class="text-button" :disabled="!selected.length" @click="assign(selected)">{{ t('imageStudio.categorize') }}</button><button class="text-button !text-red-600 dark:!text-red-400" :disabled="!selected.length" @click="removeHistory(selected)">{{ t('imageStudio.delete') }}</button></div>
          </div>
          <p v-if="studio.activeCount" role="status" class="flex items-center gap-2 border-b border-gray-100 px-4 py-3 text-xs text-primary-700 dark:border-dark-800 dark:text-primary-300"><span class="h-1.5 w-1.5 rounded-full bg-current" aria-hidden="true" />{{ t('imageStudio.queueSummary', { running: runningCount, queued: queuedCount }) }}</p>
          <div class="result-grid">
            <article v-for="task in visibleTasks" :key="task.id" class="result-card" :class="{ 'border-red-300 dark:border-red-900': task.status === 'failed' }">
              <div v-if="task.images?.length" class="relative"><button class="preview-trigger" :aria-label="t('imageStudio.previewImage')" @click="setPreview(task.images!, task.snapshot.prompt)"><BlobImage :blob="task.images[0]" :alt="task.snapshot.prompt" class="result-image" /><span class="preview-cue"><Icon name="eye" size="xs" />{{ t('imageStudio.previewImage') }}</span></button><span class="image-badge !bottom-auto top-2 bg-amber-50 text-amber-800">{{ t('imageStudio.notSaved') }}</span></div>
              <GenerationPending v-else-if="task.status === 'running' || task.status === 'queued'" :status="task.status" :elapsed="duration(task)" :title="t(`imageStudio.status.${task.status}`)" :subtitle="t(task.status === 'queued' ? 'imageStudio.pendingQueued' : 'imageStudio.pendingRunning')" :count-label="t('imageStudio.elapsedTime')" />
              <div v-else class="pending-image"><Icon name="exclamationCircle" size="lg" /><p class="mt-3 text-sm font-medium">{{ t(`imageStudio.status.${task.status}`) }}</p><p class="mt-1 text-xs tabular-nums">{{ duration(task) }}</p></div>
              <div class="space-y-3 p-3"><p class="line-clamp-3 text-sm leading-6">{{ task.snapshot.prompt }}</p><div class="text-xs text-gray-500 dark:text-dark-300">{{ task.snapshot.modelLabel || task.snapshot.model }} · {{ task.snapshot.size }} · {{ task.snapshot.quality }}</div><p v-if="task.error" class="line-clamp-2 text-xs text-red-600 dark:text-red-400">{{ formatError(task.error) }}</p><div class="flex flex-wrap gap-2"><button v-if="task.status === 'queued' || task.status === 'running'" class="btn btn-secondary btn-sm" @click="studio.cancel(task.id)">{{ t('imageStudio.cancel') }}</button><button v-else-if="task.status !== 'success'" class="btn btn-secondary btn-sm" @click="retry(task)">{{ t('imageStudio.retry') }}</button><button v-if="['failed', 'cancelled', 'interrupted'].includes(task.status)" class="text-button !text-red-600 dark:!text-red-400" @click="confirm('imageStudio.deleteTask', t('imageStudio.deleteTaskConfirm'), () => studio.dismissTask(task.id), true)">{{ t('imageStudio.delete') }}</button><button v-if="task.error" class="text-button" @click="detail = formatError(task.error)">{{ t('imageStudio.details') }}</button><template v-if="task.images?.length"><button class="btn btn-secondary btn-sm" @click="task.images!.forEach(blob => downloadImage(blob))">{{ t('imageStudio.download') }}</button><button class="btn btn-primary btn-sm" @click="studio.saveUnsaved(task)">{{ t('imageStudio.saveAgain') }}</button></template></div></div>
            </article>
            <template v-if="view === 'category' && !studio.categoryID && !search && !favoritesOnly">
              <article v-for="category in studio.categories" :key="category.id" class="result-card group">
                <button class="relative grid aspect-[4/3] w-full grid-cols-2 gap-0.5 overflow-hidden bg-gray-100 dark:bg-dark-900" :aria-label="t('imageStudio.openCategory', { name: category.name })" @click="studio.categoryID = category.id"><BlobImage v-for="(blob, index) in categoryImages(category.id)" :key="index" :blob="blob" alt="" class="h-full min-h-0 w-full object-cover" /><span v-if="!categoryImages(category.id).length" class="absolute inset-0 grid place-items-center text-gray-400"><Icon name="grid" size="xl" /></span></button>
                <div class="flex items-center justify-between gap-2 p-3"><button class="min-w-0 text-left" @click="studio.categoryID = category.id"><h3 class="truncate text-sm font-semibold">{{ category.name }}</h3><p class="mt-1 text-xs text-gray-500">{{ t('imageStudio.recordCount', { count: categoryCount(category.id) }) }}</p></button><div class="flex gap-1"><button class="icon-button" :aria-label="t('imageStudio.renameCategory')" @click="openCategory(category)"><Icon name="edit" size="sm" /></button><button class="icon-button" :aria-label="t('imageStudio.deleteCategory')" @click="confirm('imageStudio.deleteCategory', t('imageStudio.deleteCategoryConfirm'), () => studio.deleteCategory(category.id), true)"><Icon name="trash" size="sm" /></button></div></div>
              </article>
            </template>
            <article v-for="item in filteredHistory" :key="item.id" class="result-card" :data-new-result="newResultIDs.has(item.id)" :class="{ 'ring-2 ring-primary-500': selected.includes(item.id) }" @animationend.self="newResultIDs.delete(item.id)">
              <div class="relative"><button class="preview-trigger" :aria-label="t('imageStudio.previewImage')" @click="previewHistory(item)"><BlobImage :blob="item.images[0]" :alt="item.snapshot.prompt" class="result-image" /><span class="preview-cue"><Icon name="eye" size="xs" />{{ t('imageStudio.previewImage') }}</span></button><label class="absolute left-2 top-2 rounded bg-white/90 p-1.5 dark:bg-dark-900/90"><input type="checkbox" :checked="selected.includes(item.id)" :aria-label="t('imageStudio.selectImage')" class="block accent-primary-600" @change="toggle(item.id)" /></label><button class="absolute right-2 top-2 rounded bg-white/90 p-1.5 text-xs dark:bg-dark-900/90" :aria-label="item.favorite ? t('imageStudio.unfavorite') : t('imageStudio.favorite')" :aria-pressed="item.favorite" @click="studio.favorite(item)"><Icon name="badge" size="sm" :class="item.favorite ? 'fill-primary-500 text-primary-500' : 'text-gray-600 dark:text-gray-300'" /></button><div v-if="item.references.length" class="absolute bottom-2 right-2 flex gap-1"><button v-for="(blob, index) in item.references.slice(0, 3)" :key="index" type="button" data-preview-reference class="h-11 w-11 cursor-zoom-in overflow-hidden rounded border border-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500" :aria-label="t('imageStudio.previewReference', { number: index + 1 })" @click="setPreview(item.references, '', index)"><BlobImage :blob="blob" alt="" class="h-full w-full object-cover" /></button></div></div>
              <div class="space-y-2 p-3"><p class="whitespace-pre-wrap break-words text-sm leading-6" :class="{ 'line-clamp-2': !expanded.includes(item.id) }">{{ item.snapshot.prompt }}</p><div class="flex justify-end gap-3"><button class="text-button" @click="copyPrompt(item.snapshot.prompt)">{{ t('imageStudio.copy') }}</button><button class="text-button" @click="expanded = expanded.includes(item.id) ? expanded.filter(id => id !== item.id) : [...expanded, item.id]">{{ expanded.includes(item.id) ? t('imageStudio.collapse') : t('imageStudio.expand') }}</button></div><p class="text-[11px] leading-5 text-gray-500 dark:text-dark-300">{{ formatDate(item.createdAt) }} · {{ item.snapshot.modelLabel || item.snapshot.model }}<br />{{ item.snapshot.size }} · {{ item.snapshot.quality }} · {{ bytes(item.images[0]) }}</p><div class="flex flex-wrap gap-1"><span v-for="category in studio.categories.filter(category => (item.categoryIDs || [item.categoryID]).includes(category.id))" :key="category.id" class="rounded bg-primary-50 px-1.5 py-0.5 text-[10px] text-primary-700 dark:bg-primary-950 dark:text-primary-300">{{ category.name }}</span></div></div>
              <div class="flex flex-wrap gap-1 border-t border-gray-100 px-2 py-2 dark:border-dark-800"><button class="text-button" @click="studio.loadHistory(item)">{{ t('imageStudio.load') }}</button><button class="text-button" @click="retry(item)">{{ t('imageStudio.retry') }}</button><button class="text-button" :disabled="!studio.selectedModel?.max_references" @click="maskImage = item.images[0]">{{ t('imageStudio.maskEdit') }}</button><button class="text-button" @click="assign([item.id])">{{ t('imageStudio.categorize') }}</button><div class="ml-auto flex gap-1"><button class="icon-button" :aria-label="t('imageStudio.saveAsset')" @click="studio.saveAsset(item.images[0], t('imageStudio.generatedAsset', { date: formatDate(item.createdAt) }))"><Icon name="plus" size="sm" /></button><button class="icon-button" :disabled="!studio.selectedModel?.max_references" :aria-label="t('imageStudio.useReference')" @click="studio.addReferences([item.images[0]])"><Icon name="edit" size="sm" /></button><button class="icon-button" :aria-label="t('imageStudio.download')" @click="item.images.forEach(blob => downloadImage(blob))"><Icon name="download" size="sm" /></button><button class="icon-button" :aria-label="t('imageStudio.delete')" @click="removeHistory([item.id])"><Icon name="trash" size="sm" /></button></div></div>
            </article>
          </div>
          <div v-if="!filteredHistory.length && !visibleTasks.length && !(view === 'category' && !studio.categoryID && studio.categories.length)" class="flex min-h-80 flex-col items-center justify-center px-6 text-center"><Icon name="sparkles" size="xl" class="text-gray-400 dark:text-dark-500" /><h3 class="mt-5 text-base font-medium">{{ search || favoritesOnly ? t('imageStudio.noMatches') : t('imageStudio.emptyTitle') }}</h3><p class="mt-2 max-w-sm text-sm leading-6 text-gray-500 dark:text-dark-300">{{ search || favoritesOnly ? t('imageStudio.changeFilters') : t('imageStudio.emptyHint') }}</p></div>
        </section>
      </div>
    </div>
    <div v-if="showGenerationDock" ref="generationDock" data-generation-dock class="generation-dock" :class="{ 'sidebar-collapsed': app.sidebarCollapsed }" :style="{ bottom: `${keyboardInset}px` }" role="region" :aria-label="t('imageStudio.generateActions')">
      <div class="generation-dock-inner">
        <button type="button" class="dock-parameters" :aria-label="t('imageStudio.editParameters')" @click="editParameters">
          <span class="truncate text-sm font-semibold">{{ studio.selectedModel?.label || studio.selectedModel?.model || t('imageStudio.selectModel') }}</span>
          <span class="truncate text-xs text-gray-500 dark:text-dark-300">{{ t('imageStudio.batchCount', { count: studio.count }) }} · {{ t('imageStudio.editParameters') }}</span>
        </button>
        <button type="button" data-generate class="btn btn-primary min-h-11 shrink-0 px-4 sm:px-6" :disabled="!studio.canGenerate || Boolean(customSizeError)" @click="generateFromDock"><Icon name="sparkles" size="sm" />{{ generateLabel }}</button>
      </div>
    </div>
    <input ref="uploadInput" type="file" accept="image/png,image/jpeg,image/webp" multiple class="hidden" @change="upload($event)" />
    <input ref="importInput" type="file" accept="application/json,.json" class="hidden" @change="importBackup" />
    <input ref="assetUploadInput" type="file" accept="image/png,image/jpeg,image/webp" multiple class="hidden" @change="upload($event, true)" />
    <WorkflowLibrary :key="auth.user?.id" :show="libraryOpen" :prompts="studio.prompts" :initial-prompt="studio.prompt" :references="studio.references" :can-run="studio.enabled && Boolean(studio.selectedModel) && !customSizeError" @close="libraryOpen = false" @save="studio.savePrompt" @delete="studio.deletePrompt" @use="studio.prompt = $event" @run="generate" />
    <MaskEditor v-if="maskImage" :show="Boolean(maskImage)" :blob="maskImage" @close="maskImage = null" @confirm="submitEdit" />
    <BaseDialog :show="assetsOpen" :title="t('imageStudio.assets')" width="wide" @close="assetsOpen = false">
      <div class="mb-4 flex flex-wrap gap-3"><input v-model="assetSearch" class="input min-w-0 flex-1" :aria-label="t('imageStudio.searchAssets')" :placeholder="t('imageStudio.searchAssets')" /><button class="btn btn-primary" @click="assetUploadInput?.click()">{{ t('imageStudio.uploadAsset') }}</button></div>
      <div v-if="filteredAssets.length" class="grid grid-cols-2 gap-4 sm:grid-cols-3"><article v-for="asset in filteredAssets" :key="asset.id" class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700"><button type="button" data-preview-asset class="preview-trigger" :aria-label="t('imageStudio.previewAsset', { name: asset.name })" @click="setPreview(filteredAssets.map(item => item.image), filteredAssets.map(item => item.name), filteredAssets.findIndex(item => item.id === asset.id))"><BlobImage :blob="asset.image" :alt="asset.name" class="aspect-square w-full object-cover" /><span class="preview-cue"><Icon name="eye" size="xs" />{{ t('imageStudio.previewImage') }}</span></button><div class="flex items-center justify-between gap-2 p-2"><div class="min-w-0"><p class="truncate text-xs">{{ asset.name }}</p><button type="button" class="text-button mt-1 min-h-11" :disabled="!studio.selectedModel?.max_references" @click="studio.addReferences([asset.image]); assetsOpen = false">{{ t('imageStudio.useReference') }}</button></div><button class="icon-button shrink-0" :aria-label="t('imageStudio.deleteAsset')" @click="confirm('imageStudio.deleteAsset', t('imageStudio.deleteAssetConfirm'), () => studio.deleteAsset(asset.id), true)"><Icon name="trash" size="sm" /></button></div></article></div><p v-else class="py-16 text-center text-sm text-gray-500 dark:text-dark-300">{{ assetSearch ? t('imageStudio.noMatches') : t('imageStudio.noAssets') }}</p>
    </BaseDialog>
    <ImagePreview :show="Boolean(preview)" :images="preview?.images || []" :captions="preview?.captions || []" :initial-index="preview?.index || 0" @close="preview = null" @download="downloadImage" />
    <BaseDialog :show="categoryDialog" :title="editingCategory ? t('imageStudio.renameCategory') : t('imageStudio.newCategory')" width="narrow" @close="categoryDialog = false"><label class="block text-sm">{{ t('imageStudio.categoryName') }}<input v-model="categoryName" maxlength="120" class="input mt-2" @keydown.enter.prevent="saveCategory" /></label><template #footer><button class="btn btn-primary" :disabled="!categoryName.trim()" @click="saveCategory">{{ t('imageStudio.save') }}</button></template></BaseDialog>
    <BaseDialog :show="assignmentOpen" :title="t('imageStudio.categorize')" width="narrow" @close="assignmentOpen = false">
      <div class="space-y-3">
        <label v-for="category in studio.categories" :key="category.id" class="flex items-center gap-3 text-sm"><input v-model="assignmentCategories" data-category-choice type="checkbox" :value="category.id" class="rounded accent-primary-600" />{{ category.name }}</label>
        <p v-if="!studio.categories.length" class="text-sm text-gray-500 dark:text-dark-300">{{ t('imageStudio.noCategories') }}</p>
        <button class="btn btn-secondary btn-sm" @click="assignmentCategories = []">{{ t('imageStudio.moveUncategorized') }}</button>
        <p class="text-xs text-gray-500 dark:text-dark-300">{{ t('imageStudio.categoryHelp') }}</p>
      </div>
      <template #footer><button class="btn btn-primary" @click="studio.categorize(assignmentIDs, assignmentCategories); assignmentOpen = false">{{ t('imageStudio.save') }}</button></template>
    </BaseDialog>
    <BaseDialog :show="Boolean(detail)" :title="t('imageStudio.details')" width="wide" @close="detail = ''"><p class="whitespace-pre-wrap break-words text-sm leading-6">{{ detail }}</p></BaseDialog>
    <ConfirmDialog :show="Boolean(confirmation)" :title="confirmation?.title || ''" :message="confirmation?.message || ''" :danger="confirmation?.danger" @cancel="confirmation = null" @confirm="acceptConfirmation" />
  </AppLayout>
</template>

<style scoped>
.generation-dock { position: fixed; z-index: 30; left: 0; right: 0; padding: 10px 16px max(10px, env(safe-area-inset-bottom)); pointer-events: none; }
.generation-dock-inner { @apply mx-auto flex max-w-3xl items-center gap-3 rounded-xl border border-gray-200 bg-white p-3 shadow-xl dark:border-dark-600 dark:bg-dark-900; pointer-events: auto; animation: dock-arrive 180ms ease-out; }
.dock-parameters { @apply flex min-w-0 flex-1 flex-col gap-1 rounded-md text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500; min-height: 44px; justify-content: center; }
.studio-results { scroll-margin-top: 90px; }
#studio-prompt { scroll-margin-top: 90px; }
[data-new-result='true'] { animation: result-develop 650ms cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes result-confirm { from { opacity: 0.75; } to { opacity: 1; } }
@keyframes result-develop { from { opacity: 0.3; transform: scale(0.985); } to { opacity: 1; transform: scale(1); } }
@keyframes dock-arrive { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
@media (prefers-reduced-motion: reduce) { .generation-dock-inner { animation: none; } [data-new-result='true'] { animation: result-confirm 150ms ease-out; } }
.studio-layout { @apply grid gap-4; }
.studio-composer { @apply flex min-w-0 flex-col gap-5 rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900; }
.studio-results { @apply min-w-0 overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900; }
.results-toolbar { @apply border-b border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900; }
.result-grid { @apply grid grid-cols-1 gap-4 p-4 sm:grid-cols-2 2xl:grid-cols-3; }
.result-card { @apply min-w-0 overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900; }
.result-image { @apply aspect-[4/3] w-full bg-gray-100 object-cover dark:bg-dark-950; }
.pending-image { @apply flex aspect-[4/3] flex-col items-center justify-center bg-gray-50 text-gray-500 dark:bg-dark-950 dark:text-dark-300; }
.preview-trigger { @apply relative block w-full cursor-zoom-in focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500; }
.preview-cue { @apply pointer-events-none absolute bottom-2 left-2 inline-flex items-center gap-1 rounded bg-gray-900/70 px-2 py-1 text-[11px] text-white; }
.image-badge { @apply absolute bottom-2 left-2 rounded px-2 py-1 text-xs; }
.layout-button { @apply rounded-md px-3 py-1.5 text-xs text-gray-600 transition-colors hover:bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 dark:text-dark-300 dark:hover:bg-dark-700; }
.layout-button.active { @apply bg-primary-50 font-medium text-primary-700 dark:bg-primary-950 dark:text-primary-300; }
.text-button { @apply inline-flex items-center justify-center gap-1 rounded px-1 py-1 text-xs text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 disabled:cursor-not-allowed disabled:opacity-40 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white; }
.icon-button { @apply rounded p-1.5 text-gray-500 hover:bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-500 disabled:cursor-not-allowed disabled:opacity-40 dark:text-dark-300 dark:hover:bg-dark-700; }
.count-button { @apply w-11 rounded-lg border border-gray-200 text-sm disabled:opacity-40 dark:border-dark-600; }
.count-button.active { @apply border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-950 dark:text-primary-300; }
.layout-bottom .studio-results { order: 1; }
.layout-bottom .studio-composer { order: 2; }
@media (min-width: 1024px) {
  .generation-dock { left: 256px; }
  .generation-dock.sidebar-collapsed { left: 72px; }
  .layout-side { grid-template-columns: 340px minmax(0, 1fr); align-items: start; }
  .layout-side .studio-composer { max-height: calc(100dvh - 170px); overflow-y: auto; position: sticky; top: 90px; }
  .layout-side .studio-results { min-height: calc(100dvh - 170px); }
  .layout-bottom .result-grid { @apply grid-cols-3 2xl:grid-cols-4; }
  .layout-bottom .studio-composer { @apply sticky bottom-3 z-20 shadow-lg; display: grid; grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) 160px; gap: 20px; align-items: end; }
  .layout-bottom .composer-main { display: grid; grid-template-columns: minmax(0, 1fr); }
  .layout-bottom .composer-main textarea { min-height: 90px; height: 90px; }
  .layout-bottom .composer-settings { max-height: 240px; overflow-y: auto; padding-right: 5px; }
}
</style>
