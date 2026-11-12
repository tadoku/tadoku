import {
  immersionContestListLogs,
  immersionContestModerationDetachLog,
  immersionContestProfileFetchScores,
  immersionContestProfileFetchSummary,
  type ImmersionLogs,
} from 'api-client'
import type { ContestProfile, ContestSummary, Detach } from './contest-activity'
import type { LogFilters } from './contest-filters'
import { utcToday } from './format'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { draftOwnerCookie, logFailures } from './logs.server'
import { activityPageSize } from './profile'
import { sessionRequest } from './session.server'

const failed = { data: undefined, response: undefined }

/** The filters the API reads, leaving out empty ones. */
const filterQuery = (filters?: LogFilters) =>
  Object.fromEntries(Object.entries(filters ?? {}).filter(([, values]) => values.length)) as Partial<LogFilters>

/** Contest logs are public, so they are read anonymously. */
export async function readContestLogs({ contestId, userId, page, filters }: {
  contestId: string
  userId?: string
  page: number
  filters?: LogFilters
}): Promise<Read<ImmersionLogs>> {
  const { data } = await immersionContestListLogs({
    ...sessionRequest('immersionContestListLogs'),
    path: { id: contestId },
    query: {
      page_size: activityPageSize,
      page: page - 1,
      include_deleted: false,
      ...(userId ? { user_id: userId } : {}),
      ...filterQuery(filters),
    },
  }).catch(() => failed)
  return data ? { ok: true, data } : { ok: false }
}

/** A participant's scores in a contest; the API answers 404 when the contest or their registration doesn't exist. */
export async function readContestProfile(contestId: string, userId: string): Promise<ContestProfile> {
  const path = { id: contestId, user_id: userId }
  const scores = await immersionContestProfileFetchScores({ ...sessionRequest('immersionContestProfileFetchScores'), path })
    .catch(() => failed)

  if (scores.response?.status === 404) return { status: 'not-found' }
  const registration = scores.data?.registration
  const contest = registration?.contest
  if (!scores.data || !registration || !contest?.id) return { status: 'unavailable' }
  return {
    status: 'found',
    registration: { ...registration, contest: { ...contest, id: contest.id } },
    scores: scores.data.scores,
  }
}

/** A participant's contest summary under the filters, read anonymously like their scores. */
export async function readContestSummary({ contestId, userId, filters }: {
  contestId: string
  userId: string
  filters: LogFilters
}): Promise<Read<ContestSummary>> {
  const { data } = await immersionContestProfileFetchSummary({
    ...sessionRequest('immersionContestProfileFetchSummary'),
    path: { id: contestId, user_id: userId },
    query: filterQuery(filters),
  }).catch(() => failed)
  return data ? { ok: true, data: { ...data, today: utcToday() } } : { ok: false }
}

/** Removes a log from a contest for the account that opened the page; the API decides who may moderate. */
export async function detachLog({ viewerId, contestId, logId, reason }: Detach): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(viewerId)
  if (!session.ok) return session

  const { error, response } = await immersionContestModerationDetachLog({
    ...sessionRequest('immersionContestModerationDetachLog', session.data),
    path: { id: contestId, log_id: logId },
    body: { reason },
  }).catch(() => ({ error: undefined, response: undefined }))
  if (response?.status === 200 && error === undefined) return { ok: true, data: null }
  return { ok: false, reason: (response && logFailures[response.status]) || 'unavailable' }
}
