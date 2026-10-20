import { createFileRoute } from '@tanstack/react-router'
import { Login } from '../components/Credentials'
import { accountSearch } from '../lib/search'

export const Route = createFileRoute('/login')({
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Log in · Tadoku' }] }),
  component: RouteComponent,
})

function RouteComponent() {
  return <Login search={Route.useSearch()} />
}
