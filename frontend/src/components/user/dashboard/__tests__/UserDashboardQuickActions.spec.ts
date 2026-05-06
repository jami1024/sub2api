import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserDashboardQuickActions from '../UserDashboardQuickActions.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock
  })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('UserDashboardQuickActions', () => {
  it('provides a quick entry to the user guide', async () => {
    const wrapper = mount(UserDashboardQuickActions, {
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('dashboard.openUserGuide')
    expect(wrapper.text()).toContain('dashboard.learnSetupSteps')

    await wrapper.get('[data-testid="dashboard-user-guide-action"]').trigger('click')

    expect(pushMock).toHaveBeenCalledWith('/guide')
  })
})
