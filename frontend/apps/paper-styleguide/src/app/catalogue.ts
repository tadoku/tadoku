import type { CatalogDocument, CatalogRegistry } from 'paper-ui/catalog'

export type ResolvedCatalogRoute =
  | { kind: 'document'; document: CatalogDocument }
  | { kind: 'redirect'; to: string }
  | { kind: 'not-found' }

export interface NavigationGroup {
  id: string
  label: string
  documents: CatalogDocument[]
}

function normalizePath(path: string): string {
  const withoutQuery = path.split(/[?#]/u, 1)[0] || '/'
  const withLeadingSlash = withoutQuery.startsWith('/')
    ? withoutQuery
    : `/${withoutQuery}`

  return withLeadingSlash.length > 1
    ? withLeadingSlash.replace(/\/+$/u, '')
    : withLeadingSlash
}

function redirectPaths(redirect: unknown): { from?: string; to?: string } {
  if (!redirect || typeof redirect !== 'object') return {}

  const value = redirect as Record<string, unknown>
  const from = value.from ?? value.source ?? value.path
  const to = value.to ?? value.target ?? value.route

  return {
    from: typeof from === 'string' ? from : undefined,
    to: typeof to === 'string' ? to : undefined,
  }
}

export function resolveCatalogRoute(
  pathname: string,
  registry: CatalogRegistry,
): ResolvedCatalogRoute {
  const path = normalizePath(pathname)
  const document = registry.documents.find(
    (candidate) => normalizePath(candidate.route) === path,
  )

  if (document) return { kind: 'document', document }

  for (const redirect of registry.redirects) {
    const { from, to } = redirectPaths(redirect)
    if (from && to && normalizePath(from) === path) {
      return { kind: 'redirect', to: normalizePath(to) }
    }
  }

  // Aliases are a Phase 1 convenience. The registry redirect manifest remains
  // the canonical cutover contract when both are present.
  const aliasOwner = registry.documents.find((candidate) =>
    candidate.aliases.some((alias) => normalizePath(alias) === path),
  )

  return aliasOwner
    ? { kind: 'redirect', to: normalizePath(aliasOwner.route) }
    : { kind: 'not-found' }
}

function titleCase(value: string): string {
  return value
    .split('-')
    .map((word) => `${word.charAt(0).toUpperCase()}${word.slice(1)}`)
    .join(' ')
}

export function buildNavigationGroups(
  documents: readonly CatalogDocument[],
): NavigationGroup[] {
  const groups = new Map<string, CatalogDocument[]>()

  for (const document of documents) {
    const key = document.kind === 'component' ? document.category : document.kind
    const current = groups.get(key) ?? []
    current.push(document)
    groups.set(key, current)
  }

  return Array.from(groups, ([id, groupedDocuments]) => ({
    id,
    label: titleCase(id),
    documents: groupedDocuments.sort((left, right) =>
      left.name.localeCompare(right.name),
    ),
  })).sort((left, right) => left.label.localeCompare(right.label))
}

function searchableText(document: CatalogDocument): string {
  return [
    document.name,
    document.summary,
    document.route,
    document.category,
    document.kind,
    document.lifecycle,
    ...document.aliases,
    ...document.keywords,
  ]
    .join(' ')
    .toLocaleLowerCase()
}

export function searchCatalog(
  documents: readonly CatalogDocument[],
  query: string,
): CatalogDocument[] {
  const terms = query
    .trim()
    .toLocaleLowerCase()
    .split(/\s+/u)
    .filter(Boolean)

  if (terms.length === 0) return [...documents].slice(0, 8)

  return documents
    .map((document) => {
      const haystack = searchableText(document)
      const matches = terms.every((term) => haystack.includes(term))
      const exactName = document.name.toLocaleLowerCase() === terms.join(' ')
      const nameStartsWith = document.name
        .toLocaleLowerCase()
        .startsWith(terms.join(' '))

      return {
        document,
        score: matches ? (exactName ? 3 : nameStartsWith ? 2 : 1) : 0,
      }
    })
    .filter(({ score }) => score > 0)
    .sort(
      (left, right) =>
        right.score - left.score ||
        left.document.name.localeCompare(right.document.name),
    )
    .map(({ document }) => document)
}
