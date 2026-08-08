import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const shellStyles = readFileSync(
  resolve(testDirectory, '../src/styles/shell.css'),
  'utf8',
)

describe('catalogue navigation visual states', () => {
  it('keeps the inactive rail muted and reserves the action rail for active links', () => {
    expect(shellStyles).toMatch(
      /\.catalogue-nav__link\s*\{[^}]*border-inline-start:\s*3px solid var\(--paper-color-rule-subtle\)/su,
    )
    expect(shellStyles).toMatch(
      /\.catalogue-nav__link:hover,\s*\.catalogue-nav__link--active\s*\{[^}]*border-inline-start-color:\s*var\(--paper-color-action-default\)/su,
    )
  })
})
