<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { filterPrompts, loadSource, sources, type LibraryPrompt } from './catalog'
const props = withDefaults(defineProps<{ show: boolean; initialTab?: 'upstream' | 'custom' }>(), { initialTab: 'upstream' })
const emit = defineEmits<{ close: []; use: [prompt: string] }>()
const { t } = useI18n()
const tab = ref<'upstream' | 'custom'>('upstream'), category = ref(''), keyword = ref(''), tags = ref<string[]>([])
const bySource = shallowRef<Record<string, LibraryPrompt[]>>({}), loading = ref(false), failures = ref<string[]>([])
const loadedImages = ref(new Set<string>())
const limit = ref(40), pane = ref<HTMLElement>(), detail = ref<LibraryPrompt | null>(null), failedImages = ref(new Set<string>())
const items = computed(() => sources.flatMap(source => bySource.value[source.id] || []))
const available = computed(() => filterPrompts(items.value, category.value, keyword.value, []))
const availableTags = computed(() => [...new Set(available.value.flatMap(item => item.tags))])
const filtered = computed(() => filterPrompts(items.value, category.value, keyword.value, tags.value))
const visible = computed(() => filtered.value.slice(0, limit.value))
watch([category, keyword, tags], () => { limit.value = 40; if (pane.value) pane.value.scrollTop = 0 })
watch(() => props.show, show => {
  if (show) { tab.value = props.initialTab; detail.value = null; if (tab.value === 'upstream') void load() }
}, { immediate: true })
watch(tab, value => { if (props.show && value === 'upstream') void load() })
async function load(force = false) {
  if (loading.value) return
  loading.value = true; failures.value = []
  if (force) failedImages.value = new Set()
  await Promise.all(sources.map(async source => {
    try {
      const result = await loadSource(source, force)
      bySource.value = { ...bySource.value, [source.id]: result.items }
      if (result.stale) failures.value.push(source.name)
    } catch { failures.value.push(source.name) }
  }))
  loading.value = false
}
function tabKey(event: KeyboardEvent) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  tab.value = event.key === 'Home' ? 'upstream' : event.key === 'End' ? 'custom' : tab.value === 'upstream' ? 'custom' : 'upstream'
  detail.value = null
  ;(event.currentTarget as HTMLElement).querySelectorAll<HTMLButtonElement>('button')[tab.value === 'upstream' ? 0 : 1]?.focus()
}
function toggleTag(tag: string) { tags.value = tags.value.includes(tag) ? tags.value.filter(value => value !== tag) : [...tags.value, tag] }
function select(prompt: string) { emit('use', prompt); emit('close') }
function scroll() { if (pane.value && pane.value.scrollTop + pane.value.clientHeight >= pane.value.scrollHeight - 160) limit.value += 40 }
</script>
<template>
  <BaseDialog :show="show" :title="t('videoStudio.prompts')" width="full" close-tooltip :close-on-click-outside="tab==='upstream'" @close="emit('close')">
    <div class="prompt-library">
      <div class="mb-4 flex items-center justify-between gap-3 border-b border-gray-200 dark:border-dark-700">
        <div class="flex gap-5" role="tablist" @keydown="tabKey" :aria-label="t('videoStudio.prompts')">
          <button v-for="value in (['upstream','custom'] as const)" :key="value" type="button" role="tab" :aria-selected="tab===value" :tabindex="tab===value?0:-1" class="library-tab" :class="{ selected:tab===value }" @click="tab=value;detail=null">{{t(`promptLibrary.${value}`)}}</button>
        </div>
        <button v-if="tab==='upstream'" type="button" class="text-xs text-gray-500 hover:text-primary-600 disabled:opacity-50" :disabled="loading" @click="load(true)">{{t('common.refresh')}}</button>
      </div>
      <div v-if="tab==='custom'" role="tabpanel"><slot name="custom" /></div>
      <div v-else-if="detail" class="library-detail" role="tabpanel">
        <button type="button" class="mb-4 text-sm text-primary-600" @click="detail=null">← {{t('promptLibrary.back')}}</button>
        <div class="grid gap-5 md:grid-cols-2">
          <img v-if="detail.coverUrl && !failedImages.has(detail.coverUrl)" :src="detail.coverUrl" :alt="detail.title" referrerpolicy="no-referrer" class="max-h-[52dvh] w-full rounded-lg bg-gray-50 object-contain dark:bg-dark-800" @error="failedImages.add(detail.coverUrl)" />
          <div class="min-w-0 space-y-4">
            <h4 class="text-lg font-semibold">{{detail.title}}</h4>
            <p class="text-xs text-gray-500">{{detail.category}}<span v-if="detail.author"> · {{detail.author}}</span></p>
            <p v-if="detail.description" class="text-sm text-gray-500">{{detail.description}}</p>
            <p class="max-h-[35dvh] overflow-auto whitespace-pre-wrap break-words rounded-lg bg-gray-50 p-4 text-sm leading-6 dark:bg-dark-800">{{detail.prompt}}</p>
            <div class="flex flex-wrap items-center gap-4">
              <button type="button" class="btn btn-primary" @click="select(detail.prompt)">{{t('promptLibrary.use')}}</button>
              <a v-if="detail.sourceUrl" :href="detail.sourceUrl" target="_blank" rel="noopener noreferrer" class="text-sm text-primary-600">{{t('promptLibrary.source')}} ↗</a>
            </div>
          </div>
        </div>
      </div>
      <div v-else role="tabpanel" class="library-layout">
        <aside class="library-filters">
          <div role="group" :aria-label="t('promptLibrary.categories')">
            <h4 class="filter-heading">{{t('promptLibrary.categories')}}</h4>
            <div class="category-list">
              <button type="button" class="filter-option" :class="{selected:!category}" :aria-pressed="!category" @click="category='';tags=[]">{{t('promptLibrary.all')}}</button>
              <button v-for="source in sources" :key="source.id" type="button" class="filter-option" :class="{selected:category===source.id}" :aria-pressed="category===source.id" @click="category=source.id;tags=[]">{{source.name}}</button>
            </div>
          </div>
          <div class="mt-6" role="group" :aria-label="t('promptLibrary.tags')">
            <h4 class="filter-heading">{{t('promptLibrary.tags')}}</h4>
            <div class="flex flex-wrap gap-1.5">
              <button type="button" class="filter-option" :class="{selected:!tags.length}" :aria-pressed="!tags.length" @click="tags=[]">{{t('promptLibrary.all')}}</button>
              <button v-for="tag in availableTags" :key="tag" type="button" class="filter-option" :class="{selected:tags.includes(tag)}" :aria-pressed="tags.includes(tag)" @click="toggleTag(tag)">{{tag}}</button>
            </div>
          </div>
        </aside>
        <section class="flex min-h-0 min-w-0 flex-col">
          <label class="flex items-center gap-2 rounded-lg border border-gray-200 px-3 dark:border-dark-600">
            <Icon name="search" size="sm" class="shrink-0 text-gray-400" />
            <input v-model="keyword" class="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none" :aria-label="t('promptLibrary.search')" :placeholder="t('promptLibrary.search')" />
          </label>
          <div v-if="failures.length" role="status" class="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-800 dark:bg-amber-950 dark:text-amber-200">{{t('promptLibrary.loadFailed')}} {{failures.join('、')}}</div>
          <div class="my-3 flex justify-between gap-2 text-xs text-gray-500"><span>{{t('promptLibrary.count',{count:filtered.length})}}</span><span v-if="loading" role="status">{{t('common.loading')}}</span></div>
          <div ref="pane" class="min-h-0 flex-1 overflow-y-auto pr-1" @scroll="scroll">
            <div v-if="loading && !items.length" class="library-grid" aria-hidden="true"><div v-for="i in 10" :key="i" class="aspect-[3/4] rounded-lg bg-gray-100 dark:bg-dark-800" /></div>
            <div v-else class="library-grid">
              <article v-for="item in visible" :key="`${item.sourceId}:${item.id}`" class="prompt-card">
                <button type="button" class="block w-full text-left" :aria-label="t('promptLibrary.useNamed',{name:item.title})" @click="select(item.prompt)">
                  <span class="relative grid aspect-square place-items-center overflow-hidden bg-gray-50 text-gray-400 dark:bg-dark-800">
                    <span v-if="!loadedImages.has(item.coverUrl) || failedImages.has(item.coverUrl)" class="flex flex-col items-center gap-2"><Icon name="document" size="lg" /><span class="text-xs">{{t(item.coverUrl && !failedImages.has(item.coverUrl) ? 'promptLibrary.coverLoading' : 'promptLibrary.noCover')}}</span></span>
                    <img v-if="item.coverUrl && !failedImages.has(item.coverUrl)" :src="item.coverUrl" alt="" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="absolute inset-0 h-full w-full object-cover" :class="loadedImages.has(item.coverUrl)?'opacity-100':'opacity-0'" @load="loadedImages.add(item.coverUrl)" @error="failedImages.add(item.coverUrl)" />
                  </span>
                  <h4 class="truncate px-3 pb-1 pt-3 text-sm font-semibold" :title="item.title">{{item.title}}</h4>
                </button>
                <div class="flex items-center justify-between gap-2 px-3 pb-3 pt-1">
                  <button type="button" class="text-xs text-primary-600 dark:text-primary-400" @click="select(item.prompt)">{{t('promptLibrary.use')}}</button>
                  <button type="button" class="text-xs text-gray-500 hover:text-primary-600" :aria-label="t('promptLibrary.detailNamed',{name:item.title})" @click="detail=item">{{t('promptLibrary.details')}}</button>
                </div>
              </article>
            </div>
            <p v-if="!loading && !filtered.length" class="py-16 text-center text-sm text-gray-500">{{t('promptLibrary.empty')}}</p>
            <button v-if="visible.length<filtered.length" type="button" class="mt-4 w-full py-3 text-sm text-primary-600" @click="limit+=40">{{t('promptLibrary.more')}}</button>
          </div>
        </section>
      </div>
    </div>
  </BaseDialog>
</template>
<style scoped>
.library-tab { @apply border-b-2 border-transparent pb-3 text-sm font-medium text-gray-500 }
.library-tab.selected { @apply border-primary-600 text-primary-700 dark:text-primary-400 }
.library-layout { @apply grid min-h-0 gap-5 sm:grid-cols-[200px_minmax(0,1fr)]; height:62dvh }
.library-filters { @apply min-h-0 overflow-y-auto border-r border-gray-200 pr-4 dark:border-dark-700 }
.filter-heading { @apply mb-2 text-xs font-medium text-gray-400 }
.category-list { @apply flex flex-col items-start gap-1 }
.filter-option { @apply max-w-full rounded-md px-2 py-1.5 text-left text-xs leading-5 text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700; overflow-wrap:anywhere }
.filter-option.selected { @apply bg-gray-900 text-white dark:bg-gray-100 dark:text-gray-900 }
.library-grid { @apply grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 }
.prompt-card { @apply overflow-hidden rounded-lg border border-gray-200 bg-white hover:border-primary-400 focus-within:ring-2 focus-within:ring-primary-400 dark:border-dark-700 dark:bg-dark-900 }
@media(max-width:639px) {
  .library-layout { grid-template-rows:auto minmax(0,1fr); height:65dvh }
  .library-filters { max-height:140px; border-right:0; padding-right:0; border-bottom:1px solid #e5e7eb; padding-bottom:8px }
  .category-list { flex-direction:row; flex-wrap:wrap }
  .library-filters > div + div { margin-top:12px }
}
</style>
