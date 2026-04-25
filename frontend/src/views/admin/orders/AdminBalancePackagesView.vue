<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="flex items-center justify-end gap-2">
        <button @click="loadPackages" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="openEdit(null)" class="btn btn-primary">{{ t('payment.admin.createBalancePackage') }}</button>
      </div>

      <DataTable :columns="columns" :data="packages" :loading="loading">
        <template #cell-package_scope="{ value }">
          <span class="text-sm font-medium">{{ value === 'codex' ? t('payment.balancePackages.codex') : t('payment.balancePackages.general') }}</span>
        </template>
        <template #cell-price="{ value }">
          <span class="text-sm font-medium text-gray-900 dark:text-white">¥{{ Number(value ?? 0).toFixed(2) }}</span>
        </template>
        <template #cell-credit_amount="{ value }">
          <span class="text-sm font-medium text-primary-600 dark:text-primary-400">${{ Number(value ?? 0).toFixed(2) }}</span>
        </template>
        <template #cell-for_sale="{ value, row }">
          <button
            type="button"
            :class="[
              'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out',
              value ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
            ]"
            @click="toggleForSale(row)"
          >
            <span :class="[
              'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
              value ? 'translate-x-4' : 'translate-x-0'
            ]" />
          </button>
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-2">
            <button @click="openEdit(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400">
              <Icon name="edit" size="sm" />
              <span class="text-xs">{{ t('common.edit') }}</span>
            </button>
            <button @click="confirmDelete(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400">
              <Icon name="trash" size="sm" />
              <span class="text-xs">{{ t('common.delete') }}</span>
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <BalancePackageEditDialog :show="showDialog" :balance-package="editingPackage" @close="showDialog = false" @saved="loadPackages" />
    <ConfirmDialog :show="showDeleteDialog" :title="t('payment.admin.deleteBalancePackage')" :message="t('payment.admin.deleteBalancePackageConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDelete" @cancel="showDeleteDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { BalancePackage } from '@/types/payment'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import BalancePackageEditDialog from './BalancePackageEditDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const packages = ref<BalancePackage[]>([])
const showDialog = ref(false)
const showDeleteDialog = ref(false)
const editingPackage = ref<BalancePackage | null>(null)
const deletingPackageId = ref<number | null>(null)

const columns = computed((): Column[] => [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('payment.admin.balancePackageName') },
  { key: 'package_scope', label: t('admin.groups.packageScope') },
  { key: 'price', label: t('payment.admin.price') },
  { key: 'credit_amount', label: t('payment.admin.balancePackageCreditAmount') },
  { key: 'for_sale', label: t('payment.admin.forSale') },
  { key: 'sort_order', label: t('payment.admin.sortOrder') },
  { key: 'actions', label: t('common.actions') },
])

async function loadPackages() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getBalancePackages()
    packages.value = res.data || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function openEdit(pkg: BalancePackage | null) {
  editingPackage.value = pkg
  showDialog.value = true
}

async function toggleForSale(pkg: BalancePackage) {
  try {
    await adminPaymentAPI.updateBalancePackage(pkg.id, { for_sale: !pkg.for_sale })
    pkg.for_sale = !pkg.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDelete(pkg: BalancePackage) {
  deletingPackageId.value = pkg.id
  showDeleteDialog.value = true
}

async function handleDelete() {
  if (!deletingPackageId.value) return
  try {
    await adminPaymentAPI.deleteBalancePackage(deletingPackageId.value)
    appStore.showSuccess(t('common.deleted'))
    showDeleteDialog.value = false
    loadPackages()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

onMounted(loadPackages)
</script>
