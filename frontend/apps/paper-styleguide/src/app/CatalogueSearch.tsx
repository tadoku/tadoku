import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { a11y, Button, iconSizes, Modal, typography, usePaperForm } from 'paper-ui'
import type { CatalogDocument } from 'paper-ui/catalog'
import { MagnifyingGlassIcon } from 'paper-ui/icons'
import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react'
import { Link, useLocation } from 'react-router-dom'
import { searchCatalog } from './catalogue'
import { shellStyles } from '../styles/shell.styles'
import { descendantStyles } from '../descendants.styles'

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
  const methods = usePaperForm({
    defaultValues: ({ query: '' } as CatalogueSearchForm),
  })
  const query = useStore(methods.store, state => state.values['query'])
  const results = searchCatalog(documents, query)

  const changeOpen = (nextOpen: boolean) => {
    setOpen(nextOpen)
    if (!nextOpen) methods.reset({ query: '' }, { keepDefaultValues: true })
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
          variant="ghost"
          aria-label="Search Paper"
          leadingIcon={
            <MagnifyingGlassIcon
              aria-hidden="true"
              {...stylex.props(iconSizes.default)}
            />
          }
        >
          <span {...stylex.props(shellStyles.shellSearchTrigger__label)}>Search</span>
          <kbd {...stylex.props(shellStyles.shellSearchTrigger__shortcut)} aria-hidden="true">
            /
          </kbd>
        </Button>
      }
    >
      <div>
        <methods.AppForm>
          <methods.AppField name="query">{field => <field.Input ref={inputRef} xstyle={[shellStyles.searchField]} type="search" label="Search documents" placeholder="Try “color” or “foundations”" onKeyDown={(event) => {
              if (event.key === 'ArrowDown' && results.length > 0) {
                event.preventDefault()
                resultRefs.current[0]?.focus()
              } else if (event.key === 'ArrowUp' && results.length > 0) {
                event.preventDefault()
                resultRefs.current[results.length - 1]?.focus()
              }
            }} />}</methods.AppField>
        </methods.AppForm>
        <p {...stylex.props(shellStyles.searchCount, typography.metadata)} aria-live="polite">
          {results.length} {results.length === 1 ? 'result' : 'results'}
        </p>
        <ul {...stylex.props(shellStyles.searchResults)}>
          {results.map((document, index) => (
            <li key={document.id}>
              <Link
                ref={(node) => {
                  resultRefs.current[index] = node
                }}
                {...stylex.props(shellStyles.searchResult, a11y.focusRing)}
                to={document.route}
                onClick={() => changeOpen(false)}
                onKeyDown={(event) => moveResult(event, index)}
              >
                <span>
                  <strong {...stylex.props(descendantStyles.searchResultU_uStrong)}>{document.name}</strong>
                  <small {...stylex.props(descendantStyles.searchResultU_uSmall)}>{document.summary}</small>
                </span>
                <span {...stylex.props(shellStyles.lifecycleBadge)}>{document.lifecycle}</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </Modal>
  )
}
