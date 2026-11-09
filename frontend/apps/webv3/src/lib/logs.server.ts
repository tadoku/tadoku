import {
  immersionContestFindOngoingRegistrations,
  immersionLogContestRegistrationUpdate,
  immersionLogCreate,
  immersionLogDeleteById,
  immersionLogFindById,
  immersionLogGetConfigurations,
  immersionLogTagSuggestions,
  immersionLogUpdate,
  immersionProfileYearlyContestRegistrationsByUserId,
  immersionScorePreview,
  type ImmersionContestRegistration,
  type ImmersionScorePreview,
  type ImmersionTagSuggestions,
} from 'api-client'
import type { LogContestsUpdate, LogUpdate, NewLog, ScorePreviewInput } from './log-draft'
import type { ImmersionLog, LogFailure, LogOptions, LogResult, OwnerRegistrations, Registration } from './logs'
import { readAccount, readAsVisitor, readViewer, sessionCookie, sessionRequest } from './session.server'

export const logFailures: Partial<Record<number, LogFailure>> = {
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
  return { ok: false, reason: (status && logFailures[status]) || 'unavailable' }
}

export async function readLogOptions(): Promise<LogResult<LogOptions>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const [configuration, registrations] = await Promise.all([
    answer(immersionLogGetConfigurations(sessionRequest('immersionLogGetConfigurations', cookie))),
    answer(immersionContestFindOngoingRegistrations(sessionRequest('immersionContestFindOngoingRegistrations', cookie))),
  ])
  if (!configuration.ok) return configuration
  if (!registrations.ok) return registrations

  // The form needs a language, an activity and identified contests; an answer without them is malformed.
  const { languages, activities } = configuration.data
  const offered = registrations.data.registrations
  if (!languages.length || !activities.length || !offered.every(identified)) {
    return { ok: false, reason: 'unavailable' }
  }
  return { ok: true, data: { configuration: configuration.data, registrations: offered } }
}

const identified = (registration: ImmersionContestRegistration): registration is Registration =>
  registration.id !== undefined && registration.contest !== undefined

/**
 * The ongoing registrations of a log's owner, which are the only ones the API accepts for its submissions. Another
 * account, such as an administrator, reads the owner's registrations of this year and last and keeps those whose
 * contest has started and not finished by today's UTC date, as `contestwindow.Started` and `Unfinished` in the API do.
 */
export async function readOwnerRegistrations({ viewerId, ownerId }: OwnerRegistrations): Promise<LogResult<Registration[]>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }
  if (viewerId === ownerId) {
    const own = await answer(immersionContestFindOngoingRegistrations(sessionRequest('immersionContestFindOngoingRegistrations', cookie)))
    if (!own.ok) return own
    return own.data.registrations.every(identified) ? { ok: true, data: own.data.registrations } : { ok: false, reason: 'unavailable' }
  }

  const today = new Date().toISOString().slice(0, 10)
  const year = Number(today.slice(0, 4))
  const registrations: Registration[] = []
  for (const contestYear of [year, year - 1]) {
    for (let page = 0; ; page++) {
      const result = await answer(immersionProfileYearlyContestRegistrationsByUserId({
        ...sessionRequest('immersionProfileYearlyContestRegistrationsByUserId', cookie),
        path: { userId: ownerId, year: contestYear },
        query: { page_size: yearlyPageSize, page },
      }))
      if (!result.ok) return result
      if (!result.data.registrations.every(identified)) return { ok: false, reason: 'unavailable' }

      registrations.push(...result.data.registrations.filter(({ contest }) =>
        contest.contest_start <= today && contest.contest_end >= today))
      if (result.data.registrations.length < yearlyPageSize) break
    }
  }
  return { ok: true, data: registrations }
}

const yearlyPageSize = 50

/** Reads a log as the visitor. */
export function readLog(id: string): Promise<LogResult<ImmersionLog>> {
  return answer(readAsVisitor(cookie => immersionLogFindById({ ...sessionRequest('immersionLogFindById', cookie), path: { id } })))
}

/**
 * The session cookie while it belongs to the account that started a change, and that account is an administrator when
 * `admin` is set. Another account, or none, is refused before the API is asked, so a change is never made for an
 * account that did not start it.
 */
export async function draftOwnerCookie(viewerId: string, { admin = false } = {}): Promise<LogResult<string>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  // Only administrator access needs the role: the API itself refuses every write from a banned account.
  const viewer = admin ? await readViewer() : await readAccount(cookie)
  switch (viewer.status) {
    case 'anonymous': return { ok: false, reason: 'signed-out' }
    case 'unavailable': return { ok: false, reason: 'account-unavailable' }
    case 'banned': return { ok: false, reason: 'forbidden' }
  }
  if (viewer.userId !== viewerId) return { ok: false, reason: 'switched' }
  if (admin && !(viewer.status === 'signed-in' && viewer.role === 'admin')) return { ok: false, reason: 'forbidden' }
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
    ...sessionRequest('immersionLogCreate', session.data),
    body: {
      language_code: log.languageCode,
      activity_id: log.activityId,
      ...measureBody(log.measure),
      description: log.description,
      tags: log.tags,
    },
  }))
}

/**
 * Updates a log's measure, description, tags and, for a log submitted to no contest, its language. The API rescores its
 * submissions to unfinished contests.
 */
export async function updateLog(change: LogUpdate): Promise<LogResult<ImmersionLog>> {
  const session = await draftOwnerCookie(change.viewerId)
  if (!session.ok) return session

  return answer(immersionLogUpdate({
    ...sessionRequest('immersionLogUpdate', session.data),
    path: { id: change.logId },
    body: {
      language_code: change.languageCode,
      ...measureBody(change.measure),
      description: change.description,
      tags: change.tags,
    },
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
    ...sessionRequest('immersionLogContestRegistrationUpdate', session.data),
    path: { id: change.logId },
    body: { registration_ids: change.registrationIds },
  }))
}

/** Deletes a log. A log that is already gone counts as deleted. */
export async function deleteLog({ viewerId, logId }: { viewerId: string; logId: string }): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(viewerId)
  if (!session.ok) return session

  const result = await answer(immersionLogDeleteById({ ...sessionRequest('immersionLogDeleteById', session.data), path: { id: logId } }))
  return result.ok || result.reason === 'not-found' ? { ok: true, data: null } : result
}

export async function readScorePreview(input: ScorePreviewInput): Promise<LogResult<ImmersionScorePreview>> {
  return answer(immersionScorePreview({
    ...sessionRequest('immersionScorePreview', sessionCookie()),
    body: {
      language_code: input.languageCode,
      activity_id: input.activityId,
      ...measureBody(input.measure),
      tags: input.tags,
    },
  }))
}

/** The tags the viewer used most that contain `query`, with their usage counts, followed by common tags. */
export async function readTagSuggestions(query: string): Promise<LogResult<ImmersionTagSuggestions>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  return answer(immersionLogTagSuggestions({ ...sessionRequest('immersionLogTagSuggestions', cookie), query: { query } }))
}
