import { createRootRouteWithContext, HeadContent, Outlet, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import favicon from 'paper-ui/assets/brand/favicon.svg?url'
import type { ReactNode } from 'react'
import { Shell, themeBootScript } from '../components/Shell'
import { NotFound, Unavailable } from '../components/StatusViews'
import { getSession } from '../lib/session'
import 'paper-ui/styles.css'
import '../styles.css'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  beforeLoad: async () => ({ identity: await getSession() }),
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
  return <Shell identity={identity}>{identity.status === 'ready' ? <Outlet /> : <Unavailable />}</Shell>
}

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <script dangerouslySetInnerHTML={{ __html: themeBootScript }} />
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
