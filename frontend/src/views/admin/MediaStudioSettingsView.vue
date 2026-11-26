<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import ImageStudioSettingsView from './ImageStudioSettingsView.vue';
import StudioTabs from '@/features/video-studio/components/StudioTabs.vue';
const VideoStudioSettings = defineAsyncComponent(() => import('@/features/video-studio/components/VideoStudioSettings.vue'));
const route = useRoute(), router = useRouter(), tab = ref<'image' | 'video'>(route.query.tab === 'video' ? 'video' : 'image'), visitedVideo = ref(tab.value === 'video');
watch(tab, value => { if (value === 'video')
    visitedVideo.value = true; void router.replace({ query: { ...route.query, tab: value } }); });
watch(() => route.query.tab, value => { tab.value = value === 'video' ? 'video' : 'image'; });
</script>
<template>
<ImageStudioSettingsView :inactive="tab==='video'">
<template #workspace-tabs>
<StudioTabs v-model="tab" class="mx-auto max-w-6xl" />
</template>
<template #video-workspace>
<div v-if="visitedVideo" v-show="tab==='video'" id="studio-panel-video" role="tabpanel" aria-labelledby="studio-tab-video">
<VideoStudioSettings />
</div>
</template>
</ImageStudioSettingsView>
</template>
