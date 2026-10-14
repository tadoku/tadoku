import { createRootRouteWithContext, HeadContent, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { AppShell, themeBootScript } from '../components/AppShell'
import { ErrorView, NotFoundView, PendingView } from '../components/StatusViews'
import { getWebUrl } from '../lib/links'
import 'paper-ui/styles.css'
import '../styles.css'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  beforeLoad: async () => ({ webUrl: await getWebUrl() }),
  notFoundComponent: NotFoundView,
  errorComponent: ErrorView,
  pendingComponent: PendingView,
  shellComponent: RootDocument,
})

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
        <AppShell>{children}</AppShell>
        <Scripts />
      </body>
    </html>
  )
}
