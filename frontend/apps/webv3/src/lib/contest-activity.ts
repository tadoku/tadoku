import { queryOptions, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestProfileActivityRow, ImmersionContestRegistration, ImmersionContestView, ImmersionLogs, ImmersionScores } from 'api-client'
import { z } from 'zod'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { viewerQueryOptions } from './session'

/** A participant's registration with the contest it belongs to, which the contract leaves optional. */
export type ContestEntry = ImmersionContestRegistration & { contest: ImmersionContestView & { id: string } }

/** One participant's standing in one contest. Their daily scores fail on their own while the scores still show. */
export type ContestProfile =
  | { status: 'found'; registration: ContestEntry; scores: ImmersionScores; daily: Read<ImmersionContestProfileActivityRow[]> }
  | { status: 'not-found' }
  | { status: 'unavailable' }

const logsInput = z.object({
  contestId: z.guid(),
  /** Only this participant's logs. */
  userId: z.guid().optional(),
  page: z.int().min(1),
})

const profileInput = z.object({ contestId: z.guid(), userId: z.guid() })

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

const postDetach = createServerFn({ method: 'POST' })
  .inputValidator(detachSchema)
  .handler(async ({ data }) => {
    const { detachLog } = await import('./contest-activity.server')
    return detachLog(data)
  })

/** A page of a contest's logs, newest first, counted from 1. */
export const contestLogsQuery = (input: z.infer<typeof logsInput>) => queryOptions({
  queryKey: ['contest-logs', input.contestId, input.userId, input.page],
  queryFn: (): Promise<Read<ImmersionLogs>> => getContestLogs({ data: input }),
  staleTime: 30_000,
  retry: false,
})

export const contestProfileQuery = (input: z.infer<typeof profileInput>) => queryOptions({
  queryKey: ['contest-profile', input.contestId, input.userId],
  queryFn: (): Promise<ContestProfile> => getContestProfile({ data: input }),
  staleTime: 30_000,
  retry: false,
})

/**
 * Removes a log from a contest. A removal refreshes what it changes: the contest's logs, board, standings and
 * participant profiles. A failed request is `unavailable`: the log may or may not have been removed.
 */
export function useDetachLog() {
  const queryClient = useQueryClient()

  return async (detach: Detach): Promise<LogResult<null>> => {
    const result = await postDetach({ data: detach }).catch((): LogResult<null> => ({ ok: false, reason: 'unavailable' }))
    const { contestId } = detach
    if (result.ok) {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['contest-logs', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['contest-board', contestId] }),
        queryClient.invalidateQueries({ queryKey: ['standings', { kind: 'contest', contestId }] }),
        queryClient.invalidateQueries({ queryKey: ['contest-profile', contestId] }),
      ])
    } else if (result.reason === 'switched' || result.reason === 'signed-out') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}
