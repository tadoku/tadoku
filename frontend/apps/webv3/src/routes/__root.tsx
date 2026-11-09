import { createRootRouteWithContext, HeadContent, Scripts, ScriptOnce } from '@tanstack/react-router'
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
  // The canonical URL is on the web host, without the query or fragment, so every host and share link points at one page.
  head: ({ loaderData, matches }) => {
    const canonical = loaderData && `${loaderData.webUrl}${matches[matches.length - 1]?.pathname ?? '/'}`
    return {
      meta: canonical ? [{ property: 'og:site_name', content: 'Tadoku' }, { property: 'og:url', content: canonical }] : [],
      links: [
        ...(canonical ? [{ rel: 'canonical', href: canonical }] : []),
        { rel: 'icon', type: 'image/svg+xml', href: mark, media: '(prefers-color-scheme: light)' },
        { rel: 'icon', type: 'image/svg+xml', href: reversedMark, media: '(prefers-color-scheme: dark)' },
      ],
    }
  },
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
        <ScriptOnce>{themeBootScript}</ScriptOnce>
        <HeadContent />
      </head>
      <body className={bodyClassName}>
        <AppShell>{children}</AppShell>
        <Scripts />
      </body>
    </html>
  )
}
