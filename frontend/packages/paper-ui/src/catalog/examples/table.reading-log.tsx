import { useId } from 'react'
import { Button, Table, type TableColumn } from 'paper-ui'

interface ReadingRow {
  readonly id: string
  readonly title: string
  readonly language: string
  readonly progress: number
  readonly status: string
}

const readingRows: readonly ReadingRow[] = [
  {
    id: '1',
    title: 'The Housekeeper and the Professor',
    language: 'Japanese',
    progress: 184,
    status: 'Finished',
  },
  {
    id: '2',
    title: 'Convenience Store Woman',
    language: 'Japanese',
    progress: 73,
    status: 'Reading',
  },
  {
    id: '3',
    title: 'The Three-Body Problem',
    language: 'Chinese',
    progress: 42,
    status: 'Reading',
  },
]

const readingColumns: readonly TableColumn<ReadingRow>[] = [
  {
    id: 'title',
    header: 'Title',
    rowHeader: true,
    width: '18rem',
    cell: row => row.title,
  },
  { id: 'language', header: 'Language', cell: row => row.language },
  { id: 'progress', header: 'Pages', align: 'end', cell: row => row.progress },
  { id: 'status', header: 'Status', cell: row => row.status },
]

export default function Example() {
  const id = useId().replace(/:/g, '')
  return (
    <div className="paper-stack">
      <div><Button variant="outline" onClick={event => event.currentTarget.ownerDocument.getElementById(`${id}-reading-2`)?.focus()}>Find current reading</Button></div>
    <Table
      caption="Recent reading"
      rows={readingRows}
      columns={readingColumns}
      getRowKey={row => row.id}
      getRowProps={row => ({
        id: `${id}-reading-${row.id}`,
        tabIndex: -1,
        'aria-current': row.id === '2' ? 'true' : undefined,
        className: 'paper-focus-ring',
        style: row.id === '2' ? { backgroundColor: 'var(--paper-color-action-soft)' } : undefined,
      })}
    />
    </div>
  )
}
