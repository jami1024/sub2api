<template>
  <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="flex flex-col gap-4 border-b border-gray-100 px-4 py-4 dark:border-gray-700 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">上游倍率监测</h3>
        <p class="mt-1 max-w-3xl text-xs leading-5 text-gray-500 dark:text-gray-400">
          只在手动点击检测时请求上游。通过上游 /v1/usage/stats 前后差值计算倍率，并保留历史趋势。
        </p>
      </div>
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
        <label class="flex items-center gap-2 text-xs font-medium text-gray-600 dark:text-gray-300">
          模型
          <input
            :value="model"
            type="text"
            class="w-36 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @input="$emit('update:model', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <button
          type="button"
          class="rounded-lg border border-gray-200 px-3 py-2 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-dark-700"
          :disabled="loading"
          @click="$emit('refresh')"
        >
          刷新记录
        </button>
        <button
          type="button"
          class="rounded-lg bg-primary-600 px-3 py-2 text-xs font-semibold text-white hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="loading || measuring || supportedAccounts.length === 0"
          @click="$emit('measure-all')"
        >
          {{ measuring ? '检测中…' : '检测全部支持账号' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex h-40 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      加载倍率记录中…
    </div>
    <div v-else class="grid gap-5 p-4 xl:grid-cols-[minmax(0,1.35fr)_minmax(360px,0.65fr)]">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-gray-700">
          <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-700/60 dark:text-gray-400">
            <tr>
              <th class="px-3 py-3 text-left">账号</th>
              <th class="px-3 py-3 text-left">上游</th>
              <th class="px-3 py-3 text-left">Key</th>
              <th class="px-3 py-3 text-right">最新倍率</th>
              <th class="px-3 py-3 text-right">扣费增量</th>
              <th class="px-3 py-3 text-left">状态</th>
              <th class="px-3 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr v-if="accounts.length === 0">
              <td colspan="7" class="px-3 py-8 text-center text-sm text-gray-500 dark:text-gray-400">暂无可展示账号</td>
            </tr>
            <tr v-for="account in accounts" :key="account.account_id" class="hover:bg-gray-50/80 dark:hover:bg-dark-700/40">
              <td class="px-3 py-3 font-semibold text-gray-900 dark:text-white">{{ account.account_name }}</td>
              <td class="max-w-[240px] truncate px-3 py-3 text-xs text-gray-600 dark:text-gray-300" :title="account.base_url">{{ hostOf(account.base_url) }}</td>
              <td class="px-3 py-3 font-mono text-xs text-gray-500 dark:text-gray-400">{{ account.key_prefix || '—' }}</td>
              <td class="px-3 py-3 text-right font-semibold text-gray-900 dark:text-white">
                {{ formatMultiplier(account.latest_sample?.multiplier) }}
              </td>
              <td class="px-3 py-3 text-right text-xs text-gray-600 dark:text-gray-300">
                <div>标准 {{ formatCost(account.latest_sample?.standard_cost_delta) }}</div>
                <div class="text-gray-500 dark:text-gray-400">扣费 {{ formatCost(account.latest_sample?.actual_cost_delta) }}</div>
              </td>
              <td class="px-3 py-3 text-xs">
                <span
                  class="inline-flex rounded-full px-2 py-1 font-semibold"
                  :class="statusClass(account)"
                >
                  {{ statusText(account) }}
                </span>
                <div v-if="account.latest_sample?.error_message || account.skip_reason" class="mt-1 max-w-[260px] truncate text-gray-500 dark:text-gray-400" :title="account.latest_sample?.error_message || account.skip_reason">
                  {{ account.latest_sample?.error_message || account.skip_reason }}
                </div>
              </td>
              <td class="px-3 py-3 text-right">
                <button
                  type="button"
                  class="rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-gray-700 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
                  :data-testid="`measure-upstream-${account.account_id}`"
                  :disabled="!account.supported || measuringAccountId === account.account_id"
                  @click="$emit('measure-account', account.account_id)"
                >
                  {{ measuringAccountId === account.account_id ? '检测中' : '检测' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="space-y-4">
        <div class="rounded-xl border border-gray-100 bg-gray-50 p-3 dark:border-gray-700 dark:bg-dark-900/60">
          <div class="mb-3 flex items-center justify-between">
            <div class="text-xs font-semibold text-gray-700 dark:text-gray-200">倍率趋势</div>
            <div class="text-[11px] text-gray-500 dark:text-gray-400">最近 {{ successfulSamples.length }} 条成功样本</div>
          </div>
          <div v-if="successfulSamples.length === 0" class="flex h-40 items-center justify-center text-xs text-gray-500 dark:text-gray-400">
            暂无成功样本
          </div>
          <div v-else class="grid h-44 grid-cols-[3rem_minmax(0,1fr)] gap-3">
            <div class="flex flex-col justify-between py-1 text-right text-[11px] tabular-nums text-gray-500 dark:text-gray-400">
              <span>{{ formatMultiplier(maxMultiplier) }}</span>
              <span>{{ formatMultiplier(maxMultiplier / 2) }}</span>
              <span>0x</span>
            </div>
            <div class="relative h-full overflow-hidden rounded-lg border border-gray-100 bg-white dark:border-gray-700 dark:bg-dark-800">
              <div class="pointer-events-none absolute inset-x-0 top-1/2 border-t border-dashed border-gray-200 dark:border-gray-700" />
              <svg viewBox="0 0 100 45" preserveAspectRatio="none" class="h-full w-full overflow-visible">
                <polyline :points="trendPoints" fill="none" stroke="#2563eb" stroke-width="1.6" vector-effect="non-scaling-stroke" />
              </svg>
            </div>
          </div>
        </div>

        <div class="rounded-xl border border-gray-100 dark:border-gray-700">
          <div class="border-b border-gray-100 px-3 py-2 text-xs font-semibold text-gray-700 dark:border-gray-700 dark:text-gray-200">最近记录</div>
          <div class="max-h-56 divide-y divide-gray-100 overflow-y-auto dark:divide-gray-700">
            <div v-if="samples.length === 0" class="px-3 py-6 text-center text-xs text-gray-500 dark:text-gray-400">暂无记录</div>
            <div v-for="sample in samples.slice(0, 8)" :key="sample.id" class="flex items-center justify-between gap-3 px-3 py-2 text-xs">
              <div class="min-w-0">
                <div class="truncate font-semibold text-gray-800 dark:text-gray-100">{{ sample.account_name_snapshot }}</div>
                <div class="text-gray-500 dark:text-gray-400">{{ formatDate(sample.measured_at) }}</div>
              </div>
              <div class="text-right">
                <div class="font-semibold" :class="sample.status === 'success' ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-500 dark:text-gray-400'">
                  {{ sample.status === 'success' ? formatMultiplier(sample.multiplier) : sample.status }}
                </div>
                <div class="text-gray-500 dark:text-gray-400">{{ sample.model }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OpsUpstreamMultiplierAccount, OpsUpstreamMultiplierSample } from '@/api/admin/ops'

const props = defineProps<{
  model: string
  accounts: OpsUpstreamMultiplierAccount[]
  samples: OpsUpstreamMultiplierSample[]
  loading?: boolean
  measuring?: boolean
  measuringAccountId?: number | null
}>()

defineEmits<{
  'update:model': [value: string]
  refresh: []
  'measure-account': [accountID: number]
  'measure-all': []
}>()

const supportedAccounts = computed(() => props.accounts.filter(account => account.supported))
const successfulSamples = computed(() => props.samples.filter(sample => sample.status === 'success' && typeof sample.multiplier === 'number').slice().reverse())
const maxMultiplier = computed(() => Math.max(...successfulSamples.value.map(sample => sample.multiplier || 0), 0.01))
const trendPoints = computed(() => {
  const points = successfulSamples.value
  if (points.length === 0) return ''
  return points.map((sample, index) => {
    const x = points.length === 1 ? 0 : (index / (points.length - 1)) * 100
    const y = 42 - ((sample.multiplier || 0) / maxMultiplier.value) * 38
    return `${x.toFixed(2)},${Math.max(2, y).toFixed(2)}`
  }).join(' ')
})

function formatMultiplier(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${trimNumber(value)}x`
}

function formatCost(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `$${value.toFixed(6).replace(/0+$/, '').replace(/\.$/, '')}`
}

function formatDate(value?: string | null): string {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

function trimNumber(value: number): string {
  return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
}

function hostOf(raw: string): string {
  try {
    return new URL(raw).host
  } catch {
    return raw || '—'
  }
}

function statusText(account: OpsUpstreamMultiplierAccount): string {
  if (!account.supported) return '跳过'
  switch (account.latest_sample?.status) {
    case 'success':
      return '正常'
    case 'error':
      return '失败'
    case 'skipped':
      return '跳过'
    default:
      return '未检测'
  }
}

function statusClass(account: OpsUpstreamMultiplierAccount): string {
  if (!account.supported || account.latest_sample?.status === 'skipped') {
    return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-200'
  }
  if (account.latest_sample?.status === 'success') {
    return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
  }
  if (account.latest_sample?.status === 'error') {
    return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
  }
  return 'bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300'
}
</script>
