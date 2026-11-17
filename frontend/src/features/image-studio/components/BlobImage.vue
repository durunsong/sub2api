<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{ blob: Blob; alt: string; eager?: boolean }>()
const emit = defineEmits<{ load: [image: HTMLImageElement]; error: [] }>()
const url = ref('')
function release() {
  if (url.value) URL.revokeObjectURL(url.value)
  url.value = ''
}
watch(() => props.blob, (blob) => {
  release()
  url.value = URL.createObjectURL(blob)
}, { immediate: true })
onBeforeUnmount(release)
</script>

<template>
  <img :src="url" :alt="alt" :loading="eager ? 'eager' : 'lazy'" decoding="async" draggable="false" @load="emit('load', $event.target as HTMLImageElement)" @error="emit('error')" />
</template>
