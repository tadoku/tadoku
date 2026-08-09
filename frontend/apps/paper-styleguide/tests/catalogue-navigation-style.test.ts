import { describe, expect, it } from 'vitest'
import { styleguideStyles } from './style-sources'

describe('catalogue navigation visual states', () => {
  it('leaves inactive, hover, and current rails to the Paper Sidebar recipe', () => {
    expect(styleguideStyles).not.toMatch(/\.catalogue-nav__link\s*\{/u)
    expect(styleguideStyles).not.toContain('.catalogue-nav__link--active')
  })
})
