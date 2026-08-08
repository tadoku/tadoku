import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const shellStyles = readFileSync(
  resolve(testDirectory, '../src/styles/shell.css'),
  'utf8',
)

describe('component workbench visual contract', () => {
  it('frames the isolated preview without adding borders inside it', () => {
    expect(shellStyles).toMatch(
      /\.component-workbench \{[^}]*border: 1px solid var\(--paper-color-rule-default\);[^}]*overflow: hidden;/s,
    )
    expect(shellStyles).toMatch(
      /\.example-canvas \{[^}]*border: 1px solid var\(--paper-color-rule-default\);[^}]*background: var\(--paper-color-surface-raised\);/s,
    )
    expect(shellStyles).toMatch(
      /\.example-canvas__heading \{[^}]*border-block-end: 0;/s,
    )
    expect(shellStyles).toMatch(
      /\.example-canvas iframe \{[^}]*max-inline-size: none;[^}]*border: 0;/s,
    )
  })
})
