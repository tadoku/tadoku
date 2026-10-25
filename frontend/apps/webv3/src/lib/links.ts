import { rootRouteId, useLoaderData } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'
import { useMemo } from 'react'

export const getWebUrl = createServerFn({ method: 'GET' }).handler(async () => {
  const { requestUrls } = await import('./urls.server')
  const { webUrl, appOrigin, accountUrl } = requestUrls()
  const { getRequest } = await import('@tanstack/react-start/server')
  const request = new URL(getRequest().url)
  return { webUrl, appOrigin, accountUrl, requestPath: request.pathname + request.search }
})

export function useLinks() {
  const { webUrl } = useLoaderData({ from: rootRouteId })
  return useMemo(() => createLinks(webUrl), [webUrl])
}

function createLinks(base: string) {
  const legacyUrl = (path: string) => new URL(path, base).href
  return {
    legacyUrl,
    home: legacyUrl('/'),
    contests: legacyUrl('/contests/official/1'),
    leaderboard: legacyUrl('/leaderboard/latest/1'),
    guide: legacyUrl('/pages/manual'),
    logs: legacyUrl('/logs/new'),
    contest: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/leaderboard/1`),
    joinContest: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/registration`),
    profile: (id: string) => legacyUrl(`/profile/${encodeURIComponent(id)}/statistics`),
  }
}
