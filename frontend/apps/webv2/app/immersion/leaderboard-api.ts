import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { useQuery } from '@tanstack/react-query'

const root = () => `${browserAppUrls().apiEndpoint}/immersion`

const LeaderboardEntry = z.object({
  rank: z.number(),
  user_id: z.string(),
  user_display_name: z.string(),
  score: z.number(),
  is_tie: z.boolean(),
})

export type LeaderboardEntry = z.infer<typeof LeaderboardEntry>

const Leaderboard = z.object({
  entries: z.array(LeaderboardEntry),
  total_size: z.number(),
  next_page_token: z.string(),
})

export type Leaderboard = z.infer<typeof Leaderboard>

export const useContestLeaderboard = (
  opts: {
    contestId: string
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'leaderboard', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root()}/contests/${opts.contestId}/leaderboard?${new URLSearchParams(
          params,
        )}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )

export const useYearlyLeaderboard = (
  opts: {
    year: number
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'yearly', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root()}/leaderboard/yearly/${opts.year}?${new URLSearchParams(
          params,
        )}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )

export const useAllTimeLeaderboard = (
  opts: {
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'all-time', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root()}/leaderboard/global?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )
