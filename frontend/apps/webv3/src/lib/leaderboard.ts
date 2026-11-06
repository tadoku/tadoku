import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestView, ImmersionLeaderboard } from 'api-client'
import { z } from 'zod'
import type { Identified } from './home'

/** Whose scores a standings table ranks. Global boards count official contests only. */
export type StandingsScope =
  | { kind: 'contest'; contestId: string }
  | { kind: 'year'; year: number }
  | { kind: 'all-time' }

export const standingsPageSize = 50

/** The first year with official standings, where the year list starts as in Webv2. */
export const firstYear = 2020

export const leaderboardHref = {
  latest: (page = 1) => `/leaderboard/latest/${page}`,
  yearly: (year: number, page = 1) => `/leaderboard/yearly/${year}/${page}`,
  allTime: (page = 1) => `/leaderboard/all-time/${page}`,
}

/** A public read: the validated answer, or `unavailable` when the API failed or answered something unreadable. */
export type Read<T> = { ok: true; data: T } | { ok: false }

/** A contest's board: the contest whose standings and conditions the page shows. */
export interface ContestBoard {
  contest: Identified<ImmersionContestView>
}

/** A contest's board, or `none` when the contest doesn't exist: no official contest yet, or an unknown ID. */
export type BoardRead = { kind: 'board'; board: ContestBoard } | { kind: 'none' } | { kind: 'unavailable' }

const standingsInput = z.object({
  scope: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('contest'), contestId: z.guid() }),
    z.object({ kind: z.literal('year'), year: z.int() }),
    z.object({ kind: z.literal('all-time') }),
  ]),
  page: z.int().min(1),
})

const getStandings = createServerFn({ method: 'GET' })
  .inputValidator(standingsInput)
  .handler(async ({ data }) => {
    const { readStandings } = await import('./leaderboard.server')
    return readStandings(data.scope, data.page)
  })

const getLatestBoard = createServerFn({ method: 'GET' }).handler(async () => {
  const { readLatestBoard } = await import('./leaderboard.server')
  return readLatestBoard()
})

const getContestBoard = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readContestBoard } = await import('./leaderboard.server')
    return readContestBoard(data)
  })

/** One page of standings, counted from 1. */
export const standingsQuery = (scope: StandingsScope, page: number) => queryOptions({
  queryKey: ['standings', scope, page],
  queryFn: (): Promise<Read<ImmersionLeaderboard>> => getStandings({ data: { scope, page } }),
  staleTime: 30_000,
  retry: false,
})

export const latestBoardQuery = () => queryOptions({
  queryKey: ['latest-board'],
  queryFn: (): Promise<BoardRead> => getLatestBoard(),
  staleTime: 30_000,
  retry: false,
})

/** A contest's board as the viewer `viewerKey` names sees it: administrators can also open deleted contests. */
export const contestBoardQuery = (contestId: string, viewerKey: string) => queryOptions({
  queryKey: ['contest-board', contestId, viewerKey],
  queryFn: (): Promise<BoardRead> => getContestBoard({ data: contestId }),
  staleTime: 30_000,
  retry: false,
})

/** The leading integer of a path segment, as Webv2 reads its page and year segments, or `fallback`. A page path such as `2/extra` reads as 2. */
export function pathNumber(segment: string | undefined, fallback: number) {
  const value = Number.parseInt(segment ?? '', 10)
  return Number.isSafeInteger(value) ? value : fallback
}

export const pageNumber = (segment: string | undefined) => Math.max(1, pathNumber(segment, 1))

/** Where a contest's round and its registration stand on one UTC calendar day. Registration closes separately from the round. */
export interface ContestPhase {
  started: boolean
  ended: boolean
  ongoing: boolean
  /** Whether the contest accepts new participants. Registered participants can change languages until it ends. */
  registrationOpen: boolean
}

/** The phase on `today`, a `YYYY-MM-DD` UTC day: contest days run from the start's first to the end's last UTC second. */
export function contestPhase(contest: Pick<ImmersionContestView, 'contest_start' | 'contest_end' | 'registration_end'>, today: string): ContestPhase {
  const started = today >= contest.contest_start
  const ended = today > contest.contest_end
  return { started, ended, ongoing: started && !ended, registrationOpen: !ended && today <= contest.registration_end }
}

/** The UTC calendar day a query's answer was read, which decides a contest's phase on the server and after hydration. */
export const readDay = (dataUpdatedAt: number) => new Date(dataUpdatedAt).toISOString().slice(0, 10)
