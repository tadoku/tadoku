import { basename } from 'node:path'
import { describe, expect, it } from 'vitest'
import {
  expectedStyleImports,
  importedStyleSources,
  importedStyleSpecifiers,
  stylesheetNames,
} from './style-sources'

function selectorsIn(source: string): string[] {
  return [...source.matchAll(/([^{}]+)\{([^{}]*)\}/gu)].flatMap((rule) =>
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
    ['body', 'base.css'],
    ['.paper-wordmark > span', 'shell-layout.css'],
    ['.canvas-controls label', 'controls.css'],
    [
      '.catalogue-nav__group + .catalogue-nav__group',
      'navigation.css',
    ],
    ['.design-history__link', 'documents.css'],
    ['.example-canvas', 'canvas-frame.css'],
    ['.component-workbench__panel', 'workbench.css'],
    ['.preview-specimen h3', 'canvas.css'],
    ['.search-dialog__heading h2', 'overlays.css'],
    ['.preview-breakpoint__narrow', 'responsive.css'],
  ])('keeps %s owned once by %s', (selector, owner) => {
    expect(ownersFor(selector)).toEqual([owner])
  })
})
