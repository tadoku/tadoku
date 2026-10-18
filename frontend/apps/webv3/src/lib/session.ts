import { queryOptions } from '@tanstack/react-query'
import { rootRouteId, useHydrated, useLoaderData } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'
import { z } from 'zod'

export const viewerSchema = z.discriminatedUnion('status', [
  z.object({ status: z.literal('anonymous') }),
  z.object({ status: z.literal('signed-in'), userId: z.uuid(), displayName: z.string(), role: z.enum(['user', 'admin']) }),
  z.object({ status: z.literal('banned'), userId: z.uuid(), displayName: z.string() }),
  z.object({ status: z.literal('unavailable') }),
])
export type Viewer = z.infer<typeof viewerSchema>

export const getViewer = createServerFn({ method: 'GET' }).handler(async (): Promise<Viewer> => {
  const { readViewer } = await import('./session.server')
  return readViewer()
})

export const viewerQueryOptions = () => queryOptions({
  queryKey: ['viewer'], queryFn: () => getViewer(), staleTime: 30_000,
})

export function useLoginUrl() {
  const { authUrl, appOrigin, requestPath } = useLoaderData({ from: rootRouteId })
  const hydrated = useHydrated()
  return `${authUrl}/login?return_to=${encodeURIComponent(hydrated ? window.location.href : appOrigin + requestPath)}`
}
