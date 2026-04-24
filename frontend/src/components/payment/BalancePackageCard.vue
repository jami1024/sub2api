<template>
  <div
    :data-testid="`balance-package-card-${pkg.id}`"
    :aria-disabled="disabled ? 'true' : 'false'"
    :class="[
      'card h-full rounded-3xl border border-gray-100 bg-white/95 p-5 text-left shadow-sm transition duration-200 dark:border-dark-700 dark:bg-dark-900/70',
      disabled
        ? 'cursor-not-allowed opacity-70 saturate-[0.85]'
        : 'cursor-pointer hover:-translate-y-0.5 hover:shadow-lg hover:border-gray-200 dark:hover:border-dark-600'
    ]"
    @click="handleSelect"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <span :class="['inline-flex rounded-full border px-2.5 py-1 text-xs font-semibold', scopeBadgeClass]">
          {{ scopeLabel }}
        </span>
        <h3 class="mt-4 truncate text-lg font-bold text-gray-900 dark:text-white">{{ pkg.name }}</h3>
        <p
          v-if="pkg.description"
          class="mt-2 line-clamp-2 min-h-[2.5rem] text-sm leading-5 text-gray-500 dark:text-gray-400"
        >
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

    <div :data-testid="`balance-package-card-support-${pkg.id}`" class="mt-4 rounded-2xl bg-slate-50/80 px-3 py-3 text-sm dark:bg-dark-800/60">
      <div class="font-medium text-slate-700 dark:text-slate-200">
        {{ supportText }}
      </div>
      <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">
        {{ t('payment.balancePackages.noMixedScope') }}
      </div>
    </div>

    <p v-if="disabled && disabledReason" class="mt-3 text-xs leading-5 text-amber-700/90 dark:text-amber-300/90">
      {{ disabledReason }}
    </p>

    <div class="mt-5 flex items-end justify-between gap-3">
      <div>
        <div class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.amountLabel') }}</div>
        <div :data-testid="`balance-package-card-price-${pkg.id}`" class="mt-1 text-3xl font-bold text-gray-900 dark:text-white">¥{{ pkg.price.toFixed(2) }}</div>
      </div>
      <button
        type="button"
        :data-testid="disabled && canForceSwitch ? `balance-package-force-switch-${pkg.id}` : `balance-package-select-${pkg.id}`"
        :disabled="disabled && !canForceSwitch"
        class="btn shrink-0"
        :class="disabled && canForceSwitch ? 'btn-secondary' : 'btn-primary'"
        @click.stop="disabled ? handleForceSwitch() : handleSelect()"
      >
        {{
          disabled
            ? (canForceSwitch ? t('payment.balancePackages.forceSwitch') : t('payment.balancePackages.unavailable'))
            : t('payment.balancePackages.buyNow')
        }}
      </button>
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
  disabled?: boolean
  disabledReason?: string
  canForceSwitch?: boolean
}>()

const emit = defineEmits<{
  (e: 'select'): void
  (e: 'force-switch'): void
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

function handleSelect() {
  if (props.disabled) return
  emit('select')
}

function handleForceSwitch() {
  if (!props.disabled || !props.canForceSwitch) return
  emit('force-switch')
}
</script>
