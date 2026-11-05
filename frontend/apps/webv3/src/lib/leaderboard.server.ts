import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestFindLatestOfficial,
  immersionContestFindRegistration,
  immersionContestListLogs,
  immersionFetchLeaderboardForYear,
  immersionFetchLeaderboardGlobal,
  type ImmersionLeaderboard,
} from 'api-client'
import type { LatestBoard, Read, StandingsScope } from './leaderboard'
import { standingsPageSize } from './leaderboard'
import { sessionCookie, sessionRequest } from './session.server'

async function read<T>(request: Promise<{ data?: T }>): Promise<Read<T>> {
  const { data } = await request.catch(() => ({ data: undefined }))
  return data === undefined ? { ok: false } : { ok: true, data }
}

/** Standings are public, so they are read anonymously. */
export function readStandings(scope: StandingsScope, page: number): Promise<Read<ImmersionLeaderboard>> {
  const query = { page_size: standingsPageSize, page: page - 1 }
  switch (scope.kind) {
    case 'contest':
      return read(immersionContestFetchLeaderboard({ ...sessionRequest(), path: { id: scope.contestId }, query }))
    case 'year':
      return read(immersionFetchLeaderboardForYear({ ...sessionRequest(), path: { year: scope.year }, query }))
    case 'all-time':
      return read(immersionFetchLeaderboardGlobal({ ...sessionRequest(), query }))
    default: {
      const exhaustive: never = scope
      return exhaustive
    }
  }
}

export async function readLatestBoard(): Promise<LatestBoard> {
  const latest = await immersionContestFindLatestOfficial(sessionRequest()).catch(() => ({ data: undefined, response: undefined }))
  if (latest.response?.status === 404) return { kind: 'none' }
  const contest = latest.data
  if (!contest?.id) return { kind: 'unavailable' }

  const path = { id: contest.id }
  const [summary, logs] = await Promise.all([
    read(immersionContestFetchSummary({ ...sessionRequest(), path })),
    read(immersionContestListLogs({ ...sessionRequest(), path, query: { page_size: 10, page: 0, include_deleted: false } })),
  ])
  return {
    kind: 'board',
    board: {
      contest: { ...contest, id: contest.id },
      summary: summary.ok ? summary.data : null,
      recentLogs: logs.ok ? logs.data.logs : [],
    },
  }
}

/** Whether the visitor has joined the contest: 204 means they haven't. */
export async function readRegistered(contestId: string): Promise<boolean | null> {
  const cookie = sessionCookie()
  if (!cookie) return null

  const { data, response } = await immersionContestFindRegistration({ ...sessionRequest(cookie), path: { id: contestId } })
    .catch(() => ({ data: undefined, response: undefined }))
  if (response?.status === 204) return false
  return data ? true : null
}
