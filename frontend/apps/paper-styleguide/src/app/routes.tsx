import { Surface, surfaceClassName } from 'paper-ui'
import { catalogRegistry, type CatalogKind } from 'paper-ui/catalog'
import { useEffect } from 'react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { DocumentPage } from '../documentation/DocumentPage'
import { SetupPage } from '../documentation/SetupPage'
import { buildNavigationGroups, resolveCatalogRoute } from './catalogue'
import { DESIGN_HISTORY_LINKS } from './designHistory'

const INDEX_SECTIONS: readonly {
  id: string
  label: string
  kind: CatalogKind
}[] = [
  { id: 'foundations', label: 'Foundations', kind: 'foundation' },
  { id: 'components', label: 'Components', kind: 'component' },
  { id: 'patterns', label: 'Patterns', kind: 'pattern' },
  { id: 'experiments', label: 'Experiments', kind: 'experiment' },
  { id: 'governance', label: 'Governance', kind: 'governance' },
]

export function CatalogIndex() {
  const orderedDocuments = buildNavigationGroups(catalogRegistry.documents)
    .flatMap((group) => group.documents)

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
      <section className="catalogue-index__section" aria-labelledby="getting-started-title">
        <h2 id="getting-started-title" className="paper-type-section">Start using Paper</h2>
        <p>
          Load the shared stylesheet and Tailwind preset once, choose a theme
          and density, then compose pages with p-4, gap-2 and public components. Each component’s
          Examples section includes a working preview, copyable code and a props reference.
        </p>
        <p>
          <Link className="text-link paper-focus-ring" to="/setup">Set up Paper</Link>, then start with <Link className="text-link paper-focus-ring" to="/foundations/spacing-and-density">spacing and density</Link> and{' '}
          <Link className="text-link paper-focus-ring" to="/foundations/layout">layout</Link> to compose a page.
          For forms, the <Link className="text-link paper-focus-ring" to="/components/forms/input">Input guide</Link> includes
          the required React Hook Form setup. Paper owns appearance and interaction;
          your application owns routing, data and submission.
        </p>
      </section>
      {INDEX_SECTIONS.map((section) => {
        const documents = orderedDocuments.filter(
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
                  className={surfaceClassName({
                    elevation: 'floating',
                    className: 'document-card paper-focus-ring',
                  })}
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
      <section
        className="catalogue-index__section design-history"
        aria-labelledby="design-history-title"
      >
        <h2 id="design-history-title" className="paper-type-section">
          Design history
        </h2>
        <p>
          Trace the decisions, evidence, and delivery gates behind the system.
        </p>
        <ul className="design-history__links">
          {DESIGN_HISTORY_LINKS.map((entry) => (
            <li key={entry.href}>
              <a
                className={surfaceClassName({
                  className: 'design-history__link paper-focus-ring',
                })}
                href={entry.href}
              >
                <strong>{entry.label}</strong>
                <span>{entry.description}</span>
              </a>
            </li>
          ))}
        </ul>
      </section>
    </article>
  )
}

export function ResolvedCatalogueRoute() {
  const location = useLocation()

  useEffect(() => {
    if (!location.hash) {
      window.scrollTo({ top: 0, left: 0, behavior: 'instant' })
      return
    }
    let id = location.hash.slice(1)
    try {
      id = decodeURIComponent(id)
    } catch {
      return
    }
    document.getElementById(id)?.scrollIntoView?.({ block: 'start' })
  }, [location.hash, location.pathname])

  if (location.pathname === '/') return <CatalogIndex />
  if (location.pathname === '/setup') return <SetupPage />

  const resolved = resolveCatalogRoute(location.pathname, catalogRegistry)
  if (resolved.kind === 'redirect') {
    return <Navigate replace to={`${resolved.to}${location.search}${location.hash}`} />
  }
  if (resolved.kind === 'document') {
    return <DocumentPage document={resolved.document} />
  }

  return (
    <Surface as="section" accent className="not-found">
      <p className="eyebrow">404 · Outside the catalogue</p>
      <h1 className="paper-type-page">This Paper page does not exist.</h1>
      <p>
        The registry has no document or redirect for <code>{location.pathname}</code>.
      </p>
      <Link className="text-link paper-focus-ring" to="/">
        Return to the catalogue
      </Link>
    </Surface>
  )
}
