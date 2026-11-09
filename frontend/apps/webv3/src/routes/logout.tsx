import { createFileRoute, notFound } from '@tanstack/react-router'
import { Button, EmptyState } from 'paper-ui'
import { PageUnavailableError } from '../lib/page-status'
import { noindex } from '../lib/page-meta'

/**
 * The account menu's Log out form posts here. A logout the API did not confirm falls through to this route's page,
 * which answers 503 in the app shell; a GET finds no page.
 */
export const Route = createFileRoute('/logout')({
  server: {
    handlers: {
      POST: async ({ request, next }) => {
        const { logOut } = await import('../lib/logout.server')
        return (await logOut(request)) ?? next({ context: { logout: 'unavailable' as const } })
      },
    },
  },
  loader: ({ serverContext }) => {
    if (serverContext?.logout !== 'unavailable') throw notFound()
    throw new PageUnavailableError('Log out is unavailable')
  },
  head: () => ({ meta: [{ title: 'Log out · Tadoku' }, noindex] }),
  component: LogoutUnavailable,
  errorComponent: LogoutUnavailable,
})

function LogoutUnavailable() {
  return (
    <EmptyState
      headingLevel={1}
      title="Log out is unavailable"
      description="You may still be signed in. Try again, or return to the previous page."
      actions={<form method="post" action="/logout"><Button type="submit">Try again</Button></form>}
    />
  )
}
