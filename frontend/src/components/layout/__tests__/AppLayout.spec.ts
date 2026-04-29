import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppLayout.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('AppLayout desktop content offset', () => {
  it('keeps main content aligned with the compact expanded sidebar width', () => {
    expect(componentSource).toContain("sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-60'")
    expect(componentSource).not.toContain("sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64'")
  })
})
