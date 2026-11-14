import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type {
  ImmersionActivitySplit,
  ImmersionLogs,
  ImmersionProfileScores,
  ImmersionUserActivity,
  ImmersionUserProfile,
} from 'api-client'
import { z } from 'zod'

export const contestPageSize = 50
/** Profile and contest log lists both show 25 logs per page. */
export const activityPageSize = 25
export const recentLogCount = 5

/** One independently loaded part of a profile year; Webv2 reports each part's failure on its own. */
export type Section<T> = { ok: true; data: T } | { ok: false }

export type ContestHistory = {
  contests: { id: string; title: string; private: boolean; contest_start: string; contest_end: string; languages: string[] }[]
  total: number
}

export type ProfileStatistics =
  | {
    status: 'found'
    /** The server's UTC date, so the year range and daily average don't depend on the browser's clock. */
    today: string
    profile: ImmersionUserProfile
    activity: Section<ImmersionUserActivity>
    scores: Section<ImmersionProfileScores>
    split: Section<ImmersionActivitySplit>
    contests: Section<ContestHistory>
  }
  | { status: 'not-found' }
  | { status: 'unavailable' }

/** A profile year's address as the server function accepts it; a route answers 404 for anything else. */
export const profileStatisticsInput = z.object({
  userId: z.guid(),
  year: z.number().int().min(1000).max(9999),
  page: z.number().int().min(1),
})

const getProfileStatistics = createServerFn({ method: 'GET' })
  .inputValidator(profileStatisticsInput)
  .handler(async ({ data }) => {
    const { readProfileStatistics } = await import('./profile.server')
    return readProfileStatistics(data.userId, data.year, data.page)
  })

/** A profile's year as one viewer sees it: its contest history includes private contests for their participant. */
export const profileStatisticsQuery = (input: z.infer<typeof profileStatisticsInput>, viewerKey: string) => queryOptions({
  queryKey: ['profile-statistics', input.userId, input.year, input.page, viewerKey],
  queryFn: () => getProfileStatistics({ data: input }),
  retry: false,
  staleTime: 30_000,
})

/** A page of a profile's logs, newest first; its list fails on its own while the profile header still shows. */
export type ProfileActivity =
  | { status: 'found'; profile: ImmersionUserProfile; logs: Section<ImmersionLogs> }
  | { status: 'not-found' }
  | { status: 'unavailable' }

const activityInput = z.object({
  userId: z.guid(),
  page: z.number().int().min(1),
  /** Only one UTC calendar day's logs, as the statistics calendar links them. */
  date: z.iso.date().optional(),
})

const getProfileActivity = createServerFn({ method: 'GET' })
  .inputValidator(activityInput)
  .handler(async ({ data }) => {
    const { readProfileActivity } = await import('./profile.server')
    return readProfileActivity(data.userId, data.page, data.date)
  })

const getRecentLogs = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readRecentLogs } = await import('./profile.server')
    return readRecentLogs(data)
  })

/** The learner's newest logs, as the home page's recent activity shows them. */
export const recentLogsQuery = (userId: string) => queryOptions({
  queryKey: ['recent-logs', userId],
  queryFn: () => getRecentLogs({ data: userId }),
  retry: false,
  staleTime: 30_000,
})

export const profileActivityQuery = (input: z.infer<typeof activityInput>) => queryOptions({
  queryKey: ['profile-activity', input.userId, input.page, input.date],
  queryFn: () => getProfileActivity({ data: input }),
  retry: false,
  staleTime: 30_000,
})

/** The days in a year's daily scores with any activity. */
export const activeDays = (days: { score: number }[]) => days.filter(day => day.score > 0).length
