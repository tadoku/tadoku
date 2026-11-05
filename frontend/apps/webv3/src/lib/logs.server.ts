import {
  immersionContestFindOngoingRegistrations,
  immersionLogContestRegistrationUpdate,
  immersionLogCreate,
  immersionLogDeleteById,
  immersionLogFindById,
  immersionLogGetConfigurations,
  immersionLogUpdate,
  immersionScorePreview,
  type ImmersionScorePreview,
} from 'api-client'
import type { LogContestsUpdate, LogUpdate, NewLog, ScorePreviewInput } from './log-draft'
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
export async function answer<T>(request: Promise<{ data?: T; response?: Response }>): Promise<LogResult<T>> {
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

/**
 * The session cookie while it belongs to the account that started a change, and that account is an administrator when
 * `admin` is set. Another account, or none, is refused before the API is asked, so a change is never made for an
 * account that did not start it.
 */
export async function draftOwnerCookie(viewerId: string, { admin = false } = {}): Promise<LogResult<string>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const viewer = await readViewer()
  switch (viewer.status) {
    case 'anonymous': return { ok: false, reason: 'signed-out' }
    case 'unavailable': return { ok: false, reason: 'account-unavailable' }
    case 'banned': return { ok: false, reason: 'forbidden' }
  }
  if (viewer.userId !== viewerId) return { ok: false, reason: 'switched' }
  if (admin && viewer.role !== 'admin') return { ok: false, reason: 'forbidden' }
  return { ok: true, data: cookie }
}

function measureBody(measure: NewLog['measure']) {
  return measure.kind === 'amount'
    ? { amount: measure.amount, unit_key: measure.unitKey, duration_seconds: measure.seconds }
    : { duration_seconds: measure.seconds }
}

export async function saveLog(log: NewLog): Promise<LogResult<ImmersionLog>> {
  const session = await draftOwnerCookie(log.viewerId)
  if (!session.ok) return session

  return answer(immersionLogCreate({
    ...sessionRequest(session.data),
    body: {
      language_code: log.languageCode,
      activity_id: log.activityId,
      ...measureBody(log.measure),
      description: log.description,
      tags: log.tags,
      registration_ids: log.registrationIds,
    },
  }))
}

/** Updates a log's measure, description and tags. The API rescores its submissions to unfinished contests. */
export async function updateLog(change: LogUpdate): Promise<LogResult<ImmersionLog>> {
  const session = await draftOwnerCookie(change.viewerId)
  if (!session.ok) return session

  return answer(immersionLogUpdate({
    ...sessionRequest(session.data),
    path: { id: change.logId },
    body: { ...measureBody(change.measure), description: change.description, tags: change.tags },
  }))
}

/**
 * Replaces a log's submissions to unfinished contests with the chosen registrations. The API keeps submissions to finished
 * contests and their scores.
 */
export async function updateLogContests(change: LogContestsUpdate): Promise<LogResult<ImmersionLog>> {
  const session = await draftOwnerCookie(change.viewerId)
  if (!session.ok) return session

  return answer(immersionLogContestRegistrationUpdate({
    ...sessionRequest(session.data),
    path: { id: change.logId },
    body: { registration_ids: change.registrationIds },
  }))
}

/** Deletes a log. A log that is already gone counts as deleted. */
export async function deleteLog({ viewerId, logId }: { viewerId: string; logId: string }): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(viewerId)
  if (!session.ok) return session

  const result = await answer(immersionLogDeleteById({ ...sessionRequest(session.data), path: { id: logId } }))
  return result.ok || result.reason === 'not-found' ? { ok: true, data: null } : result
}

export async function readScorePreview(input: ScorePreviewInput): Promise<LogResult<ImmersionScorePreview>> {
  return answer(immersionScorePreview({
    ...sessionRequest(sessionCookie()),
    body: {
      language_code: input.languageCode,
      activity_id: input.activityId,
      ...measureBody(input.measure),
      tags: input.tags,
    },
  }))
}
