import { describe, expect, it } from 'vitest'
import { styleguideStyles } from './style-sources'

describe('isolated preview spacing contract', () => {
  it('keeps only the clearance needed for fixture focus rings and shadows', () => {
    expect(styleguideStyles).toMatch(
      /\.paper-fixture-stage \{[^}]*padding: 0\.5rem;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.example-canvas__stage \{[^}]*padding: 0\.5rem;/s,
    )
  })

  it('uses an even tighter outer gutter on narrow screens', () => {
    expect(styleguideStyles).toMatch(
      /@media \(max-width: 48rem\) \{[^}]*\.example-canvas__stage \{[^}]*padding: 0\.25rem;/s,
    )
  })
})
