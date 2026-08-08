import type { CatalogDocument } from 'paper-ui/catalog'
import cutMeterUrl from 'paper-ui/assets/brand/cut-meter.svg?no-inline'
import cutMeterReversedUrl from 'paper-ui/assets/brand/cut-meter-reversed.svg?no-inline'
import { Bars3Icon, XMarkIcon, iconClassName } from 'paper-ui/icons'
import {
  type PropsWithChildren,
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react'
import { Link, NavLink } from 'react-router-dom'
import { buildNavigationGroups } from './catalogue'
import { CatalogueSearch } from './CatalogueSearch'

interface DocsShellProps extends PropsWithChildren {
  documents: readonly CatalogDocument[]
}

interface CatalogueNavigationProps {
  documents: readonly CatalogDocument[]
  onNavigate?: () => void
}

function CatalogueNavigation({
  documents,
  onNavigate,
}: CatalogueNavigationProps) {
  return (
    <nav className="catalogue-nav" aria-label="Paper catalogue">
      {buildNavigationGroups(documents).map((group) => (
        <section key={group.id} className="catalogue-nav__group">
          <h2 className="paper-type-metadata">{group.label}</h2>
          <ul>
            {group.documents.map((document) => (
              <li key={document.id}>
                <NavLink
                  className={({ isActive }) =>
                    `catalogue-nav__link paper-focus-ring${
                      isActive ? ' catalogue-nav__link--active' : ''
                    }`
                  }
                  to={document.route}
                  onClick={onNavigate}
                >
                  <span>{document.name}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </nav>
  )
}

export function DocsShell({ documents, children }: DocsShellProps) {
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
  const mobileTriggerRef = useRef<HTMLButtonElement>(null)
  const mobileCloseRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (
      typeof window === 'undefined' ||
      typeof window.matchMedia !== 'function'
    ) {
      return
    }

    const collapsedLayout = window.matchMedia('(max-width: 64rem)')
    const closeOnDesktop = (event: MediaQueryListEvent) => {
      if (!event.matches) setMobileNavigationOpen(false)
    }

    collapsedLayout.addEventListener('change', closeOnDesktop)
    return () =>
      collapsedLayout.removeEventListener('change', closeOnDesktop)
  }, [])

  useEffect(() => {
    if (!mobileNavigationOpen) return
    mobileCloseRef.current?.focus()
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setMobileNavigationOpen(false)
        mobileTriggerRef.current?.focus()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [mobileNavigationOpen])

  function trapMobileFocus(event: ReactKeyboardEvent<HTMLElement>) {
    if (event.key !== 'Tab') return
    const focusable = Array.from(
      event.currentTarget.querySelectorAll<HTMLElement>(
        'button:not([disabled]), select:not([disabled]), a[href]',
      ),
    )
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    if (!first || !last) return

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }

  return (
    <div className="docs-shell">
      <a className="skip-link paper-focus-ring" href="#main-content">
        Skip to content
      </a>
      <header className="docs-header">
        <Link className="paper-wordmark paper-focus-ring" to="/">
          <span aria-hidden="true">
            <img
              className="paper-wordmark__mark--light"
              src={cutMeterUrl}
              alt=""
            />
            <img
              className="paper-wordmark__mark--dark"
              src={cutMeterReversedUrl}
              alt=""
            />
          </span>
          <strong>Tadoku Paper</strong>
        </Link>
        <div className="docs-header__actions">
          <CatalogueSearch documents={documents} />
          <button
            ref={mobileTriggerRef}
            className="mobile-nav-trigger paper-focus-ring"
            type="button"
            aria-label="Browse"
            aria-controls="mobile-catalogue-navigation"
            aria-expanded={mobileNavigationOpen}
            onClick={() => setMobileNavigationOpen((value) => !value)}
          >
            <Bars3Icon
              aria-hidden="true"
              className={iconClassName('default')}
            />
            <span className="mobile-nav-trigger__label">Browse</span>
          </button>
        </div>
      </header>

      <aside className="docs-sidebar">
        <CatalogueNavigation documents={documents} />
      </aside>

      <div
        className="mobile-nav-backdrop"
        role="presentation"
        data-open={mobileNavigationOpen ? 'true' : 'false'}
        aria-hidden={!mobileNavigationOpen}
        onMouseDown={(event) => {
          if (event.target === event.currentTarget) {
            setMobileNavigationOpen(false)
          }
        }}
      >
        <aside
          ref={(node) => {
            if (node) node.inert = !mobileNavigationOpen
          }}
          id="mobile-catalogue-navigation"
          className="mobile-nav-drawer paper-surface-raised paper-elevation-showcase"
          aria-label="Mobile catalogue navigation"
          onKeyDown={trapMobileFocus}
        >
          <div className="mobile-nav-drawer__heading">
            <strong className="paper-type-component">Browse Paper</strong>
            <button
              ref={mobileCloseRef}
              type="button"
              className="shell-icon-button paper-focus-ring"
              aria-label="Close navigation"
              onClick={() => {
                setMobileNavigationOpen(false)
                mobileTriggerRef.current?.focus()
              }}
            >
              <XMarkIcon
                aria-hidden="true"
                className={iconClassName('default')}
              />
            </button>
          </div>
          <CatalogueNavigation
            documents={documents}
            onNavigate={() => setMobileNavigationOpen(false)}
          />
        </aside>
      </div>

      <main id="main-content" className="docs-main" tabIndex={-1}>
        {children}
      </main>
    </div>
  )
}
