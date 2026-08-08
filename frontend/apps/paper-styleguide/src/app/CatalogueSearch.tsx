import type { CatalogDocument } from 'paper-ui/catalog'
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  iconClassName,
} from 'paper-ui/icons'
import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react'
import { Link } from 'react-router-dom'
import { searchCatalog } from './catalogue'

interface CatalogueSearchProps {
  documents: readonly CatalogDocument[]
}

function isEditableTarget(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (target.isContentEditable ||
      ['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName))
  )
}

export function CatalogueSearch({ documents }: CatalogueSearchProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const resultRefs = useRef<Array<HTMLAnchorElement | null>>([])
  const results = searchCatalog(documents, query)
  const closeSearch = (restoreFocus = false) => {
    setOpen(false)
    setQuery('')
    if (restoreFocus) triggerRef.current?.focus()
  }

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const commandSearch =
        event.key.toLocaleLowerCase() === 'k' && (event.metaKey || event.ctrlKey)
      const slashSearch = event.key === '/' && !isEditableTarget(event.target)

      if (commandSearch || slashSearch) {
        event.preventDefault()
        setOpen(true)
      } else if (event.key === 'Escape' && open) {
        event.preventDefault()
        closeSearch(true)
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open])

  useEffect(() => {
    if (open) inputRef.current?.focus()
  }, [open])

  function moveResult(
    event: ReactKeyboardEvent<HTMLElement>,
    currentIndex: number,
  ) {
    if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    const nextIndex =
      event.key === 'Home'
        ? 0
        : event.key === 'End'
          ? results.length - 1
          : (currentIndex + (event.key === 'ArrowDown' ? 1 : -1) +
              results.length) %
            results.length
    resultRefs.current[nextIndex]?.focus()
  }

  function trapDialogFocus(event: ReactKeyboardEvent<HTMLElement>) {
    if (event.key !== 'Tab') return
    const focusable = Array.from(
      event.currentTarget.querySelectorAll<HTMLElement>(
        'button:not([disabled]), input:not([disabled]), a[href]',
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
    <>
      <button
        ref={triggerRef}
        className="shell-search-trigger paper-focus-ring"
        type="button"
        aria-label="Search Paper"
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={() => setOpen(true)}
      >
        <MagnifyingGlassIcon
          aria-hidden="true"
          className={iconClassName('default')}
        />
        <span className="shell-search-trigger__label">Search Paper</span>
        <kbd aria-label="Command or Control K">⌘/Ctrl K</kbd>
      </button>

      {open ? (
        <div
          className="search-backdrop"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) closeSearch()
          }}
        >
          <section
            className="search-dialog paper-surface-raised paper-elevation-showcase"
            role="dialog"
            aria-modal="true"
            aria-labelledby="catalogue-search-title"
            onKeyDown={trapDialogFocus}
          >
            <div className="search-dialog__heading">
              <h2 id="catalogue-search-title" className="paper-type-component">
                Search Paper
              </h2>
              <button
                className="shell-icon-button paper-focus-ring"
                type="button"
                aria-label="Close search"
                onClick={() => closeSearch(true)}
              >
                <XMarkIcon
                  aria-hidden="true"
                  className={iconClassName('default')}
                />
              </button>
            </div>
            <label className="search-field">
              <span className="paper-sr-only">Search documents</span>
              <input
                ref={inputRef}
                type="search"
                value={query}
                placeholder="Try “color” or “foundations”"
                onChange={(event) => setQuery(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'ArrowDown' && results.length > 0) {
                    event.preventDefault()
                    resultRefs.current[0]?.focus()
                  } else if (event.key === 'ArrowUp' && results.length > 0) {
                    event.preventDefault()
                    resultRefs.current[results.length - 1]?.focus()
                  }
                }}
              />
            </label>
            <p className="search-count paper-type-metadata" aria-live="polite">
              {results.length} {results.length === 1 ? 'result' : 'results'}
            </p>
            <ul className="search-results">
              {results.map((document, index) => (
                <li key={document.id}>
                  <Link
                    ref={(node) => {
                      resultRefs.current[index] = node
                    }}
                    className="search-result paper-focus-ring"
                    to={document.route}
                    onClick={() => closeSearch()}
                    onKeyDown={(event) => moveResult(event, index)}
                  >
                    <span>
                      <strong>{document.name}</strong>
                      <small>{document.summary}</small>
                    </span>
                    <span className="lifecycle-badge">{document.lifecycle}</span>
                  </Link>
                </li>
              ))}
            </ul>
          </section>
        </div>
      ) : null}
    </>
  )
}
