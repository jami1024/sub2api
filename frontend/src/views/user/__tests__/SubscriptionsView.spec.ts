import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { flushPromises } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'

const routerPush = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const getMySubscriptions = vi.hoisted(() => vi.fn().mockResolvedValue([]))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: routerPush,
    }),
  }
})

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
  }),
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions,
  },
}))

describe('SubscriptionsView redirect prompt', () => {
  beforeEach(() => {
    routerPush.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    getMySubscriptions.mockReset().mockResolvedValue([])
  })

  it('shows a purchase guidance card instead of the old subscription list', async () => {
    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<div data-testid="icon-stub" />' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="subscription-redirect-card"]').text()).toContain('userSubscriptions.redirectTitle')
    expect(wrapper.text()).toContain('userSubscriptions.redirectDescription')
    expect(wrapper.text()).not.toContain('userSubscriptions.noActiveSubscriptions')
  })

  it('navigates to the purchase page when the user clicks the CTA', async () => {
    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<div data-testid="icon-stub" />' },
        },
      },
    })

    await flushPromises()
    await wrapper.get('[data-testid="subscription-redirect-action"]').trigger('click')

    expect(routerPush).toHaveBeenCalledWith('/purchase')
  })
})
