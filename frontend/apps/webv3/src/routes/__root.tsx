import { createRootRouteWithContext, HeadContent, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import mark from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import { AppShell } from '../components/AppShell'
import { ErrorView, NotFoundView, PendingView } from '../components/StatusViews'
import { getPageUrls } from '../lib/links'
import { viewerQueryOptions } from '../lib/session'
import { bodyClassName, documentClassName, themeBootScript } from 'paper-ui'
import '../styles.css'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  loader: async ({ context }) => {
    const [urls] = await Promise.all([getPageUrls(), context.queryClient.ensureQueryData(viewerQueryOptions())])
    return urls
  },
  staleTime: Infinity,
  head: () => ({
    links: [
      { rel: 'icon', type: 'image/svg+xml', href: mark, media: '(prefers-color-scheme: light)' },
      { rel: 'icon', type: 'image/svg+xml', href: reversedMark, media: '(prefers-color-scheme: dark)' },
    ],
  }),
  notFoundComponent: NotFoundView,
  errorComponent: ErrorView,
  pendingComponent: PendingView,
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={documentClassName} suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <script dangerouslySetInnerHTML={{ __html: themeBootScript }} />
        <HeadContent />
      </head>
      <body className={bodyClassName}>
        <AppShell>{children}</AppShell>
        <Scripts />
      </body>
    </html>
  )
}
