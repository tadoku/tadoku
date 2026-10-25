import { queryOptions, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type {
  ImmersionContestRegistration,
  ImmersionContestView,
  ImmersionLog,
  ImmersionLogConfigurationOptions,
} from 'api-client'
import { z } from 'zod'
import { newLogSchema, type NewLog } from './log-draft'
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

/** The form's options for the account that started a draft; they don't change while that draft is open. */
export const logOptionsQuery = (viewerId: string) => queryOptions({
  queryKey: ['log-options', viewerId],
  queryFn: () => getLogOptions(),
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
  const queryClient = useQueryClient()

  return async (log: NewLog): Promise<LogResult<ImmersionLog>> => {
    const result = await createLog({ data: log })
      .catch((): LogResult<ImmersionLog> => ({ ok: false, reason: 'unavailable' }))

    if (result.ok) {
      queryClient.setQueryData(logQuery(result.data.id, log.viewerId).queryKey, result)
    } else if (result.reason === 'switched') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}
