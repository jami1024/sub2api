<template>
  <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.providerStatus.latencyTrend') }}</h3>
      <div v-if="!loading && points.length > 0" class="grid grid-cols-3 gap-2 text-right">
        <div
          v-for="stat in latencyStats"
          :key="stat.key"
          class="rounded-lg border border-gray-100 bg-gray-50 px-2 py-1.5 dark:border-gray-700 dark:bg-dark-700/60"
        >
          <div class="flex items-center justify-end gap-1 text-[11px] font-medium text-gray-500 dark:text-gray-400">
            <i class="inline-block h-2 w-2 rounded-full" :class="stat.dotClass" />
            {{ stat.label }}
          </div>
          <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ formatMs(stat.current) }}</div>
          <div class="text-[11px] text-gray-500 dark:text-gray-400">{{ t('admin.providerStatus.peak') }} {{ formatMs(stat.peak) }}</div>
        </div>
      </div>
    </div>
    <div v-if="loading" class="flex h-56 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
    <div v-else-if="points.length === 0" class="flex h-56 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.providerStatus.empty') }}</div>
    <div v-else class="w-full">
      <div class="grid h-56 grid-cols-[3.75rem_minmax(0,1fr)] gap-3">
        <div class="flex flex-col justify-between py-1 text-right text-[11px] tabular-nums text-gray-500 dark:text-gray-400">
          <span>{{ formatMs(maxLatency) }}</span>
          <span>{{ formatMs(maxLatency / 2) }}</span>
          <span>{{ formatMs(0) }}</span>
        </div>
        <div class="relative h-full overflow-hidden rounded-xl border border-gray-100 bg-gray-50/60 dark:border-gray-700 dark:bg-dark-700/30">
          <div class="pointer-events-none absolute inset-x-0 top-1/2 border-t border-dashed border-gray-200 dark:border-gray-700" />
          <div class="pointer-events-none absolute inset-x-0 bottom-2 border-t border-gray-200 dark:border-gray-700" />
          <svg viewBox="0 0 100 45" preserveAspectRatio="none" class="h-full w-full overflow-visible">
            <polyline :points="linePoints('p50_ms')" fill="none" stroke="#0f766e" stroke-width="1.4" vector-effect="non-scaling-stroke" />
            <polyline :points="linePoints('p95_ms')" fill="none" stroke="#f59e0b" stroke-width="1.4" vector-effect="non-scaling-stroke" />
            <polyline :points="linePoints('p99_ms')" fill="none" stroke="#f97316" stroke-width="1.4" vector-effect="non-scaling-stroke" />
          </svg>
        </div>
      </div>
      <div class="mt-2 flex justify-end gap-4 text-xs text-gray-500 dark:text-gray-400">
        <span><i class="mr-1 inline-block h-2 w-2 rounded-full bg-teal-700" />P50</span>
        <span><i class="mr-1 inline-block h-2 w-2 rounded-full bg-amber-500" />P95</span>
        <span><i class="mr-1 inline-block h-2 w-2 rounded-full bg-orange-500" />P99</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsProviderStatusTimelinePoint } from '@/api/admin/ops'

const { t } = useI18n()
const props = defineProps<{
  points: OpsProviderStatusTimelinePoint[]
  loading?: boolean
}>()

type LatencyKey = 'p50_ms' | 'p95_ms' | 'p99_ms'

const maxLatency = computed(() => {
  const values = props.points.flatMap(p => [p.p50_ms, p.p95_ms, p.p99_ms].filter(isFiniteNumber))
  return Math.max(...values, 1)
})

const latencyStats = computed(() => [
  buildLatencyStat('p50_ms', 'P50', 'bg-teal-700'),
  buildLatencyStat('p95_ms', 'P95', 'bg-amber-500'),
  buildLatencyStat('p99_ms', 'P99', 'bg-orange-500'),
])

function buildLatencyStat(key: LatencyKey, label: string, dotClass: string) {
  const values = props.points.map(point => point[key]).filter(isFiniteNumber)
  return {
    key,
    label,
    dotClass,
    current: values.at(-1) ?? 0,
    peak: Math.max(...values, 0),
  }
}

function linePoints(key: LatencyKey): string {
  if (!props.points.length) return ''
  return props.points.map((point, index) => {
    const x = props.points.length === 1 ? 0 : (index / (props.points.length - 1)) * 100
    const value = point[key] ?? 0
    const y = 42 - (value / maxLatency.value) * 38
    return `${x.toFixed(2)},${Math.max(2, y).toFixed(2)}`
  }).join(' ')
}

function isFiniteNumber(value: number | null | undefined): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

function formatMs(value: number): string {
  const safeValue = Math.max(0, value)
  if (safeValue < 1000) {
    return `${Math.round(safeValue)}ms`
  }
  if (safeValue < 60_000) {
    return `${formatDecimal(safeValue / 1000)}s`
  }
  return `${formatDecimal(safeValue / 60_000)}m`
}

function formatDecimal(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}
</script>
