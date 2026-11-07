import { useStore } from '@tanstack/react-form'
import { useEffect, useId, useRef, useState } from 'react'
import { usePaperForm, Button } from 'paper-ui'
import { ChevronDownIcon, iconClassName } from 'paper-ui/icons'
import { formatDate, formatNumber, type SampleLog } from '../data'

export function RecentLogPicker({ logs, onSelect }: { logs: readonly SampleLog[]; onSelect: (log: SampleLog) => void }) {
  const search = usePaperForm({ defaultValues: { query: '' } })
  const query = useStore(search.store, state => state.values['query']).trim().toLocaleLowerCase()
  const [visible, setVisible] = useState(5)
  const details = useRef<HTMLDetailsElement>(null)
  const list = useRef<HTMLUListElement>(null)
  const firstAdded = useRef<number | null>(null)
  const id = useId()
  const sorted = logs.filter(log => [log.title, log.language, log.activity, ...log.tags].join(' ').toLocaleLowerCase().includes(query)).sort((a, b) => b.date.localeCompare(a.date))
  const shown = Math.min(visible, sorted.length)

  useEffect(() => {
    if (firstAdded.current === null || !list.current) return
    const row = list.current.children[firstAdded.current] as HTMLLIElement | undefined
    row?.querySelector('button')?.focus({ preventScroll: true })
    if (row) list.current.scrollTop += row.getBoundingClientRect().top - list.current.getBoundingClientRect().top
    firstAdded.current = null
  }, [visible])

  return <details className="min-w-0 bg-paper [&>summary]:flex [&>summary]:min-h-14 [&>summary]:cursor-pointer [&>summary]:list-none [&>summary]:items-center [&>summary]:justify-between [&>summary]:gap-4 [&>summary]:px-6 [&>summary]:py-3 [&>summary]:text-ui [&>summary]:font-semibold [&>summary]:text-ink [&>summary::-webkit-details-marker]:hidden [&>summary>svg]:shrink-0 [&>summary>svg]:text-muted [&[open]>summary>svg]:rotate-flip [&>summary:focus-visible]:outline-ring [&>summary:focus-visible]:outline-focus [&>summary:focus-visible]:outline-offset-ring [&>:not(summary)]:px-6 max-sm:[&>summary]:px-4 max-sm:[&>:not(summary)]:px-4" ref={details}>
    <summary><span>Use a recent log</span><ChevronDownIcon className={iconClassName('compact')} aria-hidden="true" /></summary>
    <search.AppForm><search.AppField name="query" listeners={{ onChange: () => {
      firstAdded.current = null
      setVisible(5)
      if (list.current) list.current.scrollTop = 0
    } }}>{field => <field.Input className="relative mb-3 mt-1 [&>label]:sr-only" type="search" label="Search recent logs" placeholder="Search recent logs…" autoComplete="off" onKeyDown={event => { if (event.key === 'Enter') event.preventDefault() }} />}</search.AppField></search.AppForm>
    {sorted.length ? <>
      <ul className="m-0 max-h-picker list-none overflow-y-auto overscroll-contain p-0 text-sm scrollbar-gutter-stable data-scrollable:h-picker" aria-label="Recent activity entries" data-scrollable={sorted.length > 5 || undefined} ref={list}>
        {sorted.slice(0, visible).map((log, index) => <li className="recent-log-picker__row m-0 grid grid-cols-content-end items-center gap-4 border-0 px-0 py-3 max-sm:gap-2 [&+&]:border-t [&+&]:border-rule-subtle" key={log.id}>
          <div className="min-w-0 [&>p]:mb-0 [&>p]:mt-0.75 [&>p]:block [&>p]:text-xs [&>p]:leading-150 [&>p]:text-muted [&>p]:wrap-anywhere">
            <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-0.75 [&>strong]:block [&>strong]:min-w-0 [&>strong]:grow [&>strong]:basis-48 [&>strong]:text-sm [&>strong]:leading-150 [&>strong]:font-semibold [&>strong]:text-ink [&>strong]:wrap-anywhere [&>time]:m-0 [&>time]:block [&>time]:shrink-0 [&>time]:whitespace-nowrap [&>time]:text-xs [&>time]:leading-150 [&>time]:text-muted"><strong id={`${id}-entry-${index}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</strong><time dateTime={log.date}>{formatDate(log.date)}</time></div>
            <p>{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
          </div>
          <Button variant="ghost" aria-label="Use this log" aria-describedby={`${id}-entry-${index}`} onClick={() => {
            if (details.current) details.current.open = false
            onSelect(log)
          }}>Use</Button>
        </li>)}
      </ul>
      <div className="flex items-center justify-between gap-3 border-t border-rule-subtle py-2 text-xs leading-150 text-muted max-sm:gap-2 [&>.paper-button]:shrink-0">
        <span role="status">{query ? `${shown} of ${sorted.length} ${sorted.length === 1 ? 'match' : 'matches'}` : `${shown} ${shown === 1 ? 'entry' : 'entries'} shown${shown === sorted.length ? ' · End of history' : ''}`}</span>
        <Button variant="ghost" disabled={shown === sorted.length} onClick={() => {
          firstAdded.current = visible
          setVisible(count => count + 5)
        }}>{shown === sorted.length ? query ? 'All matches shown' : 'All logs shown' : 'Show more'}</Button>
      </div>
    </> : <p className="m-0 py-4 text-sm text-muted" role="status">{query ? 'No matching logs.' : 'Your entries will appear here after you save your first log.'}</p>}
  </details>
}
