import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar brand initialization', () => {
  it('does not render the fallback Sub2API brand before public settings are loaded', () => {
    expect(componentSource).toContain('v-if="settingsLoaded" class="sidebar-brand-title')
    expect(componentSource).toContain('<VersionBadge v-if="settingsLoaded"')
  })
})


describe('AppSidebar user navigation', () => {
  it('does not expose the legacy my subscriptions entry in the sidebar list', () => {
    expect(componentSource).not.toContain("path: '/subscriptions'")
  })
})

describe('AppSidebar purchase menu label', () => {
  it('uses a dedicated sidebar copy key for the purchase entry', () => {
    expect(componentSource).toContain("t('nav.purchaseSubscriptionMenu')")
  })
})

describe('AppSidebar desktop width', () => {
  it('uses the compact 15rem expanded width instead of the old 16rem width', () => {
    expect(componentSource).toContain("sidebarCollapsed ? 'w-[72px]' : 'w-60'")
    expect(componentSource).not.toContain("sidebarCollapsed ? 'w-[72px]' : 'w-64'")
  })
})
