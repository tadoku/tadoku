import { Button, cn, Drawer, focusRing, Navbar, type NavigationLinkRenderer, Sidebar, type SidebarSection } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import wordmarkAccentUrl from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import wordmarkReversedUrl from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'
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
  return [{ id: 'getting-started', title: 'Getting started', links: [{ id: 'setup', label: 'Setup', href: '/setup' }] }, ...buildNavigationGroups(documents).map((group) => ({
    id: group.id,
    title: group.label,
    links: group.documents.map((document) => ({
      id: document.id,
      label: document.name,
      href: document.route,
    })),
  }))]
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
          className="mobile-nav-trigger hidden max-lg:inline-flex max-2xs:w-control max-2xs:justify-center max-2xs:px-0"
          variant="ghost"
          aria-label="Browse"
          leadingIcon={
            <Bars3Icon
              aria-hidden="true"
              className={iconClassName('default')}
            />
          }
        >
          <span className="mobile-nav-trigger__label max-2xs:hidden">Browse</span>
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
    <div className="docs-shell grid min-h-screen grid-cols-docs-shell grid-rows-docs max-lg:block [&_code]:text-link [&_code]:wrap-anywhere">
      <a className={cn(focusRing(), 'skip-link fixed start-inline top-inline z-skip -translate-y-offscreen bg-action px-field-x py-field-y text-inverse focus:translate-y-0 motion-safe:transition-transform motion-safe:duration-quick')} href="#main-content">
        Skip to content
      </a>
      <header className="docs-navbar sticky top-0 z-docs-navbar col-span-full">
        <Navbar
          brandHref="/"
          navigation={[]}
          renderLink={renderRouterLink}
          mobileNavigation={false}
          brand={
            <span className="docs-wordmark flex items-center gap-1.75" role="img" aria-label="Tadoku">
              <img
                className="block size-9 docs-wordmark__mark--ink dark:hidden"
                src={cutMeterUrl}
                alt=""
                aria-hidden="true"
                width={36}
                height={36}
              />
              <img
                className="block size-9 docs-wordmark__mark--reversed hidden dark:block"
                src={cutMeterReversedUrl}
                alt=""
                aria-hidden="true"
                width={36}
                height={36}
              />
              <img
                className="docs-wordmark__image block h-docs-wordmark-y w-docs-wordmark docs-wordmark__image--accent dark:hidden"
                src={wordmarkAccentUrl}
                alt=""
                aria-hidden="true"
                width={126.4}
                height={23.2}
              />
              <img
                className="docs-wordmark__image block h-docs-wordmark-y w-docs-wordmark docs-wordmark__image--reversed hidden dark:block"
                src={wordmarkReversedUrl}
                alt=""
                aria-hidden="true"
                width={126.4}
                height={23.2}
              />
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

      <aside className="docs-sidebar sticky top-docs-navbar h-docs-sidebar self-start overflow-y-auto border-e border-rule-subtle bg-paper px-4 pb-12 pt-5 max-lg:hidden" data-density="compact">
        <Sidebar
          sections={sections}
          currentPath={location.pathname}
          renderLink={renderRouterLink}
          label="Paper catalogue"
        />
      </aside>

      <main id="main-content" className="min-w-0 px-docs-x pb-28 pt-docs-y max-tablet:pt-8" tabIndex={-1}>
        {children}
      </main>
    </div>
  )
}
