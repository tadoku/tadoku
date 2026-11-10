import {
  AppliedFilters,
  FilterMenu,
  SectionHeading,
  listRows,
  metaLine,
  type FilterMenuHandle,
  type FilterMenuSection,
} from 'paper-ui'
import { useId, useRef, useState } from 'react'

interface Log {
  readonly id: number
  readonly title: string
  readonly language: string
  readonly activity: string
  readonly tags: readonly string[]
  readonly points: number
}

const logs: readonly Log[] = [
  { id: 1, title: 'Kiki’s Delivery Service', language: 'Japanese', activity: 'Reading', tags: ['novel'], points: 84.5 },
  { id: 2, title: 'NHK News Web Easy', language: 'Japanese', activity: 'Reading', tags: ['news'], points: 22 },
  { id: 3, title: 'Hanashite Hanashite', language: 'Japanese', activity: 'Listening', tags: ['podcast'], points: 41.2 },
  { id: 4, title: 'Yotsuba&!', language: 'Japanese', activity: 'Reading', tags: ['manga'], points: 36 },
  { id: 5, title: 'El Hobbit', language: 'Spanish', activity: 'Listening', tags: ['audiobook', 'novel'], points: 58 },
  { id: 6, title: 'Diario de clase', language: 'Spanish', activity: 'Writing', tags: [], points: 12 },
  { id: 7, title: 'Radio Ambulante', language: 'Spanish', activity: 'Listening', tags: ['podcast'], points: 47.5 },
]

type Dimension = 'language' | 'activity' | 'tag'
type Selection = Record<Dimension, readonly string[]>

const dimensions: readonly { readonly id: Dimension; readonly label: string }[] = [
  { id: 'language', label: 'Language' },
  { id: 'activity', label: 'Activity' },
  { id: 'tag', label: 'Tags' },
]
const none: Selection = { language: [], activity: [], tag: [] }
const valuesOf = (log: Log, dimension: Dimension) => (dimension === 'tag' ? log.tags : [log[dimension]])

// Values within a filter are "any of"; filters combine with "and". `except` leaves one filter out.
const matches = (log: Log, selection: Selection, except?: Dimension) =>
  dimensions.every(
    ({ id }) =>
      id === except ||
      selection[id].length === 0 ||
      valuesOf(log, id).some(value => selection[id].includes(value)),
  )

// A value's points count the logs that the other filters still allow, so the menu offers what has results.
function sectionsFor(selection: Selection): FilterMenuSection[] {
  return dimensions.map(({ id, label }) => {
    const points = new Map<string, number>()
    for (const log of logs) {
      for (const value of valuesOf(log, id)) {
        const allowed = matches(log, selection, id) ? log.points : 0
        points.set(value, (points.get(value) ?? 0) + allowed)
      }
    }
    return {
      id,
      label,
      selected: selection[id],
      values: [...points]
        .sort(([, left], [, right]) => right - left)
        .map(([value, total]) => ({ value, label: value, points: total })),
    }
  })
}

export default function FilterBarExample() {
  const [selection, setSelection] = useState<Selection>(none)
  const menu = useRef<FilterMenuHandle>(null)
  const headingId = useId()
  const rows = logs.filter(log => matches(log, selection))
  const list = listRows()
  const resultLabel = `${rows.length} ${rows.length === 1 ? 'log' : 'logs'}`
  const clear = () => {
    setSelection(none)
    menu.current?.focus()
  }
  const remove = (dimension: Dimension, value: string) => {
    const next = { ...selection, [dimension]: selection[dimension].filter(item => item !== value) }
    setSelection(next)
    if (Object.values(next).every(values => values.length === 0)) menu.current?.focus()
  }

  return (
    <section aria-labelledby={headingId} className="max-w-panel-wide">
      <SectionHeading
        level={2}
        id={headingId}
        title="Logs"
        actions={
          <FilterMenu
            ref={menu}
            sections={sectionsFor(selection)}
            onChange={(id, values) => setSelection({ ...selection, [id]: values })}
            onClear={() => setSelection(none)}
            resultLabel={resultLabel}
          />
        }
      />
      <AppliedFilters
        variant="removable"
        className="mb-3"
        filters={dimensions.flatMap(({ id }) =>
          selection[id].map(value => ({ id: `${id}:${value}`, label: value, onRemove: () => remove(id, value) })),
        )}
        onClear={clear}
        resultLabel={`${resultLabel} of ${logs.length}`}
      />
      <ul className={list.root()}>
        {rows.map(log => (
          <li key={log.id} className={list.item({ className: 'grid-cols-content-end gap-x-4' })}>
            <span>
              {log.title}
              <span className={metaLine({ className: 'block' })}>
                {[log.language, log.activity, ...log.tags].join(' · ')}
              </span>
            </span>
            <span className="font-bold tabular-nums">{log.points}</span>
          </li>
        ))}
      </ul>
    </section>
  )
}
