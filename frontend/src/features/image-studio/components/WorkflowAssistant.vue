<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { generateWorkflow, imageStudioAPI, imageToDataURL, type WorkflowDraft, type WorkflowModel } from '../api'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = withDefaults(defineProps<{ mode: 'workflow' | 'series'; initialPrompt?: string; references?: Blob[] }>(), { initialPrompt: '', references: () => [] })
const emit = defineEmits<{ workflow: [draft: WorkflowDraft]; series: [prompts: string[]] }>()
const { t } = useI18n()
const auth = useAuthStore()
const models = ref<WorkflowModel[]>([])
const modelID = ref('')
const prompt = ref(props.initialPrompt)
const count = ref(4)
const includeReferences = ref(true)
const loading = ref(false)
const busy = ref(false)
const error = ref('')
const notice = ref('')
const selected = computed(() => models.value.find(model => model.id === modelID.value))
const canSubmit = computed(() => !loading.value && !busy.value && !!selected.value && !!prompt.value.trim() &&
  (props.mode !== 'series' || (Number.isInteger(count.value) && count.value >= 1 && count.value <= 20)))
let sequence = 0
let request: AbortController | undefined
let catalog: AbortController | undefined

function cancel() {
  sequence++
  request?.abort()
  request = undefined
  busy.value = false
  notice.value = t('imageStudio.aiCancelled')
}
async function load() {
  catalog?.abort()
  const controller = new AbortController()
  catalog = controller
  models.value = []; loading.value = true; error.value = ''
  if (!auth.user?.id) { modelID.value = ''; loading.value = false; return }
  try {
    const result = await imageStudioAPI.workflowModels(controller.signal)
    if (controller.signal.aborted) return
    models.value = result.enabled ? result.models : []
    if (!models.value.some(model => model.id === modelID.value)) modelID.value = models.value[0]?.id || ''
  } catch (err) {
    if (!controller.signal.aborted) error.value = extractApiErrorMessage(err, t('imageStudio.aiFailed'))
  } finally { if (!controller.signal.aborted) loading.value = false }
}
async function generate() {
  if (!canSubmit.value || !selected.value) return
  const current = ++sequence
  const controller = new AbortController()
  request = controller
  busy.value = true; error.value = ''; notice.value = ''
  const model = selected.value
  const text = prompt.value.trim()
  const targetCount = count.value
  try {
    const images = includeReferences.value && model.supports_references ? [...props.references] : []
    if (images.length > 3 || images.some(image => image.size > 10 * 1024 * 1024)) throw new Error(t('imageStudio.aiReferenceLimit'))
    const references = await Promise.all(images.map(imageToDataURL))
    if (current !== sequence || controller.signal.aborted) return
    const result = await generateWorkflow({ model_id: model.id, mode: props.mode, prompt: text, references, ...(props.mode === 'series' ? { count: targetCount } : {}) }, controller.signal)
    if (current !== sequence || controller.signal.aborted) return
    if (props.mode === 'workflow' && 'draft' in result) emit('workflow', result.draft)
    else if (props.mode === 'series' && 'prompts' in result) emit('series', result.prompts)
    else throw new Error(t('imageStudio.aiFailed'))
  } catch (err) {
    if (current === sequence && !controller.signal.aborted) {
      const message = extractApiErrorMessage(err, t('imageStudio.aiFailed'))
      error.value = message.startsWith('imageStudio.') ? t(message) : message
    }
  } finally { if (current === sequence) { busy.value = false; request = undefined } }
}
watch(() => auth.user?.id, () => { cancel(); prompt.value = ''; notice.value = ''; void load() })
watch(() => props.initialPrompt, value => { cancel(); prompt.value = value; notice.value = '' })
onMounted(load)
onBeforeUnmount(() => { cancel(); catalog?.abort() })
</script>

<template>
  <section class="space-y-4" :aria-label="t(mode === 'workflow' ? 'imageStudio.aiCreate' : 'imageStudio.aiSeries')" :aria-busy="busy">
    <div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_8rem]">
      <label class="block min-w-0 text-sm font-medium" :class="{ 'sm:col-span-2': mode === 'workflow' }">
        {{ t('imageStudio.aiModel') }}
        <Select v-model="modelID" class="mt-2" :disabled="loading || busy || !models.length" :options="models.map(model => ({ value: model.id, label: `${model.label} · ${model.group_name}` }))" :aria-label="t('imageStudio.aiModel')" />
      </label>
      <label v-if="mode === 'series'" class="block text-sm font-medium">{{ t('imageStudio.aiCount') }}<input v-model.number="count" type="number" min="1" max="20" step="1" class="input mt-2" :disabled="busy" /></label>
    </div>
    <p v-if="loading" role="status" class="text-sm text-gray-600 dark:text-dark-300">{{ t('imageStudio.aiLoadingModels') }}</p>
    <div v-else-if="!models.length" class="space-y-2 text-sm text-gray-600 dark:text-dark-300"><p>{{ t('imageStudio.aiNoModels') }}</p><button type="button" class="btn btn-secondary btn-sm" @click="load">{{ t('imageStudio.refresh') }}</button></div>
    <label class="block text-sm font-medium">{{ t(mode === 'workflow' ? 'imageStudio.aiBrief' : 'imageStudio.aiSeriesBrief') }}<textarea v-model="prompt" rows="5" maxlength="16000" class="input mt-2" :disabled="busy" :placeholder="t('imageStudio.aiBriefPlaceholder')" /></label>
    <label v-if="references.length && selected?.supports_references" class="flex items-center gap-2 text-sm"><input v-model="includeReferences" type="checkbox" class="h-4 w-4 accent-primary-600" :disabled="busy" />{{ t('imageStudio.aiUseReferences', { count: references.length }) }}</label>
    <p v-else-if="references.length && selected" class="text-sm text-gray-600 dark:text-dark-300">{{ t('imageStudio.aiTextOnly') }}</p>
    <p class="text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('imageStudio.aiBilling') }}</p>
    <p v-if="error" role="alert" class="text-sm text-red-700 dark:text-red-300">{{ error }}</p>
    <p v-if="notice" role="status" class="text-sm text-gray-600 dark:text-dark-300">{{ notice }}</p>
    <div class="flex flex-wrap items-center justify-end gap-3">
      <span v-if="busy" role="status" class="mr-auto flex items-center gap-2 text-sm text-primary-700 dark:text-primary-300"><Icon name="refresh" size="sm" class="animate-spin motion-reduce:animate-none" />{{ t('imageStudio.aiWorking') }}</span>
      <button v-if="busy" type="button" class="btn btn-secondary" @click="cancel">{{ t('imageStudio.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="!canSubmit" @click="generate"><Icon name="sparkles" size="sm" />{{ t(mode === 'workflow' ? 'imageStudio.aiCreateDraft' : 'imageStudio.aiGenerateSeries') }}</button>
    </div>
  </section>
</template>
