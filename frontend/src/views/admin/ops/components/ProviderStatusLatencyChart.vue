<template>
  <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.providerStatus.latencyTrend') }}</h3>
    </div>
    <div v-if="loading" class="flex h-56 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
    <div v-else-if="points.length === 0" class="flex h-56 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.providerStatus.empty') }}</div>
    <div v-else class="h-56 w-full">
      <svg viewBox="0 0 100 45" preserveAspectRatio="none" class="h-full w-full overflow-visible">
        <polyline :points="linePoints('p50_ms')" fill="none" stroke="#0f766e" stroke-width="1.4" vector-effect="non-scaling-stroke" />
        <polyline :points="linePoints('p95_ms')" fill="none" stroke="#f59e0b" stroke-width="1.4" vector-effect="non-scaling-stroke" />
        <polyline :points="linePoints('p99_ms')" fill="none" stroke="#f97316" stroke-width="1.4" vector-effect="non-scaling-stroke" />
      </svg>
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

const maxLatency = computed(() => {
  const values = props.points.flatMap(p => [p.p50_ms, p.p95_ms, p.p99_ms].filter((v): v is number => typeof v === 'number'))
  return Math.max(...values, 1)
})

function linePoints(key: 'p50_ms' | 'p95_ms' | 'p99_ms'): string {
  if (!props.points.length) return ''
  return props.points.map((point, index) => {
    const x = props.points.length === 1 ? 0 : (index / (props.points.length - 1)) * 100
    const value = point[key] ?? 0
    const y = 42 - (value / maxLatency.value) * 38
    return `${x.toFixed(2)},${Math.max(2, y).toFixed(2)}`
  }).join(' ')
}
</script>
