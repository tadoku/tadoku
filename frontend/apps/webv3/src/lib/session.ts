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
  queryKey: ['viewer'], queryFn: () => getViewer(), staleTime: 30_000,
})

export function useLoginUrl() {
  const { accountUrl, appOrigin, requestPath } = useLoaderData({ from: rootRouteId })
  useRouterState({ select: state => state.location.href })
  const hydrated = useHydrated()

  return `${accountUrl}/login?return_to=${encodeURIComponent(hydrated ? window.location.href : appOrigin + requestPath)}`
}
