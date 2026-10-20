import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestFindLatestOfficial,
  immersionContestList,
  type ImmersionContest,
} from 'api-client'
import type { HomeData, Identified } from './home'
import { requestUrls } from './urls.server'

/** Home data is public, so requests carry no credentials. */
function anonymous() {
  return {
    baseUrl: requestUrls().apiUrl,
    credentials: 'omit' as const,
    headers: { Accept: 'application/json' },
    redirect: 'error' as const,
    signal: AbortSignal.timeout(10_000),
  }
}

/** The validated answer, or null when the API has nothing to show (404). */
async function found<T>(request: Promise<{ data?: T; response?: Response }>): Promise<T | null> {
  const { data, response } = await request
  if (data !== undefined) return data
  if (response?.status === 404) return null
  throw new Error('Tadoku data is unavailable. Try again.')
}

/** Home links need contest IDs; an answer without one is malformed. */
function identified<T extends { id?: string }>(contest: T | null): Identified<T> | null {
  if (contest && !contest.id) throw new Error('Tadoku data is unavailable. Try again.')
  return contest as Identified<T> | null
}

async function readNextContest(): Promise<Identified<ImmersionContest> | null> {
  const today = new Date().toISOString().slice(0, 10)
  let nextContest: ImmersionContest | null = null
  let page = 0

  while (true) {
    const result = await found(immersionContestList({
      ...anonymous(),
      query: { page_size: 50, page, official: true, include_deleted: false },
    }))
    if (!result) throw new Error('Tadoku data is unavailable. Try again.')

    for (const contest of result.contests) {
      if (contest.contest_start > today && (!nextContest || contest.contest_start < nextContest.contest_start)) {
        nextContest = contest
      }
    }

    if (!result.next_page_token || result.contests.length === 0) return identified(nextContest)
    page += 1
  }
}

export async function readHomeData(): Promise<HomeData> {
  const [latestContest, nextContest] = await Promise.all([
    found(immersionContestFindLatestOfficial(anonymous())).then(identified),
    readNextContest(),
  ])

  if (!latestContest) return { latestContest: null, summary: null, standings: null, nextContest }

  const path = { id: latestContest.id }
  const [summary, standings] = await Promise.all([
    found(immersionContestFetchSummary({ ...anonymous(), path })),
    found(immersionContestFetchLeaderboard({ ...anonymous(), path, query: { page_size: 5, page: 0 } })),
  ])

  return { latestContest, summary, standings, nextContest }
}
