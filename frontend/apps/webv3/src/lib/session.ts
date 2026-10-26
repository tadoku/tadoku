import { queryOptions } from '@tanstack/react-query'
import { rootRouteId, useHydrated, useLoaderData, useRouterState } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'

export type Viewer =
  | { status: 'anonymous' }
  | { status: 'signed-in'; userId: string; displayName: string; role: 'user' | 'admin' }
  | { status: 'banned'; userId: string; displayName: string }
  | { status: 'unavailable' }

export const getViewer = createServerFn({ method: 'GET' }).handler(async (): Promise<Viewer> => {
  const { readViewer } = await import('./session.server')
  return readViewer()
})

export const viewerQueryOptions = () => queryOptions({
  queryKey: ['viewer'],
  queryFn: () => getViewer(),
  staleTime: 30_000,
})

/** Identifies whose view a cache entry of personalized data holds. */
export function viewerKey(viewer: Viewer | undefined) {
  return viewer && 'userId' in viewer ? viewer.userId : viewer?.status ?? 'anonymous'
}

/** The authv2 login URL that returns to the current page. */
export function useLoginUrl() {
  const { accountUrl, pageUrl } = useLoaderData({ from: rootRouteId })
  const hydrated = useHydrated()
  // Re-render on every navigation: in the browser the return URL is the full current URL, fragment included, which
  // the server never receives.
  useRouterState({ select: state => state.location.href })

  const current = hydrated ? window.location.href : pageUrl
  return `${accountUrl}/login?return_to=${encodeURIComponent(current)}`
}
