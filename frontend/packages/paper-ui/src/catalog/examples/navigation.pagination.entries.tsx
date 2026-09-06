import { Pagination } from 'paper-ui'
import { useState } from 'react'

export default function PaginationExample() {
  const [page, setPage] = useState(8)
  return (
    <div className="paper-stack">
      <p role="status">
        Page {page} of 24 · Entries {(page - 1) * 10 + 1}–{page * 10}
      </p>
      <Pagination totalPages={24} currentPage={page} onPageChange={setPage} />
      <h3>One page</h3>
      <p>
        Keep pagination out of an empty result. This boundary example has no
        available previous or next page.
      </p>
      <Pagination
        label="Single-page results"
        totalPages={1}
        currentPage={1}
        getHref={() => '#results'}
      />
    </div>
  )
}
