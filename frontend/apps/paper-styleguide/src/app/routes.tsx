import * as stylex from '@stylexjs/stylex'
import { a11y, accentRail, linkStyles, Surface, surfaceStyles, typography } from 'paper-ui'
import { catalogRegistry, type CatalogKind } from 'paper-ui/catalog'
import { useEffect } from 'react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { DocumentPage } from '../documentation/DocumentPage'
import { SetupPage } from '../documentation/SetupPage'
import { buildNavigationGroups, resolveCatalogRoute } from './catalogue'
import { DESIGN_HISTORY_LINKS } from './designHistory'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

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
    <article {...stylex.props(shellStyles.catalogueIndex)}>
      <header {...stylex.props(shellStyles.catalogueIndex__hero, accentRail.base)}>
        <p {...stylex.props(shellStyles.eyebrow, descendantStyles.catalogueIndexU_uHeroU_uP)}>Tadoku design system</p>
        <h1 {...stylex.props(typography.display, descendantStyles.catalogueIndexU_uHeroU_uH1)}>Paper makes the interface legible.</h1>
        <p {...stylex.props(descendantStyles.catalogueIndexU_uHeroU_uP)}>
          Explore the visual foundations, component contracts, and product
          patterns that keep Tadoku calm, accessible, and recognizably ours.
        </p>
      </header>
      <section {...stylex.props(shellStyles.catalogueIndex__section, descendantStyles.catalogueIndexU_uSection)} aria-labelledby="getting-started-title">
        <h2 id="getting-started-title" {...stylex.props(typography.section, descendantStyles.catalogueIndexU_uSectionU_uH2)}>Start using Paper</h2>
        <p>
          Load the shared stylesheet and Tailwind preset once, choose a theme
          and density, then compose pages with p-4, gap-2 and public components. Each component’s
          Examples section includes a working preview, copyable code and a props reference.
        </p>
        <p>
          <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/setup">Set up Paper</Link>, then start with <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/spacing-and-density">spacing and density</Link> and{' '}
          <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/foundations/layout">layout</Link> to compose a page.
          For forms, the <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/components/forms/input">Input guide</Link> includes
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
            {...stylex.props(shellStyles.catalogueIndex__section, descendantStyles.catalogueIndexU_uSection)}
            aria-labelledby={`index-${section.id}-title`}
          >
            <h2
              id={`index-${section.id}-title`}
              {...stylex.props(typography.section, descendantStyles.catalogueIndexU_uSectionU_uH2)}
            >
              {section.label}
            </h2>
            <div {...stylex.props(shellStyles.documentCardGrid)}>
              {documents.map((document) => (
                <Link
                  key={document.id}
                  {...stylex.props(...surfaceStyles({ elevation: 'floating' }), shellStyles.documentCard, a11y.focusRing)}
                  to={document.route}
                >
                  <span {...stylex.props(typography.component)}>{document.name}</span>
                  <small {...stylex.props(descendantStyles.documentCardU_uSmall)}>{document.lifecycle}</small>
                  <p {...stylex.props(descendantStyles.documentCardU_uP)}>{document.summary}</p>
                </Link>
              ))}
            </div>
          </section>
        )
      })}
      <section
        {...stylex.props(shellStyles.catalogueIndex__section, shellStyles.designHistory, descendantStyles.catalogueIndexU_uSection)}
        aria-labelledby="design-history-title"
      >
        <h2 id="design-history-title" {...stylex.props(typography.section, descendantStyles.catalogueIndexU_uSectionU_uH2)}>
          Design history
        </h2>
        <p {...stylex.props(descendantStyles.designHistoryU_uP)}>
          Trace the decisions, evidence, and delivery gates behind the system.
        </p>
        <ul {...stylex.props(shellStyles.designHistory__links)}>
          {DESIGN_HISTORY_LINKS.map((entry) => (
            <li key={entry.href}>
              <a
                {...stylex.props(...surfaceStyles(), shellStyles.designHistory__link, a11y.focusRing)}
                href={entry.href}
              >
                <strong>{entry.label}</strong>
                <span {...stylex.props(descendantStyles.designHistoryU_uLinksU_uSpan)}>{entry.description}</span>
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
    <Surface as="section" accent xstyle={[shellStyles.notFound]}>
      <p {...stylex.props(shellStyles.eyebrow)}>404 · Outside the catalogue</p>
      <h1 {...stylex.props(typography.page, descendantStyles.notFoundU_uH1)}>This Paper page does not exist.</h1>
      <p>
        The registry has no document or redirect for <code>{location.pathname}</code>.
      </p>
      <Link {...stylex.props(linkStyles.base, a11y.focusRing)} to="/">
        Return to the catalogue
      </Link>
    </Surface>
  )
}
