import {
  immersionContestFindOngoingRegistrations,
  immersionLogCreate,
  immersionLogFindById,
  immersionLogGetConfigurations,
} from 'api-client'
import type { NewLog } from './log-draft'
import type { ImmersionLog, LogFailure, LogOptions, LogResult, Registration } from './logs'
import { readViewer, sessionCookie, sessionRequest } from './session.server'

const failures: Partial<Record<number, LogFailure>> = {
  400: 'invalid',
  401: 'signed-out',
  403: 'forbidden',
  404: 'not-found',
  409: 'deleting',
}

/** A generated operation's validated answer, or the failure its status means. Anything unreadable is unavailable. */
async function answer<T>(request: Promise<{ data?: T; response?: Response }>): Promise<LogResult<T>> {
  const { data, response } = await request
  if (data !== undefined) return { ok: true, data }

  const status = response && !response.ok ? response.status : undefined
  return { ok: false, reason: (status && failures[status]) || 'unavailable' }
}

export async function readLogOptions(): Promise<LogResult<LogOptions>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const [configuration, registrations] = await Promise.all([
    answer(immersionLogGetConfigurations(sessionRequest(cookie))),
    answer(immersionContestFindOngoingRegistrations(sessionRequest(cookie))),
  ])
  if (!configuration.ok) return configuration
  if (!registrations.ok) return registrations

  // The form needs a language, an activity and identified contests; an answer without them is malformed.
  const { languages, activities } = configuration.data
  const offered = registrations.data.registrations
  const identified = (registration: (typeof offered)[number]): registration is Registration =>
    registration.id !== undefined && registration.contest !== undefined
  if (!languages.length || !activities.length || !offered.every(identified)) {
    return { ok: false, reason: 'unavailable' }
  }
  return { ok: true, data: { configuration: configuration.data, registrations: offered } }
}

/** Reads a log as the visitor. A rejected session cookie falls back to the public view of the log. */
export async function readLog(id: string): Promise<LogResult<ImmersionLog>> {
  const cookie = sessionCookie()
  const result = await answer(immersionLogFindById({ ...sessionRequest(cookie), path: { id } }))

  if (!result.ok && result.reason === 'signed-out' && cookie) {
    return answer(immersionLogFindById({ ...sessionRequest(), path: { id } }))
  }
  return result
}

/** Saves the log for the account that started its draft, refusing it when another account or none is signed in. */
export async function saveLog(log: NewLog): Promise<LogResult<ImmersionLog>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const viewer = await readViewer()
  switch (viewer.status) {
    case 'anonymous': return { ok: false, reason: 'signed-out' }
    case 'unavailable': return { ok: false, reason: 'account-unavailable' }
    case 'banned': return { ok: false, reason: 'forbidden' }
  }
  if (viewer.userId !== log.viewerId) return { ok: false, reason: 'switched' }

  const measure = log.measure.kind === 'amount'
    ? { amount: log.measure.amount, unit_id: log.measure.unitId }
    : { duration_seconds: log.measure.seconds }
  return answer(immersionLogCreate({
    ...sessionRequest(cookie),
    body: {
      language_code: log.languageCode,
      activity_id: log.activityId,
      ...measure,
      description: log.description,
      tags: log.tags,
      registration_ids: log.registrationIds,
    },
  }))
}
