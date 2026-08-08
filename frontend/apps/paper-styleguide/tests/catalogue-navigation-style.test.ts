import { describe, expect, it } from 'vitest'
import { styleguideStyles } from './style-sources'

describe('catalogue navigation visual states', () => {
  it('keeps the inactive rail muted and reserves the action rail for active links', () => {
    expect(styleguideStyles).toMatch(
      /\.catalogue-nav__link\s*\{[^}]*border-inline-start:\s*3px solid var\(--paper-color-rule-subtle\)/su,
    )
    expect(styleguideStyles).toMatch(
      /\.catalogue-nav__link:hover,\s*\.catalogue-nav__link--active\s*\{[^}]*border-inline-start-color:\s*var\(--paper-color-action-default\)/su,
    )
  })
})
