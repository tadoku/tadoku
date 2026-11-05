import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type {
  ImmersionActivitySplit,
  ImmersionProfileScores,
  ImmersionUserActivity,
  ImmersionUserProfile,
} from 'api-client'
import { z } from 'zod'

export const contestPageSize = 50

/** One independently loaded part of a profile year; Webv2 reports each part's failure on its own. */
export type Section<T> = { ok: true; data: T } | { ok: false }

export type ContestHistory = { contests: { id: string; title: string; private: boolean }[]; total: number }

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

const statisticsInput = z.object({
  userId: z.guid(),
  year: z.number().int().min(1000).max(9999),
  page: z.number().int().min(1),
})

const getProfileStatistics = createServerFn({ method: 'GET' })
  .inputValidator(statisticsInput)
  .handler(async ({ data }) => {
    const { readProfileStatistics } = await import('./profile.server')
    return readProfileStatistics(data.userId, data.year, data.page)
  })

/** A profile's year as one viewer sees it: its contest history includes private contests for their participant. */
export const profileStatisticsQuery = (input: z.infer<typeof statisticsInput>, viewerKey: string) => queryOptions({
  queryKey: ['profile-statistics', input.userId, input.year, input.page, viewerKey],
  queryFn: () => getProfileStatistics({ data: input }),
  retry: false,
  staleTime: 30_000,
})
