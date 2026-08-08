import { catalogRegistry, type CatalogKind } from 'paper-ui/catalog'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { DocumentPage } from '../documentation/DocumentPage'
import { resolveCatalogRoute } from './catalogue'

const INDEX_SECTIONS: readonly {
  id: string
  label: string
  kind: CatalogKind
}[] = [
  { id: 'foundations', label: 'Foundations', kind: 'foundation' },
  { id: 'components', label: 'Components', kind: 'component' },
  { id: 'governance', label: 'Governance', kind: 'governance' },
]

export function CatalogIndex() {
  return (
    <article className="catalogue-index">
      <header className="catalogue-index__hero paper-accent-rail">
        <p className="eyebrow">Tadoku design system</p>
        <h1 className="paper-type-display">Paper makes the interface legible.</h1>
        <p>
          Explore the visual foundations, component contracts, and product
          patterns that keep Tadoku calm, accessible, and recognizably ours.
        </p>
      </header>
      {INDEX_SECTIONS.map((section) => {
        const documents = catalogRegistry.documents.filter(
          (document) => document.kind === section.kind,
        )

        if (documents.length === 0) return null

        return (
          <section
            key={section.id}
            className="catalogue-index__section"
            aria-labelledby={`index-${section.id}-title`}
          >
            <h2
              id={`index-${section.id}-title`}
              className="paper-type-section"
            >
              {section.label}
            </h2>
            <div className="document-card-grid">
              {documents.map((document) => (
                <Link
                  key={document.id}
                  className="document-card paper-surface-raised paper-elevation-floating paper-focus-ring"
                  to={document.route}
                >
                  <span className="paper-type-component">{document.name}</span>
                  <small>{document.lifecycle}</small>
                  <p>{document.summary}</p>
                </Link>
              ))}
            </div>
          </section>
        )
      })}
    </article>
  )
}

export function ResolvedCatalogueRoute() {
  const location = useLocation()

  if (location.pathname === '/') return <CatalogIndex />

  const resolved = resolveCatalogRoute(location.pathname, catalogRegistry)
  if (resolved.kind === 'redirect') {
    return <Navigate replace to={`${resolved.to}${location.search}${location.hash}`} />
  }
  if (resolved.kind === 'document') {
    return <DocumentPage document={resolved.document} />
  }

  return (
    <section className="not-found paper-accent-rail">
      <p className="eyebrow">404 · Outside the catalogue</p>
      <h1 className="paper-type-page">This Paper page does not exist.</h1>
      <p>
        The registry has no document or redirect for <code>{location.pathname}</code>.
      </p>
      <Link className="text-link paper-focus-ring" to="/">
        Return to the catalogue
      </Link>
    </section>
  )
}
