import { CodeBlock, Preview, Title } from '@components/example'
import { Pagination } from 'ui'
import { useState } from 'react'

export default function Page() {
  return (
    <>
      <h1 className="title mb-8">Pagination</h1>

      <Title>Example</Title>
      <Preview>
        <ExamplePagination />
      </Preview>
      <CodeBlock
        code={`import Pagination from '@components/Pagination'
import { useState } from 'react'

function ExamplePagination() {
  const [page, setPage] = useState(1)

  return <Pagination totalPages={30} currentPage={page} onClick={setPage} />
}`}
      />
    </>
  )
}

function ExamplePagination() {
  const [page, setPage] = useState(1)

  return <Pagination totalPages={30} currentPage={page} onClick={setPage} />
}
