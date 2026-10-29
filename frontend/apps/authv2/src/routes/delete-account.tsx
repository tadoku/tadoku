import { createFileRoute, redirect } from '@tanstack/react-router'
import { DeleteAccount } from '../components/DeleteAccount'
import { continuation, deleteAccountLogin } from '../lib/identity'
import { accountSearch } from '../lib/search'

export const Route = createFileRoute('/delete-account')({
  validateSearch: accountSearch,
  head: () => ({ meta: [{ title: 'Delete your account · Tadoku' }] }),
  // Runs while the server renders too, so guests and old logins reach login without JavaScript. An API that doesn't
  // send recent_login yet leaves the old-login check to its stale_session answer.
  beforeLoad: ({ context: { identity }, location }) => {
    if (identity.status !== 'ready') return
    const { account, recent_login } = identity.session
    if (account && recent_login !== false) return
    const next = continuation(accountSearch(location.search).return_to, identity.origin)
    throw redirect({ href: deleteAccountLogin(identity.origin, next, Boolean(account)) })
  },
  component: RouteComponent,
})

function RouteComponent() {
  return <DeleteAccount search={Route.useSearch()} />
}
