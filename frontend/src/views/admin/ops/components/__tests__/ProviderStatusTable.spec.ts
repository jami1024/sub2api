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
  duration_avg_ms: 42_000,
  duration_max_ms: 101_000,
  ttft_avg_ms: 2_400,
  ttft_p95_ms: 8_800,
  ttft_sample_count: 9,
  timeout_524_count: 2,
  timeout_524_avg_ms: 91_000,
  last_seen: '2026-06-16T06:38:45Z',
  fingerprint: {
    headers: {
      server: 'cloudflare',
      'cf-ray': 'abc-HKG',
      via: '1.1 proxy',
      'x-request-id': 'req_123',
    },
    last_seen: '2026-06-16T06:39:00Z',
  },
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
      duration_avg_ms: 42_000,
      ttft_avg_ms: 2_400,
      ttft_sample_count: 9,
      timeout_524_count: 2,
      timeout_524_avg_ms: 91_000,
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

  it('显示首响应、总耗时和 524 超时统计', () => {
    const wrapper = mount(ProviderStatusTable, {
      props: {
        loading: false,
        items: [item],
      },
    })

    expect(wrapper.text()).toContain('总耗时')
    expect(wrapper.text()).toContain('平均 42s')
    expect(wrapper.text()).toContain('最大 1.68m')
    expect(wrapper.text()).toContain('首响应')
    expect(wrapper.text()).toContain('平均 2.4s')
    expect(wrapper.text()).toContain('P95 8.8s')
    expect(wrapper.text()).toContain('样本 9')
    expect(wrapper.text()).toContain('524')
    expect(wrapper.text()).toContain('2 次')
    expect(wrapper.text()).toContain('平均 1.52m')
  })

  it('在供应商状态表格展示上游指纹摘要', async () => {
    const wrapper = mount(ProviderStatusTable, {
      props: {
        loading: false,
        items: [item],
      },
    })

    expect(wrapper.text()).toContain('admin.providerStatus.fingerprint')
    expect(wrapper.text()).toContain('server: cloudflare')
    expect(wrapper.text()).toContain('+2')

    await wrapper.get('[data-testid="provider-fingerprint-toggle"]').trigger('click')

    expect(wrapper.text()).toContain('via')
    expect(wrapper.text()).toContain('1.1 proxy')
    expect(wrapper.text()).toContain('x-request-id')
    expect(wrapper.text()).toContain('req_123')
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
