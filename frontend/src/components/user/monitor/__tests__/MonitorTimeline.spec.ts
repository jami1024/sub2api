import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import MonitorTimeline from '../MonitorTimeline.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'monitorCommon.history60pts': `近 ${params?.n} 次记录`,
        'monitorCommon.noRequestLogs': '暂无请求记录',
        'monitorCommon.nextUpdateIn': `${params?.n}s 后刷新`,
        'monitorCommon.past': 'PAST',
        'monitorCommon.now': 'NOW',
        'monitorCommon.latencyEmpty': '-',
      }
      return messages[key] ?? key
    },
  }),
}))

describe('MonitorTimeline', () => {
  it('没有时间线数据时提示暂无请求记录', () => {
    const wrapper = mount(MonitorTimeline, {
      props: {
        buckets: [],
        countdownSeconds: 58,
      },
    })

    expect(wrapper.text()).toContain('暂无请求记录')
    expect(wrapper.text()).not.toContain('近 60 次记录')
  })
})
