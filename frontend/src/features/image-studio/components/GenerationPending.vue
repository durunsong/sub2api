<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

defineProps<{
  status: 'queued' | 'running'
  elapsed: string
  title: string
  subtitle: string
  countLabel?: string
}>()

const root = ref<HTMLElement>()
const pageHidden = ref(false)
const inView = ref(true)
let observer: IntersectionObserver | undefined

function syncVisibility() { pageHidden.value = document.hidden }

onMounted(() => {
  syncVisibility()
  document.addEventListener('visibilitychange', syncVisibility)
  if (typeof IntersectionObserver !== 'undefined' && root.value) {
    observer = new IntersectionObserver(([entry]) => { inView.value = entry?.isIntersecting ?? true })
    observer.observe(root.value)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', syncVisibility)
  observer?.disconnect()
})
</script>

<template>
  <div ref="root" class="generation-pending" :data-status="status" :data-motion="pageHidden || !inView ? 'paused' : 'running'">
    <div class="pending-meta" aria-live="off">
      <span>{{ countLabel }}</span>
      <span class="pending-elapsed">{{ elapsed }}</span>
    </div>

    <div class="pending-scene" aria-hidden="true">
      <div class="scene-ground" />
      <div class="scene-frame">
        <div class="frame-sheet sheet-back" />
        <div class="frame-sheet sheet-middle" />
        <div class="frame-sheet sheet-front">
          <svg class="developing-image" viewBox="0 0 120 88" fill="none" aria-hidden="true" focusable="false">
            <path class="image-contour" d="M0 61 35 29 67 57 91 39 120 66M0 76 35 44 67 72 91 54 120 81" />
            <g class="image-fragment fragment-one"><path d="m0 61 35-32 16 14-35 32Z" fill="currentColor" opacity=".22" /></g>
            <g class="image-fragment fragment-two"><path d="m16 75 35-32 16 14-34 31H16Z" fill="currentColor" opacity=".38" /></g>
            <g class="image-fragment fragment-three"><path d="m53 70 38-31 15 14-37 35H33Z" fill="currentColor" opacity=".6" /></g>
            <g class="image-fragment fragment-four"><path d="m69 88 37-35 14 13v22Z" fill="currentColor" opacity=".28" /><circle cx="89" cy="21" r="7" fill="currentColor" opacity=".5" /></g>
          </svg>
          <div class="scan-light" />
        </div>
        <span class="frame-corner corner-top-left" />
        <span class="frame-corner corner-top-right" />
        <span class="frame-corner corner-bottom-left" />
        <span class="frame-corner corner-bottom-right" />
      </div>
    </div>

    <div class="pending-copy" role="status" aria-atomic="true">
      <p class="pending-title"><span class="state-marker" aria-hidden="true" />{{ title }}</p>
      <p class="pending-subtitle">{{ subtitle }}</p>
    </div>
  </div>
</template>

<style scoped>
.generation-pending {
  --pending-surface: theme('colors.gray.50');
  --pending-paper: theme('colors.white');
  --pending-ink: theme('colors.primary.700');
  --pending-line: theme('colors.primary.200');
  --pending-muted: theme('colors.gray.600');
  position: relative;
  display: flex;
  aspect-ratio: 4 / 3;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  overflow: hidden;
  padding: 2rem 1rem 1.25rem;
  background: var(--pending-surface);
  color: var(--pending-ink);
  container-type: inline-size;
}

.dark .generation-pending {
  --pending-surface: theme('colors.dark.950');
  --pending-paper: theme('colors.dark.900');
  --pending-ink: theme('colors.primary.300');
  --pending-line: theme('colors.primary.900');
  --pending-muted: theme('colors.dark.300');
}

.pending-meta {
  position: absolute;
  inset: 0.875rem 1rem auto;
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  color: var(--pending-muted);
  font-size: 0.6875rem;
  line-height: 1rem;
}
.pending-elapsed { flex-shrink: 0; font-variant-numeric: tabular-nums; }
.pending-scene { position: relative; width: 12.5rem; max-width: 80%; height: 7.5rem; flex-shrink: 1; }
.scene-ground {
  position: absolute;
  inset: 70% 15% 8%;
  border: 1px solid var(--pending-line);
  border-radius: 50%;
  transform: rotate(-8deg) scaleY(0.5);
  opacity: 0.45;
}
.scene-frame {
  position: absolute;
  left: 50%;
  top: 42%;
  width: 7.5rem;
  height: 5.5rem;
  transform: translate(-50%, -50%) perspective(500px) rotateX(12deg) rotateY(-18deg) rotateZ(-8deg);
  transform-style: preserve-3d;
}
.frame-sheet { position: absolute; inset: 0; border: 1px solid var(--pending-line); border-radius: 4px; background: var(--pending-paper); }
.sheet-back { transform: translate(15px, 15px); opacity: 0.4; }
.sheet-middle { transform: translate(8px, 8px); opacity: 0.65; }
.sheet-front { overflow: hidden; box-shadow: 7px 12px 18px -12px rgb(15 118 110 / 30%); }
.developing-image { display: block; width: 100%; height: 100%; }
.image-contour { stroke: currentColor; stroke-width: 0.7; opacity: 0.24; }
.image-fragment { transform-box: fill-box; transform-origin: center; }
.scan-light {
  position: absolute;
  inset: -25% 0 auto;
  height: 25%;
  border-bottom: 1px solid currentColor;
  background: linear-gradient(180deg, transparent, rgb(45 212 191 / 17%));
  opacity: 0;
}
.frame-corner { position: absolute; width: 10px; height: 10px; border-color: currentColor; border-style: solid; opacity: 0.7; }
.corner-top-left { top: -5px; left: -5px; border-width: 1px 0 0 1px; }
.corner-top-right { top: -5px; right: -5px; border-width: 1px 1px 0 0; }
.corner-bottom-left { bottom: -5px; left: -5px; border-width: 0 0 1px 1px; }
.corner-bottom-right { bottom: -5px; right: -5px; border-width: 0 1px 1px 0; }
.pending-copy { width: 100%; text-align: center; }
.pending-title { display: flex; align-items: center; justify-content: center; gap: 0.5rem; font-size: 0.875rem; font-weight: 600; line-height: 1.5rem; }
.state-marker { width: 5px; height: 5px; border-radius: 50%; background: currentColor; flex-shrink: 0; }
.pending-subtitle { max-width: 29ch; margin: 0.25rem auto 0; color: var(--pending-muted); font-size: 0.6875rem; line-height: 1.45; text-wrap: pretty; }

[data-status='queued'] .image-fragment { opacity: 0.25; }
[data-status='queued'] .state-marker { background: transparent; border: 1px solid currentColor; }
[data-status='running'] .image-fragment { animation: image-develop 4.8s cubic-bezier(0.16, 1, 0.3, 1) infinite; }
[data-status='running'] .fragment-two { animation-delay: 0.2s; }
[data-status='running'] .fragment-three { animation-delay: 0.4s; }
[data-status='running'] .fragment-four { animation-delay: 0.6s; }
[data-status='running'] .scan-light { animation: image-scan 4.8s ease-in-out infinite; }

@keyframes image-develop {
  0%, 10%, 100% { transform: translateY(5px); opacity: 0.18; }
  32%, 78% { transform: translateY(0); opacity: 1; }
}
@keyframes image-scan {
  0%, 8% { transform: translateY(0); opacity: 0; }
  16% { opacity: 0.8; }
  68% { transform: translateY(500%); opacity: 0.8; }
  78%, 100% { transform: translateY(500%); opacity: 0; }
}

[data-motion='paused'] *, [data-motion='paused'] *::before, [data-motion='paused'] *::after { animation-play-state: paused !important; }

@container (max-width: 220px) {
  .pending-scene { height: 5.5rem; }
  .scene-frame { transform: translate(-50%, -50%) scale(0.8) perspective(500px) rotateX(12deg) rotateY(-18deg) rotateZ(-8deg); }
  .pending-subtitle { font-size: 0.625rem; }
}

@media (prefers-reduced-motion: reduce) {
  .generation-pending *, .generation-pending *::before, .generation-pending *::after { animation: none !important; }
  .scan-light { display: none; }
}
</style>
