import { basename } from 'node:path'
import { describe, expect, it } from 'vitest'
import {
  expectedStyleImports,
  importedStyleSources,
  importedStyleSpecifiers,
  stylesheetNames,
} from './style-sources'

function selectorsIn(source: string): string[] {
  const withoutComments = source.replace(/\/\*[\s\S]*?\*\//gu, '')
  return [...withoutComments.matchAll(/([^{}]+)\{([^{}]*)\}/gu)].flatMap((rule) =>
    rule[1]
      .split(',')
      .map((selector) => selector.trim().replace(/\s+/gu, ' '))
      .filter((selector) => !selector.startsWith('@')),
  )
}

function ownersFor(selector: string): string[] {
  return importedStyleSources
    .filter(({ source }) => selectorsIn(source).includes(selector))
    .map(({ path }) => basename(path))
}

describe('styleguide stylesheet ownership', () => {
  it('imports every surface stylesheet exactly once in cascade order', () => {
    expect(importedStyleSpecifiers).toEqual(expectedStyleImports)
    expect(new Set(importedStyleSpecifiers).size).toBe(
      importedStyleSpecifiers.length,
    )
    expect(stylesheetNames).toEqual(
      ['shell.css', ...expectedStyleImports.map((path) => basename(path))].sort(),
    )
  })

  it.each([
    ['.skip-link:focus', 'base.css'],
    ['.docs-wordmark > span', 'shell-layout.css'],
    ['.canvas-controls .paper-field', 'controls.css'],
    ['.search-results', 'navigation.css'],
    ['.design-history__link', 'documents.css'],
    ['.example-canvas', 'canvas-frame.css'],
    ['.component-workbench__panel', 'workbench.css'],
    ['.preview-specimen h3', 'canvas.css'],
    ['.lifecycle-badge', 'overlays.css'],
    ['.preview-breakpoint__narrow', 'responsive.css'],
  ])('keeps %s owned once by %s', (selector, owner) => {
    expect(ownersFor(selector)).toEqual([owner])
  })

  it('contains no selectors left behind by migrated navigation and overlays', () => {
    const selectors = importedStyleSources.flatMap(({ source }) =>
      selectorsIn(source),
    )
    const obsoleteSelectors = selectors.filter((selector) =>
      /(?:\.mobile-nav-(?:backdrop|drawer)|\.catalogue-nav(?:__|\b)|\.search-backdrop)/u.test(
        selector,
      ),
    )

    expect(obsoleteSelectors).toEqual([])
  })

  it('does not shadow native controls or Paper component state recipes', () => {
    const selectors = importedStyleSources.flatMap(({ source }) =>
      selectorsIn(source),
    )
    const shadowSelectors = selectors.filter(
      (selector) =>
        /(?:^|[\s>:])(?:button|input|select)(?:\b|[\s:[.#>])/u.test(selector) ||
        /\.(?:paper-(?:button|select|input|tabs|sidebar|modal|drawer))(?:__|--|\b)/u.test(
          selector,
        ),
    )

    expect(shadowSelectors).toEqual([])
  })
})
