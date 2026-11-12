import { queryOptions, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestProfileSummary, ImmersionContestRegistration, ImmersionContestView, ImmersionLogs, ImmersionScores } from 'api-client'
import { z } from 'zod'
import { logFiltersSchema } from './contest-filters'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { viewerQueryOptions } from './session'

/** A participant's registration with the contest it belongs to, which the contract leaves optional. */
export type ContestEntry = ImmersionContestRegistration & { contest: ImmersionContestView & { id: string } }

/** One participant's standing in one contest. */
export type ContestProfile =
  | { status: 'found'; registration: ContestEntry; scores: ImmersionScores }
  | { status: 'not-found' }
  | { status: 'unavailable' }

const logsInput = z.object({
  contestId: z.guid(),
  /** Only this participant's logs. */
  userId: z.guid().optional(),
  page: z.int().min(1),
  filters: logFiltersSchema.optional(),
})

const profileInput = z.object({ contestId: z.guid(), userId: z.guid() })

const summaryInput = profileInput.extend({ filters: logFiltersSchema })

/** A participant's contest summary under some filters, with the UTC day it was read, where the activity chart ends. */
export type ContestSummary = ImmersionContestProfileSummary & { today: string }

/** A moderator's removal of a log from a contest, for the account that opened the page. */
export const detachSchema = z.object({
  viewerId: z.guid(),
  contestId: z.guid(),
  logId: z.guid(),
  reason: z.string().trim().min(1).max(1000),
})

export type Detach = z.infer<typeof detachSchema>

const getContestLogs = createServerFn({ method: 'GET' })
  .inputValidator(logsInput)
  .handler(async ({ data }) => {
    const { readContestLogs } = await import('./contest-activity.server')
    return readContestLogs(data)
  })

const getContestProfile = createServerFn({ method: 'GET' })
  .inputValidator(profileInput)
  .handler(async ({ data }) => {
    const { readContestProfile } = await import('./contest-activity.server')
    return readContestProfile(data.contestId, data.userId)
  })

const getContestSummary = createServerFn({ method: 'GET' })
  .inputValidator(summaryInput)
  .handler(async ({ data }) => {
    const { readContestSummary } = await import('./contest-activity.server')
    return readContestSummary(data)
  })

const postDetach = createServerFn({ method: 'POST' })
  .inputValidator(detachSchema)
  .handler(async ({ data }) => {
    const { detachLog } = await import('./contest-activity.server')
    return detachLog(data)
  })

/** While another page or other filters of the same list load, the list keeps showing the previous answer. */
const sameList = ({ contestId, userId }: { contestId: string; userId?: string }) =>
  <T>(previous: T | undefined, query?: { queryKey: readonly unknown[] }) =>
    query?.queryKey[1] === contestId && query.queryKey[2] === userId ? previous : undefined

/** A page of a contest's logs, newest first, counted from 1, optionally only those the filters match. */
export const contestLogsQuery = (input: z.infer<typeof logsInput>) => queryOptions({
  queryKey: ['contest-logs', input.contestId, input.userId, input.page, input.filters],
  queryFn: (): Promise<Read<ImmersionLogs>> => getContestLogs({ data: input }),
  placeholderData: sameList(input),
  staleTime: 30_000,
  retry: false,
})

export const contestProfileQuery = (input: z.infer<typeof profileInput>) => queryOptions({
  queryKey: ['contest-profile', input.contestId, input.userId],
  queryFn: (): Promise<ContestProfile> => getContestProfile({ data: input }),
  staleTime: 30_000,
  retry: false,
})

/** A participant's score, days and breakdowns over the logs the filters match; each breakdown ignores its own filter. */
export const contestSummaryQuery = (input: z.infer<typeof summaryInput>) => queryOptions({
  queryKey: ['contest-profile', input.contestId, input.userId, 'summary', input.filters],
  queryFn: (): Promise<Read<ContestSummary>> => getContestSummary({ data: input }),
  placeholderData: sameList(input),
  staleTime: 30_000,
  retry: false,
})

/**
 * Removes a log from a contest. A removal refreshes what it changes: the contest's logs, summary, participant profiles
 * and places, the log's submissions and the standings, every board because leaving an official contest also takes the
 * log off the yearly and all-time ones. A failed request is `unavailable`: the log may or may not have been removed.
 */
export function useDetachLog() {
  const queryClient = useQueryClient()

  return async (detach: Detach): Promise<LogResult<null>> => {
    const result = await postDetach({ data: detach }).catch((): LogResult<null> => ({ ok: false, reason: 'unavailable' }))
    const { contestId, logId } = detach
    if (result.ok) {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['contest-logs', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['standings'] }),
        queryClient.invalidateQueries({ queryKey: ['contest-profile', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['contest-entry', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['home-board', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['log', logId] }),
      ])
    } else if (result.reason === 'switched' || result.reason === 'signed-out') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}
