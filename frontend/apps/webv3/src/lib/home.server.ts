import { getCookie, setCookie } from '@tanstack/react-start/server'
import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestList,
  immersionProfileFindByUserId,
  immersionProfileYearlyActivityByUserId,
} from 'api-client'
import type { HomeBoard, NextContestRead } from './home'
import { readContestHistory } from './profile.server'
import { clearedGettingGoing } from './recent-activity'
import { publicReadRequest, sessionRequest } from './session.server'

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

const clearedCookie = (userId: string) => `home_cleared_${userId}`
/** The longest lifetime browsers keep a cookie for. */
const clearedCookieMaxAge = 400 * 86_400

/**
 * Whether the learner cleared Getting going in any year since their account was created, read from every year's
 * activity. A yes is final, so it is remembered in a cookie for the account and later calls read nothing. A read that
 * fails counts as no activity.
 */
export async function readGettingGoingCleared(userId: string): Promise<boolean> {
  if (getCookie(clearedCookie(userId))) return true

  const profile = await immersionProfileFindByUserId({ ...sessionRequest('immersionProfileFindByUserId'), path: { userId } })
    .catch(() => undefined)
  if (!profile?.data) return false

  const first = new Date(profile.data.created_at).getUTCFullYear()
  const years = Array.from({ length: Math.max(0, new Date().getUTCFullYear() - first + 1) }, (_, index) => first + index)
  const activity = await Promise.all(years.map(year => immersionProfileYearlyActivityByUserId({
    ...sessionRequest('immersionProfileYearlyActivityByUserId'),
    path: { userId, year },
  }).then(({ data }) => data?.scores ?? [], () => [])))
  const scores = new Map(activity.flat().map(day => [day.date, day.score]))

  // Contests only matter for two days in a row, so they are read only when the activity alone doesn't clear it.
  const cleared = clearedGettingGoing(scores, [])
    || (await readContests(userId, years).then(contests => clearedGettingGoing(scores, contests)))
  if (cleared) {
    setCookie(clearedCookie(userId), '1', { maxAge: clearedCookieMaxAge, path: '/', sameSite: 'lax', secure: true, httpOnly: true })
  }
  return cleared
}

/** The contests the learner registered for in `years`, private ones included, as they see their own history. */
async function readContests(userId: string, years: number[]) {
  // ponytail: the first 50 registrations of each year, as the profile lists them; page further if anyone joins more.
  const history = await Promise.all(years.map(year => readContestHistory(userId, year, 1)))
  return history.flatMap(section => (section.ok ? section.data.contests : []))
}
