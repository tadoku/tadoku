import { queryOptions, type FetchQueryOptions, type QueryClient, type QueryKey } from '@tanstack/react-query'
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

/**
 * Prefetches a read cached per viewer without waiting for the viewer: the server answers from the request's own session,
 * so only the cache key needs them. Use it in loaders so a page's public reads start alongside the viewer lookup.
 */
export async function ensureViewerData<T, Key extends QueryKey>(
  queryClient: QueryClient,
  query: (key: string) => FetchQueryOptions<T, Error, T, Key>,
): Promise<readonly [Viewer, T]> {
  const cached = queryClient.getQueryData<Viewer>(viewerQueryOptions().queryKey)
  if (cached) return [cached, await queryClient.ensureQueryData(query(viewerKey(cached)))]

  const provisional = query('pending-viewer')
  const [viewer, data] = await Promise.all([
    queryClient.ensureQueryData(viewerQueryOptions()),
    queryClient.ensureQueryData(provisional),
  ])
  queryClient.setQueryData<T>(query(viewerKey(viewer)).queryKey as QueryKey, data as T)
  queryClient.removeQueries({ queryKey: provisional.queryKey, exact: true })
  return [viewer, data]
}

/** The authv2 login URL on `accountUrl` that returns to `returnTo`. */
export const loginHref = (accountUrl: string, returnTo: string) => `${accountUrl}/login?return_to=${encodeURIComponent(returnTo)}`

/** The authv2 login URL that returns to the current page. */
export function useLoginUrl() {
  return useAccountUrls().login
}

/** The authv2 login, registration and settings URLs, each returning to the current page. */
export function useAccountUrls() {
  const { accountUrl, pageUrl } = useLoaderData({ from: rootRouteId })
  const hydrated = useHydrated()
  // Re-render on every navigation: in the browser the return URL is the full current URL, fragment included, which
  // the server never receives.
  useRouterState({ select: state => state.location.href })

  const current = hydrated ? window.location.href : pageUrl
  const returnTo = encodeURIComponent(current)
  return {
    login: loginHref(accountUrl, current),
    register: `${accountUrl}/register?return_to=${returnTo}`,
    settings: `${accountUrl}/settings?return_to=${returnTo}`,
  }
}
