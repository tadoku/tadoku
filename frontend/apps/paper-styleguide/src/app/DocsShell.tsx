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
import { useDisplayPreferences } from './useDisplayPreferences'

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
                  <small>{document.lifecycle}</small>
                </NavLink>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </nav>
  )
}

function PreferenceControls() {
  const { theme, density, setTheme, setDensity } = useDisplayPreferences()

  return (
    <div className="shell-preferences" aria-label="Display preferences">
      <label>
        <span>Theme</span>
        <select
          value={theme}
          onChange={(event) =>
            setTheme(event.target.value === 'dark' ? 'dark' : 'light')
          }
        >
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </label>
      <label>
        <span>Density</span>
        <select
          value={density}
          onChange={(event) =>
            setDensity(
              event.target.value === 'compact' ? 'compact' : 'comfortable',
            )
          }
        >
          <option value="comfortable">Comfortable</option>
          <option value="compact">Compact</option>
        </select>
      </label>
    </div>
  )
}

export function DocsShell({ documents, children }: DocsShellProps) {
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
  const mobileTriggerRef = useRef<HTMLButtonElement>(null)
  const mobileCloseRef = useRef<HTMLButtonElement>(null)

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
        <PreferenceControls />
        <CatalogueNavigation documents={documents} />
      </aside>

      {mobileNavigationOpen ? (
        <div
          className="mobile-nav-backdrop"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) {
              setMobileNavigationOpen(false)
            }
          }}
        >
          <aside
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
            <PreferenceControls />
            <CatalogueNavigation
              documents={documents}
              onNavigate={() => setMobileNavigationOpen(false)}
            />
          </aside>
        </div>
      ) : null}

      <main id="main-content" className="docs-main" tabIndex={-1}>
        {children}
      </main>
    </div>
  )
}
