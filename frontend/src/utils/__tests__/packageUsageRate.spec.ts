import { describe, expect, it } from 'vitest'
import { formatPackageGroupName, formatPackageUsageRateTitle } from '@/utils/packageUsageRate'

describe('packageUsageRate utils', () => {
  it('formats gpt pro group name for customer-facing labels', () => {
    expect(formatPackageGroupName('gpt pro')).toBe('GPT Pro')
    expect(formatPackageUsageRateTitle({
      group_id: 2,
      group_name: 'gpt pro',
      rate_multiplier: 0.8,
      rate_label: 'gpt pro 使用倍率 0.8x',
      value_lift_percent: 25,
      value_lift_label: '同样余额可多用约 25%',
    })).toBe('GPT Pro 使用倍率')
  })

  it('falls back to GPT Pro when no live group is available', () => {
    expect(formatPackageUsageRateTitle(null)).toBe('GPT Pro 使用倍率')
  })
})
