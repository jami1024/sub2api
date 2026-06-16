<template>
  <div class="space-y-5 p-4 sm:p-6">
    <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.providerStatus.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.providerStatus.description') }}</p>
      </div>
    </div>

    <ProviderStatusFilters v-model="timeRange" :loading="loading" @refresh="reload" />
    <ProviderStatusSummaryCards :items="items" />
    <ProviderStatusTable :items="items" :loading="loading" />
    <ProviderStatusLatencyChart :points="latencyPoints" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { opsAPI, type OpsProviderStatusItem, type OpsProviderStatusTimeRange } from '@/api/admin/ops'
import ProviderStatusFilters from './components/ProviderStatusFilters.vue'
import ProviderStatusSummaryCards from './components/ProviderStatusSummaryCards.vue'
import ProviderStatusTable from './components/ProviderStatusTable.vue'
import ProviderStatusLatencyChart from './components/ProviderStatusLatencyChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const timeRange = ref<OpsProviderStatusTimeRange>('1h')
const loading = ref(false)
const items = ref<OpsProviderStatusItem[]>([])
let abortController: AbortController | null = null

const latencyPoints = computed(() => {
  const busiest = [...items.value].sort((a, b) => b.request_count - a.request_count)[0]
  return busiest?.timeline || []
})

async function reload() {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  loading.value = true
  try {
    const data = await opsAPI.getProviderStatus({ time_range: timeRange.value }, { signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = data.items || []
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('admin.providerStatus.loadError')))
  } finally {
    if (abortController === ctrl) {
      loading.value = false
      abortController = null
    }
  }
}

watch(timeRange, () => {
  void reload()
})

onMounted(() => {
  void reload()
})

onBeforeUnmount(() => {
  if (abortController) abortController.abort()
})
</script>
