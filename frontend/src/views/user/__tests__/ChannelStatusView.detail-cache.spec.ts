import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ChannelStatusView from '../ChannelStatusView.vue'
import MonitorDetailDialog from '@/components/user/MonitorDetailDialog.vue'

const { listChannelMonitors, fetchDetail, showError } = vi.hoisted(() => ({
  listChannelMonitors: vi.fn(),
  fetchDetail: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/channelMonitor', () => ({
  list: listChannelMonitors,
  status: fetchDetail,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    cachedPublicSettings: { channel_monitor_enabled: true },
  }),
}))

vi.mock('@/composables/useAutoRefresh', () => ({
  useAutoRefresh: () => ({
    countdown: { value: 30 },
    enabled: { value: false },
    start: vi.fn(),
    stop: vi.fn(),
    setEnabled: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

const monitor = {
  id: 9,
  name: 'OpenAI',
  provider: 'openai',
  group_name: 'default',
  primary_model: 'gpt-5.4',
  primary_status: 'operational',
  primary_latency_ms: 120,
  primary_ping_latency_ms: null,
  availability_7d: 100,
  extra_models: [],
  timeline: [],
}

const detail = {
  id: 9,
  name: 'OpenAI',
  provider: 'openai',
  group_name: 'default',
  models: [
    {
      model: 'gpt-5.4',
      latest_status: 'operational',
      latest_latency_ms: 120,
      availability_7d: 100,
      availability_15d: 100,
      availability_30d: 100,
      avg_latency_7d_ms: 120,
    },
  ],
}

function mountView() {
  return mount(ChannelStatusView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        MonitorHero: true,
        MonitorCardGrid: {
          props: ['items'],
          emits: ['cardClick'],
          template: '<button data-testid="open-detail" @click="$emit(\'cardClick\', items[0])">open</button>',
        },
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
      },
    },
  })
}

describe('ChannelStatusView detail cache', () => {
  beforeEach(() => {
    listChannelMonitors.mockReset().mockResolvedValue({ items: [monitor] })
    fetchDetail.mockReset().mockResolvedValue(detail)
    showError.mockReset()
  })

  it('reuses cached detail when opening the same monitor repeatedly', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="open-detail"]').trigger('click')
    await flushPromises()
    wrapper.findComponent(MonitorDetailDialog).vm.$emit('close')
    await flushPromises()

    await wrapper.get('[data-testid="open-detail"]').trigger('click')
    await flushPromises()

    expect(fetchDetail).toHaveBeenCalledTimes(1)
  })
})
