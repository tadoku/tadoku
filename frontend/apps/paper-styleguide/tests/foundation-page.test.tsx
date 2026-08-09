import { render, screen, within } from '@testing-library/react'
import { catalogRegistry, type CatalogDocument } from 'paper-ui/catalog'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import wordmarkUrl from 'paper-ui/assets/brand/wordmark.svg?no-inline'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { DocsShell } from '../src/app/DocsShell'
import { CatalogIndex } from '../src/app/routes'
import { DocumentPage } from '../src/documentation/DocumentPage'

const foundationDocuments = catalogRegistry.documents.filter(
  (document) => document.kind === 'foundation',
)

const FOUNDATION_LEARNING_PATH = [
  'Principles',
  'Color',
  'Typography',
  'Spacing and density',
  'Layout',
  'Shape and borders',
  'Elevation',
  'Iconography',
  'Motion',
  'Brand',
] as const

function foundation(id: string): CatalogDocument {
  const document = foundationDocuments.find((candidate) => candidate.id === id)
  if (!document) throw new Error(`Missing foundation catalogue document: ${id}`)
  return document
}

const foundationSpecimens = [
  {
    id: 'foundation.principles',
    markers: [foundation('foundation.principles').guidance.whenToUse[0]],
  },
  {
    id: 'foundation.color',
    markers: [
      '--paper-color-surface-canvas',
      '--paper-color-text-ink',
      '--paper-color-action-default',
      '--paper-color-status-success',
      '--paper-color-chart-1',
    ],
  },
  {
    id: 'foundation.typography',
    markers: [
      'paper-type-display',
      'paper-type-page',
      'paper-type-section',
      'paper-type-component',
      'paper-type-label',
      'paper-type-metadata',
    ],
  },
  {
    id: 'foundation.spacing-and-density',
    markers: ['--paper-control-height', '2.75rem', '2.25rem'],
  },
  {
    id: 'foundation.layout',
    markers: [foundation('foundation.layout').guidance.content[0]],
  },
  {
    id: 'foundation.shape-and-borders',
    markers: [
      'paper-accent-rail',
      '--paper-border-static-width',
      '--paper-border-field-edge-width',
      '--paper-border-action-edge-width',
    ],
  },
  {
    id: 'foundation.elevation',
    markers: [
      'paper-elevation-flat',
      'paper-elevation-floating',
      'paper-elevation-showcase',
    ],
  },
  {
    id: 'foundation.iconography',
    markers: ['PlusIcon', 'CheckCircleIcon', 'paper-icon-compact'],
  },
  {
    id: 'foundation.motion',
    markers: [
      '--paper-motion-quick',
      '120ms',
      '--paper-motion-standard',
      '180ms',
      '--paper-motion-deliberate',
      '240ms',
    ],
  },
  {
    id: 'foundation.brand',
    markers: [],
  },
] as const

describe('foundation documents', () => {
  it('registers all ten canonical foundation routes', () => {
    expect(foundationDocuments.map((document) => document.route).sort()).toEqual([
      '/foundations/brand',
      '/foundations/color',
      '/foundations/elevation',
      '/foundations/iconography',
      '/foundations/layout',
      '/foundations/motion',
      '/foundations/principles',
      '/foundations/shape-and-borders',
      '/foundations/spacing-and-density',
      '/foundations/typography',
    ])
  })

  it.each(foundationDocuments)(
    'renders $name guidance as authored groups rather than flattening the registry object',
    (document) => {
      render(<DocumentPage document={document} />)

      const guidance = screen.getByRole('region', { name: 'Guidance' })
      expect(within(guidance).getByRole('heading', { level: 3, name: 'When to use' })).toBeVisible()
      expect(within(guidance).getByRole('heading', { level: 3, name: 'Avoid when' })).toBeVisible()
      expect(within(guidance).getByRole('heading', { level: 3, name: 'Guidance' })).toBeVisible()
      expect(within(guidance).getByRole('heading', { level: 3, name: 'Common mistakes' })).toBeVisible()

      const authoredItems = [
        ...document.guidance.whenToUse,
        ...document.guidance.whenNotToUse,
        ...document.guidance.content,
        ...document.guidance.commonMistakes,
      ]
      expect(
        within(guidance).getAllByRole('listitem').map((item) => item.textContent),
      ).toEqual(authoredItems)
    },
  )

  it.each(foundationSpecimens)(
    'renders a dedicated canonical specimen for $id',
    ({ id, markers }) => {
      const document = foundation(id)
      render(<DocumentPage document={document} />)

      const specimen = screen.getByRole('region', {
        name: `${document.name} specimen`,
      })
      expect(specimen).toHaveAttribute('data-foundation-specimen', id)
      for (const marker of markers) expect(specimen).toHaveTextContent(marker)
    },
  )

  it('uses the canonical packaged Cut Meter and wordmark in the Brand specimen', () => {
    render(<DocumentPage document={foundation('foundation.brand')} />)

    const specimen = screen.getByRole('region', { name: 'Brand specimen' })
    expect(within(specimen).getByRole('img', { name: 'Cut Meter' })).toHaveAttribute(
      'src',
      cutMeterUrl,
    )
    expect(within(specimen).getByRole('img', { name: 'Tadoku' })).toHaveAttribute(
      'src',
      wordmarkUrl,
    )
  })

  it.each(foundationDocuments)(
    'does not show the unrelated responsive component canvas for $name',
    (document) => {
      expect(document.fixtureIds).toEqual([])
      render(<DocumentPage document={document} />)

      expect(
        screen.queryByTitle('Paper responsive component preview'),
      ).not.toBeInTheDocument()
      expect(
        screen.queryByRole('link', { name: 'Preview canvas' }),
      ).not.toBeInTheDocument()
    },
  )

  it('uses the same curated foundation order on the index and in the sidebar', () => {
    const registryDocuments = catalogRegistry.documents as CatalogDocument[]
    const originalDocuments = [...registryDocuments]
    registryDocuments.reverse()

    try {
      render(
        <MemoryRouter>
          <DocsShell documents={registryDocuments}>
            <CatalogIndex />
          </DocsShell>
        </MemoryRouter>,
      )
    } finally {
      registryDocuments.splice(0, registryDocuments.length, ...originalDocuments)
    }

    const indexSection = screen
      .getByRole('heading', { name: 'Foundations' })
      .closest('section')
    if (!indexSection) throw new Error('Missing Foundations index section')
    const indexNames = Array.from(
      indexSection.querySelectorAll<HTMLElement>('.paper-type-component'),
      (name) => name.textContent,
    )

    const sidebar = screen.getByRole('navigation', { name: 'Paper catalogue' })
    const sidebarSection = within(sidebar)
      .getByRole('heading', { name: 'Foundation' })
      .closest('section')
    if (!sidebarSection) throw new Error('Missing Foundation sidebar section')
    const sidebarNames = within(sidebarSection)
      .getAllByRole('link')
      .map((link) => link.textContent)

    expect(indexNames).toEqual(sidebarNames)
    expect(indexNames).toEqual(FOUNDATION_LEARNING_PATH)
  })
})
