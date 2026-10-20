import { createRootRouteWithContext, HeadContent, Outlet, Scripts, redirect } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import favicon from 'paper-ui/assets/brand/favicon.svg?url'
import type { ReactNode } from 'react'
import { Shell, themeBootScript } from '../components/Shell'
import { NotFound, Unavailable } from '../components/StatusViews'
import { getSession } from '../lib/session'
import 'paper-ui/styles.css'
import '../styles.css'

/** The existing account app keeps Settings at its root and recovery at /account-recovery. */
function v1Path(pathname: string) {
  if (pathname === '/settings') return '/'
  if (pathname === '/recovery') return '/account-recovery'
  return pathname
}

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  beforeLoad: async ({ location }) => {
    const identity = await getSession()
    if (identity.status === 'ready' && identity.session.account_app === 'v1') {
      const target = new URL(identity.accountUrl)
      target.pathname = v1Path(location.pathname)
      target.search = location.searchStr
      target.hash = location.hash
      throw redirect({ href: target.href, statusCode: 307 })
    }
    return { identity }
  },
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
