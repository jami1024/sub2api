import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../RedeemView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('RedeemView layout width', () => {
  it('uses a wider left-aligned desktop layout instead of the old narrow centered container', () => {
    expect(componentSource).toContain('w-full max-w-6xl space-y-6')
    expect(componentSource).toContain(
      'grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)] lg:items-start',
    )
    expect(componentSource).not.toContain('mx-auto w-full max-w-5xl space-y-6')
    expect(componentSource).not.toContain('mx-auto max-w-2xl space-y-6')
  })
})
