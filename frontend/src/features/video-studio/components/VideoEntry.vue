<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import Icon from '@/components/icons/Icon.vue';
import VideoMedia from './VideoMedia.vue';
import type { VideoRecord } from '../types';
defineProps<{
    record: VideoRecord;
    running: boolean;
    unsaved: boolean;
}>();
const emit = defineEmits<{
    reuse: [
    ];
    retry: [
    ];
    resume: [
    ];
    pause: [
    ];
    download: [
    ];
    asset: [
    ];
    save: [
    ];
    preview: [
        index: number
    ];
}>();
const { t } = useI18n();
</script>
<template>
<article class="video-entry" :data-video-record="record.id">
  <div class="mb-1 flex justify-end">
<div class="max-w-[80%]">
<div v-if="record.references.length" class="mb-2 flex flex-wrap justify-end gap-2">
<button v-for="(r,index) in record.references" :key="r.id" type="button" class="h-20 w-20 overflow-hidden rounded-xl" :aria-label="t('videoStudio.previewReference',{n:index+1})" @click="emit('preview',index)">
<VideoMedia :blob="r.blob" kind="image" :alt="r.name" class="h-full w-full object-cover" />
</button>
</div>
<p class="whitespace-pre-wrap break-words text-right text-[15px] leading-7">{{record.prompt}}</p>
</div>
</div>
  <div class="mb-3 flex flex-wrap items-center justify-end gap-2 text-[11px] text-gray-600 dark:text-gray-400">
<time>{{new Date(record.createdAt).toLocaleString()}}</time>
<span>{{record.model.label}} · {{record.parameters.resolution}}p · {{record.parameters.seconds}}s</span>
<button type="button" class="entry-action" @click="emit('reuse')">
<Icon name="refresh" size="xs" />{{t('videoStudio.reuse')}}</button>
</div>
  <div class="max-w-[640px]">
   <template v-if="record.video">
<div class="overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
<VideoMedia :blob="record.video" kind="video" :alt="record.prompt" controls class="max-h-[440px] w-full bg-black" />
<div class="flex flex-wrap items-center justify-between gap-2 px-3 py-2.5">
<span class="text-[11px] tabular-nums text-gray-600 dark:text-gray-400">
<template v-if="record.width && record.height">{{record.width}} × {{record.height}} · </template> {{(record.video.size/1024/1024).toFixed(1)}} MB · {{((record.videoDurationMs||0)/1000).toFixed(1)}}s</span>
<div class="flex gap-2">
<button type="button" class="entry-action" @click="emit('asset')">
<Icon name="inbox" size="xs" />{{t('videoStudio.addAsset')}}</button>
<button type="button" class="entry-action" @click="emit('download')">
<Icon name="download" size="xs" />{{t('videoStudio.download')}}</button>
</div>
</div>
</div>
</template>
   <div v-else-if="running" role="status" class="video-pending">
<Icon name="refresh" class="motion-safe:animate-spin" />
<span>{{t('videoStudio.generating')}}<template v-if="record.progress"> · {{Math.round(record.progress)}}%</template>
</span>
<button type="button" class="entry-action" @click="emit('pause')">{{t('videoStudio.pause')}}</button>
</div>
   <div v-else class="rounded-xl border border-dashed border-gray-300 p-6 text-sm dark:border-dark-600">
<p>{{t(record.error || (record.taskID?'videoStudio.paused':'videoStudio.submitUncertain'))}}</p>
<div class="mt-3 flex gap-3">
<button v-if="record.taskID && record.status!=='failed'" type="button" class="entry-action" @click="emit('resume')">{{t('videoStudio.resume')}}</button>
<button type="button" class="entry-action" @click="emit('retry')">{{t('videoStudio.regenerate')}}</button>
</div>
</div>
   <div v-if="unsaved" role="alert" class="mt-2 flex items-center gap-3 text-sm text-amber-800 dark:text-amber-300">{{t('videoStudio.unsaved')}}<button type="button" class="entry-action" @click="emit('save')">{{t('videoStudio.saveAgain')}}</button>
</div>
  </div>
 </article>
</template>
<style scoped>
.entry-action {
  @apply inline-flex min-h-8 items-center gap-1.5 rounded-md px-2 text-xs hover:bg-gray-100 focus-visible:outline-primary-500 dark:hover:bg-dark-700}
.video-pending {
  @apply flex aspect-video flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-gray-300 bg-gray-50/70 text-sm text-gray-600 dark:border-dark-600 dark:bg-dark-800/70 dark:text-gray-400}
</style>
