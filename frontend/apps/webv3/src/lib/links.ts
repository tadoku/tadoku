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
