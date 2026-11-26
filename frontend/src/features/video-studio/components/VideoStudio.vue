<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import Icon from '@/components/icons/Icon.vue';
import PromptLibraryDialog from '@/features/prompt-library/PromptLibraryDialog.vue';
import BaseDialog from '@/components/common/BaseDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import ImagePreview from '@/features/image-studio/components/ImagePreview.vue';
import { createImageStudioDB } from '@/features/image-studio/storage';
import { useAuthStore } from '@/stores/auth';
import VideoComposer from './VideoComposer.vue';
import VideoEntry from './VideoEntry.vue';
import VideoMedia from './VideoMedia.vue';
import { useVideoStudio } from '../useVideoStudio';
import { videoFilename } from '../media';
import type { VideoAsset, VideoRecord, VideoReference } from '../types';
const props = withDefaults(defineProps<{
    active?: boolean;
}>(), { active: true });
const root = ref<HTMLElement>();
watch(() => props.active, value => { if (!value)
    root.value?.querySelectorAll('video').forEach(video => video.pause()); });
const { t } = useI18n(), auth = useAuthStore(), studio = reactive(useVideoStudio());
const historyOpen = ref(false), promptsOpen = ref(false), assetsOpen = ref(false), feed = ref<HTMLElement>(), libraryUpload = ref<HTMLInputElement>();
const search = ref(''), promptTitle = ref(''), promptText = ref(''), imageAssets = ref<VideoAsset[]>([]);
const preview = ref<{
    images: Blob[];
    index: number;
    captions: string[];
} | null>(null);
const videoPreview = ref<VideoAsset | null>(null);
const confirmation = ref<{
    message: string;
    action: () => void | Promise<void>;
} | null>(null);
const blank = computed(() => !studio.entries.length);
const filteredSessions = computed(() => studio.sessions.filter(s => s.title.toLowerCase().includes(search.value.toLowerCase())));
const presets = ['product', 'landscape', 'story'];
watch([() => studio.sessionID, () => studio.entries.length], async () => {
    await nextTick();
    const list = feed.value, last = list?.querySelector<HTMLElement>('.video-entry:last-child');
    if (!props.active || !list || !last) return;
    const bounds = list.getBoundingClientRect(), row = last.getBoundingClientRect();
    // Scroll only the result pane; scrollIntoView also moves the page on narrow screens.
    if (row.height > bounds.height || row.top < bounds.top) list.scrollTop += row.top - bounds.top;
    else if (row.bottom > bounds.bottom) list.scrollTop += row.bottom - bounds.bottom;
});
watch(() => auth.user?.id, () => { imageAssets.value = []; preview.value = null; videoPreview.value = null; confirmation.value = null; historyOpen.value = false; promptsOpen.value = false; assetsOpen.value = false; });
function showReferences(refs: VideoReference[], index: number) { preview.value = { images: refs.map(r => r.blob), index, captions: refs.map(r => r.name) }; }
function download(blob: Blob, name: string) { const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = name; a.click(); setTimeout(() => URL.revokeObjectURL(url), 1000); }
function retry(row: VideoRecord) { confirmation.value = { message: t('videoStudio.confirmRegenerate'), action: async () => { studio.reuse(row); await nextTick(); await studio.submit(); } }; }
function deleteSession(id: string) { confirmation.value = { message: t('videoStudio.confirmDelete'), action: () => studio.deleteSession(id) }; }
function deleteLibrary(kind: 'assets' | 'prompts', id: string) { confirmation.value = { message: t('videoStudio.confirmDeleteItem'), action: () => studio.removeLibrary(kind, id) }; }
async function confirm() { const action = confirmation.value?.action; confirmation.value = null; await action?.(); }
async function openAssets() {
    assetsOpen.value = true;
    const id = auth.user?.id;
    if (!id)
        return;
    const imageDB = createImageStudioDB(id);
    try {
        const images = await imageDB.listAssets();
        if (auth.user?.id === id)
            imageAssets.value = images.map(a => ({ id: `image:${a.id}`, name: a.name, kind: 'image', blob: a.image, createdAt: a.createdAt }));
    }
    catch {
        studio.error = 'videoStudio.imageLibraryFailed';
    }
    finally {
        imageDB.close();
    }
}
function useAsset(asset: VideoAsset) { void studio.addReferences([new File([asset.blob], asset.name, { type: asset.blob.type })]); assetsOpen.value = false; }
async function savePrompt() { await studio.savePrompt(promptTitle.value, promptText.value); promptTitle.value = ''; promptText.value = ''; }
function pause() { for (const row of studio.entries)
    studio.stop(row.id); }
</script>
<template>
<section ref="root" class="video-studio" data-video-studio>
  <div v-if="studio.loading" role="status" class="p-8 text-center">{{t('common.loading')}}</div>
  <template v-else>
   <div v-if="studio.error || studio.storageError" role="alert" class="video-alert">
<span>{{t(studio.storageError || studio.error)}}</span>
<button type="button" class="underline" @click="studio.load()">{{t('common.refresh')}}</button>
</div>
   <div v-if="studio.notice" role="status" class="px-5 py-2 text-sm text-green-700 dark:text-green-300">{{t(studio.notice)}}</div>
   <div v-if="!studio.enabled || !studio.models.length" class="px-5 py-3 text-sm text-gray-600 dark:text-gray-400">{{t(!studio.enabled?'videoStudio.disabled':'videoStudio.noModels')}}</div>
   <div class="video-workspace" :class="{'is-empty':blank}">
    <div v-if="blank" class="video-empty">
     <button type="button" class="history-trigger" @click="historyOpen=true">
<Icon name="clock" size="sm" />{{t('videoStudio.history')}}</button>
     <h2>{{t('videoStudio.greeting')}}</h2>
     <div class="video-presets">
<button v-for="preset in presets" :key="preset" type="button" @click="studio.prompt=t(`videoStudio.preset_${preset}`)">{{t(`videoStudio.preset_${preset}`)}}</button>
</div>
    </div>
    <template v-else>
<header class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-gray-100 bg-white px-4 py-3 sm:px-6 dark:border-dark-700 dark:bg-dark-900">
<h2 class="text-sm font-medium text-gray-500">{{t('videoStudio.history')}} <span class="ml-1 tabular-nums">{{studio.entries.length}}</span></h2>
<div class="flex gap-2">
<button type="button" class="session-action" @click="studio.newSession">
<Icon name="plus" size="sm" />{{t('videoStudio.newSession')}}</button>
<button type="button" class="session-action" @click="historyOpen=true">
<Icon name="clock" size="sm" />{{t('videoStudio.history')}}</button>
</div>
</header>
     <div ref="feed" class="video-feed">
<div class="mx-auto w-full max-w-[1000px]">

<div class="flex flex-col gap-4">
<VideoEntry v-for="row in studio.entries" :key="row.id" :record="row" :running="studio.active.includes(row.id)" :unsaved="!!studio.unsaved[row.id]" @reuse="studio.reuse(row)" @retry="retry(row)" @resume="studio.resume(row)" @pause="studio.stop(row.id)" @download="row.video && download(row.video,videoFilename(row.id,row.video))" @asset="studio.saveAsset(row)" @save="studio.saveUnsaved(row)" @preview="showReferences(row.references,$event)" />
</div>
</div>
</div>
    </template>
     <VideoComposer v-model:prompt="studio.prompt" v-model:model-i-d="studio.modelID" v-model:parameters="studio.parameters" :models="studio.models" :references="studio.references" :can-generate="studio.canGenerate" :running="studio.currentRunning" :inline="blank" :active="active" @generate="studio.submit" @files="studio.addReferences" @remove="studio.removeReference" @preview="showReferences(studio.references,$event)" @prompts="promptsOpen=true" @assets="openAssets" @stop="pause" />
   </div>
  </template>
  <BaseDialog :show="historyOpen" :title="t('videoStudio.history')" @close="historyOpen=false">
<div class="space-y-4">
<button type="button" class="btn btn-secondary" @click="studio.newSession();historyOpen=false">
<Icon name="plus" size="sm" />{{t('videoStudio.newSession')}}</button>
<input v-model="search" class="input" :placeholder="t('videoStudio.search')" :aria-label="t('videoStudio.search')" />
<div v-if="!filteredSessions.length" class="py-10 text-center text-gray-500">{{t('videoStudio.noHistory')}}</div>
<div v-for="s in filteredSessions" :key="s.id" class="flex items-center gap-2 border-b border-gray-100 py-3 dark:border-dark-700">
<button type="button" class="min-w-0 flex-1 text-left" @click="studio.openSession(s.id);historyOpen=false">
<strong class="block truncate text-sm">{{s.title}}</strong>
<span class="text-xs text-gray-500">{{new Date(s.updatedAt).toLocaleString()}} · {{t('videoStudio.recordCount',{count:s.count})}}</span>
</button>
<button type="button" class="session-action" :aria-label="t('videoStudio.deleteSession')" @click="deleteSession(s.id)">
<Icon name="trash" size="sm" />
</button>
</div>
</div>
</BaseDialog>
  <PromptLibraryDialog :show="promptsOpen" @close="promptsOpen=false" @use="studio.prompt=$event">
<template #custom>
<div class="space-y-4">
<div v-for="p in studio.prompts" :key="p.id" class="flex items-start gap-3 border-b py-3 dark:border-dark-700">
<button type="button" class="min-w-0 flex-1 text-left" @click="studio.prompt=p.prompt;promptsOpen=false">
<strong>{{p.title}}</strong>
<p class="mt-1 line-clamp-3 text-sm text-gray-500">{{p.prompt}}</p>
</button>
<button type="button" :aria-label="t('common.delete')" @click="deleteLibrary('prompts',p.id)">
<Icon name="trash" size="sm" />
</button>
</div>
<p v-if="!studio.prompts.length" class="text-sm text-gray-500">{{t('videoStudio.noPrompts')}}</p>
<form class="space-y-3 border-t pt-4 dark:border-dark-700" @submit.prevent="savePrompt">
<label class="block text-sm">{{t('videoStudio.promptTitle')}}<input v-model="promptTitle" class="input mt-2" maxlength="120" required />
</label>
<label class="block text-sm">{{t('videoStudio.prompt')}}<textarea v-model="promptText" class="input mt-2 min-h-24" maxlength="32000" required>
</textarea>
</label>
<div class="flex gap-2">
<button type="button" class="btn btn-secondary" @click="promptText=studio.prompt">{{t('videoStudio.useCurrentPrompt')}}</button>
<button type="submit" class="btn btn-primary" :disabled="!promptTitle.trim() || !promptText.trim()">{{t('common.save')}}</button>
</div>
</form>
</div>
</template>
</PromptLibraryDialog>
  <BaseDialog :show="assetsOpen" :title="t('videoStudio.assets')" @close="assetsOpen=false">
<div class="space-y-4">
<button type="button" class="btn btn-secondary" @click="libraryUpload?.click()">
<Icon name="upload" size="sm" />{{t('videoStudio.uploadAsset')}}</button>
<input ref="libraryUpload" hidden type="file" multiple accept="image/png,image/jpeg,image/webp,video/mp4,video/webm" @change="studio.uploadAssets(Array.from(($event.target as HTMLInputElement).files||[]));($event.target as HTMLInputElement).value=''" />
<p class="text-xs text-gray-500">{{t('videoStudio.localOnly')}}</p>
<div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
<div v-for="a in [...studio.assets,...imageAssets]" :key="a.id" class="min-w-0">
<button type="button" class="block aspect-square w-full overflow-hidden rounded-lg bg-gray-100 dark:bg-dark-700" :aria-label="t('videoStudio.previewAsset',{name:a.name})" @click="a.kind==='image'?preview={images:[a.blob],index:0,captions:[a.name]}:videoPreview=a">
<VideoMedia :blob="a.blob" :kind="a.kind" :alt="a.name" class="h-full w-full object-cover" />
</button>
<p class="mt-2 truncate text-xs">{{a.name}}</p>
<div class="mt-1 flex flex-wrap gap-2">
<button v-if="a.kind==='image'" type="button" class="text-xs text-primary-600" @click="useAsset(a)">{{t('videoStudio.useReference')}}</button>
<button v-else type="button" class="text-xs text-primary-600" @click="download(a.blob,videoFilename(a.id,a.blob))">{{t('videoStudio.download')}}</button>
<button v-if="!a.id.startsWith('image:')" type="button" class="text-xs text-red-600" @click="deleteLibrary('assets',a.id)">{{t('common.delete')}}</button>
</div>
</div>
</div>
<p v-if="!studio.assets.length&&!imageAssets.length" class="py-8 text-center text-gray-500">{{t('videoStudio.noAssets')}}</p>
</div>
</BaseDialog>
  <BaseDialog :show="!!videoPreview" :title="videoPreview?.name || t('videoStudio.videoTab')" @close="videoPreview=null">
<VideoMedia v-if="videoPreview" :blob="videoPreview.blob" kind="video" controls class="w-full" />
</BaseDialog>
  <ImagePreview v-if="preview" show :images="preview.images" :initial-index="preview.index" :captions="preview.captions" @close="preview=null" @download="download($event,`reference.${$event.type.split('/')[1] || 'png'}`)" />
  <ConfirmDialog :show="!!confirmation" :title="t('common.confirm')" :message="confirmation?.message || ''" @confirm="confirm" @cancel="confirmation=null" />
 </section>
</template>
<style scoped>
.video-studio {
  --video-background:#fff;
  @apply flex min-h-[640px] min-w-0 flex-col overflow-hidden rounded-xl border border-gray-200 bg-white text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100;
  height:calc(100dvh - 208px)}
:global(.dark .video-studio) {
  --video-background:#111827}
.video-workspace {
  @apply relative flex min-h-0 min-w-0 flex-1 flex-col;
   background-image:radial-gradient(#e5e7eb 1px,transparent 1px);
  background-size:16px 16px}
:global(.dark .video-workspace) {
  background-image:radial-gradient(rgba(245,245,244,.16) 1px,transparent 1px)}
.video-workspace.is-empty { justify-content:center; overflow-y:auto }
.video-empty {
  @apply relative flex min-h-0 flex-col items-center justify-center px-4 pt-16 sm:px-6}
.video-empty h2 {
  @apply text-center text-3xl font-semibold tracking-tight sm:text-[38px]}
.video-presets {
  @apply mt-6 flex max-w-[900px] flex-wrap items-center justify-center gap-2}
.video-presets button {
  @apply rounded-full border border-gray-200 bg-white px-4 py-2 text-sm text-gray-600 shadow-sm hover:border-gray-400 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300}
.history-trigger {
  @apply absolute right-4 top-4 inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-dark-800}
.video-feed {
  @apply min-h-0 flex-1 overflow-y-auto px-4 py-5 sm:px-6;
  overflow-anchor:none}
.session-action {
  @apply inline-flex min-h-9 items-center justify-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-sm hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-800 dark:hover:bg-dark-700}
.video-alert {
  @apply flex flex-wrap items-center justify-between gap-2 bg-amber-50 px-5 py-3 text-sm text-amber-900 dark:bg-amber-950 dark:text-amber-200}
@media(max-width:640px) {
  .video-studio {
  min-height:620px;
  height:calc(100dvh - 180px)}
.video-empty h2 {
  font-size:28px}
.video-presets {
  display:none}
}
@media(max-height:700px) {
  .video-empty {
  justify-content:flex-start}
.video-presets {
  display:none}
.video-studio {
  min-height:550px}
}
</style>
