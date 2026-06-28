<template>
  <div class="auth-root relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-gradient-to-br from-gray-50 via-indigo-50/40 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950"
    ></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Gradient Orbs -->
      <div
        class="auth-blob auth-blob-1 absolute -right-40 -top-40 h-80 w-80 rounded-full bg-indigo-400/20 blur-3xl"
      ></div>
      <div
        class="auth-blob auth-blob-2 absolute -bottom-40 -left-40 h-80 w-80 rounded-full bg-violet-500/15 blur-3xl"
      ></div>
      <div
        class="auth-blob auth-blob-3 absolute left-1/2 top-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2 rounded-full bg-indigo-300/10 blur-3xl"
      ></div>

      <!-- Grid Pattern -->
      <div
        class="auth-grid absolute inset-0 bg-[linear-gradient(rgba(99,102,241,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(99,102,241,0.04)_1px,transparent_1px)] bg-[size:64px_64px]"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div
            class="auth-logo mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl shadow-lg shadow-indigo-500/30"
          >
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="text-gradient mb-2 text-3xl font-bold">
            {{ siteName }}
          </h1>
          <p class="text-sm text-gray-500 dark:text-dark-400">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="auth-card card-glass rounded-2xl p-8 shadow-glass">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.text-gradient {
  @apply bg-clip-text text-transparent;
  background-image: linear-gradient(to right, #4f46e5, #8b5cf6);
}

/* ===== Indigo / Violet theme overrides (scoped to auth pages) ===== */
:deep(.btn-primary) {
  background-image: linear-gradient(to right, #6366f1, #7c3aed) !important;
  box-shadow: 0 6px 18px -4px rgba(99, 102, 241, 0.4) !important;
}
:deep(.btn-primary:hover) {
  background-image: linear-gradient(to right, #4f46e5, #6d28d9) !important;
  box-shadow: 0 10px 26px -6px rgba(124, 58, 237, 0.5) !important;
}
:deep(.input:focus) {
  border-color: #6366f1 !important;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.25) !important;
}
/* Text links (forgot password / sign up / sign in) */
:deep(a.text-primary-600) {
  color: #4f46e5 !important;
}
:deep(a.text-primary-600:hover) {
  color: #6366f1 !important;
}
:deep(.dark a.text-primary-600) {
  color: #a5b4fc !important;
}
:deep(.dark a.text-primary-600:hover) {
  color: #c7d2fe !important;
}

/* ===== Motion ===== */
.auth-card {
  animation: auth-card-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}
@keyframes auth-card-in {
  from {
    opacity: 0;
    transform: translateY(18px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.auth-logo {
  animation: auth-logo-in 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  transition:
    transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.35s ease;
}
.auth-logo:hover {
  transform: rotate(-6deg) scale(1.1);
  box-shadow: 0 12px 28px -8px rgba(99, 102, 241, 0.55);
}
@keyframes auth-logo-in {
  from {
    opacity: 0;
    transform: translateY(-10px) scale(0.8);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.auth-blob {
  will-change: transform;
}
.auth-blob-1 {
  animation: auth-float-1 16s ease-in-out infinite;
}
.auth-blob-2 {
  animation: auth-float-2 20s ease-in-out infinite;
}
.auth-blob-3 {
  animation: auth-float-3 22s ease-in-out infinite;
}
@keyframes auth-float-1 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(-36px, 28px) scale(1.1);
  }
}
@keyframes auth-float-2 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(40px, -28px) scale(1.12);
  }
}
@keyframes auth-float-3 {
  0%,
  100% {
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    transform: translate(-50%, -50%) scale(1.08);
  }
}
.auth-grid {
  animation: auth-grid-pan 40s linear infinite;
}
@keyframes auth-grid-pan {
  from {
    background-position: 0 0;
  }
  to {
    background-position: 64px 64px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-card,
  .auth-logo,
  .auth-blob,
  .auth-grid {
    animation: none !important;
  }
}
</style>
