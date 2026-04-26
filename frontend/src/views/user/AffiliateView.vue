<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 md:grid-cols-5">
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ formatCount(detail.aff_count) }}</p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.pendingQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-400">{{ formatRebateCurrency(detail.pending_quota || 0) }}</p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">{{ formatRebateCurrency(detail.aff_quota) }}</p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ formatRebateCurrency(detail.aff_history_quota) }}</p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.debtQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-red-600 dark:text-red-400">{{ formatRebateCurrency(detail.debt_quota || 0) }}</p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.description') }}</p>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ t('affiliate.tips.title') }}</p>
            <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
              <li>1. {{ t('affiliate.tips.line1') }}</li>
              <li>2. {{ t('affiliate.tips.line2') }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
            </ul>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.rules.title') }}</h3>
          <div class="mt-4 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <ul class="space-y-2 text-sm text-primary-800 dark:text-primary-200">
              <li>1. {{ t('affiliate.rules.line1') }}</li>
              <li>2. {{ t('affiliate.rules.line2') }}</li>
              <li>3. {{ t('affiliate.rules.line3') }}</li>
              <li>4. {{ t('affiliate.rules.line4') }}</li>
              <li>5. {{ t('affiliate.rules.line5') }}</li>
              <li>6. {{ t('affiliate.rules.line6') }}</li>
            </ul>
          </div>

          <div class="mt-4 rounded-xl border border-primary-200/80 bg-white/70 p-4 dark:border-primary-900/30 dark:bg-dark-900/30">
            <p class="text-sm font-semibold text-primary-900 dark:text-primary-100">{{ t('affiliate.rules.exampleTitle') }}</p>
            <p class="mt-2 text-sm text-primary-800 dark:text-primary-200">{{ t('affiliate.rules.exampleChain') }}</p>
            <ul class="mt-3 space-y-2 text-sm text-primary-800 dark:text-primary-200">
              <li>{{ t('affiliate.rules.exampleLevel1') }}</li>
              <li>{{ t('affiliate.rules.exampleLevel2') }}</li>
              <li>{{ t('affiliate.rules.exampleLevel3') }}</li>
            </ul>
            <p class="mt-3 text-sm text-primary-800 dark:text-primary-200">{{ t('affiliate.rules.exampleNote') }}</p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.rebates.title') }}</h3>
          <div v-if="rebateRecords.length === 0" class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.rebates.empty') }}</div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[640px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.rebates.columns.level') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.rebates.columns.sourceUser') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.rebates.columns.sourceOrder') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('common.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('common.status') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('common.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in rebateRecords" :key="item.id" class="border-b border-gray-100 last:border-b-0 dark:border-dark-800">
                  <td class="px-3 py-3">{{ rebateLevelLabel(item.level) }}</td>
                  <td class="px-3 py-3">{{ rebateSourceUserLabel(item) }}</td>
                  <td class="px-3 py-3">#{{ item.source_order_id }}</td>
                  <td class="px-3 py-3">{{ formatRebateCurrency(item.rebate_amount) }}</td>
                  <td class="px-3 py-3">{{ rebateStatusLabel(item.status) }}</td>
                  <td class="px-3 py-3">{{ formatDateTime(item.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.transfer.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button
              data-testid="affiliate-withdraw-open"
              class="btn btn-primary"
              :disabled="!canRequestWithdrawal || requestingWithdrawal"
              :aria-disabled="!canRequestWithdrawal || requestingWithdrawal"
              @click="showWithdrawDialog = true"
            >
              <Icon name="dollar" size="sm" />
              <span>{{ requestingWithdrawal ? t('affiliate.transfer.requesting') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">{{ t('affiliate.transfer.empty') }}</p>
          <p v-else-if="detail.aff_quota < withdrawalThreshold" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.transfer.thresholdHint', { amount: withdrawalThreshold.toFixed(0) }) }}
          </p>
          <p v-else-if="detail.debt_quota > 0" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ t('affiliate.transfer.debtHint') }}</p>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.manualHint') }}</p>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.withdrawals.title') }}</h3>
          <div v-if="withdrawals.length === 0" class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.withdrawals.empty') }}</div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('common.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('common.status') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('common.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in withdrawals" :key="item.id" class="border-b border-gray-100 last:border-b-0 dark:border-dark-800">
                  <td class="px-3 py-3">{{ formatRebateCurrency(item.amount) }}</td>
                  <td class="px-3 py-3">{{ rebateStatusLabel(item.status) }}</td>
                  <td class="px-3 py-3">{{ formatDateTime(item.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in detail.invitees" :key="item.user_id" class="border-b border-gray-100 last:border-b-0 dark:border-dark-800">
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <BaseDialog :show="showWithdrawDialog" :title="t('affiliate.transfer.dialogTitle')" width="normal" @close="showWithdrawDialog = false">
      <div class="space-y-4">
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.dialogDescription') }}</p>
        <div>
          <label class="input-label">{{ t('affiliate.transfer.requestAmount') }}</label>
          <input v-model.number="withdrawAmount" type="number" min="100" step="0.01" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('affiliate.transfer.requestNote') }}</label>
          <textarea v-model="withdrawNote" class="input min-h-[96px]" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="showWithdrawDialog = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="!canSubmitWithdrawDialog || requestingWithdrawal" @click="requestWithdrawal">
            {{ requestingWithdrawal ? t('affiliate.transfer.requesting') : t('affiliate.transfer.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import userAPI from '@/api/user'
import type { AffiliateRebateRecord, AffiliateWithdrawalRequest, UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const detail = ref<UserAffiliateDetail | null>(null)
const rebateRecords = ref<AffiliateRebateRecord[]>([])
const withdrawals = ref<AffiliateWithdrawalRequest[]>([])
const requestingWithdrawal = ref(false)
const withdrawalThreshold = 100
const showWithdrawDialog = ref(false)
const withdrawAmount = ref(100)
const withdrawNote = ref('')

const canRequestWithdrawal = computed(() => {
  if (!detail.value) return false
  return detail.value.aff_quota >= withdrawalThreshold && (detail.value.debt_quota || 0) <= 0
})

const canSubmitWithdrawDialog = computed(() => {
  if (!detail.value) return false
  return withdrawAmount.value >= withdrawalThreshold && withdrawAmount.value <= detail.value.aff_quota && (detail.value.debt_quota || 0) <= 0
})

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

function formatRebateCurrency(value: number): string {
  return `¥${Number(value || 0).toFixed(2)}`
}

function rebateLevelLabel(level: number): string {
  if (level === 1) return t('affiliate.rebates.level1')
  if (level === 2) return t('affiliate.rebates.level2')
  if (level === 3) return t('affiliate.rebates.level3')
  return t('affiliate.rebates.levelUnknown', { level })
}

function rebateSourceUserLabel(item: AffiliateRebateRecord): string {
  return item.source_username?.trim() || item.source_email?.trim() || '-'
}

function rebateStatusLabel(status: string): string {
  return t(`affiliate.rebates.status.${status}`)
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) loading.value = true
  try {
    const [affiliateDetail, rebateItems, withdrawalItems] = await Promise.all([
      userAPI.getAffiliateDetail(),
      userAPI.getAffiliateRebateRecords(),
      userAPI.getAffiliateWithdrawalRequests(),
    ])
    detail.value = affiliateDetail
    rebateRecords.value = rebateItems
    withdrawals.value = withdrawalItems
    if (affiliateDetail.aff_quota >= withdrawalThreshold) {
      withdrawAmount.value = Math.floor(affiliateDetail.aff_quota * 100) / 100
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) loading.value = false
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function requestWithdrawal(): Promise<void> {
  if (!detail.value || !canSubmitWithdrawDialog.value || requestingWithdrawal.value) return
  requestingWithdrawal.value = true
  try {
    await userAPI.createAffiliateWithdrawalRequest({
      amount: withdrawAmount.value,
      applicant_note: withdrawNote.value.trim() || undefined,
    })
    appStore.showSuccess(t('affiliate.transfer.requestSuccess'))
    showWithdrawDialog.value = false
    withdrawNote.value = ''
    await loadAffiliateDetail(true)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transfer.requestFailed')))
  } finally {
    requestingWithdrawal.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>
