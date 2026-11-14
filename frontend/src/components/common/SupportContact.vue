<template>
  <div
    v-if="contactInfo || contactQrCode"
    :class="hint ? 'relative z-20 mb-3 flex items-center justify-between gap-3' : 'relative'"
  >
    <p v-if="hint" class="text-sm text-gray-600 dark:text-dark-300">{{ hint }}</p>
    <div
      ref="contactRef"
      class="relative shrink-0"
      @pointerenter="onPointerEnter"
      @pointerleave="hovered = false"
      @focusout="onFocusOut"
    >
      <button
        type="button"
        class="flex min-h-10 items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-500 dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white"
        :aria-label="t('nav.supportWechat')"
        :aria-expanded="visible"
        :aria-controls="panelId"
        @click="toggle"
      >
        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M9.5 4C5.91 4 3 6.46 3 9.5c0 1.73.86 3.28 2.22 4.31L4.5 16.5l2.72-1.36c.72.2 1.48.36 2.28.36.34 0 .67-.02 1-.06A4.7 4.7 0 0 1 10.2 14c0-2.9 2.69-5.25 6-5.25.17 0 .34.01.5.02C16.05 6.16 13.07 4 9.5 4zm-2 3.25a.9.9 0 1 1 0 1.8.9.9 0 0 1 0-1.8zm4 0a.9.9 0 1 1 0 1.8.9.9 0 0 1 0-1.8zM16.2 10c-2.76 0-5 1.9-5 4.25S13.44 18.5 16.2 18.5c.55 0 1.08-.07 1.57-.2l2.23 1.12-.62-2.05A3.9 3.9 0 0 0 21.2 14.25C21.2 11.9 18.96 10 16.2 10zm-1.7 2.85a.75.75 0 1 1 0 1.5.75.75 0 0 1 0-1.5zm3.4 0a.75.75 0 1 1 0 1.5.75.75 0 0 1 0-1.5z" />
        </svg>
        <span :class="hint ? '' : 'hidden sm:inline'">{{ t(hint ? 'nav.supportWechat' : 'nav.support') }}</span>
      </button>
      <div
        v-if="visible"
        :id="panelId"
        role="region"
        :aria-label="t('nav.supportWechat')"
        class="absolute right-0 top-full z-50 w-52 max-w-[calc(100vw-2rem)] pt-2"
      >
        <div class="rounded-xl border border-gray-200 bg-white p-3 text-center shadow-lg dark:border-dark-700 dark:bg-dark-800">
          <img
            v-if="contactQrCode"
            :src="contactQrCode"
            :alt="t('nav.supportWechat')"
            class="mx-auto h-36 w-36 rounded-lg object-contain"
          />
          <p v-if="contactInfo" class="mt-2 break-all text-sm font-medium text-gray-900 dark:text-white">
            {{ contactInfo }}
          </p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('nav.supportWechat') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { onClickOutside, onKeyStroke } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

defineProps<{ hint?: string }>()

const { t } = useI18n()
const appStore = useAppStore()
const contactInfo = computed(() => appStore.contactInfo.trim())
const contactQrCode = computed(() => sanitizeUrl(appStore.contactQrCode, { allowRelative: true, allowDataUrl: true }))
const contactRef = ref<HTMLElement | null>(null)
const panelId = useId()
const pinned = ref(false)
const hovered = ref(false)
const visible = computed(() => pinned.value || hovered.value)

function close() {
  pinned.value = false
  hovered.value = false
}

function toggle() {
  pinned.value = !pinned.value
  hovered.value = false
}

function onPointerEnter(event: PointerEvent) {
  if (event.pointerType === 'mouse') hovered.value = true
}

function onFocusOut(event: FocusEvent) {
  // Clicking non-focusable QR/text clears focus; outside clicks are handled separately.
  if (event.relatedTarget && !contactRef.value?.contains(event.relatedTarget as Node)) close()
}

onClickOutside(contactRef, close)
onKeyStroke('Escape', close)
</script>
