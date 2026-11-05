import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestSummary, ImmersionContestView, ImmersionLeaderboard, ImmersionLog } from 'api-client'
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

/** A contest's board. Its summary and recent logs are extras the page leaves out when they can't be read. */
export interface ContestBoard {
  contest: Identified<ImmersionContestView>
  summary: ImmersionContestSummary | null
  recentLogs: ImmersionLog[]
}

/** The latest official contest's board, or `none` when no official contest exists yet. */
export type LatestBoard = { kind: 'board'; board: ContestBoard } | { kind: 'none' } | { kind: 'unavailable' }

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

const getRegistered = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readRegistered } = await import('./leaderboard.server')
    return readRegistered(data)
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
  queryFn: (): Promise<LatestBoard> => getLatestBoard(),
  staleTime: 30_000,
  retry: false,
})

/** Whether the signed-in viewer `viewerKey` names has joined the contest; null when that can't be read. */
export const registeredQuery = (contestId: string, viewerKey: string) => queryOptions({
  queryKey: ['registered', contestId, viewerKey],
  queryFn: (): Promise<boolean | null> => getRegistered({ data: contestId }),
  staleTime: 30_000,
  retry: false,
})

/** The leading integer of a path segment, as Webv2 reads its page and year segments, or `fallback`. A page path such as `2/extra` reads as 2. */
export function pathNumber(segment: string | undefined, fallback: number) {
  const value = Number.parseInt(segment ?? '', 10)
  return Number.isSafeInteger(value) ? value : fallback
}

export const pageNumber = (segment: string | undefined) => Math.max(1, pathNumber(segment, 1))
