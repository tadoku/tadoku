import { keepPreviousData, queryOptions, skipToken, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type {
  ImmersionContestRegistration,
  ImmersionContestView,
  ImmersionLog,
  ImmersionLogConfigurationOptions,
  ImmersionTagSuggestion,
} from 'api-client'
import { z } from 'zod'
import {
  logContestsUpdateSchema,
  logUpdateSchema,
  newLogSchema,
  scorePreviewSchema,
  type LogContestsUpdate,
  type LogUpdate,
  type NewLog,
  type ScorePreviewInput,
} from './log-draft'
import { viewerQueryOptions } from './session'

export type { ImmersionLog }

/** A registration the form can offer: the contract leaves its ID and contest optional because request bodies share it. */
export type Registration = ImmersionContestRegistration & { id: string; contest: ImmersionContestView }

/** What the activity form offers one account: languages, activities and units, and its ongoing registrations. */
export type LogOptions = { configuration: ImmersionLogConfigurationOptions; registrations: Registration[] }

export type LogFailure =
  | 'signed-out'
  | 'switched'
  | 'account-unavailable'
  | 'invalid'
  | 'forbidden'
  | 'not-found'
  | 'deleting'
  | 'unavailable'

export type LogResult<T> = { ok: true; data: T } | { ok: false; reason: LogFailure }

const getLogOptions = createServerFn({ method: 'GET' }).handler(async () => {
  const { readLogOptions } = await import('./logs.server')
  return readLogOptions()
})

const ownerRegistrationsSchema = z.object({ viewerId: z.guid(), ownerId: z.guid() })
export type OwnerRegistrations = z.infer<typeof ownerRegistrationsSchema>

const getOwnerRegistrations = createServerFn({ method: 'GET' })
  .inputValidator(ownerRegistrationsSchema)
  .handler(async ({ data }) => {
    const { readOwnerRegistrations } = await import('./logs.server')
    return readOwnerRegistrations(data)
  })

const getLog = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readLog } = await import('./logs.server')
    return readLog(data)
  })

const createLog = createServerFn({ method: 'POST' })
  .inputValidator(newLogSchema)
  .handler(async ({ data }) => {
    const { saveLog } = await import('./logs.server')
    return saveLog(data)
  })

const changeLog = createServerFn({ method: 'POST' })
  .inputValidator(logUpdateSchema)
  .handler(async ({ data }) => {
    const { updateLog } = await import('./logs.server')
    return updateLog(data)
  })

const changeLogContests = createServerFn({ method: 'POST' })
  .inputValidator(logContestsUpdateSchema)
  .handler(async ({ data }) => {
    const { updateLogContests } = await import('./logs.server')
    return updateLogContests(data)
  })

const removeLog = createServerFn({ method: 'POST' })
  .inputValidator(z.object({ viewerId: z.guid(), logId: z.guid() }))
  .handler(async ({ data }) => {
    const { deleteLog } = await import('./logs.server')
    return deleteLog(data)
  })

// A GET, because an estimate changes nothing.
const getScorePreview = createServerFn({ method: 'GET' })
  .inputValidator(scorePreviewSchema)
  .handler(async ({ data }) => {
    const { readScorePreview } = await import('./logs.server')
    return readScorePreview(data)
  })

const getTagSuggestions = createServerFn({ method: 'GET' })
  .inputValidator(z.string().max(255))
  .handler(async ({ data }) => {
    const { readTagSuggestions } = await import('./logs.server')
    return readTagSuggestions(data)
  })

/**
 * The viewer's suggested tags containing `query`, with how often they used each. A failed read suggests nothing; the
 * previous suggestions stay visible while the next ones load.
 */
export const tagSuggestionsQuery = (viewerId: string, query: string) => queryOptions({
  queryKey: ['tag-suggestions', viewerId, query],
  queryFn: async (): Promise<ImmersionTagSuggestion[]> => {
    const result = await getTagSuggestions({ data: query })
    return result.ok ? result.data.suggestions : []
  },
  placeholderData: keepPreviousData,
  retry: false,
  staleTime: 60_000,
})

/** The form's options for the account that started a draft; they don't change while that draft is open. */
export const logOptionsQuery = (viewerId: string) => queryOptions({
  queryKey: ['log-options', viewerId],
  queryFn: () => getLogOptions(),
  retry: false,
  staleTime: Infinity,
})

/** The registrations a viewer can submit a log of `ownerId` to: the owner's ongoing ones, whoever is viewing. */
export const ownerRegistrationsQuery = (viewerId: string, ownerId: string) => queryOptions({
  queryKey: ['owner-registrations', viewerId, ownerId],
  queryFn: () => getOwnerRegistrations({ data: { viewerId, ownerId } }),
  retry: false,
  staleTime: Infinity,
})

/** A log as one viewer sees it, cached per `viewerKey`: owners also see its contest submissions. */
export const logQuery = (id: string, viewerKey: string) => queryOptions({
  queryKey: ['log', id, viewerKey],
  queryFn: () => getLog({ data: id }),
  retry: false,
  staleTime: 30_000,
})

/**
 * Saves a log. A saved log is cached for its detail page; a refusal because another account signed in refreshes the
 * viewer. A failed request is reported as `unavailable`, because the log may or may not have been saved.
 */
export function useSaveLog() {
  const settle = useSettle()
  return async (log: NewLog) => settle(log.viewerId, await createLog({ data: log }).catch(unconfirmed<ImmersionLog>))
}

/** Updates a log like `useSaveLog` saves one. */
export function useUpdateLog() {
  const settle = useSettle()
  return async (change: LogUpdate) =>
    settle(change.viewerId, await changeLog({ data: change }).catch(unconfirmed<ImmersionLog>))
}

/** Changes a log's contest submissions like `useSaveLog` saves a log. Repeating the same change is safe. */
export function useUpdateLogContests() {
  const settle = useSettle()
  return async (change: LogContestsUpdate) =>
    settle(change.viewerId, await changeLogContests({ data: change }).catch(unconfirmed<ImmersionLog>))
}

/** Deletes a log for the account that viewed it. A failed request is `unavailable`: the log may or may not be gone. */
export function useDeleteLog() {
  const queryClient = useQueryClient()

  return async (viewerId: string, logId: string) => {
    const result = await removeLog({ data: { viewerId, logId } }).catch(unconfirmed<null>)
    if (!result.ok && result.reason === 'switched') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}

const unconfirmed = <T>(): LogResult<T> => ({ ok: false, reason: 'unavailable' })

function useSettle() {
  const queryClient = useQueryClient()

  return async (viewerId: string, result: LogResult<ImmersionLog>) => {
    if (result.ok) {
      queryClient.setQueryData(logQuery(result.data.id, viewerId).queryKey, result)
    } else if (result.reason === 'switched') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}

/** The API's estimate for a complete draft. The previous estimate stays visible while the next one loads. */
export const scorePreviewQuery = (input: ScorePreviewInput | undefined) => queryOptions({
  queryKey: ['score-preview', input],
  queryFn: input ? () => getScorePreview({ data: input }) : skipToken,
  placeholderData: keepPreviousData,
  retry: false,
  staleTime: Infinity,
})
