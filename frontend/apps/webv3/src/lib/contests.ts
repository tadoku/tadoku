import { queryOptions, type QueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContest } from 'api-client'
import { z } from 'zod'
import type { Identified } from './home'
import type { Read } from './leaderboard'
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
   * Language names by code and activity names by ID. As in Webv2, community and organizer lists read them for
   * signed-in visitors; otherwise, or when they can't be read, the list leaves the restrictions out.
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
