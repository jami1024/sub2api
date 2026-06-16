import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ProviderStatusTable from '../ProviderStatusTable.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const item = {
  provider: '天才程序员',
  request_count: 27,
  success_count: 15,
  failure_count: 12,
  business_limited_count: 0,
  availability: 55.56,
  error_rate: 44.44,
  p50_ms: 8_080,
  p95_ms: 188_185,
  p99_ms: 325_089,
  last_seen: '2026-06-16T06:38:45Z',
  timeline: [
    {
      bucket_start: '2026-06-16T05:12:00Z',
      request_count: 27,
      success_count: 15,
      failure_count: 12,
      availability: 55.56,
      p50_ms: 8_080,
      p95_ms: 188_185,
      p99_ms: 325_089,
    },
  ],
}

describe('ProviderStatusTable', () => {
  it('将延迟显示为更易读的秒或分钟', () => {
    const wrapper = mount(ProviderStatusTable, {
      props: {
        loading: false,
        items: [item],
      },
    })

    expect(wrapper.text()).toContain('8.08s')
    expect(wrapper.text()).toContain('3.14m')
    expect(wrapper.text()).toContain('5.42m')
    expect(wrapper.text()).not.toContain('8080ms')
  })

  it('鼠标移动到时间线点时显示详细提示信息', async () => {
    const wrapper = mount(ProviderStatusTable, {
      props: {
        loading: false,
        items: [item],
      },
    })

    await wrapper.get('[data-testid="provider-status-timeline-dot"]').trigger('mouseenter')

    expect(wrapper.text()).toContain('13:12')
    expect(wrapper.text()).toContain('27 个请求')
    expect(wrapper.text()).toContain('可用性 55.56%')
    expect(wrapper.text()).toContain('延迟: 8.08s')
    expect(wrapper.text()).toContain('OK: 15')
    expect(wrapper.text()).toContain('ERR: 12')
  })
})
