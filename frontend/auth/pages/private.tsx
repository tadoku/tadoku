import type { NextPage } from 'next'
import { useProtectedRoute } from '../src/session'

const Private: NextPage = () => {
  useProtectedRoute()

  return (
    <div>
      <h1>Private page</h1>
      This page is only available to authenticated users.
    </div>
  )
}

export default Private
