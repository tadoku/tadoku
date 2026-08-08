import type { CatalogDocument } from 'paper-ui/catalog'
import { useEffect, useRef, useState } from 'react'
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

  return (
    <>
      <button
        ref={triggerRef}
        className="shell-search-trigger paper-focus-ring"
        type="button"
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={() => setOpen(true)}
      >
        <span>Search Paper</span>
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
                ×
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
              />
            </label>
            <p className="search-count paper-type-metadata" aria-live="polite">
              {results.length} {results.length === 1 ? 'result' : 'results'}
            </p>
            <ul className="search-results">
              {results.map((document) => (
                <li key={document.id}>
                  <Link
                    className="search-result paper-focus-ring"
                    to={document.route}
                    onClick={() => closeSearch()}
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
