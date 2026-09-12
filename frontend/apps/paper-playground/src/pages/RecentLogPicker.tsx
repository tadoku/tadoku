import { useEffect, useId, useRef, useState } from 'react'
import { Button } from 'paper-ui'
import { ChevronDownIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, type SampleLog } from '../data'
import './recent-log-picker.css'

export function RecentLogPicker({ logs, onSelect }: { logs: readonly SampleLog[]; onSelect: (log: SampleLog) => void }) {
  const [visible, setVisible] = useState(5)
  const details = useRef<HTMLDetailsElement>(null)
  const list = useRef<HTMLUListElement>(null)
  const firstAdded = useRef<number | null>(null)
  const id = useId()
  const sorted = [...logs].sort((a, b) => b.date.localeCompare(a.date))
  const shown = Math.min(visible, sorted.length)

  useEffect(() => {
    if (firstAdded.current === null || !list.current) return
    const row = list.current.children[firstAdded.current] as HTMLLIElement | undefined
    row?.querySelector('button')?.focus({ preventScroll: true })
    if (row) list.current.scrollTop += row.getBoundingClientRect().top - list.current.getBoundingClientRect().top
    firstAdded.current = null
  }, [visible])

  return <details className="recent-log-picker" ref={details}>
    <summary><span>Use a recent log</span><ChevronDownIcon className="paper-icon-compact" aria-hidden="true" /></summary>
    {sorted.length ? <>
      <ul className="recent-log-picker__list" aria-label="Recent activity entries" data-scrollable={sorted.length > 5 || undefined} ref={list}>
        {sorted.slice(0, visible).map((log, index) => <li className="recent-log-picker__row log-editor__recent-row" key={log.id}>
          <div className="recent-log-picker__copy">
            <div className="recent-log-picker__entry-title"><strong id={`${id}-entry-${index}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</strong><time dateTime={log.date}>{formatDate(log.date)}</time></div>
            <p>{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
          </div>
          <Button variant="ghost" aria-label="Use this log" aria-describedby={`${id}-entry-${index}`} onClick={() => {
            if (details.current) details.current.open = false
            onSelect(log)
          }}>Use</Button>
        </li>)}
      </ul>
      <div className="recent-log-picker__footer">
        <span role="status">{shown} {shown === 1 ? 'entry' : 'entries'} shown{shown === sorted.length ? ' · End of history' : ''}</span>
        <Button variant="ghost" disabled={shown === sorted.length} onClick={() => {
          firstAdded.current = visible
          setVisible(count => count + 5)
        }}>{shown === sorted.length ? 'All logs shown' : 'Show more'}</Button>
      </div>
    </> : <p className="recent-log-picker__empty">Your entries will appear here after you save your first log.</p>}
  </details>
}
