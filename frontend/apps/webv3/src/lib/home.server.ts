import { z } from 'zod'
import {
  contestListSchema,
  contestSummarySchema,
  contestViewSchema,
  leaderboardSchema,
} from './home'
import type { Contest, HomeData } from './home'
import { requestUrls } from './urls.server'

async function readResponse<T>(path: string, schema: z.ZodType<T>): Promise<T | null> {
  const endpoint = requestUrls().apiUrl
  const response = await fetch(`${endpoint.replace(/\/$/, '')}/immersion/${path}`, {
    method: 'GET',
    credentials: 'omit',
    headers: { Accept: 'application/json' },
    redirect: 'error',
    signal: AbortSignal.timeout(10_000),
  })

  if (response.status === 404) return null
  if (!response.ok) throw new Error('Tadoku data is unavailable. Try again.')
  return schema.parse(await response.json())
}

async function readNextContest(): Promise<Contest | null> {
  const today = new Date().toISOString().slice(0, 10)
  let nextContest: Contest | null = null
  let page = 0

  while (true) {
    const result = await readResponse(
      `contests?page_size=50&page=${page}&official=true&include_deleted=false`,
      contestListSchema,
    )
    if (!result) throw new Error('Tadoku data is unavailable. Try again.')

    for (const contest of result.contests) {
      if (contest.contest_start > today && (!nextContest || contest.contest_start < nextContest.contest_start)) {
        nextContest = contest
      }
    }

    if (!result.next_page_token || result.contests.length === 0) return nextContest
    page += 1
  }
}

export async function readHomeData(): Promise<HomeData> {
  const [latestContest, nextContest] = await Promise.all([
    readResponse('contests/latest-official', contestViewSchema),
    readNextContest(),
  ])

  if (!latestContest) return { latestContest: null, summary: null, standings: null, nextContest }

  const [summary, standings] = await Promise.all([
    readResponse(`contests/${latestContest.id}/summary`, contestSummarySchema),
    readResponse(`contests/${latestContest.id}/leaderboard?page_size=5&page=0`, leaderboardSchema),
  ])

  return { latestContest, summary, standings, nextContest }
}
