import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import KeysView from '../KeysView.vue'

const authUser = vi.hoisted(() => ({
  id: 1,
  username: 'demo',
  email: 'demo@example.com',
  role: 'user',
  balance: 20,
  package_scope: 'codex',
  concurrency: 1,
  status: 'active',
  allowed_groups: null,
  balance_notify_enabled: true,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}))

const listKeys = vi.hoisted(() => vi.fn())
const getAvailableGroups = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())
const getPublicSettings = vi.hoisted(() => vi.fn())
const getDashboardApiKeysUsage = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: authUser,
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: () => false,
    nextStep: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
  },
  authAPI: {
    getPublicSettings: getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage: getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates: getUserGroupRates,
  },
}))

describe('KeysView package scope hints', () => {
  beforeEach(() => {
    listKeys.mockReset().mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0,
    })
    getAvailableGroups.mockReset().mockResolvedValue([])
    getUserGroupRates.mockReset().mockResolvedValue({})
    getPublicSettings.mockReset().mockResolvedValue({})
    getDashboardApiKeysUsage.mockReset().mockResolvedValue({ stats: {} })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('shows codex package scope hint in create key dialog', async () => {
    const wrapper = shallowMount(KeysView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="header" /><slot name="table" /><slot name="pagination" /></div>' },
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          DataTable: { template: '<div />' },
          Pagination: { template: '<div />' },
          ConfirmDialog: { template: '<div />' },
          EmptyState: { template: '<div />' },
          Select: { template: '<div />' },
          SearchInput: { template: '<div />' },
          Icon: { template: '<i />' },
          UseKeyModal: { template: '<div />' },
          EndpointPopover: { template: '<div />' },
          GroupBadge: { template: '<div />' },
          GroupOptionItem: { template: '<div />' },
        },
      },
    })
    await flushPromises()

    expect(wrapper.html()).toContain('keys.packageScopeHintCodex')
  })
})
