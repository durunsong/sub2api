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
    class="page-root relative flex min-h-screen flex-col overflow-hidden bg-[#eef3ff] text-slate-950 dark:bg-[#07111f] dark:text-white"
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
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="logo-box h-11 w-11 overflow-hidden rounded-2xl border border-white/70 bg-white/70 p-1 shadow-lg shadow-cyan-900/10 backdrop-blur-md dark:border-white/10 dark:bg-white/10">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="hidden leading-tight sm:block">
            <div class="text-sm font-semibold tracking-[0.24em] text-slate-900 dark:text-white">
              {{ siteName }}
            </div>
            <div class="text-[11px] uppercase tracking-[0.28em] text-cyan-700/80 dark:text-cyan-300/80">
              AI API 服务
            </div>
          </div>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="hover-pop rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="hover-pop rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
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
            <div
              v-reveal="{ delay: 0 }"
              class="mb-5 inline-flex items-center gap-2 rounded-full border border-cyan-300/40 bg-white/70 px-3 py-1.5 text-xs font-semibold tracking-[0.22em] text-cyan-800 shadow-sm shadow-cyan-900/5 backdrop-blur dark:border-cyan-300/20 dark:bg-white/10 dark:text-cyan-200"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 shadow-[0_0_16px_rgba(52,211,153,0.8)]"></span>
              稳定 AI API · Token 按量使用
            </div>
            <h1
              v-reveal="{ delay: 80 }"
              class="mb-5 text-5xl font-black leading-[0.95] tracking-[-0.06em] text-slate-950 dark:text-white md:text-6xl lg:text-7xl"
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
              无需自己维护账号和复杂配置，登录后即可查看额度、用量与账单。适合日常编码、自动化脚本和团队协作场景，重点就是稳定、透明、好用。
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
                按量使用 · 余额清晰 · 长期可用
              </div>
            </div>
          </div>

          <!-- Right: Terminal Animation -->
          <div v-reveal="{ delay: 220 }" class="flex justify-center lg:justify-end">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">service dashboard</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="console-kpis">
                    <div>
                      <span>ACCESS</span>
                      <strong>多模型</strong>
                    </div>
                    <div>
                      <span>RESPONSE</span>
                      <strong>快速</strong>
                    </div>
                    <div>
                      <span>STATUS</span>
                      <strong>可用</strong>
                    </div>
                  </div>
                  <div class="code-line line-1">
                    <span class="code-prompt">$</span>
                    <span class="code-cmd">curl</span>
                    <span class="code-flag">-X POST</span>
                    <span class="code-url">/v1/smart-route</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-comment"># balance checked · token usage tracked · ready to work</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">200 OK</span>
                    <span class="code-response">{ "status": "ready", "billing": "metered" }</span>
                  </div>
                  <div class="code-line line-4">
                    <span class="code-prompt">$</span>
                    <span class="cursor"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Feature Tags - Centered -->
        <div class="mb-12 flex flex-wrap items-center justify-center gap-3 md:gap-4">
          <div
            v-reveal="{ delay: 0 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-white/70 bg-white/70 px-5 py-2.5 shadow-sm shadow-cyan-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="swap" size="sm" class="text-cyan-600 dark:text-cyan-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              API 现成可用
            </span>
          </div>
          <div
            v-reveal="{ delay: 100 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-white/70 bg-white/70 px-5 py-2.5 shadow-sm shadow-emerald-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="shield" size="sm" class="text-emerald-600 dark:text-emerald-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              Token 用量清晰
            </span>
          </div>
          <div
            v-reveal="{ delay: 200 }"
            class="feature-tag inline-flex items-center gap-2.5 rounded-full border border-white/70 bg-white/70 px-5 py-2.5 shadow-sm shadow-amber-950/5 backdrop-blur-xl dark:border-white/10 dark:bg-white/10"
          >
            <Icon name="chart" size="sm" class="text-amber-600 dark:text-amber-300" />
            <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              按量更省心
            </span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-16 grid gap-5 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div
            v-reveal="{ delay: 0 }"
            class="feature-card group rounded-[2rem] border border-white/70 bg-white/70 p-7 shadow-[0_24px_80px_-48px_rgba(8,47,73,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
          >
            <div
              class="feature-icon mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-cyan-300 to-cyan-600 shadow-lg shadow-cyan-500/25 transition-transform duration-300 group-hover:scale-110 group-hover:-rotate-6"
            >
              <Icon name="server" size="lg" class="text-slate-950" />
            </div>
            <h3 class="mb-3 text-xl font-black tracking-[-0.03em] text-slate-950 dark:text-white">
              一键使用
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              登录后即可获取接入信息，适合编码工具、自动化脚本和业务项目快速开始。
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            v-reveal="{ delay: 120 }"
            class="feature-card group rounded-[2rem] border border-white/70 bg-white/70 p-7 shadow-[0_24px_80px_-48px_rgba(6,78,59,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
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
              长期可用
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              面向日常高频使用场景，减少账号、额度和配置问题带来的中断。
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            v-reveal="{ delay: 240 }"
            class="feature-card group rounded-[2rem] border border-white/70 bg-white/70 p-7 shadow-[0_24px_80px_-48px_rgba(146,64,14,0.6)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1.5 dark:border-white/10 dark:bg-white/[0.07]"
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
              账单透明
            </h3>
            <p class="text-sm leading-7 text-slate-600 dark:text-slate-400">
              Token、余额和明细统一展示，个人和团队都能看清每一次消耗。
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div v-reveal="{ delay: 0 }" class="mb-8 text-center">
          <div class="mb-3 text-xs font-bold uppercase tracking-[0.32em] text-cyan-700 dark:text-cyan-300">
            available access
          </div>
          <h2 class="mb-3 text-3xl font-black tracking-[-0.04em] text-slate-950 dark:text-white">
            可用 API 与 Token 服务
          </h2>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            一份余额，多种选择，按需调用。
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
          class="mb-16 overflow-hidden rounded-[2.5rem] border border-white/70 bg-slate-950 text-white shadow-[0_35px_120px_-60px_rgba(8,47,73,0.8)] dark:border-white/10"
        >
          <div class="relative grid gap-8 p-7 md:p-10 lg:grid-cols-[0.95fr_1.05fr]">
            <div class="panel-aurora absolute inset-0"></div>
            <div class="relative">
              <div class="mb-4 text-xs font-bold uppercase tracking-[0.34em] text-cyan-200">
                why customers choose it
              </div>
              <h2 class="mb-4 text-3xl font-black leading-tight tracking-[-0.04em] md:text-4xl">
                买到的不只是接口，而是一套省心的 AI 使用体验。
              </h2>
              <p class="max-w-xl text-sm leading-7 text-slate-300">
                {{ siteName }} 面向需要长期使用 AI API 和 Token 的用户，把接入、充值、用量查看、额度管理放在同一个入口里。少折腾配置，多把时间花在真正的工作上。
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
    <footer class="relative z-10 border-t border-gray-200/50 px-6 py-8 dark:border-dark-800/50">
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
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AI API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const rawSiteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '稳定好用的 AI API 与 Token 服务')
const siteSubtitle = computed(() => {
  const subtitle = rawSiteSubtitle.value.trim()
  if (!subtitle || /subscription api conversion|ai api gateway|sub2api/i.test(subtitle)) {
    return '稳定好用的 AI API 与 Token 服务'
  }
  return subtitle
})
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const operationCards = [
  {
    badge: '01',
    title: '一键开始',
    desc: '登录后即可查看接入信息和可用额度，新手也能快速用起来。',
  },
  {
    badge: '02',
    title: '按量更划算',
    desc: '用多少扣多少，余额、消耗和明细都能看清楚，不怕糊涂账。',
  },
  {
    badge: '03',
    title: '额度好管理',
    desc: '个人或团队都能集中查看 Token 用量，避免无感超支。',
  },
  {
    badge: '04',
    title: '长期稳定',
    desc: '面向日常高频使用场景，重点保障可用性和持续服务体验。',
  },
]

const workflowItems = [
  {
    title: '购买额度',
    desc: '按需充值或开通可用额度，费用透明，适合个人开发者和小团队。',
  },
  {
    title: '复制接入',
    desc: '在控制台获取 API Key 和接入说明，直接用于编码工具、脚本或业务项目。',
  },
  {
    title: '查看用量',
    desc: '随时查看 Token 消耗、余额和账单记录，用得明白，续费也更安心。',
  },
]

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
/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
}

.terminal-container::before {
  content: '';
  position: absolute;
  inset: -18px;
  border-radius: 34px;
  background:
    linear-gradient(135deg, rgba(34, 211, 238, 0.3), rgba(16, 185, 129, 0.18), transparent 70%),
    radial-gradient(circle at 30% 10%, rgba(251, 191, 36, 0.2), transparent 34%);
  filter: blur(10px);
  opacity: 0.9;
}

/* Terminal Window */
.terminal-window {
  position: relative;
  width: min(520px, calc(100vw - 48px));
  background:
    linear-gradient(145deg, rgba(4, 13, 27, 0.98) 0%, rgba(8, 31, 49, 0.96) 54%, rgba(2, 6, 23, 0.98) 100%);
  border-radius: 24px;
  box-shadow:
    0 30px 80px -30px rgba(8, 47, 73, 0.75),
    0 0 0 1px rgba(125, 211, 252, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.12);
  overflow: hidden;
  transform: perspective(1100px) rotateX(3deg) rotateY(-5deg);
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}

.terminal-window:hover {
  box-shadow:
    0 36px 100px -36px rgba(8, 145, 178, 0.72),
    0 0 0 1px rgba(125, 211, 252, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.16);
  transform: perspective(1100px) rotateX(0deg) rotateY(0deg) translateY(-6px);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 14px 18px;
  background: linear-gradient(90deg, rgba(15, 23, 42, 0.8), rgba(8, 47, 73, 0.38));
  border-bottom: 1px solid rgba(125, 211, 252, 0.12);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #7dd3fc;
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 22px 26px 26px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: 14px;
  line-height: 2.05;
}

.console-kpis {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 18px;
}

.console-kpis div {
  border: 1px solid rgba(125, 211, 252, 0.13);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.045);
  padding: 10px 12px;
}

.console-kpis span {
  display: block;
  margin-bottom: 2px;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: #67e8f9;
}

.console-kpis strong {
  font-size: 16px;
  color: #f8fafc;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #34d399;
  font-weight: bold;
}
.code-cmd {
  color: #67e8f9;
}
.code-flag {
  color: #fbbf24;
}
.code-url {
  color: #93c5fd;
}
.code-comment {
  color: #94a3b8;
  font-style: italic;
}
.code-success {
  color: #022c22;
  background: linear-gradient(135deg, #6ee7b7, #22d3ee);
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 800;
}
.code-response {
  color: #fde68a;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #34d399;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}

/* Dark mode adjustments */
:deep(.dark) .terminal-window {
  box-shadow:
    0 35px 90px -35px rgba(34, 211, 238, 0.45),
    0 0 0 1px rgba(125, 211, 252, 0.22),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
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
    radial-gradient(circle at 12% 16%, rgba(34, 211, 238, 0.16), transparent 30%),
    radial-gradient(circle at 82% 18%, rgba(16, 185, 129, 0.18), transparent 28%),
    radial-gradient(circle at 70% 72%, rgba(245, 158, 11, 0.14), transparent 34%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.66), rgba(219, 234, 254, 0.62));
}
:deep(.dark) .color-wash {
  background:
    radial-gradient(circle at 12% 16%, rgba(34, 211, 238, 0.16), transparent 30%),
    radial-gradient(circle at 82% 18%, rgba(16, 185, 129, 0.12), transparent 28%),
    radial-gradient(circle at 70% 72%, rgba(245, 158, 11, 0.1), transparent 34%),
    linear-gradient(135deg, rgba(2, 6, 23, 0.96), rgba(8, 47, 73, 0.82));
}

.noise-layer {
  opacity: 0.28;
  background-image:
    radial-gradient(rgba(15, 23, 42, 0.18) 0.7px, transparent 0.7px);
  background-size: 12px 12px;
  mask-image: linear-gradient(to bottom, black, transparent 88%);
}

.orb {
  filter: blur(46px);
  opacity: 0.55;
  will-change: transform;
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
@keyframes grid-pan {
  from {
    background-position: 0 0;
  }
  to {
    background-position: 64px 64px;
  }
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
:deep(.dark) .cursor-glow {
  background: radial-gradient(
    300px circle at var(--mx, 50%) var(--my, 50%),
    rgba(34, 211, 238, 0.2),
    rgba(16, 185, 129, 0.08) 40%,
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
  .code-line,
  .cursor {
    animation: none !important;
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
