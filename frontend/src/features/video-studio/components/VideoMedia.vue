<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
const props = defineProps<{
    blob: Blob;
    kind: 'image' | 'video';
    alt?: string;
    controls?: boolean;
}>();
const src = ref('');
watch(() => props.blob, blob => { if (src.value)
    URL.revokeObjectURL(src.value); src.value = URL.createObjectURL(blob); }, { immediate: true });
onBeforeUnmount(() => { if (src.value)
    URL.revokeObjectURL(src.value); });
</script>
<template>
<img v-if="kind==='image'" :src="src" :alt="alt || ''" />
<video v-else :src="src" :aria-label="alt" :controls="controls" playsinline preload="metadata" />
</template>
