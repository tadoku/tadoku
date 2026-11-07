import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { useEffect, useId, useRef, useState } from 'react'
import { Button, iconSizes, usePaperForm } from 'paper-ui'
import { ChevronDownIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, type SampleLog } from '../data'
import { logEditorStyles } from './log-editor.styles'
import { recentLogPickerStyles } from './recent-log-picker.styles'
import { recordsDescendants as descendantStyles } from './records-descendants.styles'

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

  return <details {...stylex.props(logEditorStyles.logEditor__disclosure)} ref={details}>
    <summary {...stylex.props(descendantStyles.logEditorU_uDisclosureU_uSummary, descendantStyles.logEditorU_uDisclosureU_uAny)}><span>Use a recent log</span><ChevronDownIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" /></summary>
    <search.AppForm {...stylex.props(descendantStyles.logEditorU_uDisclosureU_uAny)}><search.AppField name="query" listeners={{ onChange: () => {
      firstAdded.current = null
      setVisible(5)
      if (list.current) list.current.scrollTop = 0
    } }}>{field => <field.Input xstyle={[recentLogPickerStyles.recentLogPicker__search]} type="search" label="Search recent logs" placeholder="Search recent logs…" autoComplete="off" onKeyDown={event => { if (event.key === 'Enter') event.preventDefault() }} />}</search.AppField></search.AppForm>
    {sorted.length ? <>
      <ul {...stylex.props(recentLogPickerStyles.recentLogPicker__list, descendantStyles.logEditorU_uDisclosureU_uAny)} aria-label="Recent activity entries" data-scrollable={sorted.length > 5 || undefined} ref={list}>
        {sorted.slice(0, visible).map((log, index) => <li {...stylex.props(recentLogPickerStyles.recentLogPicker__row, descendantStyles.recentLogPickerU_uListU_uRecentLogPickerU_uRow, descendantStyles.recentLogPickerU_uRow)} key={log.id}>
          <div {...stylex.props(recentLogPickerStyles.recentLogPicker__copy)}>
            <div {...stylex.props(recentLogPickerStyles.recentLogPicker__entryTitle)}><strong {...stylex.props(descendantStyles.recentLogPickerU_uEntryTitleU_uStrong)} id={`${id}-entry-${index}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</strong><time {...stylex.props(descendantStyles.recentLogPickerU_uEntryTitleU_uTime)} dateTime={log.date}>{formatDate(log.date)}</time></div>
            <p {...stylex.props(descendantStyles.recentLogPickerU_uCopyU_uP)}>{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
          </div>
          <Button variant="ghost" aria-label="Use this log" aria-describedby={`${id}-entry-${index}`} onClick={() => {
            if (details.current) details.current.open = false
            onSelect(log)
          }}>Use</Button>
        </li>)}
      </ul>
      <div {...stylex.props(recentLogPickerStyles.recentLogPicker__footer, descendantStyles.logEditorU_uDisclosureU_uAny)}>
        <span role="status">{query ? `${shown} of ${sorted.length} ${sorted.length === 1 ? 'match' : 'matches'}` : `${shown} ${shown === 1 ? 'entry' : 'entries'} shown${shown === sorted.length ? ' · End of history' : ''}`}</span>
        <Button variant="ghost" disabled={shown === sorted.length} onClick={() => {
          firstAdded.current = visible
          setVisible(count => count + 5)
        }}>{shown === sorted.length ? query ? 'All matches shown' : 'All logs shown' : 'Show more'}</Button>
      </div>
    </> : <p {...stylex.props(recentLogPickerStyles.recentLogPicker__empty, descendantStyles.logEditorU_uDisclosureU_uAny)} role="status">{query ? 'No matching logs.' : 'Your entries will appear here after you save your first log.'}</p>}
  </details>
}
