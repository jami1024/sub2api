import { apiClient } from './client'

export interface LandingUsageRate {
  group_id: number
  group_name: string
  rate_multiplier: number
  rate_label: string
  value_lift_percent: number
  value_lift_label: string
}

export interface LandingBalancePackage {
  id: number
  name: string
  description: string
  price: number
  credit_amount: number
  package_scope: 'codex' | 'general'
  product_name: string
  display_tags: string[]
  sort_order: number
  arrival_multiplier: number
  arrival_discount: number
  arrival_discount_label: string
  effective_credit_amount: number
  effective_discount: number
  effective_discount_label: string
}

export interface LandingPackageShowcase {
  packages: LandingBalancePackage[]
  usage_rates: LandingUsageRate[]
  primary_usage_rate?: LandingUsageRate | null
}

export function getLandingPackageShowcase() {
  return apiClient.get<LandingPackageShowcase>('/payment/public/landing-packages')
}
