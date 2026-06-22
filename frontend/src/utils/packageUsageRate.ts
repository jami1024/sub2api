import type { LandingUsageRate } from '@/api/publicLanding'

export function formatPackageUsageRateTitle(rate: LandingUsageRate | null | undefined): string {
  const name = rate?.group_name?.trim() || 'GPT Pro'
  return `${formatPackageGroupName(name)} 使用倍率`
}

export function formatPackageGroupName(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => {
      const lower = part.toLowerCase()
      if (lower === 'gpt') return 'GPT'
      if (lower === 'pro') return 'Pro'
      return part
    })
    .join(' ')
}
