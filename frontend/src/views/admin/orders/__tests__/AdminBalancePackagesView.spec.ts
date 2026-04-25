import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import AdminBalancePackagesView from '../AdminBalancePackagesView.vue'

const getBalancePackages = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

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
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getBalancePackages,
    updateBalancePackage: vi.fn(),
    deleteBalancePackage: vi.fn(),
    createBalancePackage: vi.fn(),
  },
  default: {
    getBalancePackages,
    updateBalancePackage: vi.fn(),
    deleteBalancePackage: vi.fn(),
    createBalancePackage: vi.fn(),
  },
}))

describe('AdminBalancePackagesView', () => {
  beforeEach(() => {
    getBalancePackages.mockReset().mockResolvedValue({
      data: [
        {
          id: 1,
          name: 'Codex 100',
          description: 'codex package',
          price: 100,
          credit_amount: 100,
          package_scope: 'codex',
          display_tags: ['新手推荐', '1x 倍率'],
          product_name: 'Codex 100',
          for_sale: true,
          sort_order: 1,
        },
      ],
    })
    showError.mockReset()
  })

  it('loads balance packages on mount', async () => {
    const wrapper = shallowMount(AdminBalancePackagesView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          DataTable: { template: '<div class="table-stub" />', props: ['data', 'columns', 'loading'] },
          ConfirmDialog: { template: '<div />' },
          Icon: { template: '<i />' },
          BalancePackageEditDialog: { template: '<div />' },
        },
      },
    })

    await flushPromises()

    expect(getBalancePackages).toHaveBeenCalledTimes(1)
    expect(wrapper.html()).toContain('payment.admin.createBalancePackage')
  })
})
