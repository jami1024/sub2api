<template>
  <section
    data-testid="home-package-section"
    class="relative isolate overflow-hidden rounded-[2.2rem] border border-slate-200/80 bg-white/80 p-6 shadow-[0_24px_90px_rgba(15,23,42,0.08)] dark:border-white/10 dark:bg-[#08101d]/88 dark:shadow-[0_30px_100px_rgba(2,6,23,0.62)] sm:p-8"
  >
    <div
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.12),transparent_32%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.08),transparent_28%)] dark:bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.18),transparent_30%),radial-gradient(circle_at_bottom_right,rgba(125,211,252,0.14),transparent_28%)]"
    ></div>
    <div
      class="pointer-events-none absolute inset-0 opacity-30 [background-image:radial-gradient(rgba(15,23,42,0.08)_1px,transparent_1px)] [background-size:22px_22px] dark:opacity-20 dark:[background-image:radial-gradient(rgba(255,255,255,0.12)_1px,transparent_1px)]"
    ></div>

    <div class="relative">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-2xl">
          <p
            class="inline-flex rounded-full border border-sky-200/80 bg-sky-50/90 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.28em] text-sky-700 dark:border-sky-400/20 dark:bg-sky-400/10 dark:text-sky-200"
          >
            Codex Plans
          </p>
          <h2 class="mt-4 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl">Codex 额度包</h2>
          <p class="mt-3 max-w-[62ch] text-sm leading-7 text-slate-600 dark:text-slate-300">
            当前仅支持 Codex。下列档位会直接展示人民币价格和美元额度，并按 GPT-5.4 的常见使用方式估算大致可用量。
          </p>
        </div>

        <div
          class="rounded-[1.6rem] border border-slate-200/80 bg-white/85 px-4 py-3 text-sm text-slate-600 shadow-[0_16px_40px_rgba(15,23,42,0.05)] dark:border-white/10 dark:bg-white/5 dark:text-slate-300"
        >
          <p class="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400 dark:text-slate-500">估算口径</p>
          <p class="mt-2 font-medium text-slate-900 dark:text-white">按 GPT-5.4 估算</p>
          <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">按输入:输出 = 4:1 估算</p>
        </div>
      </div>

      <div class="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-[repeat(3,minmax(0,1fr))_0.9fr]">
        <article
          v-for="card in packageCards"
          :key="card.name"
          data-testid="home-package-card"
          :data-package-kind="card.kind"
          class="group flex h-full flex-col overflow-hidden rounded-[1.8rem] border p-5 transition duration-300 sm:p-6"
          :class="
            card.kind === 'codex'
              ? 'border-slate-200/80 bg-slate-50/92 shadow-[0_18px_50px_rgba(15,23,42,0.06)] hover:-translate-y-1 hover:border-slate-300 dark:border-white/10 dark:bg-slate-950/72 dark:hover:border-white/20'
              : 'border-slate-200/70 bg-white/70 shadow-[0_14px_36px_rgba(15,23,42,0.04)] dark:border-white/10 dark:bg-white/5'
          "
        >
          <template v-if="card.kind === 'codex'">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-[11px] font-semibold uppercase tracking-[0.24em] text-slate-400 dark:text-slate-500">
                  Codex Package
                </p>
                <h3 class="mt-3 text-xl font-semibold text-slate-900 dark:text-white">
                  {{ card.name }}
                </h3>
              </div>
              <span
                class="inline-flex rounded-full border border-sky-200/80 bg-sky-50 px-3 py-1 text-xs font-semibold text-sky-700 dark:border-sky-400/20 dark:bg-sky-400/10 dark:text-sky-200"
              >
                {{ card.multiplier }}
              </span>
            </div>

            <div class="mt-7 rounded-[1.5rem] bg-slate-950 px-4 py-4 text-white dark:bg-white dark:text-slate-950">
              <p class="text-xs font-medium uppercase tracking-[0.2em] text-white/60 dark:text-slate-500">按 GPT-5.4 约可使用</p>
              <p class="mt-3 text-2xl font-semibold leading-tight sm:text-[2rem]">{{ card.estimateValue }}</p>
              <p class="mt-2 text-sm text-white/70 dark:text-slate-500">tokens</p>
            </div>

            <div class="mt-6 grid gap-3 text-sm text-slate-600 dark:text-slate-300">
              <div class="flex items-center justify-between gap-3 rounded-2xl bg-white/80 px-4 py-3 dark:bg-white/5">
                <span>价格</span>
                <span class="text-base font-semibold text-slate-900 dark:text-white">{{ card.priceLabel }}</span>
              </div>
              <div class="flex items-center justify-between gap-3 rounded-2xl bg-white/80 px-4 py-3 dark:bg-white/5">
                <span>额度</span>
                <span class="text-base font-semibold text-slate-900 dark:text-white">{{ card.creditLabel }}</span>
              </div>
            </div>

            <p class="mt-5 text-sm leading-6 text-slate-500 dark:text-slate-400">
              {{ card.estimateLabel }}
            </p>
            <p class="mt-2 text-xs uppercase tracking-[0.18em] text-slate-400 dark:text-slate-500">
              {{ card.ratioLabel }}
            </p>
          </template>

          <template v-else>
            <div class="flex h-full flex-col justify-between">
              <div>
                <span
                  class="inline-flex rounded-full border border-slate-200 bg-slate-100 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500 dark:border-white/10 dark:bg-white/10 dark:text-slate-300"
                >
                  {{ card.badge }}
                </span>
                <h3 class="mt-5 text-2xl font-semibold text-slate-900 dark:text-white">{{ card.name }}</h3>
                <p class="mt-4 text-sm leading-7 text-slate-600 dark:text-slate-300">{{ card.creditLabel }}</p>
              </div>

              <div class="mt-8 rounded-[1.5rem] border border-dashed border-slate-300/80 bg-slate-50/85 px-4 py-4 dark:border-white/10 dark:bg-slate-900/70">
                <p class="text-sm font-semibold text-slate-900 dark:text-white">{{ card.priceLabel }}</p>
                <p class="mt-2 text-xs uppercase tracking-[0.18em] text-slate-400 dark:text-slate-500">
                  Claude 方案开放后将在这里补充
                </p>
              </div>
            </div>
          </template>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
type HomePackageCard = {
  name: string
  priceLabel: string
  creditLabel: string
  estimateValue?: string
  estimateLabel?: string
  ratioLabel?: string
  multiplier?: string
  badge?: string
  kind: 'codex' | 'claude-teaser'
}

const packageCards: HomePackageCard[] = [
  {
    name: '¥15 / $50 额度包',
    priceLabel: '¥15',
    creditLabel: '$50',
    estimateValue: '1000 万',
    estimateLabel: '按 GPT-5.4 约可使用 1000 万 tokens',
    ratioLabel: '按输入:输出 = 4:1 估算',
    multiplier: '1x',
    kind: 'codex'
  },
  {
    name: '¥30 / $120 额度包',
    priceLabel: '¥30',
    creditLabel: '$120',
    estimateValue: '2400 万',
    estimateLabel: '按 GPT-5.4 约可使用 2400 万 tokens',
    ratioLabel: '按输入:输出 = 4:1 估算',
    multiplier: '1x',
    kind: 'codex'
  },
  {
    name: '¥100 / $400 额度包',
    priceLabel: '¥100',
    creditLabel: '$400',
    estimateValue: '8000 万',
    estimateLabel: '按 GPT-5.4 约可使用 8000 万 tokens',
    ratioLabel: '按输入:输出 = 4:1 估算',
    multiplier: '1x',
    kind: 'codex'
  },
  {
    name: 'Claude 额度包',
    priceLabel: '敬请期待',
    creditLabel: 'Claude 额度方案正在准备中，后续开放。',
    badge: 'Coming Soon',
    kind: 'claude-teaser'
  }
]
</script>
