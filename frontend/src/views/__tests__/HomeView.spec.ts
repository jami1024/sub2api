import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'

import HomeView from '@/views/HomeView.vue'

const copyToClipboard = vi.fn().mockResolvedValue(true)
const checkAuth = vi.fn()
const fetchPublicSettings = vi.fn()

const authState = {
  isAuthenticated: false,
  isAdmin: false,
  user: { email: 'buyer@example.com' }
}

const publicSettings = ref({
  site_name: 'AigoHub',
  site_logo: '',
  site_subtitle: '',
  doc_url: '',
  home_content: ''
})

const appState = {
  publicSettingsLoaded: true,
  siteName: 'AigoHub',
  siteLogo: '',
  docUrl: '',
  cachedPublicSettings: publicSettings.value as typeof publicSettings.value | null
}

const messages: Record<string, string> = {
  'home.viewDocs': '查看文档',
  'home.switchToLight': '切换到浅色模式',
  'home.switchToDark': '切换到深色模式',
  'home.dashboard': '控制台',
  'home.login': '登录',
  'home.docs': '文档',
  'home.footer.allRightsReserved': '保留所有权利。',
  'home.landing.domainBadge': 'AIGO.RUN',
  'home.landing.title': '稳定使用，省心接入。',
  'home.landing.description': 'AigoHub 为购买用户提供更稳定的 AI 服务入口，少一点折腾，多一点省心。',
  'home.landing.primaryCta': '立即加微信',
  'home.landing.successCta': '已复制，去微信添加',
  'home.landing.wechatLabel': '微信号',
  'home.landing.copySuccess': '微信号已复制，请前往微信添加',
  'home.landing.easterEgg': '被你发现了，欢迎来聊聊。',
  'home.landing.delightPills.quickConsult': '快速咨询',
  'home.landing.whyTitle': '为什么选 AigoHub',
  'home.landing.whyItems.stableUse.title': '稳定使用',
  'home.landing.whyItems.stableUse.description': '减少中断感，日常使用更放心',
  'home.landing.whyItems.easyAccess.title': '省心接入',
  'home.landing.whyItems.easyAccess.description': '不把精力花在反复折腾上',
  'home.landing.whyItems.longTerm.title': '长期可用',
  'home.landing.whyItems.longTerm.description': '更适合希望长期稳定使用的用户',
  'home.landing.audienceTitle': '适合什么人',
  'home.landing.audienceItems.buyers': '想找稳定 AI 服务入口的人',
  'home.landing.audienceItems.simpleAccess': '不想频繁折腾接入流程的人',
  'home.landing.audienceItems.longTerm': '更在意长期体验和省心程度的人',
  'home.landing.contactTitle': '联系微信，快速咨询',
  'home.landing.contactDescription': '添加微信后可进一步了解服务内容',
  'home.landing.footerTagline': 'AigoHub，给你更放心的 AI 服务入口。'
}

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    get isAuthenticated() {
      return authState.isAuthenticated
    },
    get isAdmin() {
      return authState.isAdmin
    },
    get user() {
      return authState.user
    },
    checkAuth
  }),
  useAppStore: () => ({
    get cachedPublicSettings() {
      return appState.cachedPublicSettings
    },
    get siteName() {
      return appState.siteName
    },
    get siteLogo() {
      return appState.siteLogo
    },
    get docUrl() {
      return appState.docUrl
    },
    get publicSettingsLoaded() {
      return appState.publicSettingsLoaded
    },
    fetchPublicSettings
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copied: ref(false),
    copyToClipboard
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key
    })
  }
})

const routerLinkStub = {
  props: ['to'],
  template: `<a :href="typeof to === 'string' ? to : to.path"><slot /></a>`
}

const mountView = () =>
  mount(HomeView, {
    global: {
      stubs: {
        LocaleSwitcher: true,
        Icon: true,
        RouterLink: routerLinkStub
      }
    }
  })

describe('HomeView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(() => ({
        matches: false,
        media: '',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn()
      }))
    })
    authState.isAuthenticated = false
    authState.isAdmin = false
    publicSettings.value = {
      site_name: 'AigoHub',
      site_logo: '',
      site_subtitle: '',
      doc_url: '',
      home_content: ''
    }
    appState.cachedPublicSettings = publicSettings.value
    appState.publicSettingsLoaded = true
    appState.siteName = 'AigoHub'
    appState.siteLogo = ''
    appState.docUrl = ''
  })

  it('renders the AigoHub landing page when home_content is empty', () => {
    const wrapper = mountView()

    expect(wrapper.text()).toContain('为什么选 AigoHub')
    expect(wrapper.text()).toContain('联系微信，快速咨询')
    expect(wrapper.get('[data-testid="wechat-cta"]').text()).toContain('立即加微信')
  })

  it('copies the WeChat id when the primary CTA is clicked', async () => {
    const wrapper = mountView()

    await wrapper.get('[data-testid="wechat-cta"]').trigger('click')

    expect(copyToClipboard).toHaveBeenCalledWith('G000000000g1e', '微信号已复制，请前往微信添加')
  })

  it('切换为复制成功状态并在定时后恢复默认文案', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()

    await wrapper.get('[data-testid="wechat-cta"]').trigger('click')
    await Promise.resolve()

    expect(wrapper.get('[data-testid="wechat-cta"]').text()).toContain('已复制，去微信添加')

    await vi.advanceTimersByTimeAsync(2200)
    expect(wrapper.get('[data-testid="wechat-cta"]').text()).toContain('立即加微信')

    vi.useRealTimers()
  })

  it('连续点击 domain badge 后显示隐藏彩蛋', async () => {
    const wrapper = mountView()
    const badge = wrapper.get('[data-testid="home-domain-badge"]')

    await badge.trigger('click')
    await badge.trigger('click')
    await badge.trigger('click')
    await badge.trigger('click')
    await badge.trigger('click')

    expect(wrapper.text()).toContain('被你发现了，欢迎来聊聊。')
  })

  it('渲染 CTA 周边 delight 提示词', () => {
    const wrapper = mountView()

    expect(wrapper.get('[data-testid="hero-delight-pill-stable"]').text()).toContain('稳定使用')
    expect(wrapper.get('[data-testid="hero-delight-pill-easy"]').text()).toContain('省心接入')
    expect(wrapper.get('[data-testid="hero-delight-pill-fast"]').text()).toContain('快速咨询')
  })

  it('渲染带有生命感标记的 HeroPanel', () => {
    const wrapper = mountView()

    expect(wrapper.get('[data-testid="home-hero-panel"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="home-hero-panel-glow"]').exists()).toBe(true)
  })

  it('renders the codex package section with GPT-5.4 estimates and Claude teaser', () => {
    const wrapper = mountView()

    expect(wrapper.text()).toContain('Codex 额度包')
    expect(wrapper.text()).toContain('¥15 / $50 额度包')
    expect(wrapper.text()).toContain('¥30 / $120 额度包')
    expect(wrapper.text()).toContain('¥100 / $400 额度包')
    expect(wrapper.text()).toContain('按 GPT-5.4 约可使用 1000 万 tokens')
    expect(wrapper.text()).toContain('按输入:输出 = 4:1 估算')
    expect(wrapper.text()).toContain('Claude 额度包')
    expect(wrapper.text()).toContain('敬请期待')
  })

  it('renders three codex package cards and one claude teaser card', () => {
    const wrapper = mountView()
    const cards = wrapper.findAll('[data-testid="home-package-card"]')

    expect(cards).toHaveLength(4)
    expect(wrapper.get('[data-testid="home-package-section"]').text()).toContain('当前仅支持 Codex')
    expect(cards[3].attributes('data-package-kind')).toBe('claude-teaser')
  })

  it('renders the iframe override when home_content is an external URL', () => {
    publicSettings.value.home_content = 'https://example.com/landing'

    const wrapper = mountView()

    expect(wrapper.find('iframe').attributes('src')).toBe('https://example.com/landing')
    expect(wrapper.text()).not.toContain('稳定使用，省心接入。')
  })

  it('renders inline HTML when home_content is raw markup', () => {
    publicSettings.value.home_content = '<section id="custom-home">custom home</section>'

    const wrapper = mountView()

    expect(wrapper.html()).toContain('custom home')
    expect(wrapper.find('iframe').exists()).toBe(false)
  })

  it('keeps the dashboard entry visible for authenticated users', () => {
    authState.isAuthenticated = true

    const wrapper = mountView()

    expect(wrapper.text()).toContain('控制台')
    expect(wrapper.text()).not.toContain('登录')
  })

  it('does not flash the Sub2API fallback brand before public settings load', () => {
    appState.cachedPublicSettings = null
    appState.publicSettingsLoaded = false
    appState.siteName = 'Sub2API'

    const wrapper = mountView()

    expect(wrapper.text()).not.toContain('Sub2API')
  })
})
