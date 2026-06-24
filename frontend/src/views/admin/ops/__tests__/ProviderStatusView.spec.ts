import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProviderStatusView from '../ProviderStatusView.vue'

const mockGetProviderStatus = vi.hoisted(() => vi.fn())
const mockGetUpstreamMultiplierAccounts = vi.hoisted(() => vi.fn())
const mockGetUpstreamMultiplierSamples = vi.hoisted(() => vi.fn())
const mockMeasureUpstreamMultipliers = vi.hoisted(() => vi.fn())
const mockApplyUpstreamMultiplier = vi.hoisted(() => vi.fn())
const mockGetGroupRateRecommendations = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getProviderStatus: mockGetProviderStatus,
    getUpstreamMultiplierAccounts: mockGetUpstreamMultiplierAccounts,
    getUpstreamMultiplierSamples: mockGetUpstreamMultiplierSamples,
    measureUpstreamMultipliers: mockMeasureUpstreamMultipliers,
    applyUpstreamMultiplier: mockApplyUpstreamMultiplier,
    getGroupRateRecommendations: mockGetGroupRateRecommendations,
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
          account_rate_multiplier: 1,
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
    mockApplyUpstreamMultiplier.mockReset().mockResolvedValue({ model: 'gpt-5.4', account_id: 12, rate_multiplier: 0.12 })
    mockGetGroupRateRecommendations.mockReset().mockResolvedValue({
      params: { model: 'gpt-5.4', package_scope: 'codex', profit_margin: 0.2, safety_factor: 1.2, usage_days: 7 },
      package_basis: { package_id: 2, name: '专属包-进阶级', price: 100, credit_amount: 400, package_scope: 'codex', revenue_per_credit: 0.25 },
      groups: [
        {
          group_id: 8,
          group_name: 'gpt pro 高价',
          current_group_multiplier: 1.3,
          package_scope: 'codex',
          schedulable_account_count: 2,
          actual_blended_multiplier: 0.148,
          recommended_blended_multiplier: 0.1485,
          worst_case_multiplier: 0.18,
          minimum_group_multiplier: 0.891,
          safe_group_multiplier: 1.08,
          status: 'safe',
          accounts: [],
        },
      ],
    })
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
    expect(mockGetGroupRateRecommendations).toHaveBeenCalledWith(expect.objectContaining({ model: 'gpt-5.4', profit_margin: 0.2, safety_factor: 1.2, usage_days: 7 }))
    expect(wrapper.text()).toContain('上游倍率监测')
    expect(wrapper.text()).toContain('分组倍率与权重建议')
    expect(wrapper.text()).toContain('gpt pro 高价')
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

  it('applies the latest successful multiplier to the upstream account', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('当前账号倍率')
    expect(wrapper.text()).toContain('1x')

    await wrapper.get('[data-testid="apply-upstream-12"]').trigger('click')
    await flushPromises()

    expect(mockApplyUpstreamMultiplier).toHaveBeenCalledWith({ model: 'gpt-5.4', account_id: 12 })
    expect(mockGetUpstreamMultiplierAccounts).toHaveBeenCalledTimes(2)
    expect(mockGetUpstreamMultiplierSamples).toHaveBeenCalledTimes(2)
  })
})
