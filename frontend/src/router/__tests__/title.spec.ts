import { describe, expect, it } from 'vitest'
import router from '@/router'
import { resolveDocumentTitle } from '@/router/title'

describe('resolveDocumentTitle', () => {
  it('路由存在标题时，使用“路由标题 - 站点名”格式', () => {
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', () => {
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，回退默认站点名', () => {
    expect(resolveDocumentTitle('Dashboard', '')).toBe('Dashboard - Sub2API')
    expect(resolveDocumentTitle(undefined, '   ')).toBe('Sub2API')
  })

  it('站点名变更时仅影响后续路由标题计算', () => {
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })

  it('充值订阅页面使用页面标题文案，不使用“前往”动作文案', () => {
    const route = router.resolve('/purchase')

    expect(route.meta.titleKey).toBe('nav.purchaseSubscriptionMenu')
  })

  it('用户使用教程页面使用独立路由和标题文案', () => {
    const route = router.resolve('/guide')

    expect(route.name).toBe('UserGuide')
    expect(route.meta.requiresAuth).toBe(true)
    expect(route.meta.titleKey).toBe('userGuide.title')
    expect(route.meta.descriptionKey).toBe('userGuide.description')
  })
})
