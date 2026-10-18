import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { useQuery } from '@tanstack/react-query'
import { Logs } from '@app/immersion/api'
import { ContestRegistrationsView } from '@app/immersion/contest-api'
import { Score } from '@app/immersion/scoring-api'

const root = () => `${browserAppUrls().apiEndpoint}/immersion`

export const useProfileLogs = (
  opts: {
    pageSize: number
    page: number
    includeDeleted: boolean
    userId: string
    date?: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    [
      'profile',
      opts.userId,
      'logs',
      opts.pageSize,
      opts.page,
      opts.includeDeleted,
      opts.date,
    ],
    async (): Promise<Logs> => {
      const params = new URLSearchParams({
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        include_deleted: opts.includeDeleted.toString(),
      })
      if (opts.date) {
        params.set('date', opts.date)
      }
      const response = await fetch(
        `${root()}/users/${opts.userId}/logs?${params}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Logs.parse(await response.json())
    },
    options,
  )

const UserProfile = z.object({
  id: z.string(),
  display_name: z.string(),
  created_at: z.string(),
})

export type UserProfile = z.infer<typeof UserProfile>

export const useUserProfile = (
  opts: {
    userId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'profile'],
    async (): Promise<UserProfile> => {
      const response = await fetch(`${root()}/users/${opts.userId}/profile`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserProfile.parse(await response.json())
    },
    options,
  )

const UserActivityScore = z.object({
  date: z.string(),
  score: z.number(),
})

export type UserActivityScore = z.infer<typeof UserActivityScore>

const UserActivity = z.object({
  total_updates: z.number(),
  scores: z.array(UserActivityScore),
})

export type UserActivity = z.infer<typeof UserActivity>

export const useUserYearlyActivity = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'yearly-activity', opts.year],
    async (): Promise<UserActivity> => {
      const response = await fetch(
        `${root()}/users/${opts.userId}/activity/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserActivity.parse(await response.json())
    },
    options,
  )

const ProfileScores = z.object({
  overall_score: z.number(),
  scores: z.array(Score),
})

export type ProfileScores = z.infer<typeof ProfileScores>

export const useProfileScores = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'scores', opts.year],
    async (): Promise<ProfileScores> => {
      const response = await fetch(
        `${root()}/users/${opts.userId}/scores/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ProfileScores.parse(await response.json())
    },
    options,
  )

export const useYearlyContestRegistrations = (
  opts: {
    userId: string
    year: string | number
    pageSize: number
    page: number
  },
  options?: {
    enabled?: boolean
  },
) =>
  useQuery(
    [
      'users',
      opts.userId,
      'contest-registrations',
      opts.year,
      { pageSize: opts.pageSize, page: opts.page },
    ],
    async (): Promise<ContestRegistrationsView> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
      }
      const response = await fetch(
        `${root()}/users/${opts.userId}/contest-registrations/${
          opts.year
        }?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestRegistrationsView.parse(await response.json())
    },
    options,
  )

const ActivitySplitScore = z.object({
  activity_id: z.number(),
  activity_name: z.string(),
  score: z.number(),
})

export type ActivitySplitScore = z.infer<typeof ActivitySplitScore>

const ActivitySplit = z.object({
  activities: z.array(ActivitySplitScore),
})

export type ActivitySplit = z.infer<typeof ActivitySplit>

export const useUserYearlyActivitySplit = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'activity-split', opts.year],
    async (): Promise<ActivitySplit> => {
      const response = await fetch(
        `${root()}/users/${opts.userId}/activity-split/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ActivitySplit.parse(await response.json())
    },
    options,
  )
