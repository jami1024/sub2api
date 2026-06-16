<template>
  <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="border-b border-gray-100 px-4 py-3 dark:border-gray-700">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.providerStatus.tableTitle') }}</h3>
    </div>
    <div v-if="loading" class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="items.length === 0" class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('admin.providerStatus.empty') }}
    </div>
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-gray-700">
        <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-700/60 dark:text-gray-400">
          <tr>
            <th class="px-4 py-3 text-left">{{ t('admin.providerStatus.provider') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.providerStatus.requests') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.providerStatus.availability') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.providerStatus.errors') }}</th>
            <th class="px-4 py-3 text-right">P50</th>
            <th class="px-4 py-3 text-right">P95</th>
            <th class="px-4 py-3 text-right">P99</th>
            <th class="px-4 py-3 text-left">{{ t('admin.providerStatus.timeline') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.providerStatus.lastSeen') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
          <tr v-for="item in items" :key="item.provider" class="hover:bg-gray-50/80 dark:hover:bg-dark-700/40">
            <td class="px-4 py-3 font-semibold text-gray-900 dark:text-white">{{ item.provider || 'unknown' }}</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNumber(item.request_count) }}</td>
            <td class="px-4 py-3 text-right font-semibold" :class="availabilityClass(item.availability)">{{ item.availability.toFixed(1) }}%</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNumber(item.failure_count) }}</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatMs(item.p50_ms) }}</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatMs(item.p95_ms) }}</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatMs(item.p99_ms) }}</td>
            <td class="px-4 py-3">
              <div class="flex min-w-[160px] gap-1">
                <span
                  v-for="(point, idx) in compactTimeline(item.timeline || [])"
                  :key="`${item.provider}-${idx}`"
                  class="h-2 flex-1 rounded-full"
                  :class="timelineClass(point)"
                  :title="timelineTitle(point)"
                />
              </div>
            </td>
            <td class="px-4 py-3 text-right text-xs text-gray-500 dark:text-gray-400">{{ formatDate(item.last_seen) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { OpsProviderStatusItem, OpsProviderStatusTimelinePoint } from '@/api/admin/ops'

const { t } = useI18n()
defineProps<{
  items: OpsProviderStatusItem[]
  loading?: boolean
}>()

function formatNumber(value: number): string {
  return value.toLocaleString()
}

function formatMs(value?: number | null): string {
  return value == null ? '—' : `${value}ms`
}

function formatDate(value?: string | null): string {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

function availabilityClass(value: number): string {
  if (value >= 95) return 'text-emerald-600 dark:text-emerald-400'
  if (value >= 80) return 'text-lime-600 dark:text-lime-400'
  if (value >= 50) return 'text-orange-500 dark:text-orange-400'
  return 'text-rose-600 dark:text-rose-400'
}

function compactTimeline(points: OpsProviderStatusTimelinePoint[]): OpsProviderStatusTimelinePoint[] {
  if (points.length <= 40) return points
  const step = Math.ceil(points.length / 40)
  const out: OpsProviderStatusTimelinePoint[] = []
  for (let i = 0; i < points.length; i += step) {
    const slice = points.slice(i, i + step)
    const success = slice.reduce((sum, p) => sum + p.success_count, 0)
    const failure = slice.reduce((sum, p) => sum + p.failure_count, 0)
    out.push({
      bucket_start: slice[0].bucket_start,
      request_count: success + failure,
      success_count: success,
      failure_count: failure,
      availability: success + failure > 0 ? (success / (success + failure)) * 100 : 0,
    })
  }
  return out
}

function timelineClass(point: OpsProviderStatusTimelinePoint): string {
  if (!point || point.request_count <= 0) return 'bg-gray-200 dark:bg-gray-700'
  if (point.availability >= 95) return 'bg-emerald-500'
  if (point.availability >= 80) return 'bg-lime-500'
  if (point.availability >= 50) return 'bg-orange-500'
  return 'bg-rose-500'
}

function timelineTitle(point: OpsProviderStatusTimelinePoint): string {
  return `${new Date(point.bucket_start).toLocaleString()} · ${point.request_count} req · ${point.availability.toFixed(1)}%`
}
</script>
