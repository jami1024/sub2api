import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AffiliateView from '../AffiliateView.vue'

const {
  getAffiliateDetail,
  transferAffiliateQuota,
  createAffiliateWithdrawalRequest,
  getAffiliateWithdrawalRequests,
  showError,
  showSuccess,
  refreshUser,
  copyToClipboard,
} = vi.hoisted(() => ({
  getAffiliateDetail: vi.fn(),
  transferAffiliateQuota: vi.fn(),
  createAffiliateWithdrawalRequest: vi.fn(),
  getAffiliateWithdrawalRequests: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  refreshUser: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  default: {
    getAffiliateDetail,
    transferAffiliateQuota,
    createAffiliateWithdrawalRequest,
    getAffiliateWithdrawalRequests,
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
    'affiliate.transfer.description': '满 100 元后可申请人工提现',
    'affiliate.transfer.button': '提现',
    'affiliate.transfer.requesting': '申请中...',
    'affiliate.transfer.empty': '当前没有可提现额度',
    'affiliate.transfer.thresholdHint': '可提现返利满 100 元后才可申请',
    'affiliate.transfer.debtHint': '当前存在返利负债，请先等待后续返利抵扣完成。',
    'affiliate.transfer.manualHint': '提交申请后，管理员会线下手动打款。',
    'affiliate.transfer.requestSuccess': '提现申请已提交',
    'affiliate.transfer.requestFailed': '提现申请提交失败',
    'affiliate.transfer.dialogTitle': '申请提现',
    'affiliate.transfer.dialogDescription': '请确认本次申请金额和备注信息。',
    'affiliate.transfer.requestAmount': '申请金额',
    'affiliate.transfer.requestNote': '备注',
    'affiliate.transfer.confirm': '确认申请',
    'affiliate.withdrawals.title': '提现申请记录',
    'affiliate.withdrawals.empty': '暂无提现申请记录',
    'affiliate.stats.debtQuota': '返利负债',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => translations[key] ?? key,
    }),
  }
})

function buildDetail(overrides: Partial<{ aff_quota: number; debt_quota: number }> = {}) {
  return {
    user_id: 1,
    aff_code: 'AFFCODE123',
    inviter_id: null,
    aff_count: 3,
    aff_quota: 0,
    aff_history_quota: 18.5,
    debt_quota: 0,
    invitees: [],
    ...overrides,
  }
}

describe('AffiliateView', () => {
  beforeEach(() => {
    getAffiliateDetail.mockReset()
    transferAffiliateQuota.mockReset()
    createAffiliateWithdrawalRequest.mockReset().mockResolvedValue({
      id: 1,
      user_id: 1,
      amount: 120,
      status: 'pending',
      applicant_note: '',
      admin_note: '',
      created_at: '2026-04-25T00:00:00Z',
      updated_at: '2026-04-25T00:00:00Z',
    })
    getAffiliateWithdrawalRequests.mockReset().mockResolvedValue([])
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
          BaseDialog: { template: '<div class="modal-content"><slot /><slot name="footer" /></div>', props: ['show', 'title', 'width'] },
        },
      },
    })

    await flushPromises()

    const button = wrapper.get('[data-testid="affiliate-withdraw-open"]')
    expect(button.text()).toContain('提现')
    expect(button.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('当前没有可提现额度')
    expect(wrapper.text()).toContain('提交申请后，管理员会线下手动打款。')
  })

  it('submits a withdrawal request when available quota reaches the threshold and no debt exists', async () => {
    getAffiliateDetail
      .mockResolvedValueOnce(buildDetail({ aff_quota: 120, debt_quota: 0 }))
      .mockResolvedValueOnce(buildDetail({ aff_quota: 0, debt_quota: 0 }))

    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          BaseDialog: { template: '<div v-if="show" class="modal-content"><slot /><slot name="footer" /></div>', props: ['show', 'title', 'width'] },
        },
      },
    })

    await flushPromises()
    const button = wrapper.get('[data-testid="affiliate-withdraw-open"]')
    expect(button.attributes('disabled')).toBeUndefined()

    await button.trigger('click')
    await flushPromises()
    await wrapper.get('.modal-content .btn.btn-primary').trigger('click')
    await flushPromises()

    expect(createAffiliateWithdrawalRequest).toHaveBeenCalledWith({ amount: 120, applicant_note: undefined })
    expect(showSuccess).toHaveBeenCalledWith('提现申请已提交')
  })

  it('keeps the withdraw button disabled when affiliate debt exists', async () => {
    getAffiliateDetail.mockResolvedValue(buildDetail({ aff_quota: 120, debt_quota: 5 }))

    const wrapper = mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          BaseDialog: { template: '<div v-if="show" class="modal-content"><slot /><slot name="footer" /></div>', props: ['show', 'title', 'width'] },
        },
      },
    })

    await flushPromises()

    const button = wrapper.get('[data-testid="affiliate-withdraw-open"]')
    expect(button.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('当前存在返利负债，请先等待后续返利抵扣完成。')
  })
})
