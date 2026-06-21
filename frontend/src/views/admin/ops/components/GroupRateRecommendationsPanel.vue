<template>
  <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="flex flex-col gap-4 border-b border-gray-100 px-4 py-4 dark:border-gray-700 xl:flex-row xl:items-start xl:justify-between">
      <div>
        <p class="text-xs font-medium uppercase tracking-wide text-primary-600 dark:text-primary-300">{{ model || 'gpt-5.4' }}</p>
        <h3 class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">分组倍率与权重建议</h3>
        <p class="mt-1 max-w-3xl text-xs leading-5 text-gray-500 dark:text-gray-400">
          按余额包收入、上游真实倍率和最近使用占比估算分组成本；这里只展示建议，不会自动修改分组或账号配置。
        </p>
      </div>

      <div class="grid gap-2 sm:grid-cols-4 xl:min-w-[660px]">
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          目标利润率
          <input
            data-testid="profit-margin-input"
            :value="profitMargin"
            type="number"
            step="0.01"
            min="0.01"
            max="0.9"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitNumber('profitMargin', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          安全系数
          <input
            :value="safetyFactor"
            type="number"
            step="0.1"
            min="1"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitNumber('safetyFactor', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          使用天数
          <input
            :value="usageDays"
            type="number"
            step="1"
            min="1"
            max="30"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitInteger('update:usageDays', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <div class="flex items-end">
          <button
            type="button"
            class="w-full rounded-lg bg-primary-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loading"
            @click="emit('refresh')"
          >
            重新计算
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="flex h-40 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      加载分组建议中…
    </div>

    <div v-else class="space-y-4 p-4">
      <div class="grid gap-3 lg:grid-cols-3">
        <div class="rounded-xl border border-blue-100 bg-blue-50 px-4 py-3 text-xs text-blue-800 dark:border-blue-500/20 dark:bg-blue-500/10 dark:text-blue-200 lg:col-span-2">
          <template v-if="data?.package_basis">
            当前套餐口径：<span class="font-semibold">{{ data.package_basis.name }}</span>，{{ formatMoney(data.package_basis.price) }} 元买 {{ trimNumber(data.package_basis.credit_amount) }} 额度，单额度收入 {{ trimNumber(data.package_basis.revenue_per_credit) }}。
          </template>
          <template v-else>
            没有找到可用套餐口径，建议结果会显示为数据不足。
          </template>
        </div>
        <div class="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 text-xs text-gray-600 dark:border-gray-700 dark:bg-dark-900/60 dark:text-gray-300">
          公式：上游倍率 × 安全系数 ÷ (1 - 利润率) ÷ 单额度收入。
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-gray-700">
          <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-700/60 dark:text-gray-400">
            <tr>
              <th class="px-3 py-3 text-left">分组</th>
              <th class="px-3 py-3 text-right">当前倍率</th>
              <th class="px-3 py-3 text-right">综合成本</th>
              <th class="px-3 py-3 text-right">最坏成本</th>
              <th class="px-3 py-3 text-right">建议倍率</th>
              <th class="px-3 py-3 text-left">状态</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr v-if="groups.length === 0">
              <td colspan="6" class="px-3 py-8 text-center text-sm text-gray-500 dark:text-gray-400">暂无分组建议</td>
            </tr>
            <template v-for="group in groups" :key="group.group_id">
              <tr class="bg-white dark:bg-dark-800">
                <td class="px-3 py-3">
                  <div class="font-semibold text-gray-900 dark:text-white">{{ group.group_name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ group.schedulable_account_count }} 个账号参与建议</div>
                </td>
                <td class="px-3 py-3 text-right font-semibold text-gray-900 dark:text-white">{{ formatMultiplier(group.current_group_multiplier) }}</td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">
                  <div>{{ formatMultiplier(group.recommended_blended_multiplier ?? group.actual_blended_multiplier) }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">实际 {{ formatMultiplier(group.actual_blended_multiplier) }}</div>
                </td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">{{ formatMultiplier(group.worst_case_multiplier) }}</td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">
                  <div>最低 {{ formatMultiplier(group.minimum_group_multiplier) }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">稳妥 {{ formatMultiplier(group.safe_group_multiplier) }}</div>
                </td>
                <td class="px-3 py-3">
                  <span class="inline-flex rounded-full px-2 py-1 text-xs font-semibold" :class="statusClass(group.status)">{{ statusText(group.status) }}</span>
                </td>
              </tr>
              <tr>
                <td colspan="6" class="bg-gray-50 px-3 py-3 dark:bg-dark-900/50">
                  <div class="grid gap-2 lg:grid-cols-2 xl:grid-cols-3">
                    <div v-for="account in group.accounts" :key="account.account_id" class="rounded-lg border border-gray-100 bg-white p-3 text-xs dark:border-gray-700 dark:bg-dark-800">
                      <div class="flex items-start justify-between gap-2">
                        <div class="min-w-0">
                          <div class="truncate font-semibold text-gray-900 dark:text-white">{{ account.account_name }}</div>
                          <div class="truncate text-gray-500 dark:text-gray-400">{{ hostOf(account.base_url) }}</div>
                        </div>
                        <div class="text-right font-semibold text-gray-900 dark:text-white">{{ formatMultiplier(account.upstream_multiplier) }}</div>
                      </div>
                      <div class="mt-3 grid grid-cols-2 gap-2 text-gray-600 dark:text-gray-300">
                        <div>当前占比 {{ formatPercent(account.standard_cost_share) }}</div>
                        <div>建议权重 {{ formatPercent(account.recommended_weight) }}</div>
                        <div>当前 priority {{ account.current_priority }}</div>
                        <div>建议 priority {{ account.recommended_priority || '—' }}</div>
                      </div>
                      <div class="mt-2 text-gray-500 dark:text-gray-400">{{ account.note || '—' }}</div>
                    </div>
                  </div>
                  <div v-if="group.notes?.length" class="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">{{ group.notes.join('；') }}</div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OpsGroupRateRecommendationStatus, OpsGroupRateRecommendationsResponse } from '@/api/admin/ops'

const props = defineProps<{
  model: string
  data?: OpsGroupRateRecommendationsResponse | null
  loading?: boolean
  profitMargin: number
  safetyFactor: number
  usageDays: number
}>()

const emit = defineEmits<{
  'update:profitMargin': [value: number]
  'update:safetyFactor': [value: number]
  'update:usageDays': [value: number]
  refresh: []
}>()

const groups = computed(() => props.data?.groups || [])

function emitNumber(event: 'profitMargin' | 'safetyFactor', raw: string) {
  const value = Number(raw)
  if (!Number.isFinite(value)) return
  if (event === 'profitMargin') {
    emit('update:profitMargin', value)
  } else {
    emit('update:safetyFactor', value)
  }
  emit('refresh')
}

function emitInteger(event: 'update:usageDays', raw: string) {
  const value = Number.parseInt(raw, 10)
  if (!Number.isFinite(value)) return
  emit(event, value)
  emit('refresh')
}

function formatMultiplier(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${trimNumber(value)}x`
}

function formatPercent(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${trimNumber(value * 100)}%`
}

function formatMoney(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return trimNumber(value)
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

function statusText(status: OpsGroupRateRecommendationStatus): string {
  switch (status) {
    case 'safe':
      return '安全'
    case 'basic_safe':
      return '基本安全'
    case 'low':
      return '偏低'
    default:
      return '数据不足'
  }
}

function statusClass(status: OpsGroupRateRecommendationStatus): string {
  switch (status) {
    case 'safe':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'basic_safe':
      return 'bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300'
    case 'low':
      return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-200'
  }
}
</script>
