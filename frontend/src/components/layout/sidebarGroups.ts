export function isGroupActiveByPath(routePath: string, childPaths: string[]): boolean {
  return childPaths.includes(routePath)
}

export function isGroupExpandedByPath(
  routePath: string,
  childPaths: string[],
  itemPath: string,
  expandedGroups: Set<string>,
  collapsedGroups: Set<string>,
): boolean {
  if (collapsedGroups.has(itemPath)) return false
  return expandedGroups.has(itemPath) || isGroupActiveByPath(routePath, childPaths)
}

export function toggleGroupState(
  itemPath: string,
  isActive: boolean,
  expandedGroups: Set<string>,
  collapsedGroups: Set<string>,
): { expandedGroups: Set<string>; collapsedGroups: Set<string> } {
  const nextExpanded = new Set(expandedGroups)
  const nextCollapsed = new Set(collapsedGroups)
  const isManuallyExpanded = nextExpanded.has(itemPath)
  const isManuallyCollapsed = nextCollapsed.has(itemPath)

  if (isActive) {
    if (isManuallyCollapsed) {
      nextCollapsed.delete(itemPath)
      return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
    }
    nextExpanded.delete(itemPath)
    nextCollapsed.add(itemPath)
    return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
  }

  if (isManuallyExpanded) {
    nextExpanded.delete(itemPath)
    return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
  }

  if (isManuallyCollapsed) {
    nextCollapsed.delete(itemPath)
    return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
  }

  nextExpanded.add(itemPath)
  return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
}
