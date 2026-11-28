<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getAll } from '@/api/admin/groups';
import { extractApiErrorMessage } from '@/utils/apiError';
import Select from '@/components/common/Select.vue';
import BaseDialog from '@/components/common/BaseDialog.vue';
import Icon from '@/components/icons/Icon.vue';
import { videoStudioAPI } from '../api';
import { mergeVideoModels, setVideoVisibility } from '../adminConfig';
import type { VideoConfig, VideoModel, VideoModelSetting } from '../types';
const { t } = useI18n(), config = ref<VideoConfig>({ enabled: true, models: [] }), models = ref<VideoModel[]>([]), groups = ref<{
    id: number;
    name: string;
}[]>([]);
const loading = ref(true), saving = ref(false), loaded = ref(false), saved = ref(false), error = ref(''), search = ref(''), group = ref(0), addOpen = ref(false);
let savedTimer: ReturnType<typeof setTimeout> | undefined, disposed = false;
onBeforeUnmount(() => { disposed = true; clearTimeout(savedTimer); });
const draft = ref<VideoModelSetting>({ group_id: 0, model: '', label: '', enabled: true, protocol: 'grok' });
const rows = computed(() => mergeVideoModels(models.value, config.value.models));
const options = computed(() => groups.value.map(g => ({ value: g.id, label: g.name })));
const filtered = computed(() => rows.value.filter(r => (!group.value || r.group_id === group.value) && `${r.model} ${r.label} ${groupName(r.group_id)}`.toLowerCase().includes(search.value.toLowerCase())));
const editable = computed(() => filtered.value.filter(r => groups.value.some(g => g.id === r.group_id)));
function groupName(id: number) { return groups.value.find(g => g.id === id)?.name || models.value.find(m => m.group_id === id)?.group_name || `#${id}`; }
function same(a: VideoModelSetting, b: VideoModelSetting) { return a.group_id === b.group_id && a.model === b.model; }
function overridden(row: VideoModelSetting) { return config.value.models.some(s => same(row, s)); }
function update(row: VideoModelSetting) { saved.value = false; const index = config.value.models.findIndex(s => same(row, s)); if (index < 0)
    config.value.models.push(row);
else
    config.value.models[index] = row; }
function toggle(enabled: boolean) { config.value.models = setVideoVisibility(config.value.models, filtered.value, enabled, groups.value.map(g => g.id)); saved.value = false; }
function reset(row: VideoModelSetting) { config.value.models = config.value.models.filter(s => !same(row, s)); saved.value = false; }
function add() { if (!draft.value.group_id || !draft.value.model.trim())
    return; update({ ...draft.value, model: draft.value.model.trim(), label: draft.value.label.trim() || draft.value.model.trim() }); addOpen.value = false; draft.value = { group_id: 0, model: '', label: '', enabled: true, protocol: 'grok' }; }
async function load() { loading.value = true; error.value = ''; saved.value = false; try {
    const [settings, catalog, allGroups] = await Promise.all([videoStudioAPI.config(), videoStudioAPI.adminCatalog(), getAll()]);
    if (disposed)
        return;
    config.value = settings;
    models.value = catalog.models;
    groups.value = allGroups.filter(g => ['grok', 'composite', 'openai'].includes(g.platform));
    loaded.value = true;
}
catch (err) {
    if (!disposed)
        error.value = extractApiErrorMessage(err, t('common.error'));
}
finally {
    loading.value = false;
} }
async function save() { saving.value = true; error.value = ''; saved.value = false; try {
    const settings = await videoStudioAPI.saveConfig(config.value);
    if (disposed)
        return;
    config.value = settings;
    saved.value = true;
    clearTimeout(savedTimer);
    savedTimer = setTimeout(() => { saved.value = false; }, 5000);
    models.value = (await videoStudioAPI.adminCatalog()).models;
}
catch (err) {
    if (!disposed)
        error.value = extractApiErrorMessage(err, t('common.error'));
}
finally {
    saving.value = false;
} }
onMounted(load);
</script>
<template>
<div class="mx-auto max-w-6xl space-y-6">
  <header class="flex flex-wrap items-start justify-between gap-4">
<div>
<h1 class="text-2xl font-semibold">{{t('videoStudio.videoTab')}}</h1>
<p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-400">{{t('videoStudio.adminDescription')}}</p>
</div>
<RouterLink class="btn btn-secondary" to="/image-studio?tab=video">{{t('videoStudio.open')}}</RouterLink>
</header>
  <p v-if="error" role="alert" class="rounded-lg bg-red-50 p-4 text-sm text-red-700 dark:bg-red-950 dark:text-red-200">{{error}}</p>
<p v-if="saved" role="status" class="text-sm text-green-700 dark:text-green-300">{{t('studioAdmin.saved')}}</p>
  <div class="flex flex-wrap items-center justify-between gap-4 border-y border-gray-200 py-4 dark:border-dark-700">
<label class="flex items-center gap-3 text-sm font-medium">
<input v-model="config.enabled" type="checkbox" class="h-4 w-4 accent-primary-600" :disabled="loading||saving" />{{t('videoStudio.adminEnabled')}}</label>
<div class="flex gap-3">
<button type="button" class="btn btn-secondary" :disabled="loading||saving" @click="load">{{t('common.refresh')}}</button>
<button type="button" class="btn btn-primary" :disabled="loading||saving||!loaded" @click="save">{{t(saving?'common.saving':'common.save')}}</button>
</div>
</div>
  <div class="flex flex-wrap gap-3">
<input v-model="search" class="input max-w-md" :aria-label="t('studioAdmin.search')" :placeholder="t('studioAdmin.search')" />
<Select v-model="group" :options="[{value:0,label:t('studioAdmin.allGroups')},...options]" class="w-full sm:w-60" :aria-label="t('studioAdmin.group')" />
<button type="button" class="btn btn-secondary" :disabled="loading||saving" @click="addOpen=true">
<Icon name="plus" size="sm" />{{t('videoStudio.adminAdd')}}</button>
</div>
  <div class="flex flex-wrap items-center gap-3">
<button type="button" class="btn btn-secondary btn-sm" :disabled="loading||saving||!editable.length" @click="toggle(true)">{{t('studioAdmin.showFiltered')}}</button>
<button type="button" class="btn btn-secondary btn-sm" :disabled="loading||saving||!editable.length" @click="toggle(false)">{{t('studioAdmin.hideFiltered')}}</button>
<span class="text-sm text-gray-500">{{t('studioAdmin.bulkScope',{count:editable.length})}}</span>
</div>
  <div v-if="loading" role="status" class="py-16 text-center">{{t('common.loading')}}</div>
<div v-else class="overflow-x-auto rounded-xl border border-gray-200 dark:border-dark-700">
<table class="w-full min-w-[680px] text-left text-sm">
<thead class="bg-gray-50 text-gray-600 dark:bg-dark-800 dark:text-gray-400">
<tr>
<th class="p-4">{{t('studioAdmin.visible')}}</th>
<th class="p-4">{{t('studioAdmin.group')}}</th>
<th class="p-4">{{t('studioAdmin.model')}}</th>
<th class="p-4">{{t('studioAdmin.label')}}</th>
<th class="p-4">{{t('studioAdmin.protocol')}}</th>
<th class="p-4">{{t('studioAdmin.override')}}</th>
</tr>
</thead>
<tbody class="divide-y divide-gray-100 dark:divide-dark-700">
<tr v-for="r in filtered" :key="`${r.group_id}:${r.model}`" class="bg-white dark:bg-dark-900">
<td class="p-4">
<input type="checkbox" :checked="r.enabled" :disabled="saving||!groups.some(g=>g.id===r.group_id)" :aria-label="`${t('studioAdmin.visible')} ${r.model}`" @change="update({...r,enabled:($event.target as HTMLInputElement).checked})" />
</td>
<td class="p-4">{{groupName(r.group_id)}}</td>
<td class="break-all p-4 font-mono text-xs">{{r.model}}</td>
<td class="p-4">
<input :value="r.label" class="input min-w-36" maxlength="100" :disabled="saving||!groups.some(g=>g.id===r.group_id)" :aria-label="`${t('studioAdmin.label')} ${r.model}`" @input="update({...r,label:($event.target as HTMLInputElement).value})" />
</td>
<td class="p-4 text-xs">{{r.protocol==='muse'?'Muse':t('videoStudio.adminProtocol')}}</td>
<td class="p-4">
<button v-if="overridden(r)" type="button" class="text-primary-600 hover:underline" :disabled="saving" @click="reset(r)">{{t('studioAdmin.reset')}}</button>
<span v-else class="text-gray-500">{{t('studioAdmin.automatic')}}</span>
</td>
</tr>
<tr v-if="!filtered.length">
<td colspan="6" class="p-12 text-center text-gray-500">{{t('studioAdmin.empty')}}</td>
</tr>
</tbody>
</table>
</div>
  <p class="text-sm leading-6 text-gray-500 dark:text-gray-400">{{t('videoStudio.adminPricing')}}</p>
  <BaseDialog :show="addOpen" :title="t('videoStudio.adminAdd')" @close="addOpen=false">
<form id="video-add-model" class="space-y-5" @submit.prevent="add">
<div>
<label for="video-group" class="mb-2 block text-sm">{{t('studioAdmin.group')}}</label>
<Select id="video-group" v-model="draft.group_id" :options="options" />
</div>
<label class="block text-sm">{{t('studioAdmin.model')}}<input v-model="draft.model" class="input mt-2" maxlength="200" required />
</label>
<label class="block text-sm">{{t('studioAdmin.label')}}<input v-model="draft.label" class="input mt-2" maxlength="100" />
</label>
<label class="block text-sm">{{t('studioAdmin.protocol')}}<Select v-model="draft.protocol" class="mt-2" :options="[{value:'grok',label:'Grok'},{value:'muse',label:'Muse'}]" /></label>
<p class="text-sm text-gray-500">{{t('videoStudio.adminHint')}}</p>
</form>
<template #footer>
<button type="submit" form="video-add-model" class="btn btn-primary" :disabled="!draft.group_id||!draft.model.trim()">{{t('studioAdmin.add')}}</button>
</template>
</BaseDialog>
 </div>
</template>
