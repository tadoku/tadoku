import { z } from 'zod'
import getConfig from 'next/config'
import { useQuery } from 'react-query'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion`

export const Language = z.object({
  code: z.string(),
  name: z.string(),
})

export type Language = z.infer<typeof Language>

export const Activity = z.object({
  id: z.number(),
  name: z.string(),
  default: z.boolean().nullable().optional(),
})

export type Activity = z.infer<typeof Activity>

const ContestView = z.object({
  id: z.string().nullable(),
  contest_start: z.string(),
  contest_end: z.string(),
  registration_end: z.string(),
  title: z.string(),
  description: z.string().optional().nullable(),
  private: z.boolean(),
  official: z.boolean(),
  owner_user_id: z.string().nullable().optional(),
  owner_user_display_name: z.string().nullable().optional(),
  allowed_languages: z.array(Language).nullable().optional(),
  allowed_activities: z.array(Activity),
  deleted: z.boolean().nullable().optional(),
})

export type ContestView = z.infer<typeof ContestView>

export const useLatestOfficialContest = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findLatestOfficial'],
    async (): Promise<ContestView> => {
      const response = await fetch(`${root}/contests/latest-official`, {
        credentials: 'include',
      })

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestView.parse(await response.json())
    },
    options,
  )

export const ContestSummary = z.object({
  participant_count: z.number(),
  language_count: z.number(),
  total_score: z.number(),
})

export type ContestSummary = z.infer<typeof ContestSummary>

export const useContestSummary = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestSummary', id],
    async (): Promise<ContestSummary | undefined> => {
      const response = await fetch(`${root}/contests/${id}/summary`, {
        credentials: 'include',
      })

      if (response.status === 404) {
        return undefined
      }

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestSummary.parse(await response.json())
    },
    { ...options, retry: false },
  )

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
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'leaderboard', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
      }
      const response = await fetch(
        `${root}/contests/${opts.contestId}/leaderboard?${new URLSearchParams(
          params,
        )}`,
        { credentials: 'include' },
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
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'yearly', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
      }
      const response = await fetch(
        `${root}/leaderboard/yearly/${opts.year}?${new URLSearchParams(
          params,
        )}`,
        { credentials: 'include' },
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )
