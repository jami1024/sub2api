import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserGuideView from '../UserGuideView.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('../docs/claude-code-install.zh.md?raw', () => ({
  default: '# Claude CLI / Claude Code 安装与 API 登录指南\n\n```bash\nclaude doctor\n```\n'
}))

describe('UserGuideView', () => {
  it('renders the client usage guide with the essential setup steps', () => {
    const wrapper = mount(UserGuideView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === `string` ? to : to.path"><slot /></a>'
          },
          Icon: true
        }
      }
    })

    expect(wrapper.get('[data-testid="user-guide-view"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="user-guide-layout"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="user-guide-quick-nav"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="user-guide-content-grid"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="user-guide-claude-install-card"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="user-guide-full-doc-nav"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="user-guide-full-doc-details"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('userGuide.title')
    expect(wrapper.text()).toContain('userGuide.steps.fundAccount.title')
    expect(wrapper.text()).toContain('userGuide.steps.createKey.title')
    expect(wrapper.text()).toContain('userGuide.steps.chooseGroup.title')
    expect(wrapper.text()).toContain('userGuide.steps.configureClient.title')
    expect(wrapper.text()).toContain('userGuide.steps.checkUsage.title')
    expect(wrapper.text()).not.toContain('userGuide.claudeInstall.title')
    expect(wrapper.text()).toContain('userGuide.imageGeneration.title')
    expect(wrapper.text()).toContain('gpt-image-2')
    expect(wrapper.text()).toContain('/v1/images/generations')
    expect(wrapper.text()).toContain('image.png')
    expect(wrapper.text()).toContain('urllib.request')
    expect(wrapper.text()).toContain('json.dumps(payload)')
    expect(wrapper.text()).not.toContain('python3 -c')
    expect(wrapper.text()).not.toContain("python3 - <<'PY'")
    expect(wrapper.text()).toContain('jq -r')
    expect(wrapper.html()).toContain('Claude CLI / Claude Code 安装与 API 登录指南')
    expect(wrapper.html()).toContain('claude doctor')
    expect(wrapper.html()).toContain('href="#full-doc-section-0"')
    expect(wrapper.text()).not.toContain('Gemini CLI')
    expect(wrapper.text()).toContain('userGuide.faq.keyNotWorking.title')
  })
})
