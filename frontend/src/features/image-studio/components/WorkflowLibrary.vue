<script setup lang="ts">
import { vTooltip } from '@/components/common/tooltip'
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { ImagePrompt } from '../types'
import type { WorkflowDraft } from '../api'
import WorkflowAssistant from './WorkflowAssistant.vue'
import { extractVariables, renderTemplate, splitSeries } from './workflow'

const props = defineProps<{ show: boolean; prompts: ImagePrompt[]; initialPrompt?: string; references?: Blob[]; canRun: boolean }>()
const emit = defineEmits<{ close: []; save: [prompt: ImagePrompt]; delete: [id: string]; use: [prompt: string]; run: [prompts: string[]] }>()
const { t } = useI18n()
const search = ref('')
const category = ref('')
const tag = ref('')
const editing = ref<ImagePrompt | null>(null)
const running = ref<ImagePrompt | null>(null)
const deleteTarget = ref<ImagePrompt | null>(null)
const seriesText = ref('')
const tagText = ref('')
const values = reactive<Record<string, string>>({})
const drafts = ref<string[]>([])
const appendText = ref('')
const creating = ref(false)
const aiSeries = ref(false)
const categories = computed(() => [...new Set(props.prompts.map((item) => item.category).filter((value): value is string => Boolean(value)))])
const tags = computed(() => [...new Set(props.prompts.flatMap((item) => item.tags || []))])
const filtered = computed(() => props.prompts.filter((item) => (!category.value || item.category === category.value) && (!tag.value || item.tags?.includes(tag.value)) && `${item.title} ${item.prompt} ${(item.tags || []).join(' ')}`.toLowerCase().includes(search.value.trim().toLowerCase())))
const variables = computed(() => extractVariables([running.value?.prompt || '', ...(running.value?.seriesPrompts || [])].join('\n')))
const missingVariables = computed(() => variables.value.some((key) => !values[key]?.trim()))
const preview = computed(() => renderTemplate(running.value?.prompt || '', values))

watch(() => props.show, (show) => { if (!show) { editing.value = null; running.value = null; creating.value = false; aiSeries.value = false } })
function applyWorkflow(draft: WorkflowDraft) {
  creating.value = false
  edit({ ...draft, id: crypto.randomUUID(), createdAt: Date.now() })
}
function edit(item?: ImagePrompt) {
  editing.value = item ? { ...item } : { id: crypto.randomUUID(), title: '', prompt: props.initialPrompt || '', createdAt: Date.now() }
  seriesText.value = editing.value.seriesPrompts?.join('\n---\n') || ''
  tagText.value = editing.value.tags?.join(', ') || ''
  running.value = null
  creating.value = false
}
function save() {
  if (!editing.value?.title.trim() || !editing.value.prompt.trim()) return
  emit('save', { ...editing.value, title: editing.value.title.trim(), prompt: editing.value.prompt.trim(), category: editing.value.category?.trim(), tags: tagText.value.split(/[,，]/).map((value) => value.trim()).filter(Boolean), seriesPrompts: splitSeries(seriesText.value), updatedAt: Date.now() })
  editing.value = null
}
function duplicate(item: ImagePrompt) {
  emit('save', { ...item, id: crypto.randomUUID(), title: t('imageStudio.copyName', { name: item.title }), createdAt: Date.now() })
}
function openRunner(item: ImagePrompt) {
  running.value = item
  editing.value = null
  Object.keys(values).forEach((key) => delete values[key])
  drafts.value = []
  appendText.value = ''
  aiSeries.value = false
}
function prepareSeries() {
  if (!running.value || missingVariables.value) return
  drafts.value = (running.value.seriesPrompts?.length ? running.value.seriesPrompts : ['']).map((item) => [preview.value, renderTemplate(item, values)].filter(Boolean).join('\n\n'))
}
function move(index: number, offset: number) {
  const target = index + offset
  if (target < 0 || target >= drafts.value.length) return
  ;[drafts.value[index], drafts.value[target]] = [drafts.value[target], drafts.value[index]]
}
function run(items: string[]) {
  const prompts = items.map((item) => item.trim()).filter(Boolean)
  if (!props.canRun || !prompts.length || prompts.length > 30) return
  emit('run', prompts)
  emit('close')
}
</script>

<template>
  <BaseDialog close-tooltip :show="show" :title="t('imageStudio.workflowTitle')" width="extra-wide" @close="emit('close')">
    <div v-if="show && creating" class="mx-auto max-w-2xl space-y-5">
      <div class="flex items-center justify-between gap-3"><h3 class="text-lg font-semibold">{{ t('imageStudio.aiCreate') }}</h3><button class="btn btn-secondary" @click="creating = false">{{ t('imageStudio.back') }}</button></div>
      <WorkflowAssistant mode="workflow" :initial-prompt="initialPrompt" :references="references" @workflow="applyWorkflow" />
    </div>
    <div v-else-if="editing" class="grid gap-5 md:grid-cols-2">
      <div class="space-y-4">
        <label class="block text-sm font-medium">{{ t('imageStudio.templateName') }}<input v-model="editing.title" maxlength="120" class="input mt-2" /></label>
        <label class="block text-sm font-medium">{{ t('imageStudio.category') }}<input v-model="editing.category" maxlength="120" class="input mt-2" /></label>
        <label class="block text-sm font-medium">{{ t('imageStudio.tags') }}<input v-model="tagText" maxlength="500" class="input mt-2" :placeholder="t('imageStudio.tagsHint')" /></label>
        <label class="block text-sm font-medium">{{ t('imageStudio.templatePrompt') }}<textarea v-model="editing.prompt" rows="8" maxlength="10000" class="input mt-2" :placeholder="t('imageStudio.templatePlaceholder')" /></label>
        <p class="text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('imageStudio.variableHint') }}</p>
      </div>
      <div class="space-y-4">
        <label class="block text-sm font-medium">{{ t('imageStudio.seriesPrompts') }}<textarea v-model="seriesText" rows="15" maxlength="100000" class="input mt-2" :placeholder="t('imageStudio.seriesPlaceholder')" /></label>
        <p class="text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('imageStudio.seriesHelp') }}</p>
        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="editing = null">{{ t('imageStudio.cancel') }}</button>
          <button class="btn btn-primary" :disabled="!editing.title.trim() || !editing.prompt.trim() || splitSeries(seriesText).length > 30" @click="save">{{ t('imageStudio.saveTemplate') }}</button>
        </div>
      </div>
    </div>
    <div v-else-if="running" class="space-y-5 md:flex md:h-[calc(90dvh-10rem)] md:min-h-0 md:flex-col">
      <div class="flex shrink-0 items-center justify-between gap-3"><h3 class="min-w-0 break-words text-lg font-semibold">{{ running.title }}</h3><button class="btn btn-secondary" @click="running = null">{{ t('imageStudio.back') }}</button></div>
      <div class="grid gap-5 md:min-h-0 md:flex-1 md:grid-cols-[260px_minmax(0,1fr)]">
        <div class="min-w-0 space-y-4 md:min-h-0 md:overflow-y-auto md:overscroll-contain md:pr-2">
          <label v-for="key in variables" :key="key" class="block text-sm font-medium">{{ key }}<input v-model="values[key]" class="input mt-2" maxlength="2000" /></label>
          <p v-if="!variables.length" class="text-sm text-gray-500 dark:text-dark-300">{{ t('imageStudio.noVariables') }}</p>
          <button class="btn btn-secondary w-full" :disabled="missingVariables" @click="emit('use', preview); emit('close')">{{ t('imageStudio.usePrompt') }}</button>
          <button class="btn btn-primary w-full" :disabled="missingVariables" @click="prepareSeries">{{ t('imageStudio.prepareSeries') }}</button>
          <button class="btn btn-secondary w-full" :disabled="missingVariables" @click="aiSeries = !aiSeries"><Icon name="sparkles" size="sm" />{{ t('imageStudio.aiSeries') }}</button>
          <p class="text-xs leading-5 text-gray-500 dark:text-dark-300">{{ t('imageStudio.workflowUsesSettings') }}</p>
        </div>
        <div class="min-w-0 space-y-4 md:min-h-0 md:overflow-y-auto md:overscroll-contain md:pr-2">
          <WorkflowAssistant v-if="show && aiSeries && !missingVariables" mode="series" :initial-prompt="preview" :references="references" @series="drafts = $event; aiSeries = false" />
          <p class="whitespace-pre-wrap break-words rounded-xl bg-gray-50 p-4 text-sm leading-6 dark:bg-dark-900">{{ preview }}</p>
          <template v-if="drafts.length">
            <div class="flex flex-wrap gap-2"><input v-model="appendText" class="input min-w-0 flex-1" :aria-label="t('imageStudio.appendInstructions')" :placeholder="t('imageStudio.appendInstructions')" /><button class="btn btn-secondary" :disabled="!appendText.trim()" @click="drafts = drafts.map(item => item + '\n' + appendText.trim()); appendText = ''">{{ t('imageStudio.appendAll') }}</button></div>
            <div class="space-y-4">
              <div v-for="(_, index) in drafts" :key="index" class="space-y-2 border-b border-gray-200 pb-4 dark:border-dark-700">
                <label class="text-sm font-medium">{{ t('imageStudio.seriesItem', { number: index + 1 }) }}<textarea v-model="drafts[index]" rows="4" maxlength="10000" class="input mt-2" /></label>
                <div class="flex flex-wrap justify-end gap-2">
                  <button v-tooltip class="btn btn-secondary btn-sm" :disabled="index === 0" :aria-label="t('imageStudio.moveUp')" @click="move(index, -1)"><Icon name="arrowUp" size="sm" /></button>
                  <button v-tooltip class="btn btn-secondary btn-sm" :disabled="index === drafts.length - 1" :aria-label="t('imageStudio.moveDown')" @click="move(index, 1)"><Icon name="arrowDown" size="sm" /></button>
                  <button v-tooltip class="btn btn-secondary btn-sm" :aria-label="t('imageStudio.delete')" @click="drafts.splice(index, 1)"><Icon name="trash" size="sm" /></button>
                  <button class="btn btn-secondary btn-sm" :disabled="!canRun || !drafts[index].trim()" @click="run([drafts[index]])">{{ t('imageStudio.generateOne') }}</button>
                </div>
              </div>
            </div>
            <button class="btn btn-primary w-full" :disabled="!canRun || !drafts.some(item => item.trim()) || drafts.length > 30" @click="run(drafts)">{{ t('imageStudio.generateSeries', { count: drafts.length }) }}</button>
          </template>
        </div>
      </div>
    </div>
    <div v-else class="space-y-5">
      <div class="flex flex-wrap gap-3">
        <input v-model="search" class="input min-w-40 flex-1" :aria-label="t('imageStudio.searchTemplates')" :placeholder="t('imageStudio.searchTemplates')" />
        <Select v-model="category" class="w-40" :aria-label="t('imageStudio.category')" :options="[{ value: '', label: t('imageStudio.allCategories') }, ...categories.map(value => ({ value, label: value }))]" />
        <Select v-model="tag" class="w-36" :aria-label="t('imageStudio.tags')" :options="[{ value: '', label: t('imageStudio.allTags') }, ...tags.map(value => ({ value, label: value }))]" />
        <button class="btn btn-primary" @click="edit()"><Icon name="plus" size="sm" />{{ t('imageStudio.newTemplate') }}</button>
        <button class="btn btn-secondary" @click="creating = true"><Icon name="sparkles" size="sm" />{{ t('imageStudio.aiCreate') }}</button>
      </div>
      <p class="text-sm text-gray-500 dark:text-dark-300">{{ t('imageStudio.localTemplates') }}</p>
      <div v-if="filtered.length" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <article v-for="item in filtered" :key="item.id" class="flex flex-col border-b border-gray-200 pb-5 dark:border-dark-700">
          <h3 class="truncate font-semibold">{{ item.title }}</h3>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ item.category || t('imageStudio.uncategorized') }}<span v-if="item.seriesPrompts?.length"> · {{ t('imageStudio.seriesCount', { count: item.seriesPrompts.length }) }}</span></p>
          <p class="my-3 line-clamp-4 whitespace-pre-wrap text-sm leading-6 text-gray-600 dark:text-dark-300">{{ item.prompt }}</p>
          <div class="mt-auto flex flex-wrap gap-2">
            <button class="btn btn-primary btn-sm" @click="openRunner(item)">{{ t('imageStudio.useTemplate') }}</button>
            <button v-tooltip class="btn btn-secondary btn-sm" :aria-label="t('imageStudio.editTemplate')" @click="edit(item)"><Icon name="edit" size="sm" /></button>
            <button v-tooltip class="btn btn-secondary btn-sm" :aria-label="t('imageStudio.duplicateTemplate')" @click="duplicate(item)"><Icon name="copy" size="sm" /></button>
            <button v-tooltip class="btn btn-secondary btn-sm" :aria-label="t('imageStudio.deleteTemplate')" @click="deleteTarget = item"><Icon name="trash" size="sm" /></button>
          </div>
        </article>
      </div>
      <div v-else class="py-16 text-center text-gray-500 dark:text-dark-300">{{ search || category || tag ? t('imageStudio.noMatches') : t('imageStudio.noTemplates') }}</div>
    </div>
  </BaseDialog>
  <ConfirmDialog :show="Boolean(deleteTarget)" :title="t('imageStudio.deleteTemplate')" :message="t('imageStudio.deleteTemplateConfirm')" danger @cancel="deleteTarget = null" @confirm="deleteTarget && emit('delete', deleteTarget.id); deleteTarget = null" />
</template>
