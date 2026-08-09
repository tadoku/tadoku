import {
  Button,
  Drawer,
  Navbar,
  Sidebar,
  type NavigationLinkRenderer,
  type SidebarSection,
} from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import { Bars3Icon, iconClassName } from 'paper-ui/icons'
import { type PropsWithChildren, useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { buildNavigationGroups } from './catalogue'
import { CatalogueSearch } from './CatalogueSearch'

interface DocsShellProps extends PropsWithChildren {
  documents: readonly CatalogDocument[]
}

const renderRouterLink: NavigationLinkRenderer = ({ href, ...props }) => (
  <Link {...props} to={href} />
)

function navigationSections(
  documents: readonly CatalogDocument[],
): readonly SidebarSection[] {
  return buildNavigationGroups(documents).map((group) => ({
    id: group.id,
    title: group.label,
    links: group.documents.map((document) => ({
      id: document.id,
      label: document.name,
      href: document.route,
    })),
  }))
}

function MobileCatalogueNavigation({
  sections,
  currentPath,
}: {
  sections: readonly SidebarSection[]
  currentPath: string
}) {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    if (
      typeof window === 'undefined' ||
      typeof window.matchMedia !== 'function'
    ) {
      return
    }

    const collapsedLayout = window.matchMedia('(max-width: 64rem)')
    const closeOnDesktop = (event: MediaQueryListEvent) => {
      if (!event.matches) setOpen(false)
    }

    collapsedLayout.addEventListener('change', closeOnDesktop)
    return () =>
      collapsedLayout.removeEventListener('change', closeOnDesktop)
  }, [])

  return (
    <Drawer
      title="Browse Paper"
      closeLabel="Close navigation"
      placement="start"
      open={open}
      onOpenChange={setOpen}
      trigger={
        <Button
          className="mobile-nav-trigger"
          variant="outline"
          aria-label="Browse"
          leadingIcon={
            <Bars3Icon
              aria-hidden="true"
              className={iconClassName('default')}
            />
          }
        >
          <span className="mobile-nav-trigger__label">Browse</span>
        </Button>
      }
    >
      <Sidebar
        sections={sections}
        currentPath={currentPath}
        renderLink={renderRouterLink}
        label="Paper catalogue"
      />
    </Drawer>
  )
}

export function DocsShell({ documents, children }: DocsShellProps) {
  const location = useLocation()
  const sections = navigationSections(documents)

  return (
    <div className="docs-shell">
      <a className="skip-link paper-focus-ring" href="#main-content">
        Skip to content
      </a>
      <header className="docs-navbar">
        <Navbar
          brandHref="/"
          navigation={[]}
          renderLink={renderRouterLink}
          mobileNavigation={false}
          brand={
            <span className="docs-wordmark">
              <span aria-hidden="true">
                <img
                  className="docs-wordmark__mark--light"
                  src={cutMeterUrl}
                  alt=""
                />
                <img
                  className="docs-wordmark__mark--dark"
                  src={cutMeterReversedUrl}
                  alt=""
                />
              </span>
              <strong>Tadoku Paper</strong>
            </span>
          }
          actions={
            <>
              <CatalogueSearch documents={documents} />
              <MobileCatalogueNavigation
                key={location.key}
                sections={sections}
                currentPath={location.pathname}
              />
            </>
          }
        />
      </header>

      <aside className="docs-sidebar" data-density="compact">
        <Sidebar
          sections={sections}
          currentPath={location.pathname}
          renderLink={renderRouterLink}
          label="Paper catalogue"
        />
      </aside>

      <main id="main-content" className="docs-main" tabIndex={-1}>
        {children}
      </main>
    </div>
  )
}
