import type { PackageScope } from '@/types'

export function packageScopeLabelKey(scope: PackageScope | null | undefined): string {
  return scope === 'general'
    ? 'payment.balancePackages.general'
    : 'payment.balancePackages.codex'
}

export function packageScopeBadgeTone(scope: PackageScope | null | undefined): string {
  return scope === 'general'
    ? 'border-violet-200/90 bg-violet-50 text-violet-700 shadow-sm shadow-violet-100/80 dark:border-violet-900/60 dark:bg-violet-950/20 dark:text-violet-300 dark:shadow-none'
    : 'border-sky-200/90 bg-sky-50 text-sky-700 shadow-sm shadow-sky-100/80 dark:border-sky-900/60 dark:bg-sky-950/20 dark:text-sky-300 dark:shadow-none'
}

export function packageScopeBadgeClass(scope: PackageScope | null | undefined): string {
  return `inline-flex items-center rounded-full border font-semibold ${packageScopeBadgeTone(scope)}`
}
