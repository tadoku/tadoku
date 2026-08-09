import { describe, expect, it } from 'vitest'
import { importedStyleSources } from './style-sources'

const shellLayoutSource = importedStyleSources.find(
  ({ key }) => key === 'src/styles/shell-layout.css',
)?.source

if (shellLayoutSource === undefined) {
  throw new Error('The Paper styleguide shell layout stylesheet was not imported')
}

const shellLayoutStyles: string = shellLayoutSource

function ruleBody(selector: string) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/gu, '\\$&')
  const match = shellLayoutStyles.match(
    new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`, 'u'),
  )

  expect(match, `Expected a ${selector} rule in shell-layout.css`).not.toBeNull()
  return match?.[1] ?? ''
}

function declarationValue(rule: string, property: string) {
  const escapedProperty = property.replace(/[.*+?^${}()|[\]\\]/gu, '\\$&')
  return rule
    .match(new RegExp(`(?:^|;)\\s*${escapedProperty}\\s*:\\s*([^;]+)`, 'u'))?.[1]
    ?.trim()
    .replace(/\s+/gu, ' ')
}

describe('desktop application shell stacking contract', () => {
  it('keeps the sticky Navbar above scrolling shell regions and below overlays', () => {
    expect(declarationValue(ruleBody('.docs-navbar'), 'z-index')).toBe('30')
  })

  it('positions and sizes the sidebar below the complete bordered Navbar', () => {
    expect(
      declarationValue(ruleBody('.docs-shell'), '--docs-navbar-block-size'),
    ).toBe(
      'calc(var(--paper-control-height) + var(--paper-border-static-width))',
    )

    const sidebarRule = ruleBody('.docs-sidebar')
    expect(declarationValue(sidebarRule, 'inset-block-start')).toBe(
      'var(--docs-navbar-block-size)',
    )
    expect(declarationValue(sidebarRule, 'block-size')).toBe(
      'calc(100vh - var(--docs-navbar-block-size))',
    )
  })
})
