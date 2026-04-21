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
  <div v-else class="min-h-screen bg-slate-950 text-slate-100">
    <header class="border-b border-white/10">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-xl bg-white/5 p-1">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div>
            <p class="text-sm font-semibold text-white">{{ siteName }}</p>
            <p class="text-xs uppercase tracking-[0.2em] text-slate-400">
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
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 text-slate-300 transition hover:border-white/25 hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <button
            type="button"
            @click="toggleTheme"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 text-slate-300 transition hover:border-white/25 hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-4 py-2 text-sm font-medium text-white transition hover:bg-white/15"
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
            class="inline-flex items-center rounded-full border border-white/20 bg-white/10 px-4 py-2 text-sm font-medium text-white transition hover:bg-white/15"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-16 px-6 py-12">
      <section class="grid gap-10 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
        <div>
          <h1 class="text-4xl font-black leading-tight text-white md:text-5xl">
            {{ t('home.landing.title') }}
          </h1>
          <p class="mt-4 max-w-xl text-base leading-7 text-slate-300">
            {{ t('home.landing.description') }}
          </p>

          <div class="mt-8 flex flex-col gap-4 sm:flex-row sm:items-center">
            <button
              data-testid="wechat-cta"
              type="button"
              @click="handleWechatClick"
              class="inline-flex items-center justify-center rounded-full bg-primary-500 px-6 py-3 text-sm font-semibold text-white transition hover:bg-primary-400"
            >
              {{ t('home.landing.primaryCta') }}
            </button>
            <p class="text-sm text-slate-300">
              {{ t('home.landing.wechatLabel') }}：
              <span data-testid="wechat-id" class="font-semibold text-white">{{ WECHAT_ID }}</span>
            </p>
          </div>
        </div>

        <AigoHubHeroPanel />
      </section>

      <section>
        <h2 class="text-2xl font-bold text-white">{{ t('home.landing.whyTitle') }}</h2>
        <div class="mt-6 grid gap-4 md:grid-cols-3">
          <article
            v-for="item in whyItems"
            :key="item.title"
            class="rounded-2xl border border-white/10 bg-white/5 p-5"
          >
            <h3 class="text-lg font-semibold text-white">{{ item.title }}</h3>
            <p class="mt-2 text-sm leading-6 text-slate-300">{{ item.description }}</p>
          </article>
        </div>
      </section>

      <section class="rounded-3xl border border-white/10 bg-white/5 p-6">
        <h2 class="text-2xl font-bold text-white">{{ t('home.landing.audienceTitle') }}</h2>
        <ul class="mt-4 space-y-3">
          <li
            v-for="item in audienceItems"
            :key="item"
            class="rounded-xl border border-white/10 bg-slate-900/70 px-4 py-3 text-sm text-slate-200"
          >
            {{ item }}
          </li>
        </ul>
      </section>

      <section class="rounded-3xl border border-primary-400/30 bg-primary-500/10 p-6 text-center">
        <h2 class="text-2xl font-bold text-white">{{ t('home.landing.contactTitle') }}</h2>
        <p class="mx-auto mt-3 max-w-2xl text-sm leading-6 text-slate-200">
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

    <footer class="border-t border-white/10">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-3 px-6 py-8 text-center text-sm text-slate-400 sm:flex-row sm:text-left">
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
            class="transition hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition hover:text-white"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import AigoHubHeroPanel from '@/components/home/AigoHubHeroPanel.vue'
import Icon from '@/components/icons/Icon.vue'

const WECHAT_ID = 'G000000000g1e'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'AigoHub')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => authStore.user?.email?.charAt(0).toUpperCase() || '')
const currentYear = computed(() => new Date().getFullYear())

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

async function handleWechatClick() {
  await copyToClipboard(WECHAT_ID, t('home.landing.copySuccess'))
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
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
