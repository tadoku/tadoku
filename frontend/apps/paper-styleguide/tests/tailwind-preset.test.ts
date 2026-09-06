import { createRequire } from 'node:module'
import { readFileSync } from 'node:fs'
import postcss from 'postcss'
import tailwindcss, { type Config } from 'tailwindcss'
import { expect, it } from 'vitest'

const require = createRequire(import.meta.url)
const preset = require('../../../packages/paper-ui/tailwind-preset.cjs') as Config

it('generates ordinary Tailwind utilities from Paper tokens without removing the default scale or resetting components', async () => {
  const result = await postcss([tailwindcss({
    presets: [preset],
    content: [{ raw: 'flex gap-2 gap-inline p-4 px-8 -mt-2 md:p-6 w-5 px-5 w-full bg-paper text-ink min-h-control', extension: 'html' }],
  })]).process('@tailwind base; @tailwind utilities;', { from: undefined })
  const css = result.css
  expect(css).toMatch(/\.p-4\s*\{\s*padding: var\(--paper-space-4\)/)
  expect(css).toMatch(/\.gap-2\s*\{\s*gap: var\(--paper-space-2\)/)
  expect(css).toMatch(/\.px-8\s*\{\s*padding-left: var\(--paper-space-8\)/)
  expect(css).toContain('margin-top: calc(var(--paper-space-2) * -1)')
  expect(css).toContain('padding: var(--paper-space-6)')
  expect(css).toContain('@media (min-width: 768px)')
  expect(css).toContain('width: 1.25rem')
  expect(css).toContain('padding-left: 1.25rem')
  expect(css).toContain('width: 100%')
  expect(css).toContain('background-color: var(--paper-color-surface-paper)')
  expect(css).toContain('color: var(--paper-color-text-ink)')
  expect(css).not.toContain('border-style: solid')
  expect(css).not.toContain('font-family: ui-sans-serif')
})

it('keeps composition spacing fixed while density-sensitive utilities follow the nearest Paper boundary', async () => {
  const result = await postcss([tailwindcss({
    presets: [preset], content: [{ raw: 'p-4 gap-inline min-h-control', extension: 'html' }],
  })]).process('@tailwind utilities;', { from: undefined })
  expect(result.css).toContain('padding: var(--paper-space-4)')
  expect(result.css).toContain('gap: var(--paper-inline-gap)')
  expect(result.css).toContain('min-height: var(--paper-control-height)')
  const tokens = postcss.parse(readFileSync('../../packages/paper-ui/src/foundations/tokens.css', 'utf8'))
  const densityValues: Record<string, Record<string, string>> = {}
  tokens.walkRules(rule => {
    if (!rule.selector.includes('data-density')) return
    const values: Record<string, string> = {}
    rule.walkDecls(declaration => { values[declaration.prop] = declaration.value })
    densityValues[rule.selector.includes('comfortable') ? 'comfortable' : 'compact'] = values
  })
  expect(densityValues.comfortable['--paper-control-height']).toBe('2.75rem')
  expect(densityValues.compact['--paper-control-height']).toBe('2.25rem')
  expect(densityValues.comfortable['--paper-inline-gap']).toBe('0.625rem')
  expect(densityValues.compact['--paper-inline-gap']).toBe('0.5rem')
  expect(densityValues.comfortable).not.toHaveProperty('--paper-space-4')
  expect(densityValues.compact).not.toHaveProperty('--paper-space-4')
})
