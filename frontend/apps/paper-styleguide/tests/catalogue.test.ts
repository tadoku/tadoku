import { describe, expect, it } from 'vitest'
import type { CatalogDocument, CatalogRegistry } from 'paper-ui/catalog'
import {
  buildNavigationGroups,
  resolveCatalogRoute,
  searchCatalog,
} from '../src/app/catalogue'

function document(
  overrides: Partial<CatalogDocument> & Pick<CatalogDocument, 'id' | 'route' | 'name'>,
): CatalogDocument {
  return {
    kind: 'foundation',
    category: 'foundations',
    aliases: [],
    summary: `${overrides.name} guidance`,
    keywords: [],
    lifecycle: 'Experimental',
    owner: 'Paper team',
    reviewDate: '2026-08-08',
    sourcePath: 'src/catalog.ts',
    packageVersion: '0.1.0',
    guidance: {},
    accessibility: [],
    api: {},
    fixtureIds: [],
    dependencies: [],
    migration: {},
    changelog: [],
    behaviorTestIds: [],
    ...overrides,
  } as CatalogDocument
}

const documents = [
  document({
    id: 'foundation-color',
    route: '/foundations/color',
    name: 'Color',
    aliases: ['/colors'],
    keywords: ['theme', 'contrast'],
  }),
  document({
    id: 'foundation-type',
    route: '/foundations/typography',
    name: 'Typography',
    keywords: ['fonts'],
  }),
]

const registry = {
  documents,
  fixtures: [],
  redirects: [{ from: '/legacy-colors', to: '/foundations/color' }],
} as unknown as CatalogRegistry

describe('catalogue route resolver', () => {
  it('resolves canonical documents and normalizes trailing slashes', () => {
    expect(resolveCatalogRoute('/foundations/color/', registry)).toEqual({
      kind: 'document',
      document: documents[0],
    })
  })

  it('resolves the redirect manifest and document aliases', () => {
    expect(resolveCatalogRoute('/legacy-colors', registry)).toEqual({
      kind: 'redirect',
      to: '/foundations/color',
    })
    expect(resolveCatalogRoute('/colors', registry)).toEqual({
      kind: 'redirect',
      to: '/foundations/color',
    })
  })

  it('reports routes that do not exist', () => {
    expect(resolveCatalogRoute('/components/missing', registry)).toEqual({
      kind: 'not-found',
    })
  })
})

describe('registry-derived navigation and search', () => {
  it('builds navigation without a parallel route list', () => {
    expect(buildNavigationGroups(documents)).toEqual([
      {
        id: 'foundation',
        label: 'Foundation',
        documents,
      },
    ])
  })

  it('searches aliases, keywords, and summaries', () => {
    expect(searchCatalog(documents, 'contrast')).toEqual([documents[0]])
    expect(searchCatalog(documents, 'fonts')).toEqual([documents[1]])
    expect(searchCatalog(documents, 'missing')).toEqual([])
  })
})
