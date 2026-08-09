import { Button, Input, Modal } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import { MagnifyingGlassIcon, iconClassName } from 'paper-ui/icons'
import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { Link, useLocation } from 'react-router-dom'
import { searchCatalog } from './catalogue'

interface CatalogueSearchProps {
  documents: readonly CatalogDocument[]
}

interface CatalogueSearchForm {
  query: string
}

function isEditableTarget(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (target.isContentEditable ||
      ['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName))
  )
}

export function CatalogueSearch({ documents }: CatalogueSearchProps) {
  const location = useLocation()
  return <CatalogueSearchDialog key={location.key} documents={documents} />
}

function CatalogueSearchDialog({ documents }: CatalogueSearchProps) {
  const [open, setOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const resultRefs = useRef<Array<HTMLAnchorElement | null>>([])
  const methods = useForm<CatalogueSearchForm>({
    defaultValues: { query: '' },
  })
  const query = methods.watch('query')
  const results = searchCatalog(documents, query)

  const changeOpen = (nextOpen: boolean) => {
    setOpen(nextOpen)
    if (!nextOpen) methods.reset({ query: '' })
  }

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const commandSearch =
        event.key.toLocaleLowerCase() === 'k' && (event.metaKey || event.ctrlKey)
      const slashSearch = event.key === '/' && !isEditableTarget(event.target)

      if (commandSearch || slashSearch) {
        event.preventDefault()
        setOpen(true)
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

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

  return (
    <Modal
      title="Search Paper"
      closeLabel="Close search"
      open={open}
      onOpenChange={changeOpen}
      initialFocus={inputRef}
      footer={null}
      trigger={
        <Button
          className="shell-search-trigger"
          variant="ghost"
          aria-label="Search Paper"
          leadingIcon={
            <MagnifyingGlassIcon
              aria-hidden="true"
              className={iconClassName('default')}
            />
          }
        >
          <span className="shell-search-trigger__label">Search</span>
          <kbd className="shell-search-trigger__shortcut" aria-hidden="true">
            /
          </kbd>
        </Button>
      }
    >
      <div className="catalogue-search">
        <FormProvider {...methods}>
          <Input
            ref={inputRef}
            className="search-field"
            name="query"
            type="search"
            label="Search documents"
            placeholder="Try “color” or “foundations”"
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
        </FormProvider>
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
                onClick={() => changeOpen(false)}
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
      </div>
    </Modal>
  )
}
