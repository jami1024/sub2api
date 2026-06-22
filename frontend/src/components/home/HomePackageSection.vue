<template>
  <section
    data-testid="home-package-section"
    class="relative isolate overflow-hidden rounded-[2.2rem] border border-slate-200/80 bg-white/80 p-6 shadow-[0_24px_90px_rgba(15,23,42,0.08)] dark:border-slate-700/80 dark:bg-slate-950 dark:shadow-[0_30px_100px_rgba(2,6,23,0.72)] sm:p-8"
  >
    <div
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.12),transparent_32%),radial-gradient(circle_at_bottom_right,rgba(56,189,248,0.08),transparent_28%)] dark:bg-[radial-gradient(circle_at_top_left,rgba(14,165,233,0.22),transparent_30%),radial-gradient(circle_at_bottom_right,rgba(20,184,166,0.16),transparent_28%)]"
    ></div>
    <div
      class="pointer-events-none absolute inset-0 opacity-30 [background-image:radial-gradient(rgba(15,23,42,0.08)_1px,transparent_1px)] [background-size:22px_22px] dark:opacity-25 dark:[background-image:radial-gradient(rgba(125,211,252,0.14)_1px,transparent_1px)]"
    ></div>

    <div class="relative">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="max-w-2xl">
          <p
            class="inline-flex rounded-full border border-sky-200/80 bg-sky-50/90 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.28em] text-sky-700 dark:border-sky-300/40 dark:bg-sky-400/15 dark:text-sky-100"
          >
            Live Packages
          </p>
          <h2 class="mt-4 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl">Codex 专属余额包</h2>
          <p class="mt-3 max-w-[62ch] text-sm leading-7 text-slate-600 dark:text-slate-200">
            余额包与 OpenAI 分组倍率由后台实时同步，展示到账倍率、使用倍率和综合折扣，实际购买以充值订阅页面为准。
          </p>
          <p v-if="primaryRate" class="mt-2 text-xs font-medium text-slate-500 dark:text-sky-200/80">
            {{ primaryRate.rate_label }}<template v-if="primaryRate.value_lift_label">，{{ primaryRate.value_lift_label }}</template>
          </p>
        </div>

        <div
          class="rounded-[1.6rem] border border-slate-200/80 bg-white/85 px-4 py-3 text-sm text-slate-600 shadow-[0_16px_40px_rgba(15,23,42,0.05)] dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-200"
        >
          <p class="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400 dark:text-sky-200/70">计算口径</p>
          <p class="mt-2 font-medium text-slate-900 dark:text-white">综合折扣 = 到账余额 ÷ 使用倍率</p>
          <p class="mt-1 text-xs text-slate-500 dark:text-slate-300">到账倍率和模型使用倍率分开展示</p>
        </div>
      </div>

      <div v-if="loading" class="mt-7 grid gap-3 md:grid-cols-3">
        <div v-for="idx in 3" :key="idx" class="h-80 animate-pulse rounded-[1.65rem] bg-slate-100 dark:bg-slate-900" />
      </div>

      <div v-else class="mt-7 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="card in displayPackages"
          :key="card.id"
          data-testid="home-package-card"
          data-package-kind="codex"
          class="group flex h-full flex-col overflow-hidden rounded-[1.65rem] border p-4 transition duration-300 sm:p-5"
          :class="card.isRecommended
            ? 'border-emerald-300/80 bg-emerald-50/80 shadow-[0_20px_60px_rgba(16,185,129,0.16)] hover:-translate-y-1 dark:border-emerald-400/40 dark:bg-emerald-950/30 dark:shadow-[0_22px_60px_rgba(5,150,105,0.2)]'
            : 'border-slate-200/80 bg-slate-50/92 shadow-[0_14px_40px_rgba(15,23,42,0.05)] hover:-translate-y-1 hover:border-slate-300 dark:border-slate-700 dark:bg-slate-900/95 dark:shadow-[0_18px_44px_rgba(2,6,23,0.55)] dark:hover:border-sky-400/40'"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="tag in card.tags"
                  :key="tag"
                  class="inline-flex rounded-full border px-2.5 py-1 text-xs font-semibold"
                  :class="tagClass(tag, card.isRecommended)"
                >
                  {{ tag }}
                </span>
              </div>
              <h3 class="mt-4 text-xl font-bold leading-7 text-slate-900 dark:text-white">{{ card.name }}</h3>
              <p v-if="card.description" class="mt-2 line-clamp-2 text-sm leading-6 text-slate-500 dark:text-slate-300">{{ card.description }}</p>
            </div>
            <div class="shrink-0 text-right">
              <p class="text-xs text-slate-400 dark:text-slate-500">实付</p>
              <p class="mt-1 text-3xl font-black text-slate-950 dark:text-white">¥{{ formatAmount(card.price) }}</p>
            </div>
          </div>

          <div class="mt-5 rounded-[1.35rem] bg-slate-950 px-4 py-3.5 text-white dark:bg-slate-950 dark:ring-1 dark:ring-sky-400/15">
            <p class="text-xs font-medium uppercase tracking-[0.2em] text-white/60 dark:text-sky-100/65">到账余额</p>
            <p class="mt-2.5 text-[2.1rem] font-semibold leading-none sm:text-[2.35rem]">{{ formatAmount(card.creditAmount) }}</p>
            <p class="mt-2 text-sm text-white/70 dark:text-slate-300">余额</p>
          </div>

          <div class="mt-5 grid gap-2.5 text-sm text-slate-600 dark:text-slate-300">
            <div class="flex items-center justify-between gap-3 rounded-[1.1rem] bg-white/80 px-3.5 py-2.5 dark:bg-slate-800/90">
              <span>到账倍率</span>
              <span class="font-semibold text-slate-900 dark:text-white">{{ formatMultiplier(card.arrivalMultiplier) }}x</span>
            </div>
            <div class="flex items-center justify-between gap-3 rounded-[1.1rem] bg-white/80 px-3.5 py-2.5 dark:bg-slate-800/90">
              <span>使用倍率</span>
              <span class="font-semibold text-slate-900 dark:text-white">{{ formatMultiplier(primaryRate?.rate_multiplier ?? 1) }}x</span>
            </div>
          </div>

          <div class="mt-5 rounded-[1.2rem] border border-slate-200/80 bg-white/80 px-4 py-3 dark:border-slate-700 dark:bg-slate-800/80">
            <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">{{ card.effectiveDiscountLabel }}</p>
            <p class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-300">
              约等效 {{ formatAmount(card.effectiveCreditAmount) }} 余额，{{ card.arrivalDiscountLabel }}到账优惠。
            </p>
          </div>
        </article>
      </div>

      <p v-if="loadFailed" class="mt-4 text-xs text-amber-600 dark:text-amber-300">
        套餐实时信息暂时加载失败，登录后可在充值订阅页查看最新档位。
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getLandingPackageShowcase, type LandingBalancePackage, type LandingUsageRate } from '@/api/publicLanding'

type DisplayPackage = {
  id: number
  name: string
  description: string
  price: number
  creditAmount: number
  tags: string[]
  arrivalMultiplier: number
  arrivalDiscountLabel: string
  effectiveCreditAmount: number
  effectiveDiscountLabel: string
  isRecommended: boolean
}

const loading = ref(true)
const loadFailed = ref(false)
const packages = ref<LandingBalancePackage[]>([])
const primaryRate = ref<LandingUsageRate | null>(null)

const displayPackages = computed<DisplayPackage[]>(() =>
  packages.value
    .filter((pkg) => pkg.package_scope === 'codex')
    .map((pkg) => ({
      id: pkg.id,
      name: pkg.name,
      description: pkg.description,
      price: pkg.price,
      creditAmount: pkg.credit_amount,
      tags: normalizeTags(pkg),
      arrivalMultiplier: pkg.arrival_multiplier,
      arrivalDiscountLabel: pkg.arrival_discount_label,
      effectiveCreditAmount: pkg.effective_credit_amount,
      effectiveDiscountLabel: pkg.effective_discount_label,
      isRecommended: isRecommendedPackage(pkg)
    }))
)

onMounted(async () => {
  try {
    const response = await getLandingPackageShowcase()
    packages.value = response.data.packages || []
    primaryRate.value = response.data.primary_usage_rate || response.data.usage_rates?.[0] || null
  } catch (error) {
    loadFailed.value = true
    packages.value = []
    primaryRate.value = null
  } finally {
    loading.value = false
  }
})

function normalizeTags(pkg: LandingBalancePackage): string[] {
  const tags = [...(pkg.display_tags || [])]
  if (isRecommendedPackage(pkg) && !tags.some((tag) => tag.includes('推荐'))) {
    tags.unshift('推荐')
  }
  if (tags.length === 0) {
    tags.push('按量计费')
  }
  return tags.slice(0, 3)
}

function isRecommendedPackage(pkg: LandingBalancePackage): boolean {
  return pkg.display_tags?.some((tag) => tag.includes('推荐') || tag.includes('热销')) || pkg.effective_discount <= 2.5
}

function tagClass(tag: string, recommended: boolean): string {
  if (tag.includes('推荐') || tag.includes('热销') || recommended) {
    return 'border-emerald-200/80 bg-emerald-50 text-emerald-700 dark:border-emerald-400/30 dark:bg-emerald-400/10 dark:text-emerald-200'
  }
  return 'border-sky-200/80 bg-sky-50 text-sky-700 dark:border-sky-400/20 dark:bg-sky-400/10 dark:text-sky-200'
}

function formatAmount(value: number): string {
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}

function formatMultiplier(value: number): string {
  return formatAmount(value)
}
</script>
