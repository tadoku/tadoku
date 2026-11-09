import { QueryClient } from '@tanstack/react-query'
import { createRouter } from '@tanstack/react-router'
import { getGlobalStartContext } from '@tanstack/react-start'
import { setupRouterSsrQueryIntegration } from '@tanstack/react-router-ssr-query'
import { routeAnnouncerId } from './components/AppShell'
import { ErrorView, PendingView } from './components/StatusViews'
import { routeTree } from './routeTree.gen'

export function getRouter() {
  const queryClient = new QueryClient({
    defaultOptions: {
      dehydrate: {
        shouldDehydrateQuery: (query) => query.state.status !== 'pending',
      },
    },
  })

  const router = createRouter({
    routeTree,
    context: { queryClient },
    defaultPreload: 'intent',
    defaultErrorComponent: ErrorView,
    defaultPendingComponent: PendingView,
    defaultPendingMs: 400,
    scrollRestoration: true,
    ssr: { nonce: cspNonce() },
  })

  setupRouterSsrQueryIntegration({ router, queryClient })
  if (typeof document !== 'undefined') router.subscribe('onResolved', ({ fromLocation, pathChanged }) => {
    if (fromLocation && pathChanged) requestAnimationFrame(() => setTimeout(arrive))
  })
  return router
}

/**
 * The request's CSP nonce from `start.ts`, on the server. The browser reads it back from the page's `csp-nonce` meta tag.
 * The router can be created before the request middleware runs, such as for a server function's redirect.
 */
function cspNonce() {
  try {
    return getGlobalStartContext()?.cspNonce
  } catch {
    return undefined
  }
}

/**
 * After a client navigation, moves focus to the new page's title, or its main region without one, and announces the
 * page title, as a full page load would for keyboard and screen reader users. The first load keeps the browser's focus.
 */
function arrive() {
  const target = document.querySelector<HTMLElement>('main h1') ?? document.querySelector<HTMLElement>('main')
  if (target && !target.hasAttribute('tabindex')) target.setAttribute('tabindex', '-1')
  target?.focus({ preventScroll: true })
  const announcer = document.getElementById(routeAnnouncerId)
  if (announcer) announcer.textContent = document.title
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof getRouter>
  }
}
