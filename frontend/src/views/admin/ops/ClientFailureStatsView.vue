<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.clientFailures.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.description') }}</p>
        </div>
      </div>

      <ClientFailureStatsFilters v-model="timeRange" :loading="loading" @refresh="reload" />

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.totalFailures') }}</div>
          <div class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(totalFailures) }}</div>
        </div>
        <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.affectedUsers') }}</div>
          <div class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(items.length) }}</div>
        </div>
        <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.affectedKeys') }}</div>
          <div class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(totalKeys) }}</div>
        </div>
      </div>

      <ClientFailureStatsTable :items="items" :loading="loading" @select-user="openUserErrors" />

      <OpsErrorDetailsModal
        v-model:show="detailsOpen"
        :time-range="timeRange"
        error-type="request"
        :user-id="selectedUserId"
        initial-error-owner="client"
        initial-phase="request"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { opsAPI, type OpsClientFailureStatsItem, type OpsClientFailureStatsTimeRange } from '@/api/admin/ops'
import AppLayout from '@/components/layout/AppLayout.vue'
import OpsErrorDetailsModal from './components/OpsErrorDetailsModal.vue'
import ClientFailureStatsFilters from './components/ClientFailureStatsFilters.vue'
import ClientFailureStatsTable from './components/ClientFailureStatsTable.vue'

const { t } = useI18n()
const appStore = useAppStore()

const timeRange = ref<OpsClientFailureStatsTimeRange>('24h')
const loading = ref(false)
const items = ref<OpsClientFailureStatsItem[]>([])
const detailsOpen = ref(false)
const selectedUserId = ref<number | null>(null)
let abortController: AbortController | null = null

const totalFailures = computed(() => items.value.reduce((sum, item) => sum + item.failure_count, 0))
const totalKeys = computed(() => items.value.reduce((sum, item) => sum + item.affected_key_count, 0))

function formatNumber(value: number): string {
  return value.toLocaleString()
}

function openUserErrors(item: OpsClientFailureStatsItem) {
  selectedUserId.value = item.user_id || null
  detailsOpen.value = true
}

async function reload() {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  loading.value = true
  try {
    const data = await opsAPI.getClientFailureStats({ time_range: timeRange.value }, { signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = data.items || []
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('admin.clientFailures.loadError')))
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
