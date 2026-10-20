import {
  immersionProfileFindByUserId,
  immersionProfileListLogs,
  immersionProfileYearlyActivityByUserId,
  immersionProfileYearlyActivitySplitByUserId,
  immersionProfileYearlyContestRegistrationsByUserId,
  immersionProfileYearlyScoresByUserId,
  type ImmersionActivitySplit,
  type ImmersionActivitySplitScore,
  type ImmersionProfileScores,
  type ImmersionUserActivity,
  type ImmersionUserActivityScore,
  type ImmersionUserProfile,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useQuery } from '@tanstack/react-query'
import type { Logs } from '@app/immersion/api'
import {
  withRegistrationIds,
  type ContestRegistrationsView,
} from '@app/immersion/contest-api'

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
    async (): Promise<Logs> =>
      apiData(
        immersionProfileListLogs({
          ...browserApi(),
          path: { user_id: opts.userId },
          query: {
            page_size: opts.pageSize,
            page: opts.page - 1,
            include_deleted: opts.includeDeleted,
            ...(opts.date ? { date: opts.date } : {}),
          },
        }),
      ),
    options,
  )

export type UserProfile = ImmersionUserProfile

export const useUserProfile = (
  opts: {
    userId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'profile'],
    async (): Promise<UserProfile> =>
      apiData(
        immersionProfileFindByUserId({
          ...browserApi(),
          path: { userId: opts.userId },
        }),
      ),
    options,
  )

export type UserActivityScore = ImmersionUserActivityScore
export type UserActivity = ImmersionUserActivity

export const useUserYearlyActivity = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'yearly-activity', opts.year],
    async (): Promise<UserActivity> =>
      apiData(
        immersionProfileYearlyActivityByUserId({
          ...browserApi(),
          path: { userId: opts.userId, year: Number(opts.year) },
        }),
      ),
    options,
  )

export type ProfileScores = ImmersionProfileScores

export const useProfileScores = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'scores', opts.year],
    async (): Promise<ProfileScores> =>
      apiData(
        immersionProfileYearlyScoresByUserId({
          ...browserApi(),
          path: { userId: opts.userId, year: Number(opts.year) },
        }),
      ),
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
    async (): Promise<ContestRegistrationsView> =>
      withRegistrationIds(
        await apiData(
          immersionProfileYearlyContestRegistrationsByUserId({
            ...browserApi(),
            path: { userId: opts.userId, year: Number(opts.year) },
            query: { page_size: opts.pageSize, page: opts.page - 1 },
          }),
        ),
      ),
    options,
  )

export type ActivitySplitScore = ImmersionActivitySplitScore
export type ActivitySplit = ImmersionActivitySplit

export const useUserYearlyActivitySplit = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'activity-split', opts.year],
    async (): Promise<ActivitySplit> =>
      apiData(
        immersionProfileYearlyActivitySplitByUserId({
          ...browserApi(),
          path: { userId: opts.userId, year: Number(opts.year) },
        }),
      ),
    options,
  )
