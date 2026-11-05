import { rootRouteId, useLoaderData } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'
import { useMemo } from 'react'

/** Where links lead for this request: Webv2, the account app and the requested page itself. */
export const getPageUrls = createServerFn({ method: 'GET' }).handler(async () => {
  const { requestUrls } = await import('./urls.server')
  const { getRequest } = await import('@tanstack/react-start/server')
  const { webUrl, appOrigin, accountUrl } = requestUrls()
  const request = new URL(getRequest().url)
  return { webUrl, accountUrl, pageUrl: appOrigin + request.pathname + request.search }
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
    contests: '/contests/official/1',
    newContest: legacyUrl('/contests/new'),
    leaderboard: '/leaderboard/latest/1',
    guide: '/pages/manual',
    blog: '/blog/1',
    /** A CMS page, such as `about` or `faq`. */
    page: (slug: string) => `/pages/${encodeURIComponent(slug)}`,
    logs: '/logs/new',
    contest: (id: string, page = 1) => `/contests/${encodeURIComponent(id)}/leaderboard/${page}`,
    joinContest: (id: string) => `/contests/${encodeURIComponent(id)}/registration`,
    contestUpdates: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/updates/1`),
    contestScoring: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/scoring`),
    profile: (id: string) => `/profile/${encodeURIComponent(id)}/statistics`,
    /** A profile's updates, optionally only one UTC day's (`YYYY-MM-DD`). */
    profileUpdates: (id: string, date?: string) =>
      `/profile/${encodeURIComponent(id)}/updates/1${date ? `?date=${date}` : ''}`,
    contestProfile: (contestId: string, userId: string) =>
      legacyUrl(`/contests/${encodeURIComponent(contestId)}/profile/${encodeURIComponent(userId)}/1`),
    /** The signed-in account's own logs as a download, through Webv2's API path. */
    logExport: (format: 'csv' | 'json') => legacyUrl(`/api/internal/immersion/logs/export?format=${format}`),
  }
}
