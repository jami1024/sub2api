<template>
  <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="flex flex-col gap-1 border-b border-gray-100 px-4 py-3 dark:border-gray-700 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.clientFailures.tableTitle') }}</h3>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.tableHint') }}</p>
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.clickHint') }}</div>
    </div>

    <div v-if="loading" class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="items.length === 0" class="flex h-48 flex-col items-center justify-center gap-2 text-sm text-gray-500 dark:text-gray-400">
      <div>{{ t('admin.clientFailures.empty') }}</div>
      <div class="text-xs">{{ t('admin.clientFailures.emptyHint') }}</div>
    </div>
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-gray-700">
        <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-700/60 dark:text-gray-400">
          <tr>
            <th class="px-4 py-3 text-left">{{ t('admin.clientFailures.user') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.clientFailures.failures') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.clientFailures.keys') }}</th>
            <th class="px-4 py-3 text-left">{{ t('admin.clientFailures.topError') }}</th>
            <th class="px-4 py-3 text-left">{{ t('admin.clientFailures.endpoint') }}</th>
            <th class="px-4 py-3 text-left">{{ t('admin.clientFailures.platform') }}</th>
            <th class="px-4 py-3 text-right">{{ t('admin.clientFailures.lastSeen') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
          <tr
            v-for="item in items"
            :key="`${item.user_id || 'unknown'}-${item.user_email}`"
            class="cursor-pointer hover:bg-gray-50/80 dark:hover:bg-dark-700/40"
            @click="emit('selectUser', item)"
          >
            <td class="px-4 py-3">
              <div class="font-semibold text-gray-900 dark:text-white">{{ displayUser(item) }}</div>
              <div v-if="item.user_id" class="text-xs text-gray-500 dark:text-gray-400">ID: {{ item.user_id }}</div>
            </td>
            <td class="px-4 py-3 text-right font-semibold text-rose-600 dark:text-rose-400">{{ formatNumber(item.failure_count) }}</td>
            <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatNumber(item.affected_key_count) }}</td>
            <td class="max-w-[360px] px-4 py-3 text-gray-700 dark:text-gray-300">
              <div class="truncate" :title="item.top_error_message || '—'">{{ item.top_error_message || '—' }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.clientFailures.topErrorCount', { count: formatNumber(item.top_error_count) }) }}</div>
            </td>
            <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ item.top_inbound_endpoint || '—' }}</td>
            <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ item.top_platform || '—' }}</td>
            <td class="px-4 py-3 text-right text-xs text-gray-500 dark:text-gray-400">{{ formatDate(item.last_seen) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { OpsClientFailureStatsItem } from '@/api/admin/ops'

const { t } = useI18n()

defineProps<{
  items: OpsClientFailureStatsItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  selectUser: [item: OpsClientFailureStatsItem]
}>()

function displayUser(item: OpsClientFailureStatsItem): string {
  return item.user_email || t('admin.clientFailures.unknownUser')
}

function formatNumber(value: number): string {
  return value.toLocaleString()
}

function formatDate(value?: string | null): string {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}
</script>
