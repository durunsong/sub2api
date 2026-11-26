<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import Select from '@/components/common/Select.vue';
import Icon from '@/components/icons/Icon.vue';
import { vTooltip } from '@/components/common/tooltip';
import VideoMedia from './VideoMedia.vue';
import VideoSettings from './VideoSettings.vue';
import { inferVideoRatio } from '../parameters';
import type { VideoModel, VideoParameters, VideoReference } from '../types';
const props = defineProps<{
    prompt: string;
    modelID: string;
    models: VideoModel[];
    parameters: VideoParameters;
    references: VideoReference[];
    canGenerate: boolean;
    running: boolean;
    inline: boolean;
    active?: boolean;
}>();
const emit = defineEmits<{
    'update:prompt': [
        value: string
    ];
    'update:modelID': [
        value: string
    ];
    'update:parameters': [
        value: VideoParameters
    ];
    generate: [
    ];
    files: [
        files: File[]
    ];
    remove: [
        id: string
    ];
    preview: [
        index: number
    ];
    prompts: [
    ];
    assets: [
    ];
    stop: [
    ];
}>();
const { t } = useI18n(), upload = ref<HTMLInputElement>(), settings = ref(false), root = ref<HTMLElement>(), settingsTrigger = ref<HTMLButtonElement>(), text = ref<HTMLTextAreaElement>(), mention = ref(false);
const settingsPanel = ref<HTMLElement>(), settingsPosition = ref({ top: '16px', left: '16px' });
function positionSettings() {
    if (!settings.value || !settingsPanel.value || !settingsTrigger.value)
        return;
    const trigger = settingsTrigger.value.getBoundingClientRect();
    const panel = settingsPanel.value.getBoundingClientRect();
    settingsPosition.value = {
        left: `${Math.max(16, Math.min(trigger.right - panel.width, window.innerWidth - panel.width - 16))}px`,
        top: `${Math.max(16, Math.min(trigger.top - panel.height - 12, window.innerHeight - panel.height - 16))}px`
    };
}
async function toggleSettings() {
    settings.value = !settings.value;
    if (settings.value) {
        await nextTick();
        positionSettings();
    }
}
const options = computed(() => props.models.map(m => ({ value: m.id, label: `${m.label} · ${m.group_name}` })));
const summary = computed(() => `${props.parameters.resolution}p · ${inferVideoRatio(props.parameters.size) === 'auto' ? t('videoStudio.auto') : inferVideoRatio(props.parameters.size)} · ${props.parameters.seconds}s`);
function escape(event: KeyboardEvent) { if (settings.value || mention.value) {
    event.stopPropagation();
    event.preventDefault();
    if (settings.value)
        closeSettings();
    mention.value = false;
} }
function closeSettings() { settings.value = false; settingsTrigger.value?.focus(); }
function outside(event: PointerEvent) { if (settings.value && event.target instanceof Node && !settingsTrigger.value?.contains(event.target) && !settingsPanel.value?.contains(event.target))
    settings.value = false; }
watch(() => props.active, active => { if (active === false) { settings.value = false; mention.value = false; } });
document.addEventListener('pointerdown', outside, true);
window.addEventListener('resize', positionSettings);
window.addEventListener('scroll', positionSettings, true);
onBeforeUnmount(() => {
    document.removeEventListener('pointerdown', outside, true);
    window.removeEventListener('resize', positionSettings);
    window.removeEventListener('scroll', positionSettings, true);
});
function input(event: Event) { const value = (event.target as HTMLTextAreaElement).value; emit('update:prompt', value); mention.value = value.slice(0, (event.target as HTMLTextAreaElement).selectionStart).endsWith('@') && !!props.references.length; }
async function insertReference(index: number) { const el = text.value; if (!el)
    return; const pos = el.selectionStart; const before = props.prompt.slice(0, pos); const label = t('videoStudio.referenceNumber', { n: index + 1 }); const value = before.replace(/@$/, '') + label + ' ' + props.prompt.slice(pos); emit('update:prompt', value); mention.value = false; await nextTick(); el.focus(); el.setSelectionRange(before.length + label.length, before.length + label.length); }
function paste(event: ClipboardEvent) { const files = Array.from(event.clipboardData?.files || []).filter(f => f.type.startsWith('image/')); if (files.length) {
    event.preventDefault();
    emit('files', files);
} }
</script>
<template>
<div ref="root" class="video-composer-wrap" :class="{inline}" @keydown.esc="escape">
  <form class="video-composer" @submit.prevent="emit('generate')" @paste="paste">
   <div class="reference-strip" @dragover.prevent @drop.prevent="emit('files',Array.from($event.dataTransfer?.files || []))">
    <button class="reference-add" type="button" :aria-label="t('videoStudio.addReference')" @click="upload?.click()">
<Icon name="plus" />
</button>
    <div v-for="(item,index) in references" :key="item.id" class="video-reference group">
<button type="button" :aria-label="t('videoStudio.previewReference',{n:index+1})" @click="emit('preview',index)">
<VideoMedia :blob="item.blob" kind="image" :alt="item.name" />
</button>
<span>{{parameters.mode==='frames'?t(index===0?'videoStudio.firstFrame':'videoStudio.lastFrame'):index+1}}</span>
<button v-tooltip="t('videoStudio.removeReference')" type="button" class="reference-remove" :aria-label="t('videoStudio.removeReference')" @click="emit('remove',item.id)">
<Icon name="x" size="xs" />
</button>
</div>
    <span class="text-xs text-gray-500 dark:text-gray-400">{{t('videoStudio.referenceHint',{count:parameters.mode==='frames'?2:9})}}</span>
   </div>
   <input ref="upload" hidden type="file" accept="image/png,image/jpeg,image/webp" multiple @change="emit('files',Array.from(($event.target as HTMLInputElement).files || [])); ($event.target as HTMLInputElement).value=''" />
   <div class="relative">
<textarea ref="text" :value="prompt" maxlength="32000" class="video-prompt" :placeholder="t('videoStudio.promptPlaceholder')" :aria-label="t('videoStudio.prompt')" @input="input" @keydown.ctrl.enter.prevent="emit('generate')" @keydown.meta.enter.prevent="emit('generate')">
</textarea>
<div v-if="mention" class="mention-list" role="listbox" :aria-label="t('videoStudio.references')">
<button v-for="(reference,index) in references" :key="reference.id" type="button" role="option" :aria-selected="false" @click="insertReference(index)">{{t('videoStudio.referenceNumber',{n:index+1})}} · {{reference.name}}</button>
</div>
</div>
   <div class="video-composer-tools">
    <span class="video-kind">{{t('videoStudio.videoTab')}}</span>
    <Select :model-value="modelID" :options="options" searchable :placeholder="t('videoStudio.selectModel')" :aria-label="t('videoStudio.selectModel')" class="video-model-select" @update:model-value="emit('update:modelID',String($event || ''))" />
    <div>
<button ref="settingsTrigger" type="button" class="video-tool video-parameter-trigger" :aria-label="t('videoStudio.parameters')" :aria-expanded="settings" @click="toggleSettings">
<Icon name="cog" size="sm" />
<span>{{summary}}</span>
</button>
<Teleport to="body">
<div v-if="settings" ref="settingsPanel" class="video-settings-popover" :style="settingsPosition" role="region" :aria-label="t('videoStudio.parameters')" @keydown.esc="escape">
<VideoSettings :model-value="parameters" @update:model-value="emit('update:parameters',$event)" />
</div>
</Teleport>
</div>
    <button v-tooltip="t('videoStudio.prompts')" type="button" class="video-tool" :aria-label="t('videoStudio.prompts')" @click="emit('prompts')">
<Icon name="document" size="sm" />
</button>
    <button v-tooltip="t('videoStudio.assets')" type="button" class="video-tool" :aria-label="t('videoStudio.assets')" @click="emit('assets')">
<Icon name="inbox" size="sm" />
</button>
    <button v-if="running" type="button" class="video-send" :aria-label="t('videoStudio.pause')" @click="emit('stop')">
<Icon name="x" size="sm" />
</button>
<button v-else type="submit" class="video-send" :disabled="!canGenerate" :aria-label="t('videoStudio.generate')">
<Icon name="arrowUp" size="sm" />
</button>
   </div>
  </form>
 </div>
</template>
<style scoped>
.video-composer-wrap {
  @apply shrink-0 px-3 py-2 sm:px-6}
.video-composer-wrap:not(.inline) {
  background:var(--video-background); border-top:1px solid rgba(148,163,184,.15)}
.video-composer-wrap.inline {
  @apply mx-auto w-full max-w-[1048px] px-4 pb-6 pt-3}
.video-composer {
  @apply pointer-events-auto mx-auto max-w-[1000px] rounded-2xl border border-gray-200 bg-white px-3 py-2 dark:border-dark-600 dark:bg-dark-800;
  box-shadow:0 16px 50px rgba(0,0,0,.09)}
.reference-strip {
  @apply flex min-h-14 items-center gap-3 overflow-x-auto px-1 py-1}
.reference-add {
  @apply grid h-12 w-11 shrink-0 -rotate-6 place-items-center rounded-sm border border-dashed border-gray-300 bg-gray-50 text-gray-500 transition-transform hover:rotate-0 dark:border-dark-500 dark:bg-dark-700}
.video-reference {
  @apply relative h-12 w-11 shrink-0 overflow-hidden rounded-sm border border-gray-200 dark:border-dark-600}
.video-reference>button:first-child {
  @apply block h-full w-full}
.video-reference :deep(img) {
  @apply h-full w-full object-cover}
.video-reference>span {
  @apply absolute bottom-0 left-0 bg-black/70 px-1 text-[10px] text-white}
.reference-remove {
  @apply absolute right-0 top-0 rounded-full bg-white/90 p-1 text-gray-900 opacity-0 group-hover:opacity-100 group-focus-within:opacity-100}
.video-prompt {
  @apply block min-h-14 max-h-40 w-full resize-y border-0 bg-transparent px-1 py-1 text-[15px] leading-6 outline-none placeholder:text-gray-500 dark:placeholder:text-gray-400}
.video-composer-tools {
  @apply mt-1.5 flex flex-wrap items-center gap-2}
.video-kind {
  @apply shrink-0 rounded-full bg-gray-100 px-3 py-1.5 text-sm dark:bg-dark-700}
.video-model-select {
  @apply min-w-0 flex-1 sm:max-w-[220px] sm:min-w-[9rem]}
.video-tool {
  @apply inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-full border border-gray-200 px-2.5 text-xs hover:bg-gray-100 focus-visible:outline-primary-500 dark:border-dark-600 dark:hover:bg-dark-700}
.video-send {
  @apply ml-auto grid h-9 w-9 shrink-0 place-items-center rounded-full bg-gray-900 text-white disabled:opacity-35 dark:bg-gray-100 dark:text-gray-900}
.video-settings-popover {
  @apply fixed z-40 w-[min(340px,calc(100vw-32px))] max-h-[min(560px,calc(100dvh-32px))] overflow-auto rounded-xl border border-gray-200 bg-white p-4 text-gray-900 shadow-xl dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100}
.mention-list {
  @apply absolute bottom-full left-0 z-50 max-h-40 overflow-auto rounded-lg border bg-white p-1 shadow-lg dark:border-dark-600 dark:bg-dark-800}
.mention-list button {
  @apply block w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700}
@media(max-width:640px) {
  .video-parameter-trigger span {
  display:none}
.video-kind {
  display:none}
.video-composer-tools {
  gap:6px}
.video-tool {
  padding-left:8px;
  padding-right:8px}
}
</style>
