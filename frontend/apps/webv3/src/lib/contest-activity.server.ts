import {
  immersionContestListLogs,
  immersionContestModerationDetachLog,
  immersionContestProfileFetchActivity,
  immersionContestProfileFetchScores,
  type ImmersionLogs,
} from 'api-client'
import type { ContestProfile, Detach } from './contest-activity'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { draftOwnerCookie, logFailures } from './logs.server'
import { activityPageSize } from './profile'
import { sessionRequest } from './session.server'

const failed = { data: undefined, response: undefined }

/** Contest logs are public, so they are read anonymously. */
export async function readContestLogs({ contestId, userId, page }: { contestId: string; userId?: string; page: number }): Promise<Read<ImmersionLogs>> {
  const { data } = await immersionContestListLogs({
    ...sessionRequest(),
    path: { id: contestId },
    query: { page_size: activityPageSize, page: page - 1, include_deleted: false, ...(userId ? { user_id: userId } : {}) },
  }).catch(() => failed)
  return data ? { ok: true, data } : { ok: false }
}

/** A participant's scores in a contest; the API answers 404 when the contest or their registration doesn't exist. */
export async function readContestProfile(contestId: string, userId: string): Promise<ContestProfile> {
  const path = { id: contestId, user_id: userId }
  const [scores, daily] = await Promise.all([
    immersionContestProfileFetchScores({ ...sessionRequest(), path }).catch(() => failed),
    immersionContestProfileFetchActivity({ ...sessionRequest(), path }).catch(() => failed),
  ])

  if (scores.response?.status === 404) return { status: 'not-found' }
  const registration = scores.data?.registration
  const contest = registration?.contest
  if (!scores.data || !registration || !contest?.id) return { status: 'unavailable' }
  return {
    status: 'found',
    registration: { ...registration, contest: { ...contest, id: contest.id } },
    scores: scores.data.scores,
    daily: daily.data ? { ok: true, data: daily.data.rows } : { ok: false },
  }
}

/** Removes a log from a contest for the account that opened the page; the API decides who may moderate. */
export async function detachLog({ viewerId, contestId, logId, reason }: Detach): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(viewerId)
  if (!session.ok) return session

  const { error, response } = await immersionContestModerationDetachLog({
    ...sessionRequest(session.data),
    path: { id: contestId, log_id: logId },
    body: { reason },
  }).catch(() => ({ error: undefined, response: undefined }))
  if (response?.status === 200 && error === undefined) return { ok: true, data: null }
  return { ok: false, reason: (response && logFailures[response.status]) || 'unavailable' }
}
