import {
  immersionContestCreate,
  immersionContestCreatePermissionCheck,
  immersionContestGetConfigurations,
  immersionContestList,
  immersionContestListRegistrations,
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
import { readAsVisitor, sessionCookie, sessionRequest } from './session.server'

const failed = { data: undefined, response: undefined }

/** Reads a list as the visitor, whose session decides which private contests it includes. */
function listContests(query: NonNullable<Parameters<typeof immersionContestList>[0]>['query']) {
  return readAsVisitor(cookie => immersionContestList({ ...sessionRequest('immersionContestList', cookie), query }).catch(() => failed))
}

const identified = (contest: ImmersionContest): contest is Identified<ImmersionContest> => contest.id !== undefined

export async function readContests(scope: ContestListScope, page: number): Promise<Read<ContestList>> {
  if (scope.kind === 'joined') return readJoinedContests(page)
  const cookie = sessionCookie()
  const query = {
    page_size: contestsPageSize,
    page: page - 1,
    official: scope.kind === 'official',
    include_deleted: false,
    ...(scope.kind === 'organizer' ? { user_id: scope.userId } : {}),
  }
  const [list, options] = await Promise.all([
    listContests(query),
    scope.kind !== 'official' && cookie
      ? immersionContestGetConfigurations(sessionRequest('immersionContestGetConfigurations', cookie)).catch(() => failed)
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
 * Every contest the signed-in account registered for, newest contest first, as contest rows. The registrations carry
 * each contest's allowed activities with their names, so the rows name restrictions without another read; their
 * language lists are always empty, so rows name no language restriction.
 */
async function readJoinedContests(page: number): Promise<Read<ContestList>> {
  const result = await immersionContestListRegistrations({
    ...sessionRequest('immersionContestListRegistrations', sessionCookie()),
    query: { page_size: contestsPageSize, page: page - 1 },
  }).catch(() => failed)
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
  const { response } = await immersionContestCreatePermissionCheck(sessionRequest('immersionContestCreatePermissionCheck', cookie)).catch(() => failed)
  return response?.status === 200
}

export async function readContestEditor(): Promise<ContestEditor> {
  const cookie = sessionCookie()
  if (!cookie) return { kind: 'signed-out' }

  const [permission, options] = await Promise.all([
    immersionContestCreatePermissionCheck(sessionRequest('immersionContestCreatePermissionCheck', cookie)).catch(() => failed),
    immersionContestGetConfigurations(sessionRequest('immersionContestGetConfigurations', cookie)).catch(() => failed),
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
    ...sessionRequest('immersionContestCreate', session.data),
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
