<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('affiliateWithdrawals.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliateWithdrawals.description') }}</p>
          </div>
          <button class="btn btn-secondary" @click="loadWithdrawals">
            {{ t('common.refresh') }}
          </button>
        </div>
      </div>

      <div class="card p-0 overflow-hidden">
        <div v-if="loading" class="p-6 text-sm text-gray-500 dark:text-dark-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="withdrawals.length === 0" class="p-6 text-sm text-gray-500 dark:text-dark-400">
          {{ t('affiliateWithdrawals.empty') }}
        </div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="bg-gray-50 text-left text-gray-500 dark:bg-dark-900 dark:text-dark-400">
              <tr>
                <th class="px-4 py-3 font-medium">{{ t('common.id') }}</th>
                <th class="px-4 py-3 font-medium">{{ t('common.user') }}</th>
                <th class="px-4 py-3 font-medium">{{ t('common.amount') }}</th>
                <th class="px-4 py-3 font-medium">{{ t('common.status') }}</th>
                <th class="px-4 py-3 font-medium">{{ t('common.createdAt') }}</th>
                <th class="px-4 py-3 font-medium">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in withdrawals" :key="item.id" class="border-t border-gray-100 dark:border-dark-800">
                <td class="px-4 py-3">{{ item.id }}</td>
                <td class="px-4 py-3">{{ item.user_id }}</td>
                <td class="px-4 py-3">{{ formatCurrency(item.amount) }}</td>
                <td class="px-4 py-3">{{ item.status }}</td>
                <td class="px-4 py-3">{{ formatDateTime(item.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex gap-2">
                    <button
                      class="btn btn-secondary btn-sm"
                      :disabled="item.status !== 'pending' || actionLoadingId === item.id"
                      @click="reject(item.id)"
                    >
                      {{ t('common.reject') }}
                    </button>
                    <button
                      class="btn btn-primary btn-sm"
                      :disabled="item.status !== 'pending' || actionLoadingId === item.id"
                      @click="markPaid(item.id)"
                    >
                      {{ t('affiliateWithdrawals.markPaid') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import { adminAffiliateAPI } from '@/api/admin/affiliate'
import type { AffiliateWithdrawalRequest } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const actionLoadingId = ref<number | null>(null)
const withdrawals = ref<AffiliateWithdrawalRequest[]>([])

async function loadWithdrawals() {
  loading.value = true
  try {
    withdrawals.value = await adminAffiliateAPI.getAffiliateWithdrawals()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliateWithdrawals.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function reject(id: number) {
  actionLoadingId.value = id
  try {
    await adminAffiliateAPI.rejectAffiliateWithdrawal(id)
    appStore.showSuccess(t('affiliateWithdrawals.rejectSuccess'))
    await loadWithdrawals()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliateWithdrawals.actionFailed')))
  } finally {
    actionLoadingId.value = null
  }
}

async function markPaid(id: number) {
  actionLoadingId.value = id
  try {
    await adminAffiliateAPI.markAffiliateWithdrawalPaid(id)
    appStore.showSuccess(t('affiliateWithdrawals.markPaidSuccess'))
    await loadWithdrawals()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliateWithdrawals.actionFailed')))
  } finally {
    actionLoadingId.value = null
  }
}

onMounted(() => {
  void loadWithdrawals()
})
</script>
