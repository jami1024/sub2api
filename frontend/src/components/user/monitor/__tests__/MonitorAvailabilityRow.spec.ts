import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import MonitorAvailabilityRow from '../MonitorAvailabilityRow.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => (key === 'monitorCommon.latencyEmpty' ? '-' : key),
  }),
}))

describe('MonitorAvailabilityRow', () => {
  it('无可用率数据时只显示空值，不显示百分号', () => {
    const wrapper = mount(MonitorAvailabilityRow, {
      props: {
        windowLabel: '可用性 · 7 天',
        value: null,
      },
    })

    expect(wrapper.text()).toContain('-')
    expect(wrapper.text()).not.toContain('%')
  })

  it('有可用率数据时显示百分比', () => {
    const wrapper = mount(MonitorAvailabilityRow, {
      props: {
        windowLabel: '可用性 · 7 天',
        value: 100,
      },
    })

    expect(wrapper.text()).toContain('100.00')
    expect(wrapper.text()).toContain('%')
  })
})
