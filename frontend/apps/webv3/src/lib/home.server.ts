import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestList,
} from 'api-client'
import type { HomeBoard, NextContestRead } from './home'
import { publicReadRequest } from './session.server'

/** The validated answer, or null when the API has nothing to show (404). */
async function found<T>(request: Promise<{ data?: T; response?: Response }>): Promise<T | null> {
  const { data, response } = await request
  if (data !== undefined) return data
  if (response?.status === 404) return null
  throw new Error('Tadoku data is unavailable. Try again.')
}

/** The summary and top of the board for the latest contest; home data is public, so these reads carry no session. */
export async function readHomeBoard(contestId: string): Promise<HomeBoard> {
  const path = { id: contestId }
  const [summary, standings] = await Promise.all([
    found(immersionContestFetchSummary({ ...publicReadRequest('immersionContestFetchSummary'), path })),
    found(immersionContestFetchLeaderboard({
      ...publicReadRequest('immersionContestFetchLeaderboard'),
      path,
      query: { page_size: 5, page: 0 },
    })),
  ])
  return { summary, standings }
}

/** The soonest official contest that hasn't started, or null when none is scheduled. */
export async function readNextContest(): Promise<NextContestRead> {
  const { data } = await immersionContestList({
    ...publicReadRequest('immersionContestList'),
    query: { upcoming: true, official: true, include_deleted: false, page_size: 1 },
  }).catch(() => ({ data: undefined }))
  if (!data) return { ok: false }

  const [next] = data.contests
  if (!next) return { ok: true, data: null }
  // Home links need contest IDs; an answer without one is malformed.
  return next.id ? { ok: true, data: { ...next, id: next.id } } : { ok: false }
}
