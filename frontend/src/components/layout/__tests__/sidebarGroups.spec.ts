import { describe, expect, it } from 'vitest'
import { isGroupExpandedByPath, toggleGroupState } from '../sidebarGroups'

describe('sidebar group expansion helpers', () => {
  it('lets active groups be collapsed manually and reopened on second click', () => {
    const itemPath = '/admin/orders'
    const childPaths = ['/admin/orders/dashboard', '/admin/affiliate-withdrawals']
    const routePath = '/admin/affiliate-withdrawals'

    let expanded = new Set<string>()
    let collapsed = new Set<string>()

    expect(isGroupExpandedByPath(routePath, childPaths, itemPath, expanded, collapsed)).toBe(true)

    ;({ expandedGroups: expanded, collapsedGroups: collapsed } = toggleGroupState(itemPath, true, expanded, collapsed))
    expect(isGroupExpandedByPath(routePath, childPaths, itemPath, expanded, collapsed)).toBe(false)

    ;({ expandedGroups: expanded, collapsedGroups: collapsed } = toggleGroupState(itemPath, true, expanded, collapsed))
    expect(isGroupExpandedByPath(routePath, childPaths, itemPath, expanded, collapsed)).toBe(true)
  })

  it('toggles inactive groups with the manual expanded set only', () => {
    const itemPath = '/admin/orders'

    let expanded = new Set<string>()
    let collapsed = new Set<string>()

    ;({ expandedGroups: expanded, collapsedGroups: collapsed } = toggleGroupState(itemPath, false, expanded, collapsed))
    expect(expanded.has(itemPath)).toBe(true)
    expect(collapsed.has(itemPath)).toBe(false)

    ;({ expandedGroups: expanded, collapsedGroups: collapsed } = toggleGroupState(itemPath, false, expanded, collapsed))
    expect(expanded.has(itemPath)).toBe(false)
    expect(collapsed.has(itemPath)).toBe(false)
  })
})
