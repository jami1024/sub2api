import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ProviderStatusLatencyChart from '../ProviderStatusLatencyChart.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => {
        if (key === 'admin.providerStatus.peak') return '峰值'
        return key
      },
    }),
  }
})

const points = [
  {
    bucket_start: '2026-06-16T00:00:00Z',
    request_count: 10,
    success_count: 10,
    failure_count: 0,
    availability: 100,
    p50_ms: 120,
    p95_ms: 980,
    p99_ms: 1_200,
  },
  {
    bucket_start: '2026-06-16T00:05:00Z',
    request_count: 12,
    success_count: 11,
    failure_count: 1,
    availability: 91.7,
    p50_ms: 240,
    p95_ms: 1_500,
    p99_ms: 65_000,
  },
]

describe('ProviderStatusLatencyChart', () => {
  it('显示 P50/P95/P99 当前值和峰值数字', () => {
    const wrapper = mount(ProviderStatusLatencyChart, {
      props: {
        loading: false,
        points,
      },
    })

    expect(wrapper.text()).toContain('P50')
    expect(wrapper.text()).toContain('240ms')
    expect(wrapper.text()).toContain('峰值 240ms')
    expect(wrapper.text()).toContain('P95')
    expect(wrapper.text()).toContain('1.5s')
    expect(wrapper.text()).toContain('峰值 1.5s')
    expect(wrapper.text()).toContain('P99')
    expect(wrapper.text()).toContain('1.1m')
    expect(wrapper.text()).toContain('峰值 1.1m')
  })

  it('显示纵轴刻度数字，帮助判断趋势数值范围', () => {
    const wrapper = mount(ProviderStatusLatencyChart, {
      props: {
        loading: false,
        points,
      },
    })

    expect(wrapper.text()).toContain('0ms')
    expect(wrapper.text()).toContain('32.5s')
    expect(wrapper.text()).toContain('1.1m')
  })

  it('没有数据时保持空态，不显示指标数字', () => {
    const wrapper = mount(ProviderStatusLatencyChart, {
      props: {
        loading: false,
        points: [],
      },
    })

    expect(wrapper.text()).toContain('admin.providerStatus.empty')
    expect(wrapper.text()).not.toContain('峰值')
  })
})
