import { createRootRouteWithContext, HeadContent, Scripts } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { rootProps, themeBootScript } from 'paper-ui'
import mark from 'paper-ui/assets/brand/cut-meter.svg?url'
import reversedMark from 'paper-ui/assets/brand/cut-meter-reversed.svg?url'
import { AppShell, themeKey } from '../components/AppShell'
import { ErrorView, NotFoundView, PendingView } from '../components/StatusViews'
import { getPageUrls } from '../lib/links'
import { viewerQueryOptions } from '../lib/session'
// Paper's font faces; the StyleX plugin appends every compiled style to this stylesheet.
import 'paper-ui/fonts.css'

// The development server serves StyleX's styles separately and reloads them as sources change.
if (import.meta.env.DEV && typeof window !== 'undefined') void import('virtual:stylex:runtime')

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  loader: async ({ context }) => {
    const [urls] = await Promise.all([getPageUrls(), context.queryClient.ensureQueryData(viewerQueryOptions())])
    return urls
  },
  staleTime: Infinity,
  head: () => ({
    links: [
      ...(import.meta.env.DEV ? [{ rel: 'stylesheet', href: '/virtual:stylex.css' }] : []),
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
    <html lang="en" suppressHydrationWarning {...rootProps.html()}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <script dangerouslySetInnerHTML={{ __html: themeBootScript(themeKey) }} />
        <HeadContent />
      </head>
      <body {...rootProps.body()}>
        <AppShell>{children}</AppShell>
        <Scripts />
      </body>
    </html>
  )
}
