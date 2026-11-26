<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const props = defineProps<{
    modelValue: 'image' | 'video';
}>();
const emit = defineEmits<{
    'update:modelValue': [
        value: 'image' | 'video'
    ];
}>();
const { t } = useI18n();
function keyboard(event: KeyboardEvent) { if (['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) {
    event.preventDefault();
    const next = event.key === 'Home' ? 'image' : event.key === 'End' ? 'video' : props.modelValue === 'image' ? 'video' : 'image';
    emit('update:modelValue', next);
    (event.currentTarget as HTMLElement).querySelector<HTMLButtonElement>(`[data-studio-tab=${next}]`)?.focus();
} }
</script>
<template>
<div class="mb-5 flex flex-wrap items-center justify-between gap-3">
<div role="tablist" :aria-label="t('videoStudio.title')" class="inline-flex gap-1 rounded-xl border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-900" @keydown="keyboard">
<button v-for="tab in (['image','video'] as const)" :id="`studio-tab-${tab}`" :key="tab" type="button" role="tab" :data-studio-tab="tab" :aria-selected="modelValue===tab" :aria-controls="`studio-panel-${tab}`" :tabindex="modelValue===tab?0:-1" class="rounded-lg px-5 py-2 text-sm font-medium transition-colors focus-visible:outline-primary-500" :class="modelValue===tab?'bg-gray-100 text-gray-900 dark:bg-dark-700 dark:text-white':'text-gray-600 hover:bg-gray-50 dark:text-gray-400 dark:hover:bg-dark-800'" @click="emit('update:modelValue',tab)">{{t(`videoStudio.${tab}Tab`)}}</button>
</div>
<slot />
</div>
</template>
