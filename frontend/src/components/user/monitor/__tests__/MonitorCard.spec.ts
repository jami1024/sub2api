import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import MonitorCard from '../MonitorCard.vue'
import type { UserMonitorView } from '@/api/channelMonitor'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (key === 'monitorCommon.history60pts') return `近 ${params?.n} 次记录`
      if (key === 'monitorCommon.realtimeWindow30m') return '近 30 分钟'
      if (key === 'monitorCommon.nextUpdateIn') return `${params?.n}s 后刷新`
      if (key === 'monitorCommon.latencyEmpty') return '-'
      return key
    },
  }),
}))

const baseItem = (overrides: Partial<UserMonitorView> = {}): UserMonitorView => ({
  id: 1,
  name: '天才程序员',
  provider: 'openai',
  group_name: '',
  primary_model: 'gpt-5.5',
  primary_status: '',
  primary_latency_ms: null,
  primary_ping_latency_ms: null,
  availability_7d: 0,
  extra_models: [],
  timeline: [
    {
      status: 'failed',
      latency_ms: null,
      ping_latency_ms: null,
      checked_at: '2026-06-05T00:00:00Z',
    },
  ],
  ...overrides,
})

const stubs = {
  ProviderIcon: { template: '<span />' },
  MonitorMetricPair: { template: '<div />' },
  MonitorAvailabilityRow: {
    name: 'MonitorAvailabilityRow',
    props: ['windowLabel', 'value', 'samplesLabel'],
    template: '<div data-testid="availability">{{ value }}</div>',
  },
  MonitorTimeline: {
    name: 'MonitorTimeline',
    props: ['buckets', 'countdownSeconds'],
    template: '<div data-testid="timeline">{{ buckets.length }}</div>',
  },
}

describe('MonitorCard', () => {
  it('无当前使用日志时不把 0% 和旧历史显示成当前状态', () => {
    const wrapper = mount(MonitorCard, {
      props: {
        item: baseItem(),
        window: '7d',
        countdownSeconds: 58,
      },
      global: { stubs },
    })

    expect(wrapper.findComponent({ name: 'MonitorAvailabilityRow' }).props('value')).toBeNull()
    expect(wrapper.findComponent({ name: 'MonitorTimeline' }).props('buckets')).toEqual([])
  })

  it('有当前使用日志时保留可用率和时间线', () => {
    const item = baseItem({
      primary_status: 'operational',
      primary_latency_ms: 120,
      availability_7d: 100,
    })
    const wrapper = mount(MonitorCard, {
      props: {
        item,
        window: '7d',
        countdownSeconds: 58,
      },
      global: { stubs },
    })

    expect(wrapper.findComponent({ name: 'MonitorAvailabilityRow' }).props('value')).toBe(100)
    expect(wrapper.findComponent({ name: 'MonitorTimeline' }).props('buckets')).toEqual(item.timeline)
  })
})
