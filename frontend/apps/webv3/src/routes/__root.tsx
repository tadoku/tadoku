import { createRootRouteWithContext, HeadContent, Scripts, ScriptOnce } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
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
        // `public/favicon.svg` switches its fill with the visitor's appearance; the PNG and `/favicon.ico` are for browsers
        // without SVG icons. All are drawn from paper-ui's brand `favicon` sources.
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/png', sizes: '48x48', href: '/favicon.png' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' },
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
        {/* The header's paper surface in each appearance, so the browser's bar continues it. Kept out of head(), which
            keeps one meta per name. */}
        <meta name="theme-color" content="#f4f4f4" media="(prefers-color-scheme: light)" />
        <meta name="theme-color" content="#1c1c1c" media="(prefers-color-scheme: dark)" />
        <HeadContent />
      </head>
      <body className={bodyClassName}>
        <AppShell>{children}</AppShell>
        <Scripts />
      </body>
    </html>
  )
}
