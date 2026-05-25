import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserBalanceModal from '../UserBalanceModal.vue'
import type { AdminUser } from '@/types'

const { updateBalance } = vi.hoisted(() => ({
  updateBalance: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      updateBalance
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const user = (overrides: Partial<AdminUser> = {}): AdminUser => ({
  id: 42,
  username: 'scoped-user',
  email: 'scoped@example.com',
  role: 'user',
  balance: 20,
  package_scope: 'codex',
  concurrency: 1,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-04-17T00:00:00Z',
  updated_at: '2026-04-17T00:00:00Z',
  notes: '',
  current_concurrency: 0,
  ...overrides
})

describe('UserBalanceModal', () => {
  beforeEach(() => {
    updateBalance.mockReset()
    updateBalance.mockResolvedValue(user({ balance: 50, package_scope: 'general' }))
  })

  it('lets admins choose package scope when depositing balance', async () => {
    const wrapper = mount(UserBalanceModal, {
      props: {
        show: true,
        user: user(),
        operation: 'add'
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div v-if="show"><slot /><slot name="footer" /></div>'
          }
        }
      }
    })

    const scopeSelect = wrapper.get('[data-testid="package-scope-select"]')
    await scopeSelect.setValue('general')
    await wrapper.get('input[type="number"]').setValue('50')
    await wrapper.get('form').trigger('submit.prevent')

    expect(updateBalance).toHaveBeenCalledWith(42, 50, 'add', '', 'general')
    expect(wrapper.text()).toContain('admin.users.packageScopeSwitchWarning')
  })
})
