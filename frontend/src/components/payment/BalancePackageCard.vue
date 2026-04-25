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
      <div class="min-w-0 flex-1">
        <span :class="['inline-flex rounded-full border px-2.5 py-1 text-xs font-semibold', scopeBadgeClass]">
          {{ scopeLabel }}
        </span>
        <div v-if="pkg.display_tags?.length" class="mt-3 flex flex-wrap gap-2">
          <span
            v-for="(tag, index) in pkg.display_tags.slice(0, 1)"
            :key="`${tag}-${index}`"
            :data-testid="`balance-package-card-tag-${pkg.id}-${index}`"
            :class="[
              'inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-medium shadow-sm',
              displayTagClass(tag),
            ]"
          >
            {{ tag }}
          </span>
        </div>
        <div class="mt-4 min-h-[3.5rem]">
          <h3 class="line-clamp-2 text-lg font-bold leading-7 text-gray-900 dark:text-white">{{ pkg.name }}</h3>
        </div>
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

    <p v-if="disabled && disabledReason && !canForceSwitch" class="mt-3 text-xs leading-5 text-amber-700/90 dark:text-amber-300/90">
      {{ disabledReason }}
    </p>

    <div class="mt-5 flex items-end justify-between gap-3">
      <div class="min-w-0 flex-1">
        <div class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.amountLabel') }}</div>
        <div
          :data-testid="`balance-package-card-price-${pkg.id}`"
          class="mt-1 truncate text-3xl font-bold text-gray-900 dark:text-white"
        >
          ¥{{ pkg.price.toFixed(2) }}
        </div>
      </div>
      <button
        type="button"
        :data-testid="disabled && canForceSwitch ? `balance-package-force-switch-${pkg.id}` : `balance-package-select-${pkg.id}`"
        :disabled="disabled && !canForceSwitch"
        class="btn h-14 w-32 shrink-0 justify-center whitespace-nowrap px-4 text-sm"
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

function displayTagClass(tag: string) {
  const normalized = tag.trim().toLowerCase()
  if (normalized.includes('推荐') || normalized.includes('热销') || normalized.includes('popular')) {
    return 'border-emerald-200/80 bg-emerald-50 text-emerald-700 dark:border-emerald-900/50 dark:bg-emerald-950/20 dark:text-emerald-300 dark:shadow-none'
  }
  if (normalized.includes('倍率') || normalized.includes('1x') || normalized.includes('2x') || normalized.includes('rate')) {
    return 'border-amber-200/80 bg-amber-50 text-amber-700 dark:border-amber-900/50 dark:bg-amber-950/20 dark:text-amber-300 dark:shadow-none'
  }
  return 'border-slate-200/80 bg-slate-50 text-slate-600 dark:border-dark-600 dark:bg-dark-800 dark:text-slate-300 dark:shadow-none'
}

function handleSelect() {
  if (props.disabled) return
  emit('select')
}

function handleForceSwitch() {
  if (!props.disabled || !props.canForceSwitch) return
  emit('force-switch')
}
</script>
