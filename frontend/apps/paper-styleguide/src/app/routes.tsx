import { accentRail, cn, focusRing, Surface, surfaceClassName, typography } from 'paper-ui'
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
    <article className="mx-auto max-w-catalogue">
      <header className={cn(accentRail(), 'mb-16 max-w-record ps-4 max-tablet:mb-10 [&_h1]:mb-4 [&_h1]:mt-1.2 [&>p:last-child]:text-component [&>p:last-child]:leading-165 [&>p:last-child]:text-muted')}>
        <p className="m-0 text-metadata font-semibold tracking-caps text-link uppercase">Tadoku design system</p>
        <h1 className={typography({ role: 'display' })}>Paper makes the interface legible.</h1>
        <p>
          Explore the visual foundations, component contracts, and product
          patterns that keep Tadoku calm, accessible, and recognizably ours.
        </p>
      </header>
      <section className="catalogue-index__section [&+&]:mt-14 max-tablet:[&+&]:mt-10 [&>h2]:mb-4 [&>h2]:mt-0" aria-labelledby="getting-started-title">
        <h2 id="getting-started-title" className={typography({ role: 'section' })}>Start using Paper</h2>
        <p>
          Import Paper’s theme into Tailwind once, choose a theme
          and density, then compose pages with p-4, gap-2 and public components. Each component’s
          Examples section includes a working preview, copyable code and a props reference.
        </p>
        <p>
          <Link className={cn(focusRing(), 'text-link')} to="/setup">Set up Paper</Link>, then start with <Link className={cn(focusRing(), 'text-link')} to="/foundations/spacing-and-density">spacing and density</Link> and{' '}
          <Link className={cn(focusRing(), 'text-link')} to="/foundations/layout">layout</Link> to compose a page.
          For forms, the <Link className={cn(focusRing(), 'text-link')} to="/components/forms/input">Input guide</Link> includes
          the required TanStack Form setup. Paper owns appearance and interaction;
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
            className="catalogue-index__section [&+&]:mt-14 max-tablet:[&+&]:mt-10 [&>h2]:mb-4 [&>h2]:mt-0"
            aria-labelledby={`index-${section.id}-title`}
          >
            <h2
              id={`index-${section.id}-title`}
              className={typography({ role: 'section' })}
            >
              {section.label}
            </h2>
            <div className="grid grid-cols-cards gap-4 max-tablet:gap-3">
              {documents.map((document) => (
                <Link
                  key={document.id}
                  className={surfaceClassName({
                    elevation: 'floating',
                    className: 'document-card min-h-card text-ink no-underline max-tablet:min-h-0 max-tablet:p-3.4 [&_small]:float-end [&_small]:text-muted [&_p]:mb-0 [&_p]:text-muted max-tablet:[&_p]:hidden focus-visible:outline-ring focus-visible:outline-focus focus-visible:outline-offset-ring',
                  })}
                  to={document.route}
                >
                  <span className={typography({ role: 'component' })}>{document.name}</span>
                  <small>{document.lifecycle}</small>
                  <p>{document.summary}</p>
                </Link>
              ))}
            </div>
          </section>
        )
      })}
      <section
        className="catalogue-index__section [&+&]:mt-14 max-tablet:[&+&]:mt-10 [&>h2]:mb-4 [&>h2]:mt-0 [&>p]:max-w-docs-section [&>p]:text-muted"
        aria-labelledby="design-history-title"
      >
        <h2 id="design-history-title" className={typography({ role: 'section' })}>
          Design history
        </h2>
        <p>
          Trace the decisions, evidence, and delivery gates behind the system.
        </p>
        <ul className="mb-0 mt-5 grid list-none grid-cols-cards gap-4 p-0 [&_span]:mt-1.4 [&_span]:block [&_span]:text-muted">
          {DESIGN_HISTORY_LINKS.map((entry) => (
            <li key={entry.href}>
              <a
                className={surfaceClassName({
                  className: 'design-history__link block h-full text-ink no-underline focus-visible:outline-ring focus-visible:outline-focus focus-visible:outline-offset-ring',
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
    <Surface as="section" accent className="not-found mx-auto mb-16 max-w-catalogue ps-4 [&_h1]:mb-4 [&_h1]:mt-1.2">
      <p className="m-0 text-metadata font-semibold tracking-caps text-link uppercase">404 · Outside the catalogue</p>
      <h1 className={typography({ role: 'page' })}>This Paper page does not exist.</h1>
      <p>
        The registry has no document or redirect for <code>{location.pathname}</code>.
      </p>
      <Link className={cn(focusRing(), 'text-link')} to="/">
        Return to the catalogue
      </Link>
    </Surface>
  )
}
