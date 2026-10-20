import {
  immersionContestFetchLeaderboard,
  immersionFetchLeaderboardForYear,
  immersionFetchLeaderboardGlobal,
  type ImmersionLeaderboard,
  type ImmersionLeaderboardEntry,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useQuery } from '@tanstack/react-query'

export type LeaderboardEntry = ImmersionLeaderboardEntry
export type Leaderboard = ImmersionLeaderboard

type Filters = {
  pageSize: number
  page: number
  languageCode?: string
  activityId?: number
}

const query = (opts: Filters) => ({
  page_size: opts.pageSize,
  page: opts.page - 1,
  ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
  ...(opts.activityId ? { activity_id: opts.activityId } : {}),
})

export const useContestLeaderboard = (
  opts: Filters & { contestId: string },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'leaderboard', opts],
    async (): Promise<Leaderboard> =>
      apiData(
        immersionContestFetchLeaderboard({
          ...browserApi(),
          path: { id: opts.contestId },
          query: query(opts),
        }),
      ),
    options,
  )

export const useYearlyLeaderboard = (
  opts: Filters & { year: number },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'yearly', opts],
    async (): Promise<Leaderboard> =>
      apiData(
        immersionFetchLeaderboardForYear({
          ...browserApi(),
          path: { year: opts.year },
          query: query(opts),
        }),
      ),
    options,
  )

export const useAllTimeLeaderboard = (
  opts: Filters,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'all-time', opts],
    async (): Promise<Leaderboard> =>
      apiData(
        immersionFetchLeaderboardGlobal({ ...browserApi(), query: query(opts) }),
      ),
    options,
  )
