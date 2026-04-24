<template>
  <div data-testid="balance-package-summary" class="card rounded-3xl border border-gray-100 bg-white/95 p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900/70">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div class="min-w-0">
        <span :class="['inline-flex rounded-full border px-2.5 py-1 text-xs font-semibold', scopeBadgeClass]">
          {{ scopeLabel }}
        </span>
        <h3 class="mt-3 text-xl font-bold text-gray-900 dark:text-white">{{ pkg.name }}</h3>
        <p v-if="pkg.description" class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
          {{ pkg.description }}
        </p>
      </div>
      <div class="shrink-0 text-right">
        <div class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.creditedBalance') }}</div>
        <div class="mt-1 text-3xl font-black text-primary-600 dark:text-primary-400">
          ${{ pkg.credit_amount.toFixed(2) }}
        </div>
      </div>
    </div>

    <div class="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div class="rounded-2xl bg-slate-50/80 p-4 dark:bg-dark-800/60">
        <div class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.amountLabel') }}</div>
        <div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">¥{{ pkg.price.toFixed(2) }}</div>
      </div>
      <div class="rounded-2xl bg-slate-50/80 p-4 dark:bg-dark-800/60">
        <div class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.balancePackages.supportRange') }}</div>
        <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ supportText }}</div>
      </div>
    </div>

    <div class="mt-4 rounded-2xl bg-slate-50/80 p-4 text-sm text-slate-700 dark:bg-dark-800/60 dark:text-slate-200">
      <div class="font-medium">{{ t('payment.balancePackages.purchaseNotice') }}</div>
      <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ t('payment.balancePackages.noMixedScope') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { BalancePackage } from '@/types/payment'
import { packageScopeBadgeClass, packageScopeLabelKey } from '@/utils/packageScopeBadge'

const props = defineProps<{
  pkg: BalancePackage
}>()

const { t } = useI18n()

const scopeLabel = computed(() =>
  t(packageScopeLabelKey(props.pkg.package_scope)),
)

const supportText = computed(() =>
  props.pkg.package_scope === 'codex'
    ? t('payment.balancePackages.supportsCodexOnly')
    : t('payment.balancePackages.supportsGeneralOnly'),
)

const scopeBadgeClass = computed(() =>
  packageScopeBadgeClass(props.pkg.package_scope),
)
</script>
