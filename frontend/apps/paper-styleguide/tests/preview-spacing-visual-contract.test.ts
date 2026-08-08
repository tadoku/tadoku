import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const shellStyles = readFileSync(
  resolve(testDirectory, '../src/styles/shell.css'),
  'utf8',
)

describe('isolated preview spacing contract', () => {
  it('keeps only the clearance needed for fixture focus rings and shadows', () => {
    expect(shellStyles).toMatch(
      /\.paper-fixture-stage \{[^}]*padding: 0\.5rem;/s,
    )
    expect(shellStyles).toMatch(
      /\.example-canvas__stage \{[^}]*padding: 0\.5rem;/s,
    )
  })

  it('uses an even tighter outer gutter on narrow screens', () => {
    expect(shellStyles).toMatch(
      /@media \(max-width: 48rem\) \{[^}]*\.example-canvas__stage \{[^}]*padding: 0\.25rem;/s,
    )
  })
})
