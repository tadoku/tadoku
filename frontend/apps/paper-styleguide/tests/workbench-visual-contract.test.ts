import { describe, expect, it } from 'vitest'
import { styleguideStyles } from './style-sources'

function declarationsFor(selector: string): string[] {
  return [...styleguideStyles.matchAll(/([^{}]+)\{([^{}]*)\}/gu)]
    .filter((rule) =>
      rule[1]
        .split(',')
        .map((candidate) => candidate.trim().replace(/\s+/gu, ' '))
        .includes(selector),
    )
    .map((rule) => rule[2])
}

function borderValues(selector: string): string[] {
  return declarationsFor(selector).flatMap((declarations) =>
    [...declarations.matchAll(/border(?:-[a-z-]+)?\s*:\s*([^;}]+)/gu)].map(
      (match) => match[1].trim(),
    ),
  )
}

describe('component workbench visual contract', () => {
  it('uses the workbench as the only outer frame around the isolated preview', () => {
    expect(styleguideStyles).toMatch(
      /\.component-workbench \{[^}]*border: 1px solid var\(--paper-color-rule-default\);[^}]*overflow: hidden;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.example-canvas \{[^}]*border: 0;[^}]*background: var\(--paper-color-surface-raised\);/s,
    )
    expect(styleguideStyles).toMatch(
      /\.example-canvas__heading \{[^}]*border: 0;[^}]*border-block-end: 0;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.example-canvas__stage \{[^}]*border: 0;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.paper-fixture-stage \{[^}]*border: 0;/s,
    )
    expect(styleguideStyles).toMatch(
      /\.example-canvas iframe \{[^}]*max-inline-size: none;[^}]*border: 0;/s,
    )

    const framelessPreviewLayers = [
      '.example-canvas',
      '.example-canvas__heading',
      '.example-canvas__stage',
      '.paper-fixture-stage',
      '.example-canvas iframe',
    ]

    expect(borderValues('.component-workbench')).toEqual([
      '1px solid var(--paper-color-rule-default)',
    ])
    for (const selector of framelessPreviewLayers) {
      expect(borderValues(selector), selector).not.toHaveLength(0)
      expect(
        borderValues(selector).every((value) => value === '0'),
        selector,
      ).toBe(true)
    }
  })
})
