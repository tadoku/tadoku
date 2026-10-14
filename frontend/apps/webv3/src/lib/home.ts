import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import { z } from 'zod'

const contestBase = z.object({
  id: z.uuid(),
  title: z.string(),
  description: z.string().optional(),
  contest_start: z.iso.date(),
  contest_end: z.iso.date(),
  registration_end: z.iso.date(),
  private: z.boolean(),
  official: z.boolean(),
  owner_user_id: z.uuid().optional(),
  owner_user_display_name: z.string().optional(),
  created_at: z.iso.datetime({ offset: true }).optional(),
  updated_at: z.iso.datetime({ offset: true }).optional(),
  deleted: z.boolean().optional(),
})

export const contestSchema = contestBase.extend({
  language_code_allow_list: z.array(z.string()).nullable().transform(value => value ?? []),
  activity_type_id_allow_list: z.array(z.int()),
})

export const contestViewSchema = contestBase.extend({
  allowed_languages: z.array(z.object({ code: z.string(), name: z.string() })).nullable().transform(value => value ?? []),
  allowed_activities: z.array(z.object({
    id: z.int(),
    name: z.string(),
    default: z.boolean().optional(),
    input_type: z.enum(['amount_primary', 'time_primary']).optional(),
  })),
})

export const contestListSchema = z.object({
  contests: z.array(contestSchema),
  total_size: z.int(),
  next_page_token: z.string(),
})

export const contestSummarySchema = z.object({
  participant_count: z.int(),
  language_count: z.int(),
  total_score: z.number(),
})

export const leaderboardSchema = z.object({
  entries: z.array(z.object({
    rank: z.int(),
    user_id: z.uuid(),
    user_display_name: z.string(),
    score: z.number(),
    is_tie: z.boolean(),
  })),
  total_size: z.int(),
  next_page_token: z.string(),
})

export type Contest = z.infer<typeof contestSchema>
export type ContestView = z.infer<typeof contestViewSchema>
export type ContestSummary = z.infer<typeof contestSummarySchema>
export type Leaderboard = z.infer<typeof leaderboardSchema>

export interface HomeData {
  latestContest: ContestView | null
  summary: ContestSummary | null
  standings: Leaderboard | null
  nextContest: Contest | null
}

const getHomeData = createServerFn({ method: 'GET' }).handler(async () => {
  const { readHomeData } = await import('./home.server')
  return readHomeData()
})

export const homeQueryOptions = () => queryOptions({
  queryKey: ['home'],
  queryFn: () => getHomeData(),
  staleTime: 30_000,
  retry: false,
  retryOnMount: false,
})
