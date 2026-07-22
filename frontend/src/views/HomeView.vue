<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    ref="pageRef"
    @mousemove="onPointerMove"
    :class="{ 'is-dark': isDark }"
    class="page-root relative flex min-h-screen flex-col overflow-hidden bg-[#f7fbff] text-slate-950 dark:bg-[#07111f] dark:text-white"
  >
    <!-- Background Decorations -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="color-wash absolute inset-0"></div>
      <div class="noise-layer absolute inset-0"></div>
      <div class="orb orb-1 absolute -right-24 top-10 h-[34rem] w-[34rem] rounded-full"></div>
      <div class="orb orb-2 absolute -left-28 top-[34rem] h-[28rem] w-[28rem] rounded-full"></div>
      <div class="orb orb-3 absolute bottom-20 right-1/4 h-72 w-72 rounded-full"></div>
      <div
        class="grid-overlay absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.055)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:56px_56px] dark:bg-[linear-gradient(rgba(148,163,184,0.06)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.045)_1px,transparent_1px)]"
      ></div>
      <!-- Cursor-following spotlight -->
      <div class="cursor-glow" :class="{ 'cursor-glow-active': pointerActive }"></div>
    </div>

    <!-- Header -->
    <header class="site-header relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="logo-box h-11 w-11 overflow-hidden rounded-[15px] shadow-lg shadow-cyan-900/10">
            <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-full w-full object-cover" />
          </div>
          <div class="hidden leading-tight sm:block">
            <div class="text-sm font-semibold tracking-[0.24em] text-slate-900 dark:text-white">
              {{ siteName }}
            </div>
            <div class="text-[11px] uppercase tracking-[0.28em] text-cyan-700/80 dark:text-cyan-300/80">
              {{ t('home.showcase.headerService') }}
            </div>
          </div>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <div
            class="nav-locale rounded-full border border-slate-200/80 bg-white/90 px-1 shadow-sm shadow-slate-900/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <LocaleSwitcher />
          </div>

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-control hover-pop flex h-10 w-10 items-center justify-center rounded-full border border-slate-200/80 bg-white/90 text-slate-700 shadow-sm shadow-slate-900/5 backdrop-blur-xl transition-colors hover:bg-white hover:text-cyan-700 dark:border-white/10 dark:bg-white/10 dark:text-slate-100 dark:hover:bg-white/15 dark:hover:text-cyan-200"
            :title="t('home.viewDocs')"
            :aria-label="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="nav-control hover-pop flex h-10 w-10 items-center justify-center rounded-full border border-slate-200/80 bg-white/90 text-slate-700 shadow-sm shadow-slate-900/5 backdrop-blur-xl transition-colors hover:bg-white hover:text-cyan-700 dark:border-white/10 dark:bg-white/10 dark:text-slate-100 dark:hover:bg-white/15 dark:hover:text-cyan-200"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="hover-pop inline-flex items-center gap-1.5 rounded-full bg-slate-950 py-1 pl-1 pr-3 shadow-lg shadow-cyan-950/20 transition-colors hover:bg-slate-800 dark:bg-white dark:hover:bg-cyan-50"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-cyan-400 to-emerald-400 text-[10px] font-semibold text-slate-950"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white dark:text-slate-950">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="hover-pop inline-flex items-center rounded-full bg-slate-950 px-4 py-2 text-xs font-medium text-white shadow-lg shadow-cyan-950/20 transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-cyan-50"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 pb-16 pt-12 md:pt-16">
      <div class="mx-auto max-w-6xl">
        <!-- Hero Section - Left/Right Layout -->
        <div class="mb-12 grid items-center gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:gap-16">
          <!-- Left: Text Content -->
          <div class="text-center lg:text-left">
            <div v-reveal="{ delay: 0 }" class="mb-5">
              <div
                class="inline-flex items-center gap-2 rounded-full border border-cyan-300/40 bg-white/70 px-3 py-1.5 text-xs font-semibold tracking-[0.22em] text-cyan-800 shadow-sm shadow-cyan-900/5 backdrop-blur dark:border-cyan-300/20 dark:bg-white/10 dark:text-cyan-200"
              >
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 shadow-[0_0_16px_rgba(52,211,153,0.8)]"></span>
                {{ t('home.showcase.kicker') }}
              </div>
            </div>
            <h1
              v-reveal="{ delay: 80 }"
              class="brand-shine mx-auto mb-5 block w-fit max-w-full text-5xl font-black leading-[0.95] tracking-[-0.06em] text-slate-950 dark:text-white md:text-6xl lg:mx-0 lg:text-7xl"
              :data-text="siteName"
            >
              {{ siteName }}
            </h1>
            <p
              v-reveal="{ delay: 160 }"
              class="mx-auto mb-4 max-w-xl text-lg font-medium leading-8 text-slate-700 dark:text-slate-200 md:text-xl lg:mx-0"
            >
              {{ siteSubtitle }}
            </p>
            <p
              v-reveal="{ delay: 220 }"
              class="mx-auto mb-8 max-w-xl text-sm leading-7 text-slate-500 dark:text-slate-400 lg:mx-0"
            >
              {{ t('home.showcase.description') }}
            </p>

            <!-- CTA Button -->
            <div v-reveal="{ delay: 280 }" class="flex flex-col items-center gap-4 sm:flex-row lg:justify-start">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="cta-orbit inline-flex items-center rounded-full px-7 py-3 text-base font-semibold text-slate-950 shadow-xl shadow-cyan-950/20"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" class="cta-arrow ml-2" :stroke-width="2" />
              </router-link>
              <div class="text-xs font-medium uppercase tracking-[0.26em] text-slate-500 dark:text-slate-400">
                {{ t('home.showcase.ctaMeta') }}
              </div>
            </div>
          </div>

          <!-- Right: AI Access Pass -->
          <div v-reveal="{ delay: 220 }" class="flex justify-center lg:justify-end">
            <div class="product-cockpit">
              <div class="cockpit-frame">
                <div class="cockpit-header">
                  <div class="cockpit-brand">
                    <div class="cockpit-logo">
                      <img :src="siteLogo || '/logo.svg'" alt="" />
                    </div>
                    <div>
                      <strong>{{ siteName }}</strong>
                      <span>{{ t('home.showcase.cockpit.brandNetwork') }}</span>
                    </div>
                  </div>
                  <span class="cockpit-live"><i></i>{{ t('home.showcase.cockpit.online') }}</span>
                </div>

                <div class="access-pass">
                  <div class="pass-grid" aria-hidden="true"></div>
                  <div class="pass-glow" aria-hidden="true"></div>
                  <div class="pass-topline">
                    <span>{{ t('home.showcase.cockpit.unifiedAccess') }}</span>
                    <span>{{ t('home.showcase.cockpit.readyStatus') }}</span>
                  </div>
                  <h2>{{ t('home.showcase.cockpit.passTitle') }}</h2>
                  <p>{{ t('home.showcase.cockpit.passDescription') }}</p>
                  <div class="pass-metrics">
                    <div>
                      <span>{{ t('home.showcase.cockpit.metrics.access.label') }}</span>
                      <strong>{{ t('home.showcase.cockpit.metrics.access.value') }}</strong>
                    </div>
                    <div>
                      <span>{{ t('home.showcase.cockpit.metrics.billing.label') }}</span>
                      <strong>{{ t('home.showcase.cockpit.metrics.billing.value') }}</strong>
                    </div>
                    <div>
                      <span>{{ t('home.showcase.cockpit.metrics.service.label') }}</span>
                      <strong>{{ t('home.showcase.cockpit.metrics.service.value') }}</strong>
                    </div>
                  </div>
                </div>

                <div class="service-rail">
                  <div class="service-row">
                    <span class="service-index">01</span>
                    <div>
                      <strong>{{ t('home.showcase.cockpit.services.ready.title') }}</strong>
                      <span>{{ t('home.showcase.cockpit.services.ready.description') }}</span>
                    </div>
                    <i>{{ t('home.showcase.cockpit.services.ready.status') }}</i>
                  </div>
                  <div class="service-row">
                    <span class="service-index">02</span>
                    <div>
                      <strong>{{ t('home.showcase.cockpit.services.balance.title') }}</strong>
                      <span>{{ t('home.showcase.cockpit.services.balance.description') }}</span>
                    </div>
                    <i>{{ t('home.showcase.cockpit.services.balance.status') }}</i>
                  </div>
                  <div class="service-row">
                    <span class="service-index">03</span>
                    <div>
                      <strong>{{ t('home.showcase.cockpit.services.stable.title') }}</strong>
                      <span>{{ t('home.showcase.cockpit.services.stable.description') }}</span>
                    </div>
                    <i>{{ t('home.showcase.cockpit.services.stable.status') }}</i>
                  </div>
                </div>

                <div class="access-footer">
                  <span>{{ t('home.showcase.cockpit.footer.addCredits') }}</span>
                  <i></i>
                  <span>{{ t('home.showcase.cockpit.footer.copySetup') }}</span>
                  <i></i>
                  <span>{{ t('home.showcase.cockpit.footer.viewUsage') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Feature Tags - Centered -->
        <div class="mb-12 flex flex-wrap items-center justify-center gap-3 md:gap-4">
          <div
            v-reveal="{ delay: 0 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-slate-200/80 bg-white/90 px-5 py-2.5 shadow-sm shadow-cyan-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="swap" size="sm" class="text-cyan-600 dark:text-cyan-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              {{ t('home.showcase.tags.ready') }}
            </span>
          </div>
          <div
            v-reveal="{ delay: 100 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-slate-200/80 bg-white/90 px-5 py-2.5 shadow-sm shadow-emerald-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="shield" size="sm" class="text-emerald-600 dark:text-emerald-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              {{ t('home.showcase.tags.usage') }}
            </span>
          </div>
          <div
            v-reveal="{ delay: 200 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-slate-200/80 bg-white/90 px-5 py-2.5 shadow-sm shadow-amber-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="chart" size="sm" class="text-amber-600 dark:text-amber-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              {{ t('home.showcase.tags.value') }}
            </span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-16 grid gap-5 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div
            v-reveal="{ delay: 0 }"
            class="feature-card group rounded-[2rem] border border-slate-200/80 bg-white/95 p-7 shadow-[0_24px_80px_-48px_rgba(8,47,73,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
          >
            <div
              class="feature-icon mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-cyan-300 to-cyan-600 shadow-lg shadow-cyan-500/25 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6"
            >
              <Icon name="server" size="lg" class="text-slate-950" />
            </div>
            <h3 class="mb-3 text-xl font-black tracking-[-0.03em] text-slate-950 dark:text-white">
              {{ t('home.showcase.features.access.title') }}
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              {{ t('home.showcase.features.access.description') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            v-reveal="{ delay: 120 }"
            class="feature-card group rounded-[2rem] border border-slate-200/80 bg-white/95 p-7 shadow-[0_24px_80px_-48px_rgba(6,78,59,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
          >
            <div
              class="feature-icon mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-300 to-teal-600 shadow-lg shadow-emerald-500/25 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6"
            >
              <svg
                class="h-6 w-6 text-slate-950"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="mb-3 text-xl font-black tracking-[-0.03em] text-slate-950 dark:text-white">
              {{ t('home.showcase.features.availability.title') }}
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              {{ t('home.showcase.features.availability.description') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            v-reveal="{ delay: 240 }"
            class="feature-card group rounded-[2rem] border border-slate-200/80 bg-white/95 p-7 shadow-[0_24px_80px_-48px_rgba(146,64,14,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
          >
            <div
              class="feature-icon mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-200 to-orange-500 shadow-lg shadow-amber-500/25 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6"
            >
              <svg
                class="h-6 w-6 text-slate-950"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <h3 class="mb-3 text-xl font-black tracking-[-0.03em] text-slate-950 dark:text-white">
              {{ t('home.showcase.features.billing.title') }}
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              {{ t('home.showcase.features.billing.description') }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div v-reveal="{ delay: 0 }" class="mb-8 text-center">
          <div class="mb-3 text-xs font-bold uppercase tracking-[0.32em] text-cyan-700 dark:text-cyan-300">
            {{ t('home.showcase.providers.eyebrow') }}
          </div>
          <h2 class="mb-3 text-3xl font-black tracking-[-0.04em] text-slate-950 dark:text-white">
            {{ t('home.showcase.providers.title') }}
          </h2>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ t('home.showcase.providers.description') }}
          </p>
        </div>

        <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
          <!-- Claude - Supported -->
          <div
            v-reveal="{ delay: 0 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/70 bg-white/65 px-5 py-3 ring-1 ring-cyan-900/10 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.07] dark:ring-white/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-orange-400 to-orange-500"
            >
              <span class="text-xs font-bold text-white">C</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ t('home.providers.claude') }}</span>
            <span
              class="rounded-md bg-cyan-100 px-1.5 py-0.5 text-[10px] font-bold text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- GPT - Supported -->
          <div
            v-reveal="{ delay: 80 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/70 bg-white/65 px-5 py-3 ring-1 ring-cyan-900/10 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.07] dark:ring-white/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-green-500 to-green-600"
            >
              <span class="text-xs font-bold text-white">G</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">GPT</span>
            <span
              class="rounded-md bg-cyan-100 px-1.5 py-0.5 text-[10px] font-bold text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Gemini - Supported -->
          <div
            v-reveal="{ delay: 160 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/70 bg-white/65 px-5 py-3 ring-1 ring-cyan-900/10 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.07] dark:ring-white/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-blue-600"
            >
              <span class="text-xs font-bold text-white">G</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ t('home.providers.gemini') }}</span>
            <span
              class="rounded-md bg-cyan-100 px-1.5 py-0.5 text-[10px] font-bold text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Antigravity - Supported -->
          <div
            v-reveal="{ delay: 240 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/70 bg-white/65 px-5 py-3 ring-1 ring-cyan-900/10 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.07] dark:ring-white/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-rose-500 to-pink-600"
            >
              <span class="text-xs font-bold text-white">A</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ t('home.providers.antigravity') }}</span>
            <span
              class="rounded-md bg-cyan-100 px-1.5 py-0.5 text-[10px] font-bold text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- Kiro - Supported -->
          <div
            v-reveal="{ delay: 320 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/70 bg-white/65 px-5 py-3 ring-1 ring-cyan-900/10 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.07] dark:ring-white/10"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500 to-indigo-600"
            >
              <span class="text-xs font-bold text-white">K</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ t('home.providers.kiro') }}</span>
            <span
              class="rounded-md bg-cyan-100 px-1.5 py-0.5 text-[10px] font-bold text-cyan-700 dark:bg-cyan-400/10 dark:text-cyan-300"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- More - Coming Soon -->
          <div
            v-reveal="{ delay: 400, opacity: 0.6 }"
            class="provider-chip flex items-center gap-2 rounded-2xl border border-white/60 bg-white/40 px-5 py-3 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.04]"
          >
            <div
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-gray-500 to-gray-600"
            >
              <span class="text-xs font-bold text-white">+</span>
            </div>
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ t('home.providers.more') }}</span>
            <span
              class="rounded-md bg-slate-100 px-1.5 py-0.5 text-[10px] font-bold text-slate-500 dark:bg-white/10 dark:text-slate-400"
              >{{ t('home.providers.soon') }}</span
            >
          </div>
        </div>

        <!-- Product Narrative -->
        <section
          v-reveal="{ delay: 0 }"
          class="narrative-panel mb-16 overflow-hidden rounded-[2.5rem] border border-white/70 bg-slate-950 text-white shadow-[0_35px_120px_-60px_rgba(8,47,73,0.8)] dark:border-white/10"
        >
          <div class="relative grid gap-8 p-7 md:p-10 lg:grid-cols-[0.95fr_1.05fr]">
            <div class="panel-aurora absolute inset-0"></div>
            <div class="relative">
              <div class="mb-4 text-xs font-bold uppercase tracking-[0.34em] text-cyan-200">
                {{ t('home.showcase.narrative.eyebrow') }}
              </div>
              <h2 class="mb-4 text-3xl font-black leading-tight tracking-[-0.04em] md:text-4xl">
                {{ t('home.showcase.narrative.title') }}
              </h2>
              <p class="max-w-xl text-sm leading-7 text-slate-300">
                {{ t('home.showcase.narrative.description', { siteName }) }}
              </p>
            </div>
            <div class="relative grid gap-3 sm:grid-cols-2">
              <div
                v-for="item in operationCards"
                :key="item.title"
                class="rounded-3xl border border-white/10 bg-white/[0.07] p-5 backdrop-blur-xl"
              >
                <div class="mb-3 text-2xl">{{ item.badge }}</div>
                <h3 class="mb-2 text-base font-bold text-white">{{ item.title }}</h3>
                <p class="text-xs leading-6 text-slate-300">{{ item.desc }}</p>
              </div>
            </div>
          </div>
        </section>

        <!-- Workflow Strip -->
        <section class="grid gap-4 md:grid-cols-3">
          <div
            v-for="(item, index) in workflowItems"
            :key="item.title"
            v-reveal="{ delay: index * 90 }"
            class="workflow-card rounded-[2rem] border border-white/70 bg-white/65 p-6 backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.06]"
          >
            <div class="mb-5 flex items-center justify-between">
              <span class="text-xs font-black uppercase tracking-[0.26em] text-cyan-700 dark:text-cyan-300">
                0{{ index + 1 }}
              </span>
              <span class="h-px flex-1 bg-gradient-to-r from-cyan-300/70 to-transparent ml-4"></span>
            </div>
            <h3 class="mb-3 text-xl font-black tracking-[-0.03em] text-slate-950 dark:text-white">
              {{ item.title }}
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">{{ item.desc }}</p>
          </div>
        </section>
      </div>
    </main>

    <!-- Footer -->
    <footer class="site-footer relative z-10 border-t border-gray-200/50 px-6 py-8 dark:border-dark-800/50">
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="text-sm text-gray-500 dark:text-dark-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Scroll-reveal directive: fades & slides elements into view, with optional
// stagger delay and a custom final opacity. Respects reduced-motion preferences.
const prefersReducedMotion =
  typeof window !== 'undefined' &&
  window.matchMedia &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

const vReveal = {
  mounted(el: HTMLElement, binding: { value?: { delay?: number; opacity?: number } }) {
    const delay = binding.value?.delay ?? 0
    const finalOpacity = binding.value?.opacity ?? 1
    el.style.setProperty('--reveal-final-opacity', String(finalOpacity))

    if (prefersReducedMotion) {
      el.style.opacity = String(finalOpacity)
      return
    }

    el.classList.add('reveal')
    el.style.transitionDelay = `${delay}ms`

    const reveal = () => el.classList.add('reveal-visible')
    const observer = new IntersectionObserver(
      (entries, obs) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            reveal()
            obs.unobserve(entry.target)
          }
        })
      },
      { threshold: 0.15 }
    )
    observer.observe(el)
  },
}

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const rawSiteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '')
const siteSubtitle = computed(() => {
  const subtitle = rawSiteSubtitle.value.trim()
  if (!subtitle || /subscription(?:\s+to)?\s+api\s+conversion|ai api gateway|sub2api/i.test(subtitle)) {
    return t('home.showcase.defaultSubtitle')
  }
  return subtitle
})
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const operationCards = computed(() => [
  {
    badge: '01',
    title: t('home.showcase.operationCards.start.title'),
    desc: t('home.showcase.operationCards.start.description'),
  },
  {
    badge: '02',
    title: t('home.showcase.operationCards.value.title'),
    desc: t('home.showcase.operationCards.value.description'),
  },
  {
    badge: '03',
    title: t('home.showcase.operationCards.balance.title'),
    desc: t('home.showcase.operationCards.balance.description'),
  },
  {
    badge: '04',
    title: t('home.showcase.operationCards.stability.title'),
    desc: t('home.showcase.operationCards.stability.description'),
  },
])

const workflowItems = computed(() => [
  {
    title: t('home.showcase.workflow.addCredits.title'),
    desc: t('home.showcase.workflow.addCredits.description'),
  },
  {
    title: t('home.showcase.workflow.copySetup.title'),
    desc: t('home.showcase.workflow.copySetup.description'),
  },
  {
    title: t('home.showcase.workflow.viewUsage.title'),
    desc: t('home.showcase.workflow.viewUsage.description'),
  },
])

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Cursor-following spotlight: track pointer position as CSS vars on the root,
// throttled via requestAnimationFrame for smoothness.
const pageRef = ref<HTMLElement | null>(null)
const pointerActive = ref(false)
let rafId = 0
function onPointerMove(e: MouseEvent) {
  if (prefersReducedMotion) return
  const el = pageRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  if (!pointerActive.value) pointerActive.value = true
  if (rafId) return
  rafId = requestAnimationFrame(() => {
    el.style.setProperty('--mx', `${x}px`)
    el.style.setProperty('--my', `${y}px`)
    rafId = 0
  })
}

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* AI Access Pass */
.product-cockpit {
  position: relative;
  isolation: isolate;
  width: min(570px, calc(100vw - 48px));
}

.product-cockpit::before,
.product-cockpit::after {
  content: '';
  position: absolute;
  pointer-events: none;
  z-index: -1;
}

.product-cockpit::before {
  inset: 12% -4% -8% 18%;
  border-radius: 46px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.32), rgba(16, 185, 129, 0.26));
  filter: blur(42px);
}

.product-cockpit::after {
  top: -18px;
  right: 26px;
  height: 72px;
  width: 72px;
  border: 1px solid rgba(8, 145, 178, 0.18);
  border-radius: 22px;
  transform: rotate(14deg);
}

.cockpit-frame {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.09);
  border-radius: 34px;
  background: rgba(255, 255, 255, 0.86);
  box-shadow:
    0 40px 100px -54px rgba(8, 47, 73, 0.62),
    inset 0 1px 0 rgba(255, 255, 255, 0.92);
  padding: 22px;
  backdrop-filter: blur(28px);
  transform: perspective(1400px) rotateX(1.5deg) rotateY(-3deg);
  transition:
    transform 0.35s cubic-bezier(0.22, 1, 0.36, 1),
    box-shadow 0.35s ease;
}

.cockpit-frame:hover {
  box-shadow:
    0 46px 120px -58px rgba(8, 145, 178, 0.75),
    inset 0 1px 0 rgba(255, 255, 255, 0.94);
  transform: perspective(1400px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

.page-root.is-dark .cockpit-frame {
  border-color: rgba(103, 232, 249, 0.14);
  background:
    linear-gradient(145deg, rgba(8, 18, 31, 0.98), rgba(5, 14, 25, 0.96)),
    radial-gradient(circle at 18% 0%, rgba(34, 211, 238, 0.12), transparent 38%);
  box-shadow:
    0 42px 110px -48px rgba(6, 182, 212, 0.36),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.page-root.is-dark .cockpit-frame:hover {
  box-shadow:
    0 48px 130px -52px rgba(34, 211, 238, 0.52),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.cockpit-header,
.cockpit-brand,
.cockpit-brand > div:last-child,
.cockpit-live,
.pass-topline,
.pass-metrics,
.service-row,
.access-footer {
  display: flex;
  align-items: center;
}

.cockpit-header {
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 18px;
}

.cockpit-brand {
  gap: 11px;
  min-width: 0;
}

.cockpit-logo {
  display: grid;
  flex: 0 0 auto;
  height: 38px;
  width: 38px;
  place-items: center;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 13px;
  background: white;
  padding: 4px;
  box-shadow: 0 10px 28px -18px rgba(8, 47, 73, 0.8);
}

.cockpit-logo img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.cockpit-brand > div:last-child {
  align-items: flex-start;
  flex-direction: column;
  min-width: 0;
}

.cockpit-brand strong {
  max-width: 230px;
  overflow: hidden;
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cockpit-brand span {
  margin-top: 2px;
  color: #64748b;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.2em;
}

.page-root.is-dark .cockpit-brand strong {
  color: #f8fafc;
}

.page-root.is-dark .cockpit-brand span {
  color: #64748b;
}

.cockpit-live {
  flex: 0 0 auto;
  gap: 7px;
  border: 1px solid rgba(16, 185, 129, 0.18);
  border-radius: 999px;
  background: rgba(236, 253, 245, 0.86);
  padding: 7px 10px;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.cockpit-live i {
  height: 6px;
  width: 6px;
  border-radius: 999px;
  background: #10b981;
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.12), 0 0 16px rgba(16, 185, 129, 0.65);
}

.page-root.is-dark .cockpit-live {
  border-color: rgba(52, 211, 153, 0.18);
  background: rgba(16, 185, 129, 0.08);
  color: #6ee7b7;
}

.access-pass {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  min-height: 246px;
  border: 1px solid rgba(125, 211, 252, 0.16);
  border-radius: 28px;
  background:
    radial-gradient(circle at 84% 22%, rgba(45, 212, 191, 0.3), transparent 27%),
    linear-gradient(135deg, #07111f 0%, #0a2434 58%, #0d3a42 100%);
  padding: 26px;
  color: white;
  box-shadow:
    0 28px 70px -40px rgba(8, 47, 73, 0.95),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.pass-grid,
.pass-glow {
  position: absolute;
  pointer-events: none;
  z-index: -1;
}

.pass-grid {
  inset: 0;
  opacity: 0.34;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.08) 1px, transparent 1px);
  background-size: 34px 34px;
  mask-image: linear-gradient(115deg, transparent 8%, black 42%, transparent 94%);
}

.pass-glow {
  right: -48px;
  bottom: -72px;
  height: 230px;
  width: 230px;
  border-radius: 50%;
  background: rgba(45, 212, 191, 0.18);
  filter: blur(16px);
}

.pass-topline {
  justify-content: space-between;
  gap: 16px;
  color: #67e8f9;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.2em;
}

.access-pass h2 {
  margin: 36px 0 10px;
  font-size: clamp(26px, 2.65vw, 34px);
  font-weight: 950;
  letter-spacing: -0.05em;
  line-height: 1.12;
  white-space: nowrap;
}

.access-pass > p {
  margin: 0;
  color: #a5f3fc;
  font-size: 13px;
  line-height: 1.8;
}

.pass-metrics {
  justify-content: space-between;
  gap: 14px;
  margin-top: 30px;
  padding-top: 18px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.pass-metrics > div {
  min-width: 0;
}

.pass-metrics span,
.pass-metrics strong {
  display: block;
}

.pass-metrics span {
  color: #64748b;
  font-size: 9px;
  font-weight: 900;
  letter-spacing: 0.18em;
}

.pass-metrics strong {
  margin-top: 5px;
  color: #f8fafc;
  font-size: 12px;
  font-weight: 900;
}

.service-rail {
  margin-top: 16px;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.58);
}

.page-root.is-dark .service-rail {
  border-color: rgba(148, 163, 184, 0.13);
  background: rgba(15, 30, 46, 0.72);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.035);
}

.service-row {
  gap: 13px;
  min-height: 62px;
  padding: 12px 15px;
}

.service-row + .service-row {
  border-top: 1px solid rgba(15, 23, 42, 0.07);
}

.page-root.is-dark .service-row + .service-row {
  border-top-color: rgba(148, 163, 184, 0.1);
}

.service-index {
  display: grid;
  flex: 0 0 auto;
  height: 32px;
  width: 32px;
  place-items: center;
  border-radius: 11px;
  background: #ecfeff;
  color: #0891b2;
  font-size: 10px;
  font-weight: 950;
}

.page-root.is-dark .service-index {
  background: rgba(34, 211, 238, 0.1);
  color: #67e8f9;
}

.service-row > div {
  min-width: 0;
  flex: 1;
}

.service-row strong,
.service-row > div span {
  display: block;
}

.service-row strong {
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
}

.service-row > div span {
  margin-top: 3px;
  overflow: hidden;
  color: #64748b;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.service-row > i {
  color: #0f766e;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 9px;
  font-style: normal;
  font-weight: 900;
  letter-spacing: 0.12em;
}

.page-root.is-dark .service-row strong {
  color: #f1f5f9;
}

.page-root.is-dark .service-row > div span {
  color: #94a3b8;
}

.page-root.is-dark .service-row > i {
  color: #5eead4;
}

.access-footer {
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
  color: #475569;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.access-footer i {
  position: relative;
  height: 1px;
  width: 32px;
  background: linear-gradient(90deg, rgba(8, 145, 178, 0.18), rgba(8, 145, 178, 0.72));
}

.access-footer i::after {
  content: '';
  position: absolute;
  right: 0;
  top: -2px;
  height: 5px;
  width: 5px;
  border-top: 1px solid #0891b2;
  border-right: 1px solid #0891b2;
  transform: rotate(45deg);
}

.page-root.is-dark .access-footer {
  color: #94a3b8;
}

@media (max-width: 640px) {
  .cockpit-frame {
    padding: 16px;
    transform: none;
  }

  .cockpit-brand strong {
    max-width: 140px;
  }

  .access-pass {
    padding: 22px;
  }

  .access-pass h2 {
    font-size: clamp(24px, 8vw, 30px);
    white-space: normal;
  }

  .pass-metrics {
    align-items: flex-start;
    flex-direction: column;
  }

  .access-footer {
    gap: 7px;
  }

  .access-footer i {
    width: 18px;
  }
}

/* ===== Scroll Reveal ===== */
.reveal {
  opacity: 0;
  transform: translateY(24px);
  transition:
    opacity 0.7s cubic-bezier(0.22, 1, 0.36, 1),
    transform 0.7s cubic-bezier(0.22, 1, 0.36, 1);
  will-change: opacity, transform;
}
.reveal-visible {
  opacity: var(--reveal-final-opacity, 1);
  transform: translateY(0);
}

/* ===== Atmospheric Background ===== */
.color-wash {
  background:
    radial-gradient(circle at 10% 18%, rgba(34, 211, 238, 0.11), transparent 28%),
    radial-gradient(circle at 86% 18%, rgba(20, 184, 166, 0.12), transparent 30%),
    radial-gradient(circle at 70% 72%, rgba(245, 158, 11, 0.08), transparent 34%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(235, 248, 255, 0.9));
}
.page-root.is-dark .color-wash {
  background:
    radial-gradient(circle at 14% 12%, rgba(34, 211, 238, 0.12), transparent 27%),
    radial-gradient(circle at 86% 16%, rgba(16, 185, 129, 0.1), transparent 24%),
    radial-gradient(circle at 66% 70%, rgba(14, 116, 144, 0.1), transparent 30%),
    linear-gradient(145deg, #050a12 0%, #071421 48%, #05101b 100%);
}

.noise-layer {
  opacity: 0.18;
  background-image:
    radial-gradient(rgba(15, 23, 42, 0.18) 0.7px, transparent 0.7px);
  background-size: 12px 12px;
  mask-image: linear-gradient(to bottom, black, transparent 88%);
}

.page-root.is-dark .noise-layer {
  opacity: 0.24;
  background-image: radial-gradient(rgba(148, 163, 184, 0.14) 0.65px, transparent 0.65px);
}

.orb {
  filter: blur(46px);
  opacity: 0.26;
  will-change: transform;
}
.page-root.is-dark .orb {
  opacity: 0.3;
}
.orb-1 {
  background: #22d3ee;
  animation: float-1 16s ease-in-out infinite;
}
.orb-2 {
  background: #34d399;
  animation: float-2 20s ease-in-out infinite;
}
.orb-3 {
  background: #f59e0b;
  animation: float-3 18s ease-in-out infinite;
}

@keyframes float-1 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(-40px, 30px) scale(1.1);
  }
}
@keyframes float-2 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(50px, -30px) scale(1.12);
  }
}
@keyframes float-3 {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(30px, 40px) scale(0.92);
  }
}
/* ===== Grid Drift ===== */
.grid-overlay {
  animation: grid-pan 40s linear infinite;
}

.page-root.is-dark .grid-overlay {
  opacity: 0.72;
}
@keyframes grid-pan {
  from {
    background-position: 0 0;
  }
  to {
    background-position: 64px 64px;
  }
}

/* ===== Theme Surfaces ===== */
.site-header {
  transition:
    background-color 0.3s ease,
    border-color 0.3s ease;
}

.nav-control,
.nav-locale {
  min-height: 40px;
}

.nav-control :deep(svg) {
  height: 19px;
  width: 19px;
  stroke-width: 1.8;
}

.nav-control:focus-visible,
.nav-locale:focus-within,
.cta-orbit:focus-visible {
  outline: 3px solid rgba(6, 182, 212, 0.22);
  outline-offset: 3px;
}

.page-root.is-dark .site-header {
  border-bottom: 1px solid rgba(148, 163, 184, 0.07);
  background: linear-gradient(180deg, rgba(5, 10, 18, 0.82), rgba(5, 10, 18, 0.24));
  backdrop-filter: blur(18px);
}

.page-root.is-dark .nav-control,
.page-root.is-dark .nav-locale {
  border-color: rgba(125, 211, 252, 0.13);
  background: rgba(8, 22, 36, 0.84);
  box-shadow:
    0 16px 40px -28px rgba(34, 211, 238, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.055);
}

.page-root.is-dark .nav-control {
  color: #cbd5e1;
}

.page-root.is-dark .nav-control:hover {
  border-color: rgba(103, 232, 249, 0.3);
  background: rgba(11, 34, 52, 0.96);
  color: #67e8f9;
}

.page-root.is-dark .feature-tag,
.page-root.is-dark .feature-card,
.page-root.is-dark .provider-chip,
.page-root.is-dark .workflow-card {
  border-color: rgba(125, 211, 252, 0.11);
  background: linear-gradient(145deg, rgba(13, 29, 45, 0.9), rgba(7, 18, 31, 0.86));
  box-shadow:
    0 28px 70px -48px rgba(0, 0, 0, 0.9),
    inset 0 1px 0 rgba(255, 255, 255, 0.045);
}

.page-root.is-dark .feature-tag {
  background: rgba(9, 25, 40, 0.82);
}

.page-root.is-dark .feature-card:hover,
.page-root.is-dark .provider-chip:hover,
.page-root.is-dark .workflow-card:hover {
  border-color: rgba(103, 232, 249, 0.25);
  box-shadow:
    0 32px 82px -46px rgba(6, 182, 212, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.page-root.is-dark .narrative-panel {
  border-color: rgba(103, 232, 249, 0.13);
  background:
    radial-gradient(circle at 8% 0%, rgba(34, 211, 238, 0.12), transparent 28%),
    linear-gradient(135deg, #07111f, #081827 62%, #07121e);
  box-shadow: 0 38px 110px -62px rgba(34, 211, 238, 0.42);
}

.page-root.is-dark .site-footer {
  border-top-color: rgba(148, 163, 184, 0.08);
  background: rgba(3, 8, 15, 0.32);
}

/* ===== Feature Tags ===== */
.feature-tag {
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}
.feature-tag:hover {
  transform: translateY(-3px);
  box-shadow: 0 14px 30px -18px rgba(8, 145, 178, 0.55);
}

/* ===== Feature Icon subtle float on card hover ===== */
.feature-icon {
  transition:
    transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.35s ease;
}

/* ===== Provider Chips ===== */
.provider-chip {
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}
.provider-chip:hover {
  transform: translateY(-3px) scale(1.03);
  box-shadow: 0 18px 36px -22px rgba(8, 145, 178, 0.62);
}

/* ===== Brand text shine ===== */
.brand-shine {
  position: relative;
  display: block;
  width: fit-content;
  max-width: 100%;
}

.brand-shine::after {
  content: attr(data-text);
  position: absolute;
  inset: 0;
  /* ponytail: 非 repeating + 周期末停顿，避免斜角 repeat 接缝闪一下 */
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
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  -webkit-text-fill-color: transparent;
  pointer-events: none;
  animation: brand-text-shine 4.5s ease-in-out infinite;
}

@keyframes brand-text-shine {
  0%,
  18% {
    background-position: 120% 50%;
  }
  52%,
  100% {
    background-position: -20% 50%;
  }
}

/* ===== CTA Button ===== */
.cta-orbit {
  position: relative;
  overflow: hidden;
  background-image: linear-gradient(135deg, #a7f3d0, #67e8f9 48%, #fcd34d);
  transition:
    background-image 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.2s ease;
}
.cta-orbit:hover {
  box-shadow: 0 18px 42px -18px rgba(8, 145, 178, 0.8);
  transform: translateY(-2px);
}
.cta-orbit::after {
  content: '';
  position: absolute;
  top: 0;
  left: -150%;
  width: 60%;
  height: 100%;
  background: linear-gradient(
    120deg,
    transparent,
    rgba(255, 255, 255, 0.45),
    transparent
  );
  transform: skewX(-20deg);
  animation: cta-sweep 3.5s ease-in-out infinite;
}
@keyframes cta-sweep {
  0%,
  60% {
    left: -150%;
  }
  100% {
    left: 150%;
  }
}
.cta-arrow {
  transition: transform 0.3s ease;
}
.cta-orbit:hover .cta-arrow {
  transform: translateX(4px);
}

/* ===== Cursor-following Spotlight ===== */
.cursor-glow {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 0.4s ease;
  background: radial-gradient(
    320px circle at var(--mx, 50%) var(--my, 50%),
    rgba(34, 211, 238, 0.18),
    rgba(16, 185, 129, 0.08) 40%,
    transparent 70%
  );
}
.cursor-glow-active {
  opacity: 1;
}
.page-root.is-dark .cursor-glow {
  background: radial-gradient(
    300px circle at var(--mx, 50%) var(--my, 50%),
    rgba(34, 211, 238, 0.12),
    rgba(16, 185, 129, 0.05) 40%,
    transparent 70%
  );
}

.panel-aurora {
  background:
    radial-gradient(circle at 20% 20%, rgba(34, 211, 238, 0.22), transparent 28%),
    radial-gradient(circle at 78% 0%, rgba(16, 185, 129, 0.18), transparent 34%),
    radial-gradient(circle at 70% 82%, rgba(251, 191, 36, 0.12), transparent 38%);
}

.workflow-card {
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}
.workflow-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 24px 70px -45px rgba(8, 47, 73, 0.72);
}

/* ===== Header Hover Pop ===== */
.hover-pop {
  transition:
    transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}
.hover-pop:hover {
  transform: translateY(-2px) scale(1.06);
}
.hover-pop:active {
  transform: translateY(0) scale(0.96);
}

.logo-box {
  transition:
    transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.35s ease;
}
.logo-box:hover {
  transform: rotate(-6deg) scale(1.12);
  box-shadow: 0 16px 32px -18px rgba(8, 145, 178, 0.75);
}

/* ===== Reduced Motion ===== */
@media (prefers-reduced-motion: reduce) {
  .orb,
  .grid-overlay,
  .cta-orbit::after,
  .brand-shine::after {
    animation: none !important;
  }
  .brand-shine::after {
    content: none;
  }
  .cursor-glow {
    display: none;
  }
  .reveal {
    opacity: var(--reveal-final-opacity, 1);
    transform: none;
    transition: none;
  }
}
</style>
