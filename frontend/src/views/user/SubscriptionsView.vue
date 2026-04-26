<template>
  <AppLayout>
    <div class="mx-auto max-w-3xl">
      <div
        data-testid="subscription-redirect-card"
        class="card flex flex-col items-center px-6 py-14 text-center sm:px-10"
      >
        <div
          class="mb-5 flex h-16 w-16 items-center justify-center rounded-full bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-300"
        >
          <Icon name="creditCard" size="xl" />
        </div>
        <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('userSubscriptions.redirectTitle') }}
        </h2>
        <p class="mt-3 max-w-xl text-sm leading-6 text-gray-500 dark:text-dark-400 sm:text-base">
          {{ t('userSubscriptions.redirectDescription') }}
        </p>
        <button
          data-testid="subscription-redirect-action"
          class="btn btn-primary mt-8 px-6 py-3 text-sm font-medium sm:text-base"
          @click="router.push('/purchase')"
        >
          {{ t('userSubscriptions.redirectAction') }}
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()
</script>
