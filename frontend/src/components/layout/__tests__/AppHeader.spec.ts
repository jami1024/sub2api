import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AppHeader from '@/components/layout/AppHeader.vue'

const pushMock = vi.hoisted(() => vi.fn())
const authState = vi.hoisted(() => ({
  user: {
    username: 'alice',
    email: 'alice@example.com',
    role: 'user',
    balance: 12.34,
    package_scope: 'codex' as 'codex' | 'general' | null,
    avatar_url: '',
  },
  isAdmin: false,
  isSimpleMode: false,
  logout: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({
    meta: {
      title: 'Dashboard',
      description: 'Overview',
    },
    name: 'Dashboard',
    params: {},
  }),
  RouterLink: {
    props: ['to'],
    template: '<a :href="String(to)"><slot /></a>',
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => {
        if (key === 'profile.balanceMode') return 'Balance Mode'
        if (key === 'profile.balanceModeCodex') return 'Codex'
        if (key === 'profile.balanceModeGeneral') return 'General'
        if (key === 'common.balance') return 'Balance'
        return key
      },
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    toggleMobileSidebar: vi.fn(),
    contactInfo: '',
    docUrl: '',
    cachedPublicSettings: null,
  }),
  useAuthStore: () => authState,
  useOnboardingStore: () => ({
    replay: vi.fn(),
  }),
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: [],
  }),
}))

describe('AppHeader balance mode', () => {
  beforeEach(() => {
    pushMock.mockReset()
    authState.user.package_scope = 'codex'
  })

  it('shows balance mode badge in desktop balance area', () => {
    const wrapper = mount(AppHeader, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: true,
          SubscriptionProgressMini: true,
          AnnouncementBell: true,
          Icon: true,
        },
      },
    })

    expect(wrapper.get('[data-testid="app-header-balance-mode-desktop"]').text()).toContain('Codex')
  })

  it('shows balance mode in mobile dropdown balance section', async () => {
    const wrapper = mount(AppHeader, {
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          LocaleSwitcher: true,
          SubscriptionProgressMini: true,
          AnnouncementBell: true,
          Icon: true,
        },
      },
    })

    await wrapper.find('button[aria-label="User Menu"]').trigger('click')

    expect(wrapper.get('[data-testid="app-header-balance-mode-mobile"]').text()).toContain('Balance Mode')
    expect(wrapper.get('[data-testid="app-header-balance-mode-mobile"]').text()).toContain('Codex')
  })
})
