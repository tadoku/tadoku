import { describe, expect, it } from 'vitest'
import { importedStyleSources, styleguideStyles } from './style-sources'

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

function conditionalRuleBodies(
  source: string,
  conditional: 'container' | 'media',
): string[] {
  const bodies: string[] = []
  const openingPattern = new RegExp(`@${conditional}\\s*[^{}]+\\{`, 'gu')

  for (const opening of source.matchAll(openingPattern)) {
    const bodyStart = (opening.index ?? 0) + opening[0].length
    let depth = 1
    let cursor = bodyStart

    while (depth > 0 && cursor < source.length) {
      if (source[cursor] === '{') depth += 1
      if (source[cursor] === '}') depth -= 1
      cursor += 1
    }

    bodies.push(source.slice(bodyStart, cursor - 1))
  }

  return bodies
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

  it('lets settings fields shrink inside narrow canvas grid tracks', () => {
    const controlsStyles = importedStyleSources.find(
      ({ key }) => key === 'src/styles/controls.css',
    )?.source
    const fieldMinimum = controlsStyles?.match(
      /\.canvas-controls \.paper-field\s*\{[^}]*min-inline-size:\s*([^;]+)/su,
    )?.[1].trim()

    expect(fieldMinimum).toBe('0')
  })

  it('chooses settings columns from the canvas width, not the browser width', () => {
    const canvasStyles = importedStyleSources.find(
      ({ key }) => key === 'src/styles/canvas.css',
    )?.source
    const viewportRules = conditionalRuleBodies(canvasStyles ?? '', 'media')
    const containerRules = conditionalRuleBodies(
      canvasStyles ?? '',
      'container',
    )

    expect(canvasStyles).toBeDefined()
    expect(viewportRules).not.toEqual(
      expect.arrayContaining([expect.stringMatching(/\.canvas-controls\b/u)]),
    )
    expect(containerRules).toEqual(
      expect.arrayContaining([expect.stringMatching(/\.canvas-controls\b/u)]),
    )
  })

  it('provides a flex action row that opts fixture buttons out of form-grid stretching', () => {
    const formDeclarations = declarationsFor(
      '.paper-fixture-stage form',
    ).join('\n')
    const actionRowDeclarations = declarationsFor('.paper-fixture-row').join(
      '\n',
    )

    expect(formDeclarations).toMatch(/display:\s*grid;/u)
    expect(actionRowDeclarations).toMatch(/display:\s*flex;/u)
    expect(actionRowDeclarations).toMatch(/flex-wrap:\s*wrap;/u)
    expect(actionRowDeclarations).toMatch(/align-items:\s*center;/u)
  })
})
