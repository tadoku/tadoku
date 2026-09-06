import { Table, type TableColumn } from 'paper-ui'

interface ReadingRow {
  readonly id: string
  readonly title: string
  readonly language: string
  readonly progress: number
  readonly status: string
}

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
  return (
    <Table
      caption="Recent reading"
      rows={[]}
      columns={readingColumns}
      emptyMessage="No reading logged yet."
    />
  )
}
