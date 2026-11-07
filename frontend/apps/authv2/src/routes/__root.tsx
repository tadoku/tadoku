import { createRootRouteWithContext, HeadContent, Outlet, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import { rootProps, themeBootScript } from 'paper-ui'
import favicon from 'paper-ui/assets/brand/favicon.svg?url'
import type { ReactNode } from 'react'
import { Shell, themeKey } from '../components/Shell'
import { NotFound, Unavailable } from '../components/StatusViews'
import { getSession } from '../lib/session'
// Paper's font faces; the StyleX plugin appends every compiled style to this stylesheet.
import 'paper-ui/fonts.css'

// The development server serves StyleX's styles separately and reloads them as sources change.
if (import.meta.env.DEV && typeof window !== 'undefined') void import('virtual:stylex:runtime')

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  // /deleted confirms a deletion the API has already accepted, so it neither reads nor waits for the session.
  beforeLoad: async ({ location }) => ({ identity: await getSession({ data: location.pathname !== '/deleted' }) }),
  head: () => ({
    meta: [{ title: 'Account · Tadoku' }, { name: 'robots', content: 'noindex, nofollow' }],
    links: [
      ...(import.meta.env.DEV ? [{ rel: 'stylesheet', href: '/virtual:stylex.css' }] : []),
      { rel: 'icon', type: 'image/svg+xml', href: favicon },
    ],
  }),
  component: AccountLayout,
  shellComponent: RootDocument,
  notFoundComponent: NotFound,
})

function AccountLayout() {
  const { identity } = Route.useRouteContext()
  return <Shell identity={identity}>{identity.status === 'unavailable' ? <Unavailable /> : <Outlet />}</Shell>
}

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning {...rootProps.html()}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <script dangerouslySetInnerHTML={{ __html: themeBootScript(themeKey) }} />
        <HeadContent />
      </head>
      <body {...rootProps.body()}>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
