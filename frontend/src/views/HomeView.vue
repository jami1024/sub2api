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
    class="min-h-screen bg-gradient-to-b from-slate-50 via-white to-slate-100 text-slate-900 dark:from-slate-950 dark:via-slate-950 dark:to-slate-900 dark:text-slate-100"
  >
    <header class="border-b border-slate-200/70 dark:border-white/10">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-2xl bg-slate-900/5 p-1 dark:bg-white/5">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div>
            <p class="text-sm font-semibold tracking-[0.01em] text-slate-900 dark:text-white">{{ siteName }}</p>
            <p class="text-[11px] uppercase tracking-[0.28em] text-slate-500 dark:text-slate-400">
              {{ t('home.landing.domainBadge') }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 transition hover:border-slate-400 hover:text-slate-900 dark:border-white/10 dark:text-slate-300 dark:hover:border-white/25 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <button
            type="button"
            @click="toggleTheme"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 transition hover:border-slate-400 hover:text-slate-900 dark:border-white/10 dark:text-slate-300 dark:hover:border-white/25 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-2 rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm transition hover:bg-slate-50 dark:border-white/20 dark:bg-white/10 dark:text-white dark:hover:bg-white/15"
          >
            <span
              class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-500/90 text-xs font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            {{ t('home.dashboard') }}
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-800 shadow-sm transition hover:bg-slate-50 dark:border-white/20 dark:bg-white/10 dark:text-white dark:hover:bg-white/15"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-20 px-6 py-16 lg:py-20">
      <section class="grid gap-14 lg:grid-cols-[1.15fr_0.85fr] lg:items-center">
        <div class="max-w-2xl">
          <button
            data-testid="home-domain-badge"
            type="button"
            @click="handleBadgeClick"
            class="inline-flex items-center rounded-full border border-slate-300/80 bg-white/80 px-4 py-2 text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-500 shadow-sm transition duration-300 hover:-translate-y-0.5 hover:border-slate-400 hover:text-slate-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-400/60 dark:border-white/10 dark:bg-white/5 dark:text-slate-300 dark:hover:border-white/20 dark:hover:text-white"
          >
            {{ t('home.landing.domainBadge') }}
          </button>
          <p class="mt-8 max-w-[34rem] text-[15px] leading-8 text-slate-600 dark:text-slate-300 md:text-lg">
            {{ t('home.landing.description') }}
          </p>

          <div class="relative mt-10 max-w-xl">
            <div
              class="pointer-events-none absolute inset-x-6 -top-5 bottom-0 rounded-full bg-primary-400/20 blur-3xl transition duration-300 dark:bg-primary-500/20"
              :class="isCtaHighlighted ? 'scale-100 opacity-100' : 'scale-95 opacity-0'"
            />
            <div class="relative z-10 mb-4 flex flex-wrap gap-3">
              <span
                v-for="(pill, index) in delightPills"
                :key="pill.key"
                :data-testid="`hero-delight-pill-${pill.key}`"
                class="rounded-full border border-slate-200 bg-white/80 px-3 py-2 text-xs font-medium text-slate-600 shadow-sm transition duration-300 dark:border-white/10 dark:bg-white/5 dark:text-slate-200"
                :class="[
                  isCtaHighlighted ? 'translate-y-0 scale-100 border-primary-300/60 dark:border-primary-400/40' : 'translate-y-0',
                  prefersReducedMotion ? '' : 'home-delight-pill'
                ]"
                :style="prefersReducedMotion ? undefined : { animationDelay: `${index * 160}ms` }"
              >
                {{ pill.label }}
              </span>
            </div>
            <button
              data-testid="wechat-cta"
              type="button"
              @click="handleWechatClick"
              @mouseenter="isCtaHovered = true"
              @mouseleave="isCtaHovered = false"
              @focus="isCtaHovered = true"
              @blur="isCtaHovered = false"
              class="group relative inline-flex min-h-12 items-center justify-center overflow-hidden rounded-full px-7 py-3 text-sm font-semibold text-white transition duration-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-400/60"
              :class="
                isWechatCopied
                  ? 'bg-emerald-500 shadow-[0_18px_40px_rgba(16,185,129,0.28)]'
                  : 'bg-slate-950 shadow-[0_16px_40px_rgba(15,23,42,0.18)] hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-primary-500 dark:shadow-[0_18px_40px_rgba(59,130,246,0.28)] dark:hover:bg-primary-400'
              "
            >
              <span class="pointer-events-none absolute inset-0 rounded-full border border-white/15"></span>
              <span
                class="pointer-events-none absolute inset-0 rounded-full border border-white/30 transition duration-500"
                :class="isWechatCopied ? 'scale-[1.08] opacity-100' : 'scale-100 opacity-0'"
              ></span>
              <span
                class="home-cta-sheen pointer-events-none absolute inset-y-0 left-[-35%] w-24 -skew-x-12 bg-white/20 blur-xl transition duration-700"
                :class="isCtaHighlighted ? 'translate-x-[260%] opacity-100' : 'translate-x-0 opacity-0'"
              ></span>
              <span class="relative z-10 flex items-center gap-2">
                <span>{{ ctaLabel }}</span>
                <span
                  v-if="isWechatCopied"
                  class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-white/20 text-xs font-bold"
                  aria-hidden="true"
                >
                  ✓
                </span>
              </span>
            </button>

            <p
              v-if="easterEggVisible"
              data-testid="home-easter-egg"
              class="mt-3 text-sm text-slate-500 transition duration-300 dark:text-slate-300"
            >
              {{ easterEggMessage }}
            </p>
          </div>
        </div>

        <AigoHubHeroPanel :reduced-motion="prefersReducedMotion" />
      </section>

      <HomePackageSection />

      <section>
        <h2 class="text-2xl font-bold text-slate-900 dark:text-white">{{ t('home.landing.whyTitle') }}</h2>
        <div class="mt-6 grid gap-4 md:grid-cols-3">
          <article
            v-for="item in whyItems"
            :key="item.title"
            class="rounded-2xl border border-slate-200 bg-white/70 p-5 shadow-sm dark:border-white/10 dark:bg-white/5 dark:shadow-none"
          >
            <h3 class="text-lg font-semibold text-slate-900 dark:text-white">{{ item.title }}</h3>
            <p class="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{{ item.description }}</p>
          </article>
        </div>
      </section>

      <section class="rounded-3xl border border-slate-200 bg-white/70 p-6 shadow-sm dark:border-white/10 dark:bg-white/5 dark:shadow-none">
        <h2 class="text-2xl font-bold text-slate-900 dark:text-white">{{ t('home.landing.audienceTitle') }}</h2>
        <ul class="mt-4 space-y-3">
          <li
            v-for="item in audienceItems"
            :key="item"
            class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700 dark:border-white/10 dark:bg-slate-900/70 dark:text-slate-200"
          >
            {{ item }}
          </li>
        </ul>
      </section>

      <section class="rounded-3xl border border-primary-400/30 bg-primary-500/10 p-6 text-center">
        <h2 class="text-2xl font-bold text-slate-900 dark:text-white">{{ t('home.landing.contactTitle') }}</h2>
        <p class="mx-auto mt-3 max-w-2xl text-sm leading-6 text-slate-700 dark:text-slate-200">
          {{ t('home.landing.contactDescription') }}
        </p>
        <button
          type="button"
          @click="handleWechatClick"
          class="mt-6 inline-flex items-center justify-center rounded-full bg-primary-500 px-6 py-3 text-sm font-semibold text-white transition hover:bg-primary-400"
        >
          {{ t('home.landing.primaryCta') }}
        </button>
      </section>
    </main>

    <footer class="border-t border-slate-200/80 dark:border-white/10">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-3 px-6 py-8 text-center text-sm text-slate-500 dark:text-slate-400 sm:flex-row sm:text-left">
        <div>
          <p>{{ t('home.landing.footerTagline') }}</p>
          <p class="mt-1">&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        </div>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition hover:text-slate-800 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import AigoHubHeroPanel from '@/components/home/AigoHubHeroPanel.vue'
import HomePackageSection from '@/components/home/HomePackageSection.vue'
import Icon from '@/components/icons/Icon.vue'

const WECHAT_ID = 'G000000000g1e'
const CTA_SUCCESS_RESET_MS = 2200
const EASTER_EGG_TRIGGER_COUNT = 5
const EASTER_EGG_RESET_MS = 3200

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const siteName = computed(() => {
  if (appStore.cachedPublicSettings?.site_name) return appStore.cachedPublicSettings.site_name
  if (!appStore.publicSettingsLoaded) return ''
  return appStore.siteName || 'Sub2API'
})
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isCtaHovered = ref(false)
const isWechatCopied = ref(false)
const easterEggClicks = ref(0)
const easterEggVisible = ref(false)
const prefersReducedMotion = ref(false)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => authStore.user?.email?.charAt(0).toUpperCase() || '')
const currentYear = computed(() => new Date().getFullYear())
const ctaLabel = computed(() => (isWechatCopied.value ? t('home.landing.successCta') : t('home.landing.primaryCta')))
const easterEggMessage = computed(() => t('home.landing.easterEgg'))
const isCtaHighlighted = computed(() => isCtaHovered.value || isWechatCopied.value)
const delightPills = computed(() => [
  { key: 'stable', label: t('home.landing.whyItems.stableUse.title') },
  { key: 'easy', label: t('home.landing.whyItems.easyAccess.title') },
  { key: 'fast', label: t('home.landing.delightPills.quickConsult') }
])

let ctaResetTimer: number | null = null
let easterEggTimer: number | null = null

const whyItems = computed(() => [
  {
    title: t('home.landing.whyItems.stableUse.title'),
    description: t('home.landing.whyItems.stableUse.description')
  },
  {
    title: t('home.landing.whyItems.easyAccess.title'),
    description: t('home.landing.whyItems.easyAccess.description')
  },
  {
    title: t('home.landing.whyItems.longTerm.title'),
    description: t('home.landing.whyItems.longTerm.description')
  }
])

const audienceItems = computed(() => [
  t('home.landing.audienceItems.buyers'),
  t('home.landing.audienceItems.simpleAccess'),
  t('home.landing.audienceItems.longTerm')
])

function clearCtaResetTimer() {
  if (!ctaResetTimer) return
  window.clearTimeout(ctaResetTimer)
  ctaResetTimer = null
}

function clearEasterEggTimer() {
  if (!easterEggTimer) return
  window.clearTimeout(easterEggTimer)
  easterEggTimer = null
}

function scheduleCtaReset() {
  clearCtaResetTimer()
  ctaResetTimer = window.setTimeout(() => {
    isWechatCopied.value = false
    ctaResetTimer = null
  }, CTA_SUCCESS_RESET_MS)
}

function scheduleEasterEggHide() {
  clearEasterEggTimer()
  easterEggTimer = window.setTimeout(() => {
    easterEggVisible.value = false
    easterEggClicks.value = 0
    easterEggTimer = null
  }, EASTER_EGG_RESET_MS)
}

async function handleWechatClick() {
  await copyToClipboard(WECHAT_ID, t('home.landing.copySuccess'))
  isWechatCopied.value = true
  scheduleCtaReset()
}

function handleBadgeClick() {
  easterEggClicks.value += 1
  if (easterEggClicks.value < EASTER_EGG_TRIGGER_COUNT) return

  easterEggVisible.value = true
  scheduleEasterEggHide()
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  prefersReducedMotion.value = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

onBeforeUnmount(() => {
  clearCtaResetTimer()
  clearEasterEggTimer()
})
</script>

<style scoped>
.home-delight-pill {
  animation: home-delight-pill-float 6.4s ease-in-out infinite;
}

@keyframes home-delight-pill-float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-delight-pill {
    animation: none;
  }
}
</style>
