import {
  immersionProfileYearlyContestRegistrationsByUserId,
  immersionContestCreate,
  immersionContestCreatePermissionCheck,
  immersionContestGetConfigurations,
  immersionContestList,
  type ImmersionContest,
  type ImmersionContestView,
} from 'api-client'
import type { NewContest } from './contest-form'
import type { ContestEditor, ContestList, ContestListScope } from './contests'
import { contestsPageSize } from './contests'
import type { Identified } from './home'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { answer, draftOwnerCookie } from './logs.server'
import { sessionCookie, sessionRequest } from './session.server'

const failed = { data: undefined, response: undefined }

/**
 * Reads a list as the visitor, whose session decides which private contests it includes. A rejected session cookie
 * falls back to the public list.
 */
async function listContests(query: NonNullable<Parameters<typeof immersionContestList>[0]>['query'], cookie?: string) {
  const list = (session?: string) => immersionContestList({ ...sessionRequest(session), query }).catch(() => failed)
  const result = await list(cookie)
  return result.response?.status === 401 && cookie ? list() : result
}

const identified = (contest: ImmersionContest): contest is Identified<ImmersionContest> => contest.id !== undefined

export async function readContests(scope: ContestListScope, page: number): Promise<Read<ContestList>> {
  if (scope.kind === 'participant') return readJoinedContests(scope.userId, scope.year, page)
  const cookie = sessionCookie()
  const query = {
    page_size: contestsPageSize,
    page: page - 1,
    official: scope.kind === 'official',
    include_deleted: false,
    ...(scope.kind === 'organizer' ? { user_id: scope.userId } : {}),
  }
  const [list, options] = await Promise.all([
    listContests(query, cookie),
    scope.kind !== 'official' && cookie
      ? immersionContestGetConfigurations(sessionRequest(cookie)).catch(() => failed)
      : failed,
  ])

  // Rows link to their contest, so an answer without contest IDs is malformed.
  const contests = list.data?.contests
  if (!list.data || !contests?.every(identified)) return { ok: false }
  const names = options.data && {
    languages: Object.fromEntries(options.data.languages.map(language => [language.code, language.name])),
    activities: Object.fromEntries(options.data.activities.map(activity => [String(activity.id), activity.name])),
  }
  return { ok: true, data: { contests, total: list.data.total_size, names: names ?? null } }
}

/**
 * The contests one account registered for in a year, newest first, as contest rows. The registrations carry each
 * contest's allowed languages and activities with their names, so the rows name restrictions without another read.
 */
async function readJoinedContests(userId: string, year: number, page: number): Promise<Read<ContestList>> {
  const read = (cookie?: string) => immersionProfileYearlyContestRegistrationsByUserId({
    ...sessionRequest(cookie),
    path: { userId, year },
    query: { page_size: contestsPageSize, page: page - 1 },
  }).catch(() => failed)
  const cookie = sessionCookie()
  let result = await read(cookie)
  if (cookie && result.response?.status === 401) result = await read()
  const views = result.data?.registrations.map(registration => registration.contest)
  if (!result.data || !views?.every((view): view is Identified<ImmersionContestView> => view?.id !== undefined)) {
    return { ok: false }
  }

  const contests = views.map(({ allowed_languages: languages, allowed_activities: activities, ...contest }) => ({
    ...contest,
    language_code_allow_list: languages?.map(language => language.code) ?? null,
    activity_type_id_allow_list: activities.map(activity => activity.id),
  }))
  const names = {
    languages: Object.fromEntries(views.flatMap(view => view.allowed_languages ?? []).map(language => [language.code, language.name])),
    activities: Object.fromEntries(views.flatMap(view => view.allowed_activities).map(activity => [String(activity.id), activity.name])),
  }
  return { ok: true, data: { contests, total: result.data.total_size, names } }
}

/** The API answers 200 when the visitor may create a contest; anything else hides the action, as in Webv2. */
export async function readCanCreateContest(): Promise<boolean> {
  const cookie = sessionCookie()
  if (!cookie) return false
  const { response } = await immersionContestCreatePermissionCheck(sessionRequest(cookie)).catch(() => failed)
  return response?.status === 200
}

export async function readContestEditor(): Promise<ContestEditor> {
  const cookie = sessionCookie()
  if (!cookie) return { kind: 'signed-out' }

  const [permission, options] = await Promise.all([
    immersionContestCreatePermissionCheck(sessionRequest(cookie)).catch(() => failed),
    immersionContestGetConfigurations(sessionRequest(cookie)).catch(() => failed),
  ])
  switch (permission.response?.status) {
    case 200: return options.data?.activities.length ? { kind: 'allowed', options: options.data } : { kind: 'unavailable' }
    case 401: return { kind: 'signed-out' }
    case 403: return { kind: 'forbidden' }
    default: return { kind: 'unavailable' }
  }
}

/** Creates a contest for the account that started the draft, and answers the new contest's ID. */
export async function saveContest(contest: NewContest): Promise<LogResult<string>> {
  const session = await draftOwnerCookie(contest.viewerId)
  if (!session.ok) return session

  const { kind } = contest
  const result = await answer(immersionContestCreate({
    ...sessionRequest(session.data),
    body: {
      title: contest.title,
      description: contest.description,
      contest_start: contest.contestStart,
      contest_end: contest.contestEnd,
      registration_end: contest.registrationEnd,
      official: kind.kind === 'official',
      private: kind.kind === 'community' && kind.unlisted,
      language_code_allow_list: kind.kind === 'community' ? kind.languageCodes : [],
      activity_type_id_allow_list: contest.activityIds,
    },
  }))
  if (!result.ok) return result
  // The contest page needs the new contest's ID, so an answer without one is malformed.
  return result.data.id ? { ok: true, data: result.data.id } : { ok: false, reason: 'unavailable' }
}
