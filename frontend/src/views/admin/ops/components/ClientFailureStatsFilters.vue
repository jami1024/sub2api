<template>
  <div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-gray-200 bg-white p-3 shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="flex flex-wrap gap-2">
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        class="rounded-xl px-3 py-1.5 text-sm font-medium transition-colors"
        :class="modelValue === option.value
          ? 'bg-primary-600 text-white shadow-sm'
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
        @click="emit('update:modelValue', option.value)"
      >
        {{ option.label }}
      </button>
    </div>
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-xl border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-dark-700"
      :disabled="loading"
      @click="emit('refresh')"
    >
      <span :class="loading ? 'animate-spin' : ''">↻</span>
      {{ t('admin.clientFailures.refresh') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { OpsClientFailureStatsTimeRange } from '@/api/admin/ops'

const { t } = useI18n()

defineProps<{
  modelValue: OpsClientFailureStatsTimeRange
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: OpsClientFailureStatsTimeRange]
  refresh: []
}>()

const options: Array<{ value: OpsClientFailureStatsTimeRange; label: string }> = [
  { value: '15m', label: t('admin.clientFailures.ranges.15m') },
  { value: '1h', label: t('admin.clientFailures.ranges.1h') },
  { value: '6h', label: t('admin.clientFailures.ranges.6h') },
  { value: '24h', label: t('admin.clientFailures.ranges.24h') },
  { value: '7d', label: t('admin.clientFailures.ranges.7d') },
]
</script>
