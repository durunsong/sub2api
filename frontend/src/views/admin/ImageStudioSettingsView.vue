<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">{{ t('studioAdmin.title') }}</h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-400">{{ t('studioAdmin.description') }}</p>
        </div>
        <RouterLink class="btn btn-secondary" to="/image-studio">{{ t('studioAdmin.open') }}</RouterLink>
      </header>
      <p v-if="error" role="alert" class="rounded-lg bg-red-50 p-4 text-sm text-red-700 dark:bg-red-950 dark:text-red-200">{{ error }}</p>
      <p v-if="saved" role="status" class="text-sm text-green-700 dark:text-green-300">{{ t('studioAdmin.saved') }}</p>
      <div class="flex flex-wrap items-center justify-between gap-4 border-y border-gray-200 py-4 dark:border-dark-700">
        <label class="flex items-center gap-3 text-sm font-medium">
          <input v-model="config.enabled" type="checkbox" class="h-4 w-4 rounded accent-primary-600" :disabled="loading || saving" />
          {{ t('studioAdmin.enabled') }}
        </label>
        <div class="flex gap-3">
          <button class="btn btn-secondary" :disabled="loading || saving" @click="load">{{ t('common.refresh') }}</button>
          <button class="btn btn-primary" :disabled="loading || saving || !loaded" @click="save">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </div>
      <div class="flex flex-wrap gap-3">
        <input v-model="search" class="input max-w-md" :placeholder="t('studioAdmin.search')" :aria-label="t('studioAdmin.search')" />
        <Select v-model="groupFilter" class="w-full sm:w-60" :options="[{value:0,label:t('studioAdmin.allGroups')},...groupOptions]" :aria-label="t('studioAdmin.group')" />
        <button class="btn btn-secondary" :disabled="loading || saving" @click="showAdd = true"><Icon name="plus" size="sm" />{{ t('studioAdmin.add') }}</button>
      </div>
      <div v-if="loading" role="status" class="py-16 text-center text-gray-500">{{ t('common.loading') }}</div>
      <div v-else class="overflow-x-auto rounded-xl border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[680px] text-left text-sm">
          <thead class="bg-gray-50 text-gray-600 dark:bg-dark-800 dark:text-gray-400">
            <tr><th class="p-4">{{ t('studioAdmin.visible') }}</th><th class="p-4">{{ t('studioAdmin.group') }}</th><th class="p-4">{{ t('studioAdmin.model') }}</th><th class="p-4">{{ t('studioAdmin.label') }}</th><th class="p-4">{{ t('studioAdmin.protocol') }}</th><th class="p-4">{{ t('studioAdmin.override') }}</th></tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="item in filtered" :key="`${item.group_id}:${item.model}`" class="bg-white dark:bg-dark-900">
              <td class="p-4"><input type="checkbox" :checked="item.enabled" :aria-label="`${t('studioAdmin.visible')} ${item.model}`" :disabled="saving" class="h-4 w-4 accent-primary-600" @change="update({...item,enabled:($event.target as HTMLInputElement).checked})" /></td>
              <td class="p-4">{{ groupName(item.group_id) }}</td>
              <td class="p-4"><span class="break-all font-mono text-xs">{{ item.model }}</span></td>
              <td class="p-4"><input :value="item.label" :aria-label="`${t('studioAdmin.label')} ${item.model}`" class="input min-w-36" maxlength="100" :disabled="saving" @input="update({...item,label:($event.target as HTMLInputElement).value})" /></td>
              <td class="p-4 text-xs">{{ textModels.some(model => model.group_id === item.group_id && model.model === item.model) ? t('studioAdmin.textModel') : item.protocol === 'gemini' ? 'Gemini' : 'Images API' }}</td>
              <td class="p-4"><button v-if="hasOverride(item)" class="text-primary-600 hover:underline dark:text-primary-400" :disabled="saving" @click="reset(item)">{{ t('studioAdmin.reset') }}</button><span v-else class="text-gray-500">{{ t('studioAdmin.automatic') }}</span></td>
            </tr>
            <tr v-if="!filtered.length"><td colspan="6" class="p-12 text-center text-gray-500">{{ t('studioAdmin.empty') }}</td></tr>
          </tbody>
        </table>
      </div>
      <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">{{ t('studioAdmin.pricingNote') }}</p>
      <BaseDialog :show="showAdd" :title="t('studioAdmin.add')" @close="showAdd = false">
        <form id="studio-add-model" class="space-y-5" @submit.prevent="add">
          <div><label for="studio-group" class="mb-2 block text-sm">{{ t('studioAdmin.group') }}</label><Select id="studio-group" v-model="draft.group_id" :options="groupOptions" :aria-label="t('studioAdmin.group')" /></div>
          <div><label for="studio-model" class="mb-2 block text-sm">{{ t('studioAdmin.model') }}</label><input id="studio-model" v-model="draft.model" class="input" maxlength="200" required :placeholder="t('studioAdmin.modelHint')" /></div>
          <div><label for="studio-label" class="mb-2 block text-sm">{{ t('studioAdmin.label') }}</label><input id="studio-label" v-model="draft.label" class="input" maxlength="100" /></div>
          <div><label for="studio-protocol" class="mb-2 block text-sm">{{ t('studioAdmin.protocol') }}</label><Select id="studio-protocol" v-model="draft.protocol" :options="[{value:'images',label:'Images API'},{value:'gemini',label:'Gemini generateContent'}]" :aria-label="t('studioAdmin.protocol')" /></div>
          <p class="text-sm text-gray-500">{{ t('studioAdmin.customNote') }}</p>
        </form>
        <template #footer><button type="submit" form="studio-add-model" class="btn btn-primary" :disabled="!draft.group_id || !draft.model.trim()">{{ t('studioAdmin.add') }}</button></template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { getAll } from '@/api/admin/groups'
import { imageStudioAPI, type StudioConfig, type StudioModel, type StudioModelSetting, type WorkflowModel } from '@/features/image-studio/api'
import { mergeModelSettings, updateModelSetting } from '@/features/image-studio/adminConfig'
import { extractApiErrorMessage } from '@/utils/apiError'
const { t } = useI18n()
const config = ref<StudioConfig>({ enabled: true, models: [] })
const candidates = ref<StudioModel[]>([])
const textModels = ref<WorkflowModel[]>([])
const groups = ref<{id:number;name:string}[]>([])
const loading = ref(true), loaded = ref(false), saving = ref(false), saved = ref(false), error = ref('')
const search = ref(''), groupFilter = ref(0), showAdd = ref(false)
const draft = ref<StudioModelSetting>({ group_id: 0, model: '', label: '', enabled: true, protocol: 'images' })
const groupOptions = computed(() => groups.value.map(group => ({ value: group.id, label: group.name })))
const rows = computed(() => mergeModelSettings([...candidates.value, ...textModels.value], config.value.models))
const filtered = computed(() => rows.value.filter(item => (!groupFilter.value || groupFilter.value === item.group_id) && `${item.model} ${item.label} ${groupName(item.group_id)}`.toLowerCase().includes(search.value.toLowerCase())))
function groupName(id: number) { return groups.value.find(group => group.id === id)?.name || candidates.value.find(model => model.group_id === id)?.group_name || `#${id}` }
function same(a:StudioModelSetting,b:StudioModelSetting) { return a.group_id === b.group_id && a.model === b.model }
function hasOverride(item:StudioModelSetting) { return config.value.models.some(value => same(value,item)) }
function update(item:StudioModelSetting) { saved.value = false; config.value.models = updateModelSetting(config.value.models,item) }
function reset(item:StudioModelSetting) { saved.value = false; config.value.models = config.value.models.filter(value => !same(value,item)) }
function add() { const item = {...draft.value, model:draft.value.model.trim(),label:draft.value.label.trim() || draft.value.model.trim()}; update(item); showAdd.value=false; draft.value={group_id:0,model:'',label:'',enabled:true,protocol:'images'} }
async function load() {
  loading.value=true; error.value=''; saved.value=false
  try { const [settings,catalog,textCatalog,allGroups] = await Promise.all([imageStudioAPI.config(),imageStudioAPI.adminCatalog(),imageStudioAPI.adminWorkflowModels(),getAll()]); config.value=settings; candidates.value=catalog.models; textModels.value=textCatalog.models; groups.value=allGroups; loaded.value=true }
  catch(err) { error.value=extractApiErrorMessage(err,t('common.error')) }
  finally { loading.value=false }
}
async function save() {
  saving.value=true; error.value=''; saved.value=false
  try { config.value=await imageStudioAPI.saveConfig(config.value); saved.value=true; const [images,text] = await Promise.all([imageStudioAPI.adminCatalog(),imageStudioAPI.adminWorkflowModels()]); candidates.value=images.models; textModels.value=text.models }
  catch(err) { error.value=extractApiErrorMessage(err,t('common.error')) }
  finally { saving.value=false }
}
onMounted(load)
</script>
