import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContest, ImmersionContestSummary, ImmersionContestView, ImmersionLeaderboard } from 'api-client'

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
