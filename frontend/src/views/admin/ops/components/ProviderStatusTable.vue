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
            <th class="px-4 py-3 text-right">总耗时</th>
            <th class="px-4 py-3 text-right">分层首响应</th>
            <th class="px-4 py-3 text-right">524</th>
            <th class="px-4 py-3 text-left">{{ t('admin.providerStatus.fingerprint') }}</th>
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
            <td class="px-4 py-3 text-right text-xs text-gray-700 dark:text-gray-300">
              <div>平均 {{ formatMs(item.duration_avg_ms) }}</div>
              <div class="text-gray-500 dark:text-gray-400">最大 {{ formatMs(item.duration_max_ms) }}</div>
            </td>
            <td class="px-4 py-3 text-right text-xs text-gray-700 dark:text-gray-300">
              <div>客户端视角 {{ formatMs(item.ttft_avg_ms) }}</div>
              <div class="text-gray-500 dark:text-gray-400">我站→上游 {{ formatMs(item.upstream_ttft_avg_ms) }}</div>
              <div class="text-gray-500 dark:text-gray-400">网关处理 {{ formatMs(item.gateway_ttft_avg_ms) }}</div>
              <div class="text-gray-400 dark:text-gray-500">P95 {{ formatMs(item.ttft_p95_ms) }}</div>
              <div class="text-gray-400 dark:text-gray-500">样本 {{ formatNumber(item.ttft_sample_count || 0) }}</div>
            </td>
            <td class="px-4 py-3 text-right text-xs text-gray-700 dark:text-gray-300">
              <div>{{ formatNumber(item.timeout_524_count || 0) }} 次</div>
              <div class="text-gray-500 dark:text-gray-400">平均 {{ formatMs(item.timeout_524_avg_ms) }}</div>
            </td>
            <td class="px-4 py-3 align-top text-xs">
              <div v-if="fingerprintEntries(item).length" class="min-w-[180px] space-y-1">
                <div
                  v-for="[key, value] in fingerprintPreviewEntries(item)"
                  :key="key"
                  class="max-w-[260px] truncate font-mono text-gray-700 dark:text-gray-200"
                  :title="`${key}: ${value}`"
                >
                  <span class="text-gray-400">{{ key }}:</span> {{ value }}
                </div>
                <button
                  v-if="fingerprintEntries(item).length > 2"
                  type="button"
                  data-testid="provider-fingerprint-toggle"
                  class="mt-1 rounded-md bg-gray-100 px-2 py-1 text-[11px] font-semibold text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600"
                  @click="toggleFingerprint(item.provider)"
                >
                  {{ expandedFingerprintProvider === item.provider ? t('common.collapse') || '收起' : `+${fingerprintEntries(item).length - 2}` }}
                </button>
                <div v-if="expandedFingerprintProvider === item.provider" class="mt-2 rounded-lg bg-gray-50 p-2 dark:bg-dark-900">
                  <div
                    v-for="[key, value] in fingerprintEntries(item)"
                    :key="`full-${key}`"
                    class="grid grid-cols-[96px_1fr] gap-2 py-0.5 font-mono"
                  >
                    <span class="truncate text-gray-400">{{ key }}</span>
                    <span class="break-all text-gray-700 dark:text-gray-200">{{ value }}</span>
                  </div>
                  <div v-if="item.fingerprint?.last_seen" class="mt-1 text-gray-400">
                    {{ t('admin.providerStatus.fingerprintSeenAt') }} {{ formatDate(item.fingerprint.last_seen) }}
                  </div>
                </div>
              </div>
              <span v-else class="text-gray-400">{{ t('admin.providerStatus.noFingerprint') }}</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex min-w-[160px] gap-1">
                <span
                  v-for="(point, idx) in compactTimeline(item.timeline || [])"
                  :key="`${item.provider}-${idx}`"
                  data-testid="provider-status-timeline-dot"
                  class="relative h-2 flex-1 rounded-full"
                  :class="timelineClass(point)"
                  :title="timelineTitle(point)"
                  @mouseenter="showTimelineTooltip($event, item.provider, idx, point)"
                  @mousemove="moveTimelineTooltip"
                  @mouseleave="hideTimelineTooltip"
                />
              </div>
              <div
                v-if="hoveredTimeline && hoveredTimeline.provider === item.provider"
                class="pointer-events-none fixed z-50 rounded-lg bg-gray-950 px-3 py-2 text-xs leading-6 text-white shadow-xl"
                :style="tooltipStyle"
              >
                <div>{{ formatTimelineTime(hoveredTimeline.point.bucket_start) }}</div>
                <div>{{ formatNumber(hoveredTimeline.point.request_count) }} 个请求</div>
                <div>可用性 {{ formatPercent(hoveredTimeline.point.availability) }}</div>
                <div>延迟: {{ formatMs(hoveredTimeline.point.p50_ms) }}</div>
                <div>总耗时均值: {{ formatMs(hoveredTimeline.point.duration_avg_ms) }}</div>
                <div>客户端视角首响应: {{ formatMs(hoveredTimeline.point.ttft_avg_ms) }} / 样本 {{ formatNumber(hoveredTimeline.point.ttft_sample_count || 0) }}</div>
                <div>我站→上游: {{ formatMs(hoveredTimeline.point.upstream_ttft_avg_ms) }}</div>
                <div>网关处理/下发: {{ formatMs(hoveredTimeline.point.gateway_ttft_avg_ms) }}</div>
                <div>524: {{ formatNumber(hoveredTimeline.point.timeout_524_count || 0) }} 次 / 平均 {{ formatMs(hoveredTimeline.point.timeout_524_avg_ms) }}</div>
                <div class="flex gap-2">
                  <span class="text-emerald-400">OK: {{ formatNumber(hoveredTimeline.point.success_count) }}</span>
                  <span class="text-rose-400">ERR: {{ formatNumber(hoveredTimeline.point.failure_count) }}</span>
                </div>
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
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsProviderStatusItem, OpsProviderStatusTimelinePoint } from '@/api/admin/ops'

const { t } = useI18n()
defineProps<{
  items: OpsProviderStatusItem[]
  loading?: boolean
}>()

const hoveredTimeline = ref<{
  provider: string
  index: number
  point: OpsProviderStatusTimelinePoint
  x: number
  y: number
} | null>(null)
const expandedFingerprintProvider = ref<string | null>(null)

const tooltipStyle = computed(() => {
  if (!hoveredTimeline.value) return {}
  return {
    left: `${hoveredTimeline.value.x}px`,
    top: `${hoveredTimeline.value.y}px`,
    transform: 'translate(12px, -100%)',
  }
})

function formatNumber(value: number): string {
  return value.toLocaleString()
}

function formatMs(value?: number | null): string {
  if (value == null) return '—'
  if (value < 1000) return `${Math.round(value)}ms`
  if (value < 60_000) return `${trimFixed(value / 1000)}s`
  return `${trimFixed(value / 60_000)}m`
}

function formatDate(value?: string | null): string {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

function formatTimelineTime(value: string): string {
  return new Date(value).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatPercent(value: number): string {
  return `${trimFixed(value)}%`
}

function trimFixed(value: number): string {
  return value.toFixed(2).replace(/\.?0+$/, '')
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
      p50_ms: firstDefinedNumber(slice.map(point => point.p50_ms)),
      p95_ms: firstDefinedNumber(slice.map(point => point.p95_ms)),
      p99_ms: firstDefinedNumber(slice.map(point => point.p99_ms)),
      duration_avg_ms: weightedAverageMs(slice, 'duration_avg_ms', 'request_count'),
      ttft_avg_ms: weightedAverageMs(slice, 'ttft_avg_ms', 'ttft_sample_count'),
      ttft_sample_count: slice.reduce((sum, p) => sum + (p.ttft_sample_count || 0), 0),
      upstream_ttft_avg_ms: weightedAverageMs(slice, 'upstream_ttft_avg_ms', 'ttft_sample_count'),
      gateway_ttft_avg_ms: weightedAverageMs(slice, 'gateway_ttft_avg_ms', 'ttft_sample_count'),
      timeout_524_count: slice.reduce((sum, p) => sum + (p.timeout_524_count || 0), 0),
      timeout_524_avg_ms: weightedAverageMs(slice, 'timeout_524_avg_ms', 'timeout_524_count'),
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
  return `${formatTimelineTime(point.bucket_start)} · ${point.request_count} 个请求 · 可用性 ${formatPercent(point.availability)} · 延迟: ${formatMs(point.p50_ms)} · 总耗时: ${formatMs(point.duration_avg_ms)} · 客户端视角首响应: ${formatMs(point.ttft_avg_ms)} · 我站→上游: ${formatMs(point.upstream_ttft_avg_ms)} · 网关处理: ${formatMs(point.gateway_ttft_avg_ms)} · 524: ${point.timeout_524_count || 0} 次 · OK: ${point.success_count} · ERR: ${point.failure_count}`
}

function showTimelineTooltip(event: MouseEvent, provider: string, index: number, point: OpsProviderStatusTimelinePoint) {
  hoveredTimeline.value = {
    provider,
    index,
    point,
    x: event.clientX,
    y: event.clientY - 8,
  }
}

function moveTimelineTooltip(event: MouseEvent) {
  if (!hoveredTimeline.value) return
  hoveredTimeline.value = {
    ...hoveredTimeline.value,
    x: event.clientX,
    y: event.clientY - 8,
  }
}

function hideTimelineTooltip() {
  hoveredTimeline.value = null
}

function fingerprintEntries(item: OpsProviderStatusItem): Array<[string, string]> {
  const headers = item.fingerprint?.headers || {}
  return Object.entries(headers)
    .filter(([, value]) => String(value || '').trim() !== '')
    .sort(([a], [b]) => fingerprintHeaderPriority(a) - fingerprintHeaderPriority(b) || a.localeCompare(b))
}

function fingerprintPreviewEntries(item: OpsProviderStatusItem): Array<[string, string]> {
  return fingerprintEntries(item).slice(0, 2)
}

function fingerprintHeaderPriority(header: string): number {
  const order = ['server', 'via', 'cf-ray', 'cf-cache-status', 'x-request-id', 'openai-processing-ms']
  const idx = order.indexOf(header)
  return idx >= 0 ? idx : order.length
}

function toggleFingerprint(provider: string) {
  expandedFingerprintProvider.value = expandedFingerprintProvider.value === provider ? null : provider
}

function weightedAverageMs(points: OpsProviderStatusTimelinePoint[], valueKey: 'duration_avg_ms' | 'ttft_avg_ms' | 'upstream_ttft_avg_ms' | 'gateway_ttft_avg_ms' | 'timeout_524_avg_ms', weightKey: 'request_count' | 'ttft_sample_count' | 'timeout_524_count'): number | null {
  let weighted = 0
  let weightSum = 0
  for (const point of points) {
    const value = point[valueKey]
    const weight = point[weightKey] || 0
    if (typeof value === 'number' && Number.isFinite(value) && weight > 0) {
      weighted += value * weight
      weightSum += weight
    }
  }
  return weightSum > 0 ? Math.round(weighted / weightSum) : null
}

function firstDefinedNumber(values: Array<number | null | undefined>): number | null {
  for (const value of values) {
    if (typeof value === 'number' && Number.isFinite(value)) {
      return value
    }
  }
  return null
}
</script>
