import { createFileRoute } from '@tanstack/react-router'
import { Register } from '../components/Credentials'
import { accountSearch } from '../lib/search'

export const Route = createFileRoute('/register')({
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Create an account · Tadoku' }] }),
  component: RouteComponent,
})

function RouteComponent() {
  return <Register search={Route.useSearch()} />
}
