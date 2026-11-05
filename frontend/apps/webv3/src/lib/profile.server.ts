import {
  immersionProfileFindByUserId,
  immersionProfileYearlyActivityByUserId,
  immersionProfileYearlyActivitySplitByUserId,
  immersionProfileYearlyContestRegistrationsByUserId,
  immersionProfileYearlyScoresByUserId,
} from 'api-client'
import { contestPageSize, type ContestHistory, type ProfileStatistics, type Section } from './profile'
import { sessionCookie, sessionRequest } from './session.server'

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
  const cookie = sessionCookie()
  const read = (cookie?: string) => immersionProfileYearlyContestRegistrationsByUserId({
    ...sessionRequest(cookie),
    path: { userId, year },
    query: { page_size: contestPageSize, page: page - 1 },
  })

  let answer = await read(cookie).catch(() => undefined)
  if (cookie && answer?.response?.status === 401) answer = await read().catch(() => undefined)
  const data = answer?.data
  if (!data) return { ok: false }

  const contests = []
  for (const { contest } of data.registrations) {
    // Request bodies share the registration schema, so its contest is optional there; history needs it.
    if (!contest?.id) return { ok: false }
    contests.push({ id: contest.id, title: contest.title, private: contest.private })
  }
  return { ok: true, data: { contests, total: data.total_size } }
}

export async function readProfileStatistics(userId: string, year: number, page: number): Promise<ProfileStatistics> {
  const path = { userId, year }
  const [profile, activity, scores, split, contests] = await Promise.all([
    immersionProfileFindByUserId({ ...sessionRequest(), path: { userId } }).catch(() => undefined),
    section(immersionProfileYearlyActivityByUserId({ ...sessionRequest(), path })),
    section(immersionProfileYearlyScoresByUserId({ ...sessionRequest(), path })),
    section(immersionProfileYearlyActivitySplitByUserId({ ...sessionRequest(), path })),
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
