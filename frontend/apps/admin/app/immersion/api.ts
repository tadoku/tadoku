import {
  immersionContestFetchLeaderboard,
  immersionContestFetchSummary,
  immersionContestFindLatestOfficial,
  immersionFetchLeaderboardForYear,
  type ImmersionActivity,
  type ImmersionContestSummary,
  type ImmersionContestView,
  type ImmersionLanguage,
  type ImmersionLeaderboard,
  type ImmersionLeaderboardEntry,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useQuery } from '@tanstack/react-query'

export type Language = ImmersionLanguage
export type Activity = ImmersionActivity
export type ContestView = ImmersionContestView

export const useLatestOfficialContest = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findLatestOfficial'],
    async (): Promise<ContestView> =>
      apiData(immersionContestFindLatestOfficial(browserApi())),
    options,
  )

export type ContestSummary = ImmersionContestSummary

export const useContestSummary = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestSummary', id],
    async (): Promise<ContestSummary | null> => {
      const result = await immersionContestFetchSummary({
        ...browserApi(),
        path: { id },
      })
      return result.response?.status === 404 ? null : apiData(result)
    },
    { ...options, retry: false },
  )

export type LeaderboardEntry = ImmersionLeaderboardEntry
export type Leaderboard = ImmersionLeaderboard

export const useContestLeaderboard = (
  opts: {
    contestId: string
    pageSize: number
    page: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'leaderboard', opts],
    async (): Promise<Leaderboard> =>
      apiData(
        immersionContestFetchLeaderboard({
          ...browserApi(),
          path: { id: opts.contestId },
          query: { page_size: opts.pageSize, page: opts.page - 1 },
        }),
      ),
    options,
  )

export const useYearlyLeaderboard = (
  opts: {
    year: number
    pageSize: number
    page: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'yearly', opts],
    async (): Promise<Leaderboard> =>
      apiData(
        immersionFetchLeaderboardForYear({
          ...browserApi(),
          path: { year: opts.year },
          query: { page_size: opts.pageSize, page: opts.page - 1 },
        }),
      ),
    options,
  )
