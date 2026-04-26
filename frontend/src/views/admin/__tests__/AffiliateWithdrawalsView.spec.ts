import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AffiliateWithdrawalsView from '../AffiliateWithdrawalsView.vue'

const getAffiliateWithdrawals = vi.hoisted(() => vi.fn())
const rejectAffiliateWithdrawal = vi.hoisted(() => vi.fn())
const markAffiliateWithdrawalPaid = vi.hoisted(() => vi.fn())
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

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/api/admin/affiliate', () => ({
  adminAffiliateAPI: {
    getAffiliateWithdrawals,
    rejectAffiliateWithdrawal,
    markAffiliateWithdrawalPaid,
  },
}))

vi.mock('@/utils/format', () => ({
  formatCurrency: (value: number) => `$${value.toFixed(2)}`,
  formatDateTime: (value: string) => value,
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

describe('AffiliateWithdrawalsView', () => {
  beforeEach(() => {
    getAffiliateWithdrawals.mockReset().mockResolvedValue([
      {
        id: 1,
        user_id: 2,
        amount: 120,
        status: 'pending',
        applicant_note: '',
        admin_note: '',
        created_at: '2026-04-25T00:00:00Z',
        updated_at: '2026-04-25T00:00:00Z',
      },
    ])
    rejectAffiliateWithdrawal.mockReset().mockResolvedValue({})
    markAffiliateWithdrawalPaid.mockReset().mockResolvedValue({})
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads withdrawal requests on mount', async () => {
    const wrapper = mount(AffiliateWithdrawalsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(getAffiliateWithdrawals).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('affiliateWithdrawals.title')
    expect(wrapper.text()).toContain('¥120.00')
  })

  it('marks a pending withdrawal as paid', async () => {
    const wrapper = mount(AffiliateWithdrawalsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()
    const buttons = wrapper.findAll('button.btn.btn-primary.btn-sm')
    await buttons[0].trigger('click')
    await flushPromises()

    expect(markAffiliateWithdrawalPaid).toHaveBeenCalledWith(1)
    expect(showSuccess).toHaveBeenCalledWith('affiliateWithdrawals.markPaidSuccess')
  })
})
