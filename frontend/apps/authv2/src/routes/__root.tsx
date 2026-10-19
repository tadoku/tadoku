import { createRootRouteWithContext, HeadContent, Scripts, Outlet, redirect } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { getSession } from '../lib/session'
import 'paper-ui/styles.css'
import '../styles.css'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  beforeLoad: async ({ location }) => {
    const identity = await getSession()
    if (identity.status === 'ready' && identity.session.account_app === 'v1') {
      const target = new URL(identity.accountUrl)
      target.pathname = location.pathname
      target.search = location.searchStr
      target.hash = location.hash
      throw redirect({ href: target.href, statusCode: 307 })
    }
    return { identity }
  },
  head: () => ({ meta: [{ title: 'Tadoku · Account' }, { name: 'robots', content: 'noindex, nofollow' }] }),
  component: AccountLayout,
  shellComponent: RootDocument,
  notFoundComponent: () => <main className="account-content"><h1>Page not found</h1><a href="/">Account home</a></main>,
})

function AccountLayout() {
  const { identity } = Route.useRouteContext()
  return <div className="account-shell">
    <header><a href={identity.webUrl} aria-label="Tadoku home">Tadoku</a></header>
    <main id="main" className="account-content"><Outlet /></main>
  </div>
}

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <HeadContent />
      </head>
      <body><a className="skip-link" href="#main">Skip to content</a>{children}<Scripts /></body>
    </html>
  )
}
