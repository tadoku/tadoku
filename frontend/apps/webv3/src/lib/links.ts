import { rootRouteId, useLoaderData } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'
import { useMemo } from 'react'

/** Where links lead for this request: the public web origin, the account and admin apps and the requested page itself. */
export const getPageUrls = createServerFn({ method: 'GET' }).handler(async () => {
  const { requestUrls } = await import('./urls.server')
  const { getRequest } = await import('@tanstack/react-start/server')
  const { webUrl, appOrigin, accountUrl, adminUrl } = requestUrls()
  const request = new URL(getRequest().url)
  return { webUrl, accountUrl, adminUrl, pageUrl: appOrigin + request.pathname + request.search }
})

export function useLinks() {
  const { webUrl } = useLoaderData({ from: rootRouteId })
  return useMemo(() => createLinks(webUrl), [webUrl])
}

function createLinks(base: string) {
  return {
    home: '/',
    contests: '/contests/official/1',
    newContest: '/contests/new',
    leaderboard: '/leaderboard/latest/1',
    guide: '/pages/manual',
    blog: '/blog/1',
    /** A CMS page, such as `about` or `faq`. */
    page: (slug: string) => `/pages/${encodeURIComponent(slug)}`,
    logs: '/logs/new',
    contest: (id: string, page = 1) => `/contests/${encodeURIComponent(id)}/leaderboard/${page}`,
    joinContest: (id: string) => `/contests/${encodeURIComponent(id)}/registration`,
    contestUpdates: (id: string, page = 1) => `/contests/${encodeURIComponent(id)}/updates/${page}`,
    contestScoring: (id: string) => `/contests/${encodeURIComponent(id)}/scoring`,
    profile: (id: string) => `/profile/${encodeURIComponent(id)}/statistics`,
    /** A profile's updates, optionally only one UTC day's (`YYYY-MM-DD`). */
    profileUpdates: (id: string, date?: string) =>
      `/profile/${encodeURIComponent(id)}/updates/1${date ? `?date=${date}` : ''}`,
    contestProfile: (contestId: string, userId: string, page = 1) =>
      `/contests/${encodeURIComponent(contestId)}/profile/${encodeURIComponent(userId)}/${page}`,
    /** The signed-in account's own logs as a download, from the API gateway on the public web origin. */
    logExport: (format: 'csv' | 'json') => new URL(`/api/internal/immersion/logs/export?format=${format}`, base).href,
  }
}
