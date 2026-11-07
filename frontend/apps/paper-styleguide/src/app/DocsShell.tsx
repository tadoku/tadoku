import * as stylex from '@stylexjs/stylex'
import { a11y, Button, Drawer, iconSizes, Navbar, Sidebar, type NavigationLinkRenderer, type SidebarSection } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import wordmarkAccentUrl from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import wordmarkReversedUrl from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'
import { Bars3Icon } from 'paper-ui/icons'
import { type PropsWithChildren, useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { buildNavigationGroups } from './catalogue'
import { CatalogueSearch } from './CatalogueSearch'
import { shellStyles } from '../styles/shell.styles'

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
          xstyle={[shellStyles.mobileNavTrigger]}
          variant="ghost"
          aria-label="Browse"
          leadingIcon={
            <Bars3Icon
              aria-hidden="true"
              {...stylex.props(iconSizes.default)}
            />
          }
        >
          <span {...stylex.props(shellStyles.mobileNavTrigger__label)}>Browse</span>
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
    <div {...stylex.props(shellStyles.docsShell)}>
      <a {...stylex.props(shellStyles.skipLink, a11y.focusRing)} href="#main-content">
        Skip to content
      </a>
      <header {...stylex.props(shellStyles.docsNavbar)}>
        <Navbar
          brandHref="/"
          navigation={[]}
          renderLink={renderRouterLink}
          mobileNavigation={false}
          brand={
            <span {...stylex.props(shellStyles.docsWordmark)} role="img" aria-label="Tadoku">
              <img
                {...stylex.props(shellStyles.docsWordmark__mark, shellStyles.docsWordmark__mark_ink)}
                src={cutMeterUrl}
                alt=""
                aria-hidden="true"
                width={36}
                height={36}
              />
              <img
                {...stylex.props(shellStyles.docsWordmark__mark, shellStyles.docsWordmark__mark_reversed)}
                src={cutMeterReversedUrl}
                alt=""
                aria-hidden="true"
                width={36}
                height={36}
              />
              <img
                {...stylex.props(shellStyles.docsWordmark__image, shellStyles.docsWordmark__image_accent)}
                src={wordmarkAccentUrl}
                alt=""
                aria-hidden="true"
                width={126.4}
                height={23.2}
              />
              <img
                {...stylex.props(shellStyles.docsWordmark__image, shellStyles.docsWordmark__image_reversed)}
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

      <aside {...stylex.props(shellStyles.docsSidebar)} data-density="compact">
        <Sidebar
          sections={sections}
          currentPath={location.pathname}
          renderLink={renderRouterLink}
          label="Paper catalogue"
        />
      </aside>

      <main id="main-content" {...stylex.props(shellStyles.docsMain)} tabIndex={-1}>
        {children}
      </main>
    </div>
  )
}
