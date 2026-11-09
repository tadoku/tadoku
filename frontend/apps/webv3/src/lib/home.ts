import { queryOptions, useQuery } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import { z } from 'zod'
import type { ImmersionContest, ImmersionContestSummary, ImmersionContestView, ImmersionLeaderboard } from 'api-client'
import { latestBoardQuery, type Read } from './leaderboard'
import { profileStatisticsQuery, recentLogsQuery } from './profile'
import { viewerKey, type Viewer } from './session'

/** A contest with its ID. Contest bodies share the schema, so the contract leaves `id` optional. */
export type Identified<T> = T & { id: string }

export interface HomeBoard {
  summary: ImmersionContestSummary | null
  standings: ImmersionLeaderboard | null
}

export interface HomeData extends HomeBoard {
  latestContest: Identified<ImmersionContestView> | null
}

/** The next official round: `ok: false` when it can't be read, `null` data when none is scheduled. */
export type NextContestRead = Read<Identified<ImmersionContest> | null>

const getHomeBoard = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readHomeBoard } = await import('./home.server')
    return readHomeBoard(data)
  })

const getNextContest = createServerFn({ method: 'GET' }).handler(async () => {
  const { readNextContest } = await import('./home.server')
  return readNextContest()
})

/** The summary and top standings for the latest contest, read as soon as that contest is known. */
export const homeBoardQuery = (contestId: string) => queryOptions({
  queryKey: ['home-board', contestId],
  queryFn: () => getHomeBoard({ data: contestId }),
  staleTime: 30_000,
  retry: false,
  retryOnMount: false,
})

export const nextContestQuery = () => queryOptions({
  queryKey: ['next-contest'],
  queryFn: (): Promise<NextContestRead> => getNextContest(),
  staleTime: 30_000,
  retry: false,
  retryOnMount: false,
})

/** The latest contest with its board, as one view for the home page. */
export function useHomeData() {
  const latest = useQuery(latestBoardQuery())
  const contest = latest.data?.kind === 'board' ? latest.data.board.contest : null
  const board = useQuery({ ...homeBoardQuery(contest?.id ?? ''), enabled: contest !== null })

  let data: HomeData | undefined
  if (latest.data?.kind === 'none') data = { latestContest: null, summary: null, standings: null }
  else if (contest && board.data) data = { latestContest: contest, ...board.data }

  return {
    data,
    dataUpdatedAt: latest.dataUpdatedAt,
    isError: latest.isError || latest.data?.kind === 'unavailable' || board.isError,
    isFetching: latest.isFetching || board.isFetching,
    refetch: () => {
      void latest.refetch()
      if (contest) void board.refetch()
    },
  }
}

/** The home page's reads for a signed-in learner: this year's statistics and their newest logs. */
export const memberDashboardQueries = (viewer: Extract<Viewer, { status: 'signed-in' }>, year: number) => [
  profileStatisticsQuery({ userId: viewer.userId, year, page: 1 }, viewerKey(viewer)),
  recentLogsQuery(viewer.userId),
] as const
