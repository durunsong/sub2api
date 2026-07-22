<template>
  <div
    ref="authRootRef"
    class="auth-root relative flex min-h-screen items-center justify-center overflow-hidden px-5 py-24"
    :class="{
      'auth-pointer-active': pointerActive,
      'auth-immersive': immersive,
      'auth-dark': isDark
    }"
    @mousemove="onPointerMove"
    @mouseleave="pointerActive = false"
  >
    <div class="auth-background absolute inset-0"></div>

    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="auth-blob auth-blob-1 absolute -right-40 -top-40 h-96 w-96 rounded-full"></div>
      <div class="auth-blob auth-blob-2 absolute -bottom-44 -left-40 h-96 w-96 rounded-full"></div>
      <div class="auth-blob auth-blob-3 absolute left-1/2 top-1/2 h-[30rem] w-[30rem] -translate-x-1/2 -translate-y-1/2 rounded-full"></div>
      <div class="auth-grid absolute inset-0"></div>
      <div class="auth-noise absolute inset-0"></div>
      <div class="auth-cursor-glow" :class="{ 'auth-cursor-glow-active': pointerActive }"></div>
    </div>

    <header v-if="immersive" class="auth-header absolute inset-x-0 top-0 z-20 px-6 py-5">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <router-link to="/" class="auth-header-brand flex items-center gap-3">
          <div class="auth-header-logo h-10 w-10 overflow-hidden rounded-[14px]">
            <img :src="siteLogo || '/logo.svg'" alt="" class="h-full w-full object-cover" />
          </div>
          <div>
            <strong>{{ siteName }}</strong>
            <span>{{ t('auth.showcase.headerService') }}</span>
          </div>
        </router-link>
        <div class="auth-header-actions flex items-center gap-2">
          <div class="auth-locale rounded-full px-1">
            <LocaleSwitcher />
          </div>
          <button
            type="button"
            class="auth-theme-toggle flex h-10 w-10 items-center justify-center rounded-full"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
        </div>
      </nav>
    </header>

    <div
      class="relative z-10 w-full"
      :class="immersive ? 'auth-stage mx-auto max-w-6xl' : 'max-w-md'"
    >
      <section v-if="immersive" class="auth-showcase">
        <div class="auth-kicker"><i></i>{{ t('auth.showcase.kicker') }}</div>
        <h1>
          <span>{{ t('auth.showcase.titleLine1') }}</span>
          <span class="auth-title-shine" :data-text="t('auth.showcase.titleLine2')">
            {{ t('auth.showcase.titleLine2') }}
          </span>
        </h1>
        <p>{{ siteSubtitle }}</p>

        <div class="auth-access-pass">
          <div class="auth-pass-grid" aria-hidden="true"></div>
          <div class="auth-pass-topline">
            <span>{{ siteName }}</span>
            <span>{{ t('auth.showcase.accessReady') }}</span>
          </div>
          <h2>{{ t('auth.showcase.passTitle') }}</h2>
          <div class="auth-pass-metrics">
            <div>
              <span>{{ t('auth.showcase.metrics.access.label') }}</span>
              <strong>{{ t('auth.showcase.metrics.access.value') }}</strong>
            </div>
            <div>
              <span>{{ t('auth.showcase.metrics.billing.label') }}</span>
              <strong>{{ t('auth.showcase.metrics.billing.value') }}</strong>
            </div>
            <div>
              <span>{{ t('auth.showcase.metrics.service.label') }}</span>
              <strong>{{ t('auth.showcase.metrics.service.value') }}</strong>
            </div>
          </div>
        </div>

        <div class="auth-trust-row">
          <span><i></i>{{ t('auth.showcase.trust.ready') }}</span>
          <span><i></i>{{ t('auth.showcase.trust.usage') }}</span>
          <span><i></i>{{ t('auth.showcase.trust.billing') }}</span>
        </div>
      </section>

      <section class="auth-form-column">
        <div v-if="!immersive" class="mb-8 text-center">
          <div class="auth-logo mb-4 inline-flex h-16 w-16 overflow-hidden rounded-[20px]">
            <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-full w-full object-cover" />
          </div>
          <h1 class="text-gradient mb-2 text-3xl font-bold">{{ siteName }}</h1>
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ siteSubtitle }}</p>
        </div>

        <div ref="authCardRef" class="auth-card rounded-[2rem] p-8">
          <slot />
        </div>

        <div class="auth-footer mt-6 text-center text-sm">
          <slot name="footer" />
        </div>

        <div class="auth-copyright mt-7 text-center text-xs">
          &copy; {{ currentYear }} {{ siteName }}
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const props = withDefaults(defineProps<{
  immersive?: boolean
}>(), {
  immersive: false
})

const { t } = useI18n()
const appStore = useAppStore()
const immersive = computed(() => props.immersive)
const authRootRef = ref<HTMLElement | null>(null)
const authCardRef = ref<HTMLElement | null>(null)
const pointerActive = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))
const prefersReducedMotion = typeof window !== 'undefined'
  && window.matchMedia('(prefers-reduced-motion: reduce)').matches
let rafId = 0

const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const rawSiteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '')
const siteSubtitle = computed(() => {
  const subtitle = rawSiteSubtitle.value.trim()
  if (!subtitle || /subscription(?:\s+to)?\s+api\s+conversion|ai api gateway|sub2api/i.test(subtitle)) {
    return t('auth.showcase.defaultSubtitle')
  }
  return subtitle
})

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  initTheme()
  appStore.fetchPublicSettings()
})

onUnmounted(() => {
  if (rafId) cancelAnimationFrame(rafId)
})

function toggleTheme(): void {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme(): void {
  const savedTheme = localStorage.getItem('theme')
  const shouldUseDark = savedTheme === 'dark'
    || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  isDark.value = shouldUseDark
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

function onPointerMove(e: MouseEvent) {
  if (prefersReducedMotion) return
  const el = authRootRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  if (!pointerActive.value) pointerActive.value = true
  if (rafId) return
  rafId = requestAnimationFrame(() => {
    el.style.setProperty('--mx', `${x}px`)
    el.style.setProperty('--my', `${y}px`)
    const card = authCardRef.value
    if (card) {
      const cardRect = card.getBoundingClientRect()
      card.style.setProperty('--card-mx', `${e.clientX - cardRect.left}px`)
      card.style.setProperty('--card-my', `${e.clientY - cardRect.top}px`)
    }
    rafId = 0
  })
}
</script>

<style scoped>
.auth-root {
  color: #0f172a;
}

.auth-background {
  background:
    radial-gradient(circle at 12% 16%, rgba(34, 211, 238, 0.12), transparent 28%),
    radial-gradient(circle at 88% 14%, rgba(16, 185, 129, 0.13), transparent 28%),
    linear-gradient(145deg, #f8fcff 0%, #eff9fc 48%, #f7fbff 100%);
}

.auth-root.auth-dark .auth-background {
  background:
    radial-gradient(circle at 14% 12%, rgba(34, 211, 238, 0.12), transparent 27%),
    radial-gradient(circle at 86% 16%, rgba(16, 185, 129, 0.1), transparent 24%),
    linear-gradient(145deg, #050a12 0%, #071421 48%, #05101b 100%);
}

.auth-grid {
  opacity: 0.75;
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.04) 1px, transparent 1px);
  background-size: 56px 56px;
  animation: auth-grid-pan 40s linear infinite;
}

.auth-root.auth-dark .auth-grid {
  opacity: 0.68;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.055) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.04) 1px, transparent 1px);
}

.auth-noise {
  opacity: 0.16;
  background-image: radial-gradient(rgba(15, 23, 42, 0.18) 0.65px, transparent 0.65px);
  background-size: 12px 12px;
  mask-image: linear-gradient(to bottom, black, transparent 92%);
}

.auth-root.auth-dark .auth-noise {
  opacity: 0.22;
  background-image: radial-gradient(rgba(148, 163, 184, 0.13) 0.65px, transparent 0.65px);
}

.text-gradient {
  @apply bg-clip-text text-transparent;
  background-image: linear-gradient(110deg, #0e7490, #0f766e);
}

.auth-header {
  border-bottom: 1px solid transparent;
}

.auth-root.auth-dark .auth-header {
  border-bottom-color: rgba(148, 163, 184, 0.07);
  background: linear-gradient(180deg, rgba(5, 10, 18, 0.78), rgba(5, 10, 18, 0.16));
  backdrop-filter: blur(18px);
}

.auth-header-brand strong,
.auth-header-brand span {
  display: block;
}

.auth-header-brand strong {
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.18em;
}

.auth-header-brand span {
  margin-top: 2px;
  color: #0e7490;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.22em;
}

.auth-root.auth-dark .auth-header-brand strong {
  color: #f8fafc;
}

.auth-root.auth-dark .auth-header-brand span {
  color: #67e8f9;
}

.auth-header-logo {
  box-shadow: 0 14px 30px -18px rgba(8, 47, 73, 0.7);
  transition: transform 0.3s ease;
}

.auth-header-brand:hover .auth-header-logo {
  transform: rotate(-5deg) scale(1.06);
}

.auth-locale,
.auth-theme-toggle {
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.88);
  color: #334155;
  box-shadow: 0 12px 34px -26px rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(16px);
}

.auth-root.auth-dark .auth-locale,
.auth-root.auth-dark .auth-theme-toggle {
  border-color: rgba(125, 211, 252, 0.13);
  background: rgba(8, 22, 36, 0.84);
  color: #cbd5e1;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.auth-theme-toggle {
  transition:
    transform 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.auth-theme-toggle:hover {
  border-color: rgba(8, 145, 178, 0.28);
  color: #0891b2;
  transform: translateY(-2px);
}

.auth-root.auth-dark .auth-theme-toggle:hover {
  border-color: rgba(103, 232, 249, 0.28);
  color: #67e8f9;
}

.auth-stage {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(420px, 0.82fr);
  align-items: center;
  gap: 72px;
}

.auth-showcase {
  max-width: 590px;
}

.auth-kicker {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  border: 1px solid rgba(8, 145, 178, 0.18);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.66);
  padding: 7px 12px;
  color: #0e7490;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.16em;
  box-shadow: 0 12px 34px -26px rgba(8, 47, 73, 0.6);
  backdrop-filter: blur(14px);
}

.auth-kicker i {
  height: 6px;
  width: 6px;
  border-radius: 999px;
  background: #10b981;
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.1), 0 0 14px rgba(16, 185, 129, 0.65);
}

.auth-root.auth-dark .auth-kicker {
  border-color: rgba(103, 232, 249, 0.16);
  background: rgba(8, 22, 36, 0.72);
  color: #67e8f9;
}

.auth-showcase > h1 {
  margin: 26px 0 18px;
  color: #081321;
  font-size: clamp(42px, 3.75vw, 54px);
  font-weight: 950;
  letter-spacing: -0.065em;
  line-height: 0.98;
}

.auth-showcase > h1 span {
  display: block;
  white-space: nowrap;
}

.auth-showcase > h1 span + span {
  margin-top: 1rem;
}

.auth-title-shine {
  position: relative;
}

.auth-title-shine::after {
  content: attr(data-text);
  position: absolute;
  inset: 0;
  /* ponytail: 与首页 brand-shine 同方案，斜角 repeating 循环会接缝闪 */
  background-image: linear-gradient(
    110deg,
    transparent 0%,
    transparent 40%,
    #67e8f9 46%,
    #5eead4 50%,
    #fcd34d 54%,
    transparent 60%,
    transparent 100%
  );
  background-size: 220% 100%;
  background-repeat: no-repeat;
  background-position: 120% 50%;
  background-clip: text;
  color: transparent;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  pointer-events: none;
  animation: auth-title-shine 4.5s ease-in-out infinite;
}

@keyframes auth-title-shine {
  0%,
  18% {
    background-position: 120% 50%;
  }
  52%,
  100% {
    background-position: -20% 50%;
  }
}

.auth-root.auth-dark .auth-showcase > h1 {
  color: #f8fafc;
}

.auth-showcase > p {
  max-width: 520px;
  margin: 0;
  color: #526277;
  font-size: 15px;
  line-height: 1.9;
}

.auth-root.auth-dark .auth-showcase > p {
  color: #94a3b8;
}

.auth-access-pass {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  margin-top: 34px;
  border: 1px solid rgba(125, 211, 252, 0.15);
  border-radius: 28px;
  background:
    radial-gradient(circle at 85% 20%, rgba(45, 212, 191, 0.26), transparent 28%),
    linear-gradient(135deg, #07111f 0%, #0a2434 58%, #0d3a42 100%);
  padding: 24px;
  color: white;
  box-shadow: 0 30px 78px -45px rgba(8, 47, 73, 0.85);
}

.auth-pass-grid {
  position: absolute;
  inset: 0;
  z-index: -1;
  opacity: 0.32;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.08) 1px, transparent 1px);
  background-size: 32px 32px;
  mask-image: linear-gradient(115deg, transparent 10%, black 42%, transparent 94%);
}

.auth-pass-topline,
.auth-pass-metrics,
.auth-trust-row {
  display: flex;
  align-items: center;
}

.auth-pass-topline {
  justify-content: space-between;
  gap: 16px;
  color: #67e8f9;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.18em;
}

.auth-access-pass h2 {
  margin: 30px 0 24px;
  font-size: 25px;
  font-weight: 950;
  letter-spacing: -0.045em;
}

.auth-pass-metrics {
  justify-content: space-between;
  gap: 20px;
  padding-top: 17px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.auth-pass-metrics span,
.auth-pass-metrics strong {
  display: block;
}

.auth-pass-metrics span {
  color: #64748b;
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.18em;
}

.auth-pass-metrics strong {
  margin-top: 5px;
  color: #f8fafc;
  font-size: 12px;
  font-weight: 900;
}

.auth-trust-row {
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 20px;
  color: #526277;
  font-size: 11px;
  font-weight: 800;
}

.auth-trust-row span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.auth-trust-row i {
  height: 5px;
  width: 5px;
  border-radius: 999px;
  background: #14b8a6;
}

.auth-root.auth-dark .auth-trust-row {
  color: #94a3b8;
}

.auth-form-column {
  min-width: 0;
}

.auth-immersive .auth-form-column {
  width: 100%;
  max-width: 480px;
  justify-self: end;
}

.auth-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.9);
  box-shadow:
    0 36px 100px -56px rgba(8, 47, 73, 0.62),
    inset 0 1px 0 rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(28px);
  animation: auth-card-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.auth-root.auth-dark .auth-card {
  border-color: rgba(103, 232, 249, 0.13);
  background: linear-gradient(145deg, rgba(10, 24, 39, 0.96), rgba(6, 17, 29, 0.94));
  box-shadow:
    0 38px 110px -52px rgba(6, 182, 212, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.auth-card::before {
  content: '';
  position: absolute;
  inset: -1px;
  pointer-events: none;
  opacity: 0;
  background: radial-gradient(
    190px circle at var(--card-mx, 50%) var(--card-my, 50%),
    rgba(34, 211, 238, 0.12),
    rgba(16, 185, 129, 0.05) 38%,
    transparent 72%
  );
  transition: opacity 0.35s ease;
}

.auth-pointer-active .auth-card::before {
  opacity: 1;
}

.auth-footer,
.auth-copyright {
  color: #64748b;
}

.auth-root.auth-dark .auth-footer,
.auth-root.auth-dark .auth-copyright {
  color: #64748b;
}

:deep(.auth-view-heading) {
  text-align: left;
}

:deep(.auth-view-title) {
  color: #0f172a;
  font-size: 29px;
  font-weight: 950;
  letter-spacing: -0.04em;
}

.auth-root.auth-dark :deep(.auth-view-title) {
  color: #f8fafc;
}

:deep(.auth-view-subtitle) {
  margin-top: 8px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
}

.auth-root.auth-dark :deep(.auth-view-subtitle) {
  color: #94a3b8;
}

:deep(.input-label) {
  margin-bottom: 8px;
  color: #334155;
  font-size: 12px;
  font-weight: 800;
}

.auth-root.auth-dark :deep(.input-label) {
  color: #cbd5e1;
}

:deep(.input) {
  min-height: 50px;
  border-color: rgba(15, 23, 42, 0.1);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.88);
  color: #0f172a;
}

.auth-root.auth-dark :deep(.input) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(5, 15, 27, 0.72);
  color: #f1f5f9;
}

:deep(.input:focus) {
  border-color: #06b6d4 !important;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.16) !important;
}

:deep(.input-hint) {
  margin-top: 7px;
  color: #64748b;
  line-height: 1.6;
}

:deep(.btn) {
  min-height: 50px;
  border-radius: 14px;
  font-weight: 800;
}

:deep(.btn-primary) {
  position: relative;
  overflow: hidden;
  background-image: linear-gradient(110deg, #67e8f9, #5eead4 54%, #fcd34d) !important;
  color: #06202b !important;
  box-shadow: 0 16px 36px -20px rgba(8, 145, 178, 0.78) !important;
  transition:
    background-image 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.2s ease;
}

:deep(.btn-primary:hover) {
  background-image: linear-gradient(110deg, #a5f3fc, #6ee7b7 54%, #fde68a) !important;
  box-shadow: 0 20px 42px -20px rgba(8, 145, 178, 0.9) !important;
  transform: translateY(-1px);
}

:deep(.btn-primary)::after {
  content: '';
  position: absolute;
  top: 0;
  left: -150%;
  width: 60%;
  height: 100%;
  background: linear-gradient(
    120deg,
    transparent,
    rgba(103, 232, 249, 0.28),
    rgba(94, 234, 212, 0.58),
    rgba(252, 211, 77, 0.64),
    transparent
  );
  pointer-events: none;
  transform: skewX(-20deg);
  animation: auth-cta-sweep 3.5s ease-in-out infinite;
}

:deep(.btn-primary:not(:disabled):active) {
  transform: translateY(0) scale(0.98);
}

@keyframes auth-cta-sweep {
  0%,
  60% {
    left: -150%;
  }
  100% {
    left: 150%;
  }
}

:deep(.btn-secondary) {
  border-color: rgba(15, 23, 42, 0.1);
  background: rgba(255, 255, 255, 0.74);
  box-shadow: none;
}

.auth-root.auth-dark :deep(.btn-secondary) {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(8, 22, 36, 0.78);
  color: #e2e8f0;
}

:deep(a.text-primary-600) {
  color: #0e7490 !important;
}

:deep(a.text-primary-600:hover) {
  color: #0891b2 !important;
}

.auth-root.auth-dark :deep(a.text-primary-600) {
  color: #67e8f9 !important;
}

.auth-root.auth-dark :deep(a.text-primary-600:hover) {
  color: #a5f3fc !important;
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
  box-shadow: 0 12px 28px -8px rgba(8, 145, 178, 0.46);
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
  opacity: 0.24;
  filter: blur(52px);
  will-change: transform;
}

.auth-root.auth-dark .auth-blob {
  opacity: 0.3;
}

.auth-blob-1 {
  background: #22d3ee;
  animation: auth-float-1 16s ease-in-out infinite;
}

.auth-blob-2 {
  background: #34d399;
  animation: auth-float-2 20s ease-in-out infinite;
}

.auth-blob-3 {
  background: #0e7490;
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
.auth-cursor-glow {
  position: absolute;
  inset: 0;
  opacity: 0;
  transition: opacity 0.4s ease;
  background: radial-gradient(
    300px circle at var(--mx, 50%) var(--my, 50%),
    rgba(34, 211, 238, 0.13),
    rgba(16, 185, 129, 0.05) 42%,
    transparent 72%
  );
}

.auth-cursor-glow-active {
  opacity: 1;
}

@keyframes auth-grid-pan {
  from {
    background-position: 0 0;
  }
  to {
    background-position: 64px 64px;
  }
}

@media (max-width: 1023px) {
  .auth-stage {
    display: block;
    max-width: 500px;
  }

  .auth-showcase {
    display: none;
  }

  .auth-immersive .auth-form-column {
    max-width: none;
  }
}

@media (max-width: 640px) {
  .auth-root {
    padding: 90px 16px 42px;
  }

  .auth-header {
    padding: 16px;
  }

  .auth-header-brand span {
    display: none;
  }

  .auth-card {
    border-radius: 24px;
    padding: 24px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-card,
  .auth-logo,
  .auth-blob,
  .auth-grid,
  .auth-title-shine::after,
  :deep(.btn-primary)::after {
    animation: none !important;
  }
  .auth-title-shine::after {
    content: none;
  }
  .auth-cursor-glow,
  .auth-card::before {
    display: none;
  }
}
</style>
