import { createRootRouteWithContext, HeadContent, Outlet, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import favicon from 'paper-ui/assets/brand/favicon.svg?url'
import type { ReactNode } from 'react'
import { Shell, themeBootScript } from '../components/Shell'
import { NotFound, Unavailable } from '../components/StatusViews'
import { getSession } from '../lib/session'
import { bodyClassName, documentClassName } from 'paper-ui'
import '../styles.css'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  // /deleted confirms a deletion the API has already accepted, so it neither reads nor waits for the session.
  beforeLoad: async ({ location }) => ({ identity: await getSession({ data: location.pathname !== '/deleted' }) }),
  head: () => ({
    meta: [{ title: 'Account · Tadoku' }, { name: 'robots', content: 'noindex, nofollow' }],
    links: [{ rel: 'icon', type: 'image/svg+xml', href: favicon }],
  }),
  component: AccountLayout,
  shellComponent: RootDocument,
  notFoundComponent: NotFound,
})

function AccountLayout() {
  const { identity } = Route.useRouteContext()
  return <Shell identity={identity}>{identity.status === 'unavailable' ? <Unavailable /> : <Outlet />}</Shell>
}

// Account pages colour every plain link; Paper's buttons keep their own colours.
const htmlClassName = `${documentClassName} [:where(&_a:not(.paper-button))]:text-link`

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={htmlClassName} suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <script dangerouslySetInnerHTML={{ __html: themeBootScript }} />
        <HeadContent />
      </head>
      <body className={bodyClassName}>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
