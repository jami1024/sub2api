import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import GroupRateRecommendationsPanel from '../GroupRateRecommendationsPanel.vue'
import type { OpsGroupRateRecommendationsResponse } from '@/api/admin/ops'

const sample: OpsGroupRateRecommendationsResponse = {
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
      accounts: [
        {
          account_id: 9,
          account_name: '天才程序员',
          base_url: 'https://api.dzzzz.cf',
          key_prefix: 'sk-f642',
          schedulable: true,
          status: 'active',
          current_priority: 1,
          binding_priority: 1,
          upstream_multiplier: 0.135,
          multiplier_status: 'success',
          request_count: 70,
          request_share: 0.7,
          standard_cost: 70,
          standard_cost_share: 0.7,
          recommended_weight: 0.7,
          recommended_priority: 1,
          participates_in_advice: true,
          note: '成本较低，建议主力',
        },
      ],
    },
  ],
}

describe('GroupRateRecommendationsPanel', () => {
  it('renders package basis, group recommendation, and account weight advice', () => {
    const wrapper = mount(GroupRateRecommendationsPanel, {
      props: {
        model: 'gpt-5.4',
        data: sample,
        loading: false,
        profitMargin: 0.2,
        safetyFactor: 1.2,
        usageDays: 7,
      },
    })

    expect(wrapper.text()).toContain('分组倍率与权重建议')
    expect(wrapper.text()).toContain('专属包-进阶级')
    expect(wrapper.text()).toContain('gpt pro 高价')
    expect(wrapper.text()).toContain('1.3x')
    expect(wrapper.text()).toContain('1.08x')
    expect(wrapper.text()).toContain('天才程序员')
    expect(wrapper.text()).toContain('70%')
  })

  it('emits refresh when parameters change', async () => {
    const wrapper = mount(GroupRateRecommendationsPanel, {
      props: {
        model: 'gpt-5.4',
        data: sample,
        loading: false,
        profitMargin: 0.2,
        safetyFactor: 1.2,
        usageDays: 7,
      },
    })

    await wrapper.get('[data-testid="profit-margin-input"]').setValue('0.25')
    expect(wrapper.emitted('update:profitMargin')?.[0]).toEqual([0.25])
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })
})
