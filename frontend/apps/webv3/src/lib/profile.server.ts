import {
  immersionProfileFindByUserId,
  immersionProfileListLogs,
  immersionProfileYearlyActivityByUserId,
  immersionProfileYearlyActivitySplitByUserId,
  immersionProfileYearlyContestRegistrationsByUserId,
  immersionProfileYearlyScoresByUserId,
  type ImmersionLogs,
} from 'api-client'
import {
  activityPageSize,
  contestPageSize,
  recentLogCount,
  type ContestHistory,
  type ProfileActivity,
  type ProfileStatistics,
  type Section,
} from './profile'
import { readAsVisitor, sessionRequest } from './session.server'

/** A generated operation's validated answer; any refusal, failure or malformed answer leaves the section unavailable. */
async function section<T>(request: Promise<{ data?: T }>): Promise<Section<T>> {
  const { data } = await request.catch(() => ({ data: undefined }))
  return data === undefined ? { ok: false } : { ok: true, data }
}

/**
 * Contest history includes private contests for their participant and administrators, so it is read with the
 * visitor's session. A rejected session cookie falls back to the public history.
 */
async function readContestHistory(userId: string, year: number, page: number): Promise<Section<ContestHistory>> {
  const answer = await readAsVisitor(cookie => immersionProfileYearlyContestRegistrationsByUserId({
    ...sessionRequest('immersionProfileYearlyContestRegistrationsByUserId', cookie),
    path: { userId, year },
    query: { page_size: contestPageSize, page: page - 1 },
  }).catch(() => ({ data: undefined, response: undefined })))
  const data = answer.data
  if (!data) return { ok: false }

  const contests = []
  for (const { contest, languages } of data.registrations) {
    // Request bodies share the registration schema, so its contest is optional there; history needs it.
    if (!contest?.id) return { ok: false }
    contests.push({
      id: contest.id,
      title: contest.title,
      private: contest.private,
      contest_start: contest.contest_start,
      contest_end: contest.contest_end,
      languages: languages.map(language => language.name),
    })
  }
  return { ok: true, data: { contests, total: data.total_size } }
}

export async function readProfileStatistics(userId: string, year: number, page: number): Promise<ProfileStatistics> {
  const path = { userId, year }
  const [profile, activity, scores, split, contests] = await Promise.all([
    immersionProfileFindByUserId({ ...sessionRequest('immersionProfileFindByUserId'), path: { userId } }).catch(() => undefined),
    section(immersionProfileYearlyActivityByUserId({ ...sessionRequest('immersionProfileYearlyActivityByUserId'), path })),
    section(immersionProfileYearlyScoresByUserId({ ...sessionRequest('immersionProfileYearlyScoresByUserId'), path })),
    section(immersionProfileYearlyActivitySplitByUserId({ ...sessionRequest('immersionProfileYearlyActivitySplitByUserId'), path })),
    readContestHistory(userId, year, page),
  ])

  if (profile?.response?.status === 404) return { status: 'not-found' }
  if (!profile?.data) return { status: 'unavailable' }
  return {
    status: 'found',
    today: new Date().toISOString().slice(0, 10),
    profile: profile.data,
    activity,
    scores,
    split,
    contests,
  }
}

/** The newest logs for the home page, without the profile the page's statistics already read. */
export function readRecentLogs(userId: string): Promise<Section<ImmersionLogs>> {
  return section(immersionProfileListLogs({
    ...sessionRequest('immersionProfileListLogs'),
    path: { user_id: userId },
    query: { page_size: recentLogCount, page: 0, include_deleted: false },
  }))
}

export async function readProfileActivity(userId: string, page: number, date?: string): Promise<ProfileActivity> {
  const [profile, logs] = await Promise.all([
    immersionProfileFindByUserId({ ...sessionRequest('immersionProfileFindByUserId'), path: { userId } }).catch(() => undefined),
    section(immersionProfileListLogs({
      ...sessionRequest('immersionProfileListLogs'),
      path: { user_id: userId },
      query: { page_size: activityPageSize, page: page - 1, include_deleted: false, ...(date ? { date } : {}) },
    })),
  ])

  if (profile?.response?.status === 404) return { status: 'not-found' }
  if (!profile?.data) return { status: 'unavailable' }
  return { status: 'found', profile: profile.data, logs }
}
