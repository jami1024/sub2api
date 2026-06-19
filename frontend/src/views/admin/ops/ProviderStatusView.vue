<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.providerStatus.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-gray-400">{{ t('admin.providerStatus.description') }}</p>
        </div>
      </div>

      <ProviderStatusFilters v-model="timeRange" :loading="loading" @refresh="reload" />
      <ProviderStatusSummaryCards :items="items" />
      <ProviderStatusTable :items="items" :loading="loading" />
      <ProviderStatusLatencyChart :points="latencyPoints" :loading="loading" />
      <UpstreamMultiplierPanel
        v-model:model="multiplierModel"
        :accounts="multiplierAccounts"
        :samples="multiplierSamples"
        :loading="multiplierLoading"
        :measuring="multiplierMeasuring"
        :measuring-account-id="multiplierMeasuringAccountId"
        @refresh="loadUpstreamMultipliers"
        @measure-account="measureUpstreamAccount"
        @measure-all="measureAllUpstreamAccounts"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  opsAPI,
  type OpsProviderStatusItem,
  type OpsProviderStatusTimeRange,
  type OpsUpstreamMultiplierAccount,
  type OpsUpstreamMultiplierSample,
} from '@/api/admin/ops'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProviderStatusFilters from './components/ProviderStatusFilters.vue'
import ProviderStatusSummaryCards from './components/ProviderStatusSummaryCards.vue'
import ProviderStatusTable from './components/ProviderStatusTable.vue'
import ProviderStatusLatencyChart from './components/ProviderStatusLatencyChart.vue'
import UpstreamMultiplierPanel from './components/UpstreamMultiplierPanel.vue'

const { t } = useI18n()
const appStore = useAppStore()

const timeRange = ref<OpsProviderStatusTimeRange>('1h')
const loading = ref(false)
const items = ref<OpsProviderStatusItem[]>([])
const multiplierModel = ref('gpt-5.4')
const multiplierLoading = ref(false)
const multiplierMeasuring = ref(false)
const multiplierMeasuringAccountId = ref<number | null>(null)
const multiplierAccounts = ref<OpsUpstreamMultiplierAccount[]>([])
const multiplierSamples = ref<OpsUpstreamMultiplierSample[]>([])
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

async function loadUpstreamMultipliers() {
  multiplierLoading.value = true
  try {
    const model = multiplierModel.value.trim() || 'gpt-5.4'
    const [accounts, samples] = await Promise.all([
      opsAPI.getUpstreamMultiplierAccounts({ model }),
      opsAPI.getUpstreamMultiplierSamples({ model, limit: 100 }),
    ])
    multiplierAccounts.value = accounts.accounts || []
    multiplierSamples.value = samples.samples || []
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '加载上游倍率记录失败'))
  } finally {
    multiplierLoading.value = false
  }
}

async function measureUpstreamAccount(accountID: number) {
  multiplierMeasuring.value = true
  multiplierMeasuringAccountId.value = accountID
  try {
    const model = multiplierModel.value.trim() || 'gpt-5.4'
    await opsAPI.measureUpstreamMultipliers({ model, account_ids: [accountID] })
    await loadUpstreamMultipliers()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '上游倍率检测失败'))
  } finally {
    multiplierMeasuring.value = false
    multiplierMeasuringAccountId.value = null
  }
}

async function measureAllUpstreamAccounts() {
  const ids = multiplierAccounts.value.filter(account => account.supported).map(account => account.account_id)
  if (ids.length === 0) return
  if (!window.confirm(`将对 ${ids.length} 个上游账号发起真实小请求，确认继续？`)) return
  multiplierMeasuring.value = true
  try {
    const model = multiplierModel.value.trim() || 'gpt-5.4'
    await opsAPI.measureUpstreamMultipliers({ model, account_ids: ids })
    await loadUpstreamMultipliers()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '上游倍率检测失败'))
  } finally {
    multiplierMeasuring.value = false
  }
}

watch(timeRange, () => {
  void reload()
})

watch(multiplierModel, () => {
  void loadUpstreamMultipliers()
})

onMounted(() => {
  void reload()
  void loadUpstreamMultipliers()
})

onBeforeUnmount(() => {
  if (abortController) abortController.abort()
})
</script>
