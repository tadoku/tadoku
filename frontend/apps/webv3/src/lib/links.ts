import { useRouteContext } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'

export const getWebUrl = createServerFn({ method: 'GET' }).handler(async () => {
  const { requestUrls } = await import('./urls.server')
  return requestUrls().webUrl
})

export function useLinks() {
  const { webUrl } = useRouteContext({ strict: false }) as { webUrl?: string }
  if (!webUrl) throw new Error('Missing Webv3 request URLs')
  return createLinks(webUrl)
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

export function legacyUrl(path: string) {
  const base = import.meta.env.VITE_TADOKU_WEB_URL || 'https://tadoku.dev.lab'
  return new URL(path, base).href
}

export const links = {
  home: legacyUrl('/'),
  contests: legacyUrl('/contests/official/1'),
  leaderboard: legacyUrl('/leaderboard/latest/1'),
  guide: legacyUrl('/pages/manual'),
  logs: legacyUrl('/logs/new'),
  contest: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/leaderboard/1`),
  joinContest: (id: string) => legacyUrl(`/contests/${encodeURIComponent(id)}/registration`),
  profile: (id: string) => legacyUrl(`/profile/${encodeURIComponent(id)}/statistics`),
}
