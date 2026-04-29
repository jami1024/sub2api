import { describe, expect, it } from 'vitest'
import zh from '../zh'
import en from '../en'

const affiliateSettingsKeys = [
  'title',
  'description',
  'enabled',
  'enabledHint',
  'rebateRate',
  'rebateRateHint',
  'freezeHours',
  'freezeHoursDesc',
  'durationDays',
  'durationDaysDesc',
  'perInviteeCap',
  'perInviteeCapDesc',
  'customUsers.title',
  'customUsers.description',
  'customUsers.addButton',
  'customUsers.searchPlaceholder',
  'customUsers.batchButton',
  'customUsers.col.email',
  'customUsers.col.username',
  'customUsers.col.code',
  'customUsers.col.rate',
  'customUsers.col.actions',
  'customUsers.empty',
  'customUsers.customBadge',
  'customUsers.useGlobal',
  'customUsers.totalLabel',
  'customUsers.resetTitle',
  'customUsers.resetMessage',
  'modal.addTitle',
  'modal.editTitle',
  'modal.userLabel',
  'modal.changeUser',
  'modal.userPlaceholder',
  'modal.codeLabel',
  'modal.codePlaceholder',
  'modal.codeHint',
  'modal.rateLabel',
  'modal.ratePlaceholder',
  'modal.rateHint',
  'modal.errorEmpty',
  'modal.errorBadRate',
  'batchModal.title',
  'batchModal.hint',
  'batchModal.placeholder',
  'batchModal.clearHint',
]

function readPath(obj: unknown, path: string): unknown {
  return path.split('.').reduce<unknown>((current, key) => {
    if (!current || typeof current !== 'object') return undefined
    return (current as Record<string, unknown>)[key]
  }, obj)
}

describe('admin affiliate settings i18n', () => {
  it.each([
    ['zh', zh],
    ['en', en],
  ])('defines all affiliate feature setting labels for %s', (_locale, messages) => {
    for (const key of affiliateSettingsKeys) {
      const value = readPath(messages, `admin.settings.features.affiliate.${key}`)
      expect(value, key).toEqual(expect.any(String))
      expect(value).not.toBe('')
      expect(value).not.toContain('admin.settings.features.affiliate')
    }
  })
})
