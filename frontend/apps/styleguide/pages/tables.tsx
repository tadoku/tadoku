import { CodeBlock, Preview, Title } from '@components/example'
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
  RowData,
} from '@tanstack/react-table'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Table</h1>

      <Title>Example</Title>
      <Preview>
        <ExampleTable />
      </Preview>
      <CodeBlock
        code={`import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
  RowData,
} from '@tanstack/react-table'

declare module '@tanstack/react-table' {
  interface ColumnMeta<TData extends RowData, TValue> {
    className?: string
  }
}

const data: Ranking[] = [
  { rank: '1', user: 'powz', score: 5054.25054 },
  { rank: '2', user: 'Bijak', score: 3605.23605 },
  { rank: '3', user: 'ShockOLatte', score: 2518.72518 },
  { rank: '4', user: 'Ludie', score: 2517.32517 },
  { rank: '5', user: 'Chamsae', score: 2434.42434 },
  { rank: '6', user: 'Salome', score: 2107.12107 },
  { rank: '7', user: 'mmmm', score: 2060.1206 },
  { rank: '8', user: 'Yaku', score: 1667.21667 },
  { rank: '9', user: 'Socks', score: 1635.81635 },
  { rank: '10', user: 'clair', score: 1592.91592 },
]

interface Ranking {
  rank: string
  user: string
  score: number
}

const columnHelper = createColumnHelper<Ranking>()

const columns = [
  columnHelper.accessor('rank', {
    header: () => 'Rank',
    size: 5,
    meta: {
      className: 'justify-center',
    },
  }),
  columnHelper.accessor('user', {
    header: () => 'User',
  }),
  columnHelper.accessor('score', {
    header: () => 'Score',
    size: 20,
    cell: info => Math.round(info.getValue()),
    meta: {
      className: 'justify-end text-right',
    },
  }),
]

function ExampleTable() {
  const table = useReactTable({
    defaultColumn: {
      minSize: 0,
      size: 0,
    },
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <table className="w-full border-collapse">
      <thead>
        {table.getHeaderGroups().map(headerGroup => (
          <tr key={headerGroup.id}>
            {headerGroup.headers.map(header => (
              <th
                key={header.id}
                style={{
                  width: header.getSize() !== 0 ? header.getSize() : undefined,
                }}
                className={\`subtitle px-4 h-14 inline-flex items-center text-left \${
                  header.column.columnDef.meta?.className
                }\`}
              >
                {header.isPlaceholder
                  ? null
                  : flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map(row => (
          <tr
            key={row.id}
            className="font-bold text-xl hover:bg-primary hover:text-white h-14"
          >
            {row.getVisibleCells().map(cell => (
              <td key={cell.id} className="border-0 p-0">
                <a
                  href={\`#\${row.renderValue('rank')}\`}
                  className={\`reset h-14 px-4 flex items-center \${
                    cell.column.columnDef.meta?.className
                  }\`}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </a>
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}`}
      />
    </>
  )
}

declare module '@tanstack/react-table' {
  interface ColumnMeta<TData extends RowData, TValue> {
    className?: string
  }
}

const data: Ranking[] = [
  { rank: '1', user: 'powz', score: 5054.25054 },
  { rank: '2', user: 'Bijak', score: 3605.23605 },
  { rank: '3', user: 'ShockOLatte', score: 2518.72518 },
  { rank: '4', user: 'Ludie', score: 2517.32517 },
  { rank: '5', user: 'Chamsae', score: 2434.42434 },
  { rank: '6', user: 'Salome', score: 2107.12107 },
  { rank: '7', user: 'mmmm', score: 2060.1206 },
  { rank: '8', user: 'Yaku', score: 1667.21667 },
  { rank: '9', user: 'Socks', score: 1635.81635 },
  { rank: '10', user: 'clair', score: 1592.91592 },
]

interface Ranking {
  rank: string
  user: string
  score: number
}

const columnHelper = createColumnHelper<Ranking>()

const columns = [
  columnHelper.accessor('rank', {
    header: () => 'Rank',
    size: 5,
    meta: {
      className: 'justify-center',
    },
  }),
  columnHelper.accessor('user', {
    header: () => 'User',
  }),
  columnHelper.accessor('score', {
    header: () => 'Score',
    size: 20,
    cell: info => Math.round(info.getValue()),
    meta: {
      className: 'justify-end text-right',
    },
  }),
]

function ExampleTable() {
  const table = useReactTable({
    defaultColumn: {
      minSize: 0,
      size: 0,
    },
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <table className="w-full border-collapse">
      <thead>
        {table.getHeaderGroups().map(headerGroup => (
          <tr key={headerGroup.id}>
            {headerGroup.headers.map(header => (
              <th
                key={header.id}
                style={{
                  width: header.getSize() !== 0 ? header.getSize() : undefined,
                }}
                className={`subtitle px-4 h-14 inline-flex items-center text-left ${header.column.columnDef.meta?.className}`}
              >
                {header.isPlaceholder
                  ? null
                  : flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map(row => (
          <tr
            key={row.id}
            className="font-bold text-xl hover:bg-primary hover:text-white h-14"
          >
            {row.getVisibleCells().map(cell => (
              <td key={cell.id} className="border-0 p-0">
                <a
                  href={`#${row.renderValue('rank')}`}
                  className={`reset h-14 px-4 flex items-center ${cell.column.columnDef.meta?.className}`}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </a>
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
