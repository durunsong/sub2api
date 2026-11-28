<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { computeVideoSize, inferVideoRatio, normalizeParameters, normalizeModelParameters, videoRatios } from '../parameters';
import type { VideoParameters } from '../types';
const props = defineProps<{
    modelValue: VideoParameters;
    protocol?: string;
}>();
const emit = defineEmits<{
    'update:modelValue': [
        value: VideoParameters
    ];
}>();
const { t } = useI18n();
const ratio = computed(() => inferVideoRatio(props.modelValue.size));
const dimensions = computed(() => props.modelValue.size.split('x'));
function update(patch: Partial<VideoParameters>) { emit('update:modelValue', normalizeModelParameters(normalizeParameters({ ...props.modelValue, ...patch }), props.protocol)); }
function resolution(value: string) { update({ resolution: value, size: computeVideoSize(value, ratio.value) }); }
function dimension(index: number, event: Event) { const values = [...dimensions.value]; values[index] = String(Math.max(1, Math.min(8192, Number((event.target as HTMLInputElement).value) || 720))); update({ size: values.join('x') }); }
</script>
<template>
<div class="video-settings space-y-5">
  <fieldset>
<legend>{{t('videoStudio.resolution')}}</legend>
<div class="grid grid-cols-4 gap-2">
<button v-for="p in ['480','720','1080']" :key="p" type="button" class="video-pill" :aria-pressed="modelValue.resolution===p" @click="resolution(p)">{{p}}p</button>
<label v-if="protocol!=='muse'" class="video-resolution">
<input :value="modelValue.resolution" type="number" min="1" max="4320" :aria-label="t('videoStudio.customResolution')" @change="resolution(($event.target as HTMLInputElement).value)" />
<span>p</span>
</label>
</div>
</fieldset>
  <fieldset v-if="protocol!=='muse'">
<legend>{{t('videoStudio.size')}}</legend>
<div class="grid grid-cols-2 gap-3">
<label v-for="(name,index) in ['W','H']" :key="name" class="video-dimension">
<span>{{name}}</span>
<input type="number" min="1" max="8192" :value="dimensions[index] || ''" :disabled="ratio==='auto'" :aria-label="t(index===0?'videoStudio.width':'videoStudio.height')" @change="dimension(index,$event)" />
</label>
</div>
</fieldset>
  <fieldset>
<legend>{{t('videoStudio.ratio')}}</legend>
<div class="grid grid-cols-4 gap-2">
<button v-for="r in (protocol==='muse'?['16:9','9:16','auto']:videoRatios)" :key="r" type="button" class="video-ratio" :aria-pressed="ratio===r" @click="update({size:computeVideoSize(modelValue.resolution,r)})">
<span v-if="r!=='auto'" class="ratio-shape" :style="{aspectRatio:r.replace(':','/'),width:Number(r.split(':')[0])>=Number(r.split(':')[1])?'26px':'18px'}">
</span>
<span>{{r==='auto'?t('videoStudio.auto'):r}}</span>
</button>
</div>
</fieldset>
  <fieldset>
<legend>{{t('videoStudio.seconds')}}</legend>
<div class="flex items-center gap-3">
<input class="min-w-0 flex-1 accent-primary-600" type="range" min="4" :max="protocol==='muse'?15:30" :value="modelValue.seconds" :aria-label="t('videoStudio.seconds')" @input="update({seconds:Number(($event.target as HTMLInputElement).value)})" />
<input class="input w-16" type="number" min="4" :max="protocol==='muse'?15:30" :value="modelValue.seconds" :aria-label="t('videoStudio.secondsInput')" @change="update({seconds:Number(($event.target as HTMLInputElement).value)})" />
<span>s</span>
</div>
</fieldset>
  <fieldset>
<legend>{{t('videoStudio.mode')}}</legend>
<div class="grid grid-cols-2 gap-2">
<button v-for="mode in (protocol==='muse' ? ['reference'] as const : ['frames','reference'] as const)" :key="mode" type="button" class="video-pill" :aria-pressed="modelValue.mode===mode" @click="update({mode})">{{t(`videoStudio.${mode}`)}}</button>
</div>
</fieldset>
  <p class="text-xs leading-5 text-gray-600 dark:text-gray-400">{{t('videoStudio.providerParameters')}}</p>
 </div>
</template>
<style scoped>
legend {
  @apply mb-2.5 text-xs font-medium text-gray-600 dark:text-gray-400}
.video-pill {
  @apply min-h-9 rounded-full border border-gray-300 px-2 text-sm hover:bg-gray-100 focus-visible:outline-primary-500 dark:border-dark-600 dark:hover:bg-dark-700}
.video-pill[aria-pressed=true],.video-ratio[aria-pressed=true] {
  @apply border-gray-900 dark:border-gray-100}
.video-resolution {
  @apply flex overflow-hidden rounded-full border border-gray-300 text-sm dark:border-dark-600}
.video-resolution input {
  @apply min-w-0 w-full bg-transparent px-1 text-center}
.video-resolution span {
  @apply self-center pr-2}
.video-dimension {
  @apply flex h-9 items-center rounded-xl bg-gray-100 text-sm dark:bg-dark-700}
.video-dimension span {
  @apply px-3 text-gray-500}
.video-dimension input {
  @apply w-full min-w-0 bg-transparent pr-2 disabled:opacity-40}
.video-ratio {
  @apply flex h-[72px] flex-col items-center justify-center gap-1.5 rounded-xl border border-gray-300 text-sm hover:bg-gray-100 dark:border-dark-600 dark:hover:bg-dark-700}
.ratio-shape {
  @apply max-h-[26px] rounded-sm border-2 border-current}
</style>
