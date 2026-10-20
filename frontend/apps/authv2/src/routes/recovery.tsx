import { createFileRoute } from '@tanstack/react-router'
import { Recovery } from '../components/Recovery'
import { accountSearch } from '../lib/search'

export const Route = createFileRoute('/recovery')({
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Reset your password · Tadoku' }] }),
  component: RouteComponent,
})

function RouteComponent() {
  return <Recovery search={Route.useSearch()} />
}
