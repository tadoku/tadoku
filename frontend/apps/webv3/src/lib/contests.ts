import { queryOptions, useQueryClient, type QueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContest, ImmersionContestConfigurationOptions } from 'api-client'
import { z } from 'zod'
import { newContestSchema, type NewContest } from './contest-form'
import type { Identified } from './home'
import type { Read } from './leaderboard'
import type { LogResult } from './logs'
import { viewerKey, viewerQueryOptions, type Viewer } from './session'

/** Which contests a list holds. An organizer's list holds the community contests one account created. */
export type ContestListScope = { kind: 'official' } | { kind: 'community' } | { kind: 'organizer'; userId: string }

export const contestsPageSize = 25

export type ContestCollection = 'official' | 'community' | 'mine'

export const contestsHref: Record<ContestCollection, (page?: number) => string> = {
  official: (page = 1) => `/contests/official/${page}`,
  community: (page = 1) => `/contests/user-contests/${page}`,
  mine: (page = 1) => `/contests/my-contests/${page}`,
}

export interface ContestList {
  /** In the API's order, newest first. */
  contests: Identified<ImmersionContest>[]
  total: number
  /**
   * Language names by code and activity names by ID, read for community and organizer lists when the visitor sends a
   * session cookie; null otherwise or when they can't be read. Screens show them to signed-in viewers only, as Webv2.
   */
  names: { languages: Record<string, string>; activities: Record<string, string> } | null
}

const listInput = z.object({
  scope: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('official') }),
    z.object({ kind: z.literal('community') }),
    z.object({ kind: z.literal('organizer'), userId: z.guid() }),
  ]),
  page: z.int().min(1),
})

const getContests = createServerFn({ method: 'GET' })
  .inputValidator(listInput)
  .handler(async ({ data }) => {
    const { readContests } = await import('./contests.server')
    return readContests(data.scope, data.page)
  })

const getCanCreateContest = createServerFn({ method: 'GET' }).handler(async () => {
  const { readCanCreateContest } = await import('./contests.server')
  return readCanCreateContest()
})

/**
 * One page of a contest list, counted from 1, as the viewer `forViewer` names sees it: the API adds a signed-in
 * account's own private contests, and every private contest for administrators.
 */
export const contestsQuery = (scope: ContestListScope, page: number, forViewer: string) => queryOptions({
  queryKey: ['contests', scope, page, forViewer],
  queryFn: (): Promise<Read<ContestList>> => getContests({ data: { scope, page } }),
  staleTime: 30_000,
  retry: false,
})

/** Whether the signed-in viewer may create a contest, which shows the create action. */
export const canCreateContestQuery = (forViewer: string) => queryOptions({
  queryKey: ['can-create-contest', forViewer],
  queryFn: (): Promise<boolean> => getCanCreateContest(),
  staleTime: 30_000,
  retry: false,
})

/** The list a collection shows this viewer; null while the viewer has no account to organize contests with. */
export function contestScope(collection: ContestCollection, viewer: Viewer | undefined): ContestListScope | null {
  if (collection !== 'mine') return { kind: collection }
  return viewer && 'userId' in viewer ? { kind: 'organizer', userId: viewer.userId } : null
}

/** Prefetches a collection page and, for signed-in viewers, whether they may create contests. */
export async function loadContests(queryClient: QueryClient, collection: ContestCollection, page: number) {
  const viewer = await queryClient.ensureQueryData(viewerQueryOptions())
  const scope = contestScope(collection, viewer)
  await Promise.all([
    scope ? queryClient.ensureQueryData(contestsQuery(scope, page, viewerKey(viewer))) : null,
    viewer.status === 'signed-in' ? queryClient.ensureQueryData(canCreateContestQuery(viewerKey(viewer))) : null,
  ])
}

/**
 * Whether the signed-in viewer may create a contest, and the languages and activities to offer when they may. The API
 * refuses accounts younger than a month and accounts that reached their yearly limit.
 */
export type ContestEditor =
  | { kind: 'allowed'; options: ImmersionContestConfigurationOptions }
  | { kind: 'forbidden' }
  | { kind: 'signed-out' }
  | { kind: 'unavailable' }

const getContestEditor = createServerFn({ method: 'GET' }).handler(async () => {
  const { readContestEditor } = await import('./contests.server')
  return readContestEditor()
})

const createContest = createServerFn({ method: 'POST' })
  .inputValidator(newContestSchema)
  .handler(async ({ data }) => {
    const { saveContest } = await import('./contests.server')
    return saveContest(data)
  })

/** The contest form for the account that started the draft; its options don't change while that draft is open. */
export const contestEditorQuery = (viewerId: string) => queryOptions({
  queryKey: ['contest-editor', viewerId],
  queryFn: (): Promise<ContestEditor> => getContestEditor(),
  retry: false,
  staleTime: Infinity,
})

/**
 * Creates a contest for the account that started the draft and answers its ID. A created contest marks contest lists
 * and creation permissions stale; a refusal because another account signed in refreshes the viewer. A failed request is `unavailable`, because the contest may or may not exist.
 */
export function useCreateContest() {
  const queryClient = useQueryClient()

  return async (contest: NewContest): Promise<LogResult<string>> => {
    const result = await createContest({ data: contest })
      .catch((): LogResult<string> => ({ ok: false, reason: 'unavailable' }))
    if (result.ok) {
      // Lists gain the contest and the allowance shrinks; refetch them when next shown, not under the open form.
      await Promise.all(['contests', 'can-create-contest', 'contest-editor'].map(key =>
        queryClient.invalidateQueries({ queryKey: [key], refetchType: 'none' })))
    } else if (result.reason === 'switched') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    }
    return result
  }
}
