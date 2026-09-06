import { readFileSync } from 'node:fs'
import { catalogRegistry } from '../src/catalog'
import { expect, it } from 'vitest'

it('publishes the exact TSX file used for every live example', () => {
  for (const fixture of catalogRegistry.fixtures) {
    const source = readFileSync(
      `src/catalog/examples/${fixture.id}.tsx`,
      'utf8',
    )
    expect(fixture.code, fixture.id).toBe(source)
    expect(source, fixture.id).toContain('export default function')
    expect(source, fixture.id).not.toMatch(/from ["']\.\.?\//)
  }
})
