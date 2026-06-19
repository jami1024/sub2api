import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProviderStatusView from '../ProviderStatusView.vue'

const mockGetProviderStatus = vi.hoisted(() => vi.fn())
const mockGetUpstreamMultiplierAccounts = vi.hoisted(() => vi.fn())
const mockGetUpstreamMultiplierSamples = vi.hoisted(() => vi.fn())
const mockMeasureUpstreamMultipliers = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getProviderStatus: mockGetProviderStatus,
    getUpstreamMultiplierAccounts: mockGetUpstreamMultiplierAccounts,
    getUpstreamMultiplierSamples: mockGetUpstreamMultiplierSamples,
    measureUpstreamMultipliers: mockMeasureUpstreamMultipliers,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: {
    template: '<div data-testid="app-layout"><slot /></div>',
  },
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

const sample = {
  start_time: '2026-06-16T00:00:00Z',
  end_time: '2026-06-16T01:00:00Z',
  bucket_seconds: 60,
  items: [
    {
      provider: 'gzw plus',
      request_count: 120,
      success_count: 118,
      failure_count: 2,
      business_limited_count: 0,
      availability: 98.3,
      error_rate: 1.7,
      p50_ms: 180,
      p95_ms: 420,
      p99_ms: 900,
      last_seen: '2026-06-16T00:59:00Z',
      timeline: [
        {
          bucket_start: '2026-06-16T00:00:00Z',
          request_count: 10,
          success_count: 10,
          failure_count: 0,
          availability: 100,
          p50_ms: 120,
          p95_ms: 240,
          p99_ms: 300,
        },
      ],
    },
  ],
  timeline: [],
}

function mountView() {
  return mount(ProviderStatusView, {
    global: {
      stubs: {
        ProviderStatusFilters: {
          props: ['modelValue'],
          emits: ['update:modelValue', 'refresh'],
          template: '<div><button data-testid="range-24h" @click="$emit(\'update:modelValue\', \'24h\')">range</button><button data-testid="refresh" @click="$emit(\'refresh\')">refresh</button></div>',
        },
        ProviderStatusSummaryCards: { template: '<div data-testid="summary" />' },
        ProviderStatusTable: { template: '<div data-testid="table" />' },
        ProviderStatusLatencyChart: { template: '<div data-testid="chart" />' },
      },
    },
  })
}

describe('ProviderStatusView', () => {
  beforeEach(() => {
    mockGetProviderStatus.mockReset().mockResolvedValue(sample)
    mockGetUpstreamMultiplierAccounts.mockReset().mockResolvedValue({
      model: 'gpt-5.4',
      accounts: [
        {
          account_id: 12,
          account_name: 'xixi',
          platform: 'openai',
          base_url: 'https://xixiapi.cc',
          key_prefix: 'sk-live-',
          supported: true,
          latest_sample: {
            id: 2,
            account_id: 12,
            account_name_snapshot: 'xixi',
            platform: 'openai',
            base_url_snapshot: 'https://xixiapi.cc',
            key_prefix_snapshot: 'sk-live-',
            model: 'gpt-5.4',
            status: 'success',
            standard_cost_delta: 0.1,
            actual_cost_delta: 0.012,
            multiplier: 0.12,
            measured_at: '2026-06-19T10:00:00Z',
            created_at: '2026-06-19T10:00:00Z',
          },
        },
      ],
    })
    mockGetUpstreamMultiplierSamples.mockReset().mockResolvedValue({
      model: 'gpt-5.4',
      samples: [
        {
          id: 2,
          account_id: 12,
          account_name_snapshot: 'xixi',
          platform: 'openai',
          base_url_snapshot: 'https://xixiapi.cc',
          key_prefix_snapshot: 'sk-live-',
          model: 'gpt-5.4',
          status: 'success',
          standard_cost_delta: 0.1,
          actual_cost_delta: 0.012,
          multiplier: 0.12,
          measured_at: '2026-06-19T10:00:00Z',
          created_at: '2026-06-19T10:00:00Z',
        },
      ],
    })
    mockMeasureUpstreamMultipliers.mockReset().mockResolvedValue({ model: 'gpt-5.4', samples: [] })
  })

  it('loads provider status and reloads when range changes', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(mockGetProviderStatus).toHaveBeenCalledWith(expect.objectContaining({ time_range: '1h' }), expect.any(Object))
    expect(wrapper.find('[data-testid="summary"]').exists()).toBe(true)

    await wrapper.get('[data-testid="range-24h"]').trigger('click')
    await flushPromises()

    expect(mockGetProviderStatus).toHaveBeenCalledWith(expect.objectContaining({ time_range: '24h' }), expect.any(Object))
  })

  it('renders inside the admin app layout', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="app-layout"]').exists()).toBe(true)
  })

  it('loads multiplier monitor inside provider status page and measures selected account', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(mockGetUpstreamMultiplierAccounts).toHaveBeenCalledWith({ model: 'gpt-5.4' })
    expect(mockGetUpstreamMultiplierSamples).toHaveBeenCalledWith({ model: 'gpt-5.4', limit: 100 })
    expect(wrapper.text()).toContain('上游倍率监测')
    expect(wrapper.text()).toContain('xixi')
    expect(wrapper.text()).toContain('0.12x')
    expect(wrapper.text()).toContain('sk-li…')
    expect(wrapper.text()).not.toContain('sk-live-')

    await wrapper.get('[data-testid="measure-upstream-12"]').trigger('click')
    await flushPromises()

    expect(mockMeasureUpstreamMultipliers).toHaveBeenCalledWith({ model: 'gpt-5.4', account_ids: [12] })
    expect(mockGetUpstreamMultiplierAccounts).toHaveBeenCalledTimes(2)
    expect(mockGetUpstreamMultiplierSamples).toHaveBeenCalledTimes(2)
  })
})
