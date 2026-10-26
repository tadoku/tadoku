import { createFileRoute, redirect } from '@tanstack/react-router'
import { DeleteAccount } from '../components/DeleteAccount'
import { continuation, deleteAccountLogin } from '../lib/identity'
import { accountSearch } from '../lib/search'

export const Route = createFileRoute('/delete-account')({
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Delete your account · Tadoku' }] }),
  // Runs while the server renders too, so guests reach login without JavaScript.
  beforeLoad: ({ context: { identity }, location }) => {
    if (identity.status !== 'ready' || identity.session.account) return
    const next = continuation(accountSearch(location.search).return_to, identity.origin)
    throw redirect({ href: deleteAccountLogin(identity.origin, next, false) })
  },
  component: RouteComponent,
})

function RouteComponent() {
  return <DeleteAccount search={Route.useSearch()} />
}
