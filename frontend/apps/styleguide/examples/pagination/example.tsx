import { Pagination } from 'ui'
import { useState } from 'react'

export default function PaginationExample() {
  const [page, setPage] = useState(1)

  return <Pagination totalPages={30} currentPage={page} onClick={setPage} />
}
