import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AffiliateView from '../AffiliateView.vue'

const {
  getAffiliateDetail,
  transferAffiliateQuota,
  showError,
  showSuccess,
  refreshUser,
  copyToClipboard,
} = vi.hoisted(() => ({
  getAffiliateDetail: vi.fn(),
  transferAffiliateQuota: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  refreshUser: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  default: {
    getAffiliateDetail,
    transferAffiliateQuota,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/utils/format', () => ({
  formatCurrency: (value: number) => `$${value.toFixed(2)}`,
  formatDateTime: (value: string) => value,
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  const translations: Record<string, string> = {
    'affiliate.transfer.title': '返利额度提现',
    'affiliate.transfer.description': '当前返利额度暂不支持页面直接提现',
    'affiliate.transfer.button': '提现',
    'affiliate.transfer.transferring': '提现处理中...',
    'affiliate.transfer.empty': '当前没有可提现额度',
    'affiliate.transfer.contactHint': '如需微信提现，请联系管理员。',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => translations[key] ?? key,
    }),
  }
})

function buildDetail(overrides: Partial<{ aff_quota: number }> = {}) {
  return {
    user_id: 1,
    aff_code: 'AFFCODE123',
    inviter_id: null,
    aff_count: 3,
    aff_quota: 0,
    aff_history_quota: 18.5,
    invitees: [],
    ...overrides,
  }
}

describe('AffiliateView', () => {
  beforeEach(() => {
    getAffiliateDetail.mockReset()
    transferAffiliateQuota.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    refreshUser.mockReset()
    copyToClipboard.mockReset()
  })

  it('renders a disabled withdraw button plus empty and contact hints when no affiliate quota is available', async () => {
    getAffiliateDetail.mockResolvedValue(buildDetail({ aff_quota: 0 }))

    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
        },
      },
    })

    await flushPromises()

    const button = wrapper.get('button.btn.btn-primary')
    expect(button.text()).toContain('提现')
    expect(button.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('当前没有可提现额度')
    expect(wrapper.text()).toContain('如需微信提现，请联系管理员。')
  })
})
