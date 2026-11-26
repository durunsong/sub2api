<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import ImageStudioView from './ImageStudioView.vue';
import StudioTabs from '@/features/video-studio/components/StudioTabs.vue';
import Icon from '@/components/icons/Icon.vue';
const VideoStudio = defineAsyncComponent(() => import('@/features/video-studio/components/VideoStudio.vue'));
const route = useRoute(), router = useRouter(), { t } = useI18n();
const tab = ref<'image' | 'video'>(route.query.tab === 'video' ? 'video' : 'image'), visitedVideo = ref(tab.value === 'video'), fullscreen = ref(false);
watch(tab, value => { if (value === 'video')
    visitedVideo.value = true; void router.replace({ query: { ...route.query, tab: value } }); });
watch(() => route.query.tab, value => { tab.value = value === 'video' ? 'video' : 'image'; });
function escape(event: KeyboardEvent) { if (event.key === 'Escape' && tab.value === 'video' && !document.querySelector('[role="dialog"]') && !event.defaultPrevented)
    fullscreen.value = false; }
document.addEventListener('keydown', escape);
onBeforeUnmount(() => document.removeEventListener('keydown', escape));
</script>
<template>
<ImageStudioView :inactive="tab==='video'" :workspace-fullscreen="fullscreen">
<template #workspace-tabs>
<StudioTabs v-model="tab">
<button v-if="tab==='video'" type="button" class="inline-flex min-h-11 items-center gap-2 px-3 text-sm text-primary-700 dark:text-primary-400" :aria-pressed="fullscreen" @click="fullscreen=!fullscreen">
<Icon :name="fullscreen?'minimize':'maximize'" size="md" />{{t(fullscreen?'videoStudio.exitFullscreen':'videoStudio.fullscreen')}}</button>
</StudioTabs>
</template>
<template #video-workspace>
<div v-if="visitedVideo" v-show="tab==='video'" id="studio-panel-video" role="tabpanel" aria-labelledby="studio-tab-video" :class="{'video-fullscreen':fullscreen}">
<VideoStudio :active="tab==='video'" />
</div>
</template>
</ImageStudioView>
</template>
<style scoped>
.video-fullscreen :deep(.video-studio) {
  height:calc(100dvh - 132px)}
</style>
