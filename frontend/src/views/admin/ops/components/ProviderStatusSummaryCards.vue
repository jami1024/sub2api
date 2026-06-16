<template>
  <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div v-for="card in cards" :key="card.label" class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-dark-800">
      <div class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ card.label }}</div>
      <div class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ card.value }}</div>
      <div v-if="card.hint" class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ card.hint }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsProviderStatusItem } from '@/api/admin/ops'

const { t } = useI18n()
const props = defineProps<{ items: OpsProviderStatusItem[] }>()

const totals = computed(() => {
  const request = props.items.reduce((sum, item) => sum + item.request_count, 0)
  const success = props.items.reduce((sum, item) => sum + item.success_count, 0)
  const failure = props.items.reduce((sum, item) => sum + item.failure_count, 0)
  const providers = props.items.length
  const availability = request > 0 ? (success / Math.max(success + failure, 1)) * 100 : 0
  const errorRate = request > 0 ? (failure / Math.max(success + failure, 1)) * 100 : 0
  return { request, success, failure, providers, availability, errorRate }
})

const cards = computed(() => [
  { label: t('admin.providerStatus.totalRequests'), value: formatNumber(totals.value.request), hint: '' },
  { label: t('admin.providerStatus.overallAvailability'), value: `${totals.value.availability.toFixed(1)}%`, hint: '' },
  { label: t('admin.providerStatus.errorRate'), value: `${totals.value.errorRate.toFixed(1)}%`, hint: '' },
  { label: t('admin.providerStatus.providerCount'), value: formatNumber(totals.value.providers), hint: '' },
])

function formatNumber(value: number): string {
  return value.toLocaleString()
}
</script>
