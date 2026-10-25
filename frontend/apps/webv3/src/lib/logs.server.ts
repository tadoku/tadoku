import {
  immersionContestFindOngoingRegistrations,
  immersionLogCreate,
  immersionLogFindById,
  immersionLogGetConfigurations,
} from 'api-client'
import type { Configuration, Failure, ImmersionLog, LogFormValue, Registration, Result } from './logs'
import { readViewer, sessionCookie, sessionRequest } from './session.server'

const failures: Partial<Record<number, Failure>> = {
  400: 'invalid',
  401: 'signed-out',
  403: 'forbidden',
  404: 'not-found',
  409: 'deleting',
}

/** A generated operation's validated answer, or the failure its status means. Anything unreadable is unavailable. */
async function answer<T>(request: Promise<{ data?: T; response?: Response }>): Promise<Result<T>> {
  const { data, response } = await request
  if (data !== undefined) return { ok: true, data }
  return { ok: false, reason: (response && !response.ok && failures[response.status]) || 'unavailable' }
}

export async function readLogFormData(): Promise<Result<{ configuration: Configuration; registrations: Registration[] }>> {
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
  if (!languages.length || !activities.length || !offered.every(identified)) return { ok: false, reason: 'unavailable' }
  return { ok: true, data: { configuration: configuration.data, registrations: offered } }
}

export async function readLog(id: string): Promise<Result<ImmersionLog>> {
  const cookie = sessionCookie()
  const result = await answer(immersionLogFindById({ ...sessionRequest(cookie), path: { id } }))
  return !result.ok && result.reason === 'signed-out' && cookie
    ? answer(immersionLogFindById({ ...sessionRequest(), path: { id } }))
    : result
}

export async function saveLog(value: LogFormValue & { viewerId: string }): Promise<Result<ImmersionLog>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }
  const viewer = await readViewer()
  if (viewer.status === 'anonymous') return { ok: false, reason: 'signed-out' }
  if (viewer.status === 'unavailable') return { ok: false, reason: 'account-unavailable' }
  if (viewer.status === 'banned') return { ok: false, reason: 'forbidden' }
  if (viewer.userId !== value.viewerId) return { ok: false, reason: 'switched' }

  return answer(immersionLogCreate({
    ...sessionRequest(cookie),
    body: {
      language_code: value.languageCode,
      activity_id: value.activityId,
      tags: value.tags,
      description: value.description,
      registration_ids: value.registrationIds,
      ...(value.measure.kind === 'amount'
        ? { amount: value.measure.amount, unit_id: value.measure.unitId }
        : { duration_seconds: Math.round(value.measure.minutes * 60) }),
    },
  }))
}
