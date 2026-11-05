import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestFindById,
  immersionContestFindLatestOfficial,
  immersionContestListLogs,
  immersionFetchLeaderboardForYear,
  immersionFetchLeaderboardGlobal,
  type ImmersionContestView,
  type ImmersionLeaderboard,
} from 'api-client'
import type { BoardRead, Read, StandingsScope } from './leaderboard'
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

const failed = { data: undefined, response: undefined }

export async function readLatestBoard(): Promise<BoardRead> {
  return readBoard(await immersionContestFindLatestOfficial(sessionRequest()).catch(() => failed))
}

/** Reads the contest as the visitor, whose role decides whether a deleted contest is found; a rejected session reads it anonymously. */
export async function readContestBoard(id: string): Promise<BoardRead> {
  const cookie = sessionCookie()
  const find = (session?: string) => immersionContestFindById({ ...sessionRequest(session), path: { id } }).catch(() => failed)
  const found = await find(cookie)
  return readBoard(found.response?.status === 401 && cookie ? await find() : found)
}

async function readBoard(found: { data?: ImmersionContestView; response?: Response }): Promise<BoardRead> {
  if (found.response?.status === 404) return { kind: 'none' }
  const contest = found.data
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
