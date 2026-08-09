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
  it('leaves the sole outer frame to Paper Surface', () => {
    expect(styleguideStyles).toMatch(
      /\.component-workbench \{[^}]*overflow: clip;/s,
    )
    expect(declarationsFor('.component-workbench').join('\n')).not.toMatch(
      /(?:border|background)\s*:/u,
    )

    const framelessPreviewLayers = [
      '.example-canvas',
      '.canvas-controls',
      '.example-canvas__stage',
      '.paper-fixture-stage',
      '.example-canvas iframe',
    ]

    for (const selector of framelessPreviewLayers) {
      expect(
        borderValues(selector).every((value) => value === '0'),
        selector,
      ).toBe(true)
      expect(declarationsFor(selector).join('\n'), selector).not.toMatch(
        /background\s*:/u,
      )
    }
  })

  it('keeps settings responsive and the stage as the preview overflow boundary', () => {
    expect(styleguideStyles).toMatch(
      /\.canvas-controls \{[^}]*display: grid;/su,
    )
    expect(styleguideStyles).toMatch(/@media \(min-width: 80rem\)/u)
    expect(styleguideStyles).toMatch(/@container \(min-width: 54rem\)/u)
    expect(declarationsFor('.example-canvas__stage').join('\n')).toMatch(
      /overflow: auto;/u,
    )
    expect(declarationsFor('.component-workbench__panel').join('\n')).not.toMatch(
      /min-block-size/u,
    )
    expect(declarationsFor('.example-canvas__stage').join('\n')).not.toMatch(
      /min-block-size/u,
    )
  })
})
