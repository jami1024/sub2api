<template>
  <BaseDialog :show="show" :title="balancePackage ? t('payment.admin.editBalancePackage') : t('payment.admin.createBalancePackage')" width="wide" @close="emit('close')">
    <form id="balance-package-form" @submit.prevent="handleSave" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.balancePackageName') }} <span class="text-red-500">*</span></label>
          <input v-model="form.name" type="text" class="input" required />
        </div>
        <div>
          <label class="input-label">{{ t('admin.groups.packageScope') }} <span class="text-red-500">*</span></label>
          <Select v-model="form.package_scope" :options="packageScopeOptions" />
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('payment.admin.planDescription') }}</label>
        <textarea v-model="form.description" rows="2" class="input" />
      </div>
      <div>
        <label class="input-label">{{ t('payment.admin.balancePackageDisplayTags') }}</label>
        <div class="space-y-2">
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(tag, index) in form.display_tags"
              :key="`${tag}-${index}`"
              class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-3 py-1 text-sm text-slate-700 dark:bg-dark-700 dark:text-slate-200"
            >
              {{ tag }}
              <button type="button" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-100" @click="removeDisplayTag(index)">×</button>
            </span>
          </div>
          <div class="flex gap-2">
            <input
              v-model="displayTagInput"
              type="text"
              class="input"
              :placeholder="t('payment.admin.balancePackageDisplayTagsPlaceholder')"
              @keydown.enter.prevent="addDisplayTag"
            />
            <button type="button" class="btn btn-secondary" :disabled="form.display_tags.length >= 3" @click="addDisplayTag">
              {{ t('common.add') }}
            </button>
          </div>
          <p class="input-hint">{{ t('payment.admin.balancePackageDisplayTagsHint') }}</p>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.price') }} <span class="text-red-500">*</span></label>
          <input v-model.number="form.price" type="number" step="0.01" min="0.01" class="input" required />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.balancePackageCreditAmount') }} <span class="text-red-500">*</span></label>
          <input v-model.number="form.credit_amount" type="number" step="0.00000001" min="0.00000001" class="input" required />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.balancePackageProductName') }}</label>
          <input v-model="form.product_name" type="text" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.sortOrder') }}</label>
          <input v-model.number="form.sort_order" type="number" min="0" class="input" />
        </div>
      </div>
      <div class="flex items-center gap-3">
        <label class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.admin.forSale') }}</label>
        <button
          type="button"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out',
            form.for_sale ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
          ]"
          @click="form.for_sale = !form.for_sale"
        >
          <span :class="[
            'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
            form.for_sale ? 'translate-x-5' : 'translate-x-0'
          ]" />
        </button>
      </div>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" @click="emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="balance-package-form" :disabled="saving" class="btn btn-primary">{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, computed, watch, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { BalancePackage } from '@/types/payment'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'

const props = defineProps<{
  show: boolean
  balancePackage: BalancePackage | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const { t } = useI18n()
const appStore = useAppStore()
const saving = ref(false)
const displayTagInput = ref('')

const form = reactive({
  name: '',
  description: '',
  price: 0,
  credit_amount: 0,
  package_scope: 'codex' as 'codex' | 'general',
  product_name: '',
  display_tags: [] as string[],
  sort_order: 0,
  for_sale: true,
})

const packageScopeOptions = computed(() => [
  { value: 'codex', label: t('payment.balancePackages.codex') },
  { value: 'general', label: t('payment.balancePackages.general') },
])

watch(() => props.show, (visible) => {
  if (!visible) return
  if (props.balancePackage) {
    Object.assign(form, {
      ...props.balancePackage,
      display_tags: props.balancePackage.display_tags || [],
    })
  } else {
    Object.assign(form, {
      name: '',
      description: '',
      price: 0,
      credit_amount: 0,
      package_scope: 'codex',
      product_name: '',
      display_tags: [],
      sort_order: 0,
      for_sale: true,
    })
  }
  displayTagInput.value = ''
})

function addDisplayTag() {
  const tag = displayTagInput.value.trim()
  if (!tag || form.display_tags.includes(tag) || form.display_tags.length >= 3) return
  form.display_tags.push(tag)
  displayTagInput.value = ''
}

function removeDisplayTag(index: number) {
  form.display_tags.splice(index, 1)
}

async function handleSave() {
  if (!form.name.trim() || !form.price || form.price <= 0 || !form.credit_amount || form.credit_amount <= 0) {
    appStore.showError(t('common.error'))
    return
  }
  saving.value = true
  try {
    const payload = { ...form, display_tags: [...form.display_tags] }
    if (props.balancePackage) {
      await adminPaymentAPI.updateBalancePackage(props.balancePackage.id, payload)
    } else {
      await adminPaymentAPI.createBalancePackage(payload)
    }
    appStore.showSuccess(t('common.saved'))
    emit('close')
    emit('saved')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}
</script>
