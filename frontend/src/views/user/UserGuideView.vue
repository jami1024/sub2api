<template>
  <AppLayout>
    <div data-testid="user-guide-view" class="space-y-8">
      <section class="rounded-3xl border border-primary-100 bg-primary-50/70 p-6 shadow-sm dark:border-primary-900/30 dark:bg-primary-950/20 sm:p-8">
        <div class="max-w-4xl">
          <p class="mb-3 inline-flex rounded-full bg-white px-3 py-1 text-xs font-semibold uppercase tracking-wide text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300">
            {{ t('userGuide.badge') }}
          </p>
          <h1 class="text-3xl font-bold tracking-tight text-gray-950 dark:text-white sm:text-4xl">{{ t('userGuide.title') }}</h1>
          <p class="mt-4 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300 sm:text-base">
            {{ t('userGuide.hero') }}
          </p>
          <div class="mt-6 flex flex-wrap gap-3">
            <RouterLink
              to="/keys"
              class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary-700"
            >
              {{ t('userGuide.actions.createKey') }}
              <Icon name="arrowRight" size="sm" />
            </RouterLink>
            <RouterLink
              to="/purchase"
              class="inline-flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-sm font-semibold text-gray-700 ring-1 ring-gray-200 transition hover:text-primary-700 dark:bg-dark-800 dark:text-dark-200 dark:ring-dark-700 dark:hover:text-primary-300"
            >
              {{ t('userGuide.actions.recharge') }}
            </RouterLink>
          </div>
        </div>
      </section>

      <div data-testid="user-guide-layout" class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_20rem] xl:items-start">
        <main class="space-y-8">
          <section id="quick-start" class="rounded-3xl border border-gray-100 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800 sm:p-6">
            <div class="mb-5 flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-primary-600 dark:text-primary-300">{{ t('userGuide.quickStart.kicker') }}</p>
                <h2 class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('userGuide.quickStart.title') }}</h2>
              </div>
              <p class="max-w-xl text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('userGuide.quickStart.description') }}</p>
            </div>

            <div class="grid gap-3 md:grid-cols-2">
              <article
                v-for="(step, index) in steps"
                :key="step.key"
                class="flex gap-4 rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/60"
                :class="index === 0 ? 'md:col-span-2' : ''"
              >
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-white text-sm font-bold text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-300">
                  {{ index + 1 }}
                </div>
                <div>
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t(`userGuide.steps.${step.key}.title`) }}</h3>
                  <p class="mt-1 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t(`userGuide.steps.${step.key}.body`) }}</p>
                </div>
              </article>
            </div>
          </section>

          <section id="client-setup" data-testid="user-guide-content-grid">
            <div class="rounded-3xl border border-gray-100 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('userGuide.client.title') }}</h2>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('userGuide.client.description') }}</p>

              <div class="mt-5 grid gap-3 md:grid-cols-3">
                <div
                  v-for="client in clients"
                  :key="client.key"
                  class="flex items-start gap-4 rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/60 md:flex-col md:gap-3"
                >
                  <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-300">
                    <Icon :name="client.icon" size="md" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t(`userGuide.client.${client.key}.title`) }}</h3>
                    <p class="mt-1 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t(`userGuide.client.${client.key}.body`) }}</p>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section id="image-generation" class="rounded-3xl border border-gray-100 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('userGuide.imageGeneration.title') }}</h2>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('userGuide.imageGeneration.description') }}</p>

            <div class="mt-5 grid gap-4 lg:grid-cols-2">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('userGuide.imageGeneration.jqExample') }}</p>
                <div class="mt-3 overflow-hidden rounded-2xl bg-gray-950 text-gray-100">
                  <pre class="overflow-x-auto p-4 text-xs leading-6"><code>curl "$BASE_URL/v1/images/generations" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "画一只在月光下写代码的橘猫",
    "size": "1024x1024",
    "quality": "high",
    "response_format": "b64_json"
  }' \
  | jq -r '.data[0].b64_json' \
  | base64 --decode > image.png</code></pre>
                </div>
              </div>

              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('userGuide.imageGeneration.noJqExample') }}</p>
                <div class="mt-3 overflow-hidden rounded-2xl bg-gray-950 text-gray-100">
                  <pre class="overflow-x-auto p-4 text-xs leading-6"><code>import base64
import json
import os
import pathlib
import urllib.request

base_url = os.environ["BASE_URL"]
api_key = os.environ["API_KEY"]

payload = {
    "model": "gpt-image-2",
    "prompt": "画一只在月光下写代码的橘猫",
    "size": "1024x1024",
    "quality": "high",
    "response_format": "b64_json",
}

request = urllib.request.Request(
    f"{base_url}/v1/images/generations",
    data=json.dumps(payload).encode("utf-8"),
    headers={
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    },
    method="POST",
)

with urllib.request.urlopen(request) as response:
    result = json.loads(response.read().decode("utf-8"))

image_data = base64.b64decode(result["data"][0]["b64_json"])
pathlib.Path("image.png").write_bytes(image_data)
print("saved image.png")</code></pre>
                </div>
              </div>
            </div>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('userGuide.imageGeneration.note') }}</p>
          </section>

          <section id="faq" class="rounded-3xl border border-gray-100 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('userGuide.faq.title') }}</h2>
            <div class="mt-5 grid gap-4 md:grid-cols-3">
              <div v-for="item in faqItems" :key="item.key" class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/60">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t(`userGuide.faq.${item.key}.title`) }}</h3>
                <p class="mt-1 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t(`userGuide.faq.${item.key}.body`) }}</p>
              </div>
            </div>
          </section>

          <details
            id="full-doc"
            data-testid="user-guide-full-doc-details"
            class="group rounded-3xl border border-gray-100 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800"
          >
            <summary class="flex cursor-pointer list-none flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('userGuide.fullClaudeDoc.title') }}</h2>
                <p class="mt-1 text-sm text-gray-600 dark:text-dark-300">{{ t('userGuide.fullClaudeDoc.description') }}</p>
              </div>
              <span class="inline-flex items-center gap-2 rounded-xl bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 transition group-open:bg-primary-50 group-open:text-primary-700 dark:bg-dark-700 dark:text-dark-200 dark:group-open:bg-primary-900/30 dark:group-open:text-primary-300">
                {{ t('userGuide.fullClaudeDoc.toggle') }}
                <Icon name="chevronDown" size="sm" class="transition group-open:rotate-180" />
              </span>
            </summary>
            <div class="mt-6 border-t border-gray-100 pt-6 dark:border-dark-700">
              <nav
                v-if="fullDocHeadings.length"
                data-testid="user-guide-full-doc-nav"
                class="mb-6 rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/60"
              >
                <p class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('userGuide.fullClaudeDoc.navTitle') }}</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <a
                    v-for="heading in fullDocHeadings"
                    :key="heading.id"
                    :href="`#${heading.id}`"
                    class="rounded-xl bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-sm ring-1 ring-gray-100 transition hover:text-primary-600 dark:bg-dark-800 dark:text-dark-200 dark:ring-dark-700 dark:hover:text-primary-300"
                  >
                    {{ heading.text }}
                  </a>
                </div>
              </nav>
              <div class="markdown-body" v-html="claudeInstallDocHtml"></div>
            </div>
          </details>
        </main>

        <aside class="hidden xl:block">
          <div class="sticky top-6 rounded-3xl border border-gray-100 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <p class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('userGuide.quickNav.title') }}</p>
            <div class="mt-4 space-y-1">
              <a
                v-for="item in quickNavItems"
                :key="item.href"
                :href="item.href"
                class="block rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-50 hover:text-primary-600 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-primary-300"
              >
                {{ t(`userGuide.quickNav.${item.key}`) }}
              </a>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import claudeInstallDoc from './docs/claude-code-install.zh.md?raw'

const { t } = useI18n()

interface FullDocHeading {
  id: string
  text: string
  depth: number
}

const stripMarkdown = (value: string) => value
  .replace(/`([^`]+)`/g, '$1')
  .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
  .replace(/[*_~#>]/g, '')
  .trim()

const fullDocHeadings = computed<FullDocHeading[]>(() => claudeInstallDoc
  .split('\n')
  .map((line) => {
    const match = /^(#{1,2})\s+(.+)$/.exec(line)

    if (!match) {
      return null
    }

    return {
      depth: match[1].length,
      text: stripMarkdown(match[2]),
    }
  })
  .filter((heading): heading is Omit<FullDocHeading, 'id'> => Boolean(heading))
  .map((heading, index) => ({
    ...heading,
    id: `full-doc-section-${index}`,
  })))

const claudeInstallDocHtml = computed(() => {
  const renderer = new marked.Renderer()
  let headingIndex = 0

  renderer.heading = ({ tokens, depth }) => {
    const text = marked.parseInline(tokens.map((token) => token.raw).join(''))
    const id = depth <= 2 ? fullDocHeadings.value[headingIndex]?.id : undefined

    if (depth <= 2) {
      headingIndex += 1
    }

    return id ? `<h${depth} id="${id}">${text}</h${depth}>\n` : `<h${depth}>${text}</h${depth}>\n`
  }

  return DOMPurify.sanitize(marked.parse(claudeInstallDoc, {
    breaks: true,
    gfm: true,
    renderer,
  }) as string)
})

const steps = [
  { key: 'fundAccount' },
  { key: 'createKey' },
  { key: 'chooseGroup' },
  { key: 'configureClient' },
  { key: 'checkUsage' },
] as const

const clients = [
  { key: 'codex', icon: 'terminal' },
  { key: 'claude', icon: 'chat' },
  { key: 'openai', icon: 'cloud' },
] as const

const quickNavItems = [
  { key: 'quickStart', href: '#quick-start' },
  { key: 'clientSetup', href: '#client-setup' },
  { key: 'imageGeneration', href: '#image-generation' },
  { key: 'faq', href: '#faq' },
  { key: 'fullDoc', href: '#full-doc' },
] as const

const faqItems = [
  { key: 'keyNotWorking' },
  { key: 'noGroup' },
  { key: 'billing' },
] as const
</script>

<style scoped>
.markdown-body {
  @apply text-[15px] leading-[1.75] text-gray-700 dark:text-gray-300;
}

.markdown-body :deep(h1) {
  @apply mb-6 mt-2 border-b border-gray-200 pb-3 text-2xl font-bold text-gray-900 dark:border-dark-600 dark:text-white;
}

.markdown-body :deep(h2) {
  @apply mb-4 mt-8 border-b border-gray-100 pb-2 text-xl font-bold text-gray-900 dark:border-dark-700 dark:text-white;
}

.markdown-body :deep(h3) {
  @apply mb-3 mt-6 text-lg font-semibold text-gray-900 dark:text-white;
}

.markdown-body :deep(h4) {
  @apply mb-2 mt-5 text-base font-semibold text-gray-900 dark:text-white;
}

.markdown-body :deep(p) {
  @apply mb-4 leading-relaxed;
}

.markdown-body :deep(a) {
  @apply font-medium text-blue-600 underline decoration-blue-600/30 decoration-2 underline-offset-2 transition-all hover:decoration-blue-600 dark:text-blue-400 dark:decoration-blue-400/30 dark:hover:decoration-blue-400;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  @apply mb-4 ml-6 space-y-2;
}

.markdown-body :deep(ul) {
  @apply list-disc;
}

.markdown-body :deep(ol) {
  @apply list-decimal;
}

.markdown-body :deep(li) {
  @apply pl-2 leading-relaxed;
}

.markdown-body :deep(blockquote) {
  @apply my-5 rounded-2xl border border-blue-100 bg-blue-50/70 px-5 py-4 text-gray-700 dark:border-blue-900/40 dark:bg-blue-900/10 dark:text-gray-300;
}

.markdown-body :deep(code) {
  @apply rounded-lg bg-gray-100 px-2 py-1 font-mono text-[13px] text-pink-600 dark:bg-dark-700 dark:text-pink-400;
}

.markdown-body :deep(pre) {
  @apply my-5 overflow-x-auto rounded-xl border border-gray-200 bg-gray-950 p-5 dark:border-dark-600;
}

.markdown-body :deep(pre code) {
  @apply bg-transparent p-0 text-[13px] text-gray-100;
}

.markdown-body :deep(hr) {
  @apply my-8 border-0 border-t border-gray-200 dark:border-dark-700;
}

.markdown-body :deep(table) {
  @apply mb-5 block w-full overflow-x-auto rounded-xl border border-gray-200 dark:border-dark-600;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  @apply border-b border-r border-gray-200 px-4 py-3 text-left dark:border-dark-600;
}

.markdown-body :deep(th:last-child),
.markdown-body :deep(td:last-child) {
  @apply border-r-0;
}

.markdown-body :deep(th) {
  @apply bg-gray-50 font-semibold text-gray-900 dark:bg-dark-900 dark:text-white;
}

.markdown-body :deep(strong) {
  @apply font-semibold text-gray-900 dark:text-white;
}
</style>
