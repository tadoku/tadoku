import {
  immersionContestFetchLeaderboard,
  immersionContestFetchLeaderboardEntry,
  immersionContestFindById,
  immersionContestFindLatestOfficial,
  immersionFetchLeaderboardForYear,
  immersionFetchLeaderboardGlobal,
  type ImmersionContestView,
  type ImmersionLeaderboard,
  type ImmersionLeaderboardEntry,
} from 'api-client'
import type { BoardRead, Read, StandingsScope } from './leaderboard'
import { standingsPageSize } from './leaderboard'
import { readAsVisitor, sessionRequest } from './session.server'

async function read<T>(request: Promise<{ data?: T }>): Promise<Read<T>> {
  const { data } = await request.catch(() => ({ data: undefined }))
  return data === undefined ? { ok: false } : { ok: true, data }
}

/** Standings are public, so they are read anonymously. */
export function readStandings(scope: StandingsScope, page: number): Promise<Read<ImmersionLeaderboard>> {
  const query = { page_size: standingsPageSize, page: page - 1 }
  switch (scope.kind) {
    case 'contest':
      return read(immersionContestFetchLeaderboard({ ...sessionRequest('immersionContestFetchLeaderboard'), path: { id: scope.contestId }, query }))
    case 'year':
      return read(immersionFetchLeaderboardForYear({ ...sessionRequest('immersionFetchLeaderboardForYear'), path: { year: scope.year }, query }))
    case 'all-time':
      return read(immersionFetchLeaderboardGlobal({ ...sessionRequest('immersionFetchLeaderboardGlobal'), query }))
    default: {
      const exhaustive: never = scope
      return exhaustive
    }
  }
}

const failed = { data: undefined, response: undefined }

export async function readLatestBoard(): Promise<BoardRead> {
  return readBoard(await immersionContestFindLatestOfficial(sessionRequest('immersionContestFindLatestOfficial')).catch(() => failed))
}

/** Reads the contest as the visitor, whose role decides whether a deleted contest is found. */
export async function readContestBoard(id: string): Promise<BoardRead> {
  return readBoard(await readAsVisitor(cookie =>
    immersionContestFindById({ ...sessionRequest('immersionContestFindById', cookie), path: { id } }).catch(() => failed)))
}

/** One participant's place on a contest's board at any rank, or null when they aren't on it. */
export async function readContestEntry(contestId: string, userId: string): Promise<Read<ImmersionLeaderboardEntry | null>> {
  const { data, response } = await immersionContestFetchLeaderboardEntry({
    ...sessionRequest('immersionContestFetchLeaderboardEntry'),
    path: { id: contestId, user_id: userId },
  }).catch(() => failed)
  if (response?.status === 204 || response?.status === 404) return { ok: true, data: null }
  return data ? { ok: true, data } : { ok: false }
}

function readBoard(found: { data?: ImmersionContestView; response?: Response }): BoardRead {
  if (found.response?.status === 404) return { kind: 'none' }
  const contest = found.data
  if (!contest?.id) return { kind: 'unavailable' }
  return { kind: 'board', board: { contest: { ...contest, id: contest.id } } }
}
