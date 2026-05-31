import { getConfiguredTableDefaultPageSize, normalizeTablePageSize } from '@/utils/tablePreferences'

const STORAGE_KEY = 'table-page-size'
const SOURCE_STORAGE_KEY = 'table-page-size-source'
const USER_SOURCE_VALUE = 'user'

export function getPersistedPageSize(fallback = getConfiguredTableDefaultPageSize()): number {
  if (typeof window !== 'undefined') {
    try {
      const stored = window.localStorage.getItem(STORAGE_KEY)
      const source = window.localStorage.getItem(SOURCE_STORAGE_KEY)
      if (stored !== null && source !== USER_SOURCE_VALUE) {
        const parsed = Number(stored)
        if (Number.isFinite(parsed)) {
          return normalizeTablePageSize(parsed)
        }
      }
    } catch (error) {
      console.warn('Failed to read persisted page size:', error)
    }
  }
  return normalizeTablePageSize(getConfiguredTableDefaultPageSize() || fallback)
}

export function setPersistedPageSize(size: number): void {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(STORAGE_KEY, String(size))
  } catch (error) {
    console.warn('Failed to persist page size:', error)
  }
}
