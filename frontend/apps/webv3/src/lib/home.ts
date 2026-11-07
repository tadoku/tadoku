import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContest, ImmersionContestSummary, ImmersionContestView, ImmersionLeaderboard } from 'api-client'
import { profileActivityQuery, profileStatisticsQuery } from './profile'
import { viewerKey, type Viewer } from './session'

/** A contest with its ID. Contest bodies share the schema, so the contract leaves `id` optional. */
export type Identified<T> = T & { id: string }

export interface HomeData {
  latestContest: Identified<ImmersionContestView> | null
  summary: ImmersionContestSummary | null
  standings: ImmersionLeaderboard | null
  nextContest: Identified<ImmersionContest> | null
}

const getHomeData = createServerFn({ method: 'GET' }).handler(async () => {
  const { readHomeData } = await import('./home.server')
  return readHomeData()
})

export const homeQueryOptions = () => queryOptions({
  queryKey: ['home'],
  queryFn: () => getHomeData(),
  staleTime: 30_000,
  retry: false,
  retryOnMount: false,
})

/** The home page's reads for a signed-in learner: this year's statistics and their newest logs. */
export const memberDashboardQueries = (viewer: Extract<Viewer, { status: 'signed-in' }>, year: number) => [
  profileStatisticsQuery({ userId: viewer.userId, year, page: 1 }, viewerKey(viewer)),
  profileActivityQuery({ userId: viewer.userId, page: 1 }),
] as const

export const dayNumber = (from: string, to: string) => Math.round((Date.parse(to) - Date.parse(from)) / 86_400_000) + 1

export type Phase = 'open' | 'live' | 'upcoming' | 'ended' | 'none' | 'unavailable' | 'loading'

export function phaseOf(data: HomeData | undefined, failed: boolean, today: string): Phase {
  if (failed) return 'unavailable'
  if (!data) return 'loading'
  const contest = data.latestContest
  if (!contest) return 'none'
  if (today < contest.contest_start) return 'upcoming'
  if (today > contest.contest_end) return 'ended'
  return today <= contest.registration_end ? 'open' : 'live'
}
