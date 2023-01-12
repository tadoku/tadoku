import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery } from 'react-query'
import { ContestFormSchema } from '@app/contests/ContestForm'
import { DateTime } from 'luxon'
import { ContestRegistrationFormSchema } from './ContestRegistration'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion/contests`

export const Language = z.object({
  code: z.string(),
  name: z.string(),
})

export type Language = z.infer<typeof Language>

export const Activity = z.object({
  id: z.number(),
  name: z.string(),
  default: z.boolean().nullable().optional(),
})

export type Activity = z.infer<typeof Activity>

const ContestConfigurationOptions = z.object({
  languages: z.array(Language),
  activities: z.array(Activity),
  can_create_official_round: z.boolean(),
})

export type ContestConfigurationOptions = z.infer<
  typeof ContestConfigurationOptions
>

export const useContestConfigurationOptions = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'configuration-options'],
    async (): Promise<ContestConfigurationOptions> => {
      const response = await fetch(`${root}/configuration-options`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestConfigurationOptions.parse(await response.json())
    },
    options,
  )

export const useCreateContest = (onSuccess: (id: string) => void) =>
  useMutation({
    mutationFn: async (contest: ContestFormSchema) => {
      const payload = {
        ...contest,
        activity_type_id_allow_list: contest.activity_type_id_allow_list.map(
          it => it.id,
        ),
        language_code_allow_list: contest.language_code_allow_list.map(
          it => it.code,
        ),
      }

      const res = await fetch(root, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      return await res.json()
    },
    onSuccess(data) {
      onSuccess(data.id)
    },
  })

const Contest = z.object({
  id: z.string(),
  contest_start: z.string(),
  contest_end: z.string(),
  registration_end: z.string(),
  title: z.string(),
  description: z.string().optional().nullable(),
  private: z.boolean(),
  official: z.boolean(),
  language_code_allow_list: z.array(z.string()).nullable(),
  activity_type_id_allow_list: z.array(z.number()),
  deleted: z.boolean(),
})

export type Contest = z.infer<typeof Contest>

const Contests = z.object({
  contests: z.array(Contest),
  total_size: z.number(),
  next_page_token: z.string(),
})

export type Contests = z.infer<typeof Contests>

export const useContestList = (
  opts: {
    pageSize: number
    page: number
    includeDeleted: boolean
    official: boolean
    userId?: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'list', opts],
    async (): Promise<Contests> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        official: opts.official.toString(),
        include_deleted: opts.includeDeleted.toString(),
        ...(opts.userId ? { user_id: opts.userId.toString() } : {}),
      }
      const response = await fetch(`${root}?${new URLSearchParams(params)}`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return Contests.parse(await response.json())
    },
    options,
  )

const ContestView = z.object({
  id: z.string().nullable(),
  contest_start: z.string(),
  contest_end: z.string(),
  registration_end: z.string(),
  title: z.string(),
  description: z.string().optional().nullable(),
  private: z.boolean(),
  official: z.boolean(),
  owner_user_id: z.string().nullable().optional(),
  owner_user_display_name: z.string().nullable().optional(),
  allowed_languages: z.array(Language).nullable().optional(),
  allowed_activities: z.array(Activity),
  deleted: z.boolean().nullable().optional(),
})

export type ContestView = z.infer<typeof ContestView>

export const useContest = (id: string, options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findByID', id],
    async (): Promise<ContestView> => {
      const response = await fetch(`${root}/${id}`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestView.parse(await response.json())
    },
    options,
  )

export const useLatestOfficialContest = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findLatestOfficial'],
    async (): Promise<ContestView> => {
      const response = await fetch(`${root}/latest-official`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestView.parse(await response.json())
    },
    options,
  )

export const ContestRegistrationView = z.object({
  id: z.string(),
  contest_id: z.string(),
  user_id: z.string(),
  user_display_name: z.string(),
  languages: z.array(
    z.object({
      code: z.string(),
      name: z.string(),
    }),
  ),
  contest: ContestView.nullable().optional(),
})

export type ContestRegistrationView = z.infer<typeof ContestRegistrationView>

export const useContestRegistration = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestRegistrationForUser', id],
    async (): Promise<ContestRegistrationView | undefined> => {
      const response = await fetch(`${root}/${id}/registration`)

      if (response.status === 404) {
        return undefined
      }

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestRegistrationView.parse(await response.json())
    },
    { ...options, retry: false },
  )

export const useContestRegistrationUpdate = (onSuccess: () => void) =>
  useMutation({
    mutationFn: async (registration: ContestRegistrationFormSchema) => {
      const res = await fetch(
        `${root}/${registration.contest_id}/registration`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            language_codes: registration.new_languages.map(it => it.code),
          }),
        },
      )
      return
    },
    onSuccess() {
      onSuccess()
    },
  })

const LeaderboardEntry = z.object({
  rank: z.number(),
  user_id: z.string(),
  user_display_name: z.string(),
  score: z.number(),
  is_tie: z.boolean(),
})

export type LeaderboardEntry = z.infer<typeof LeaderboardEntry>

const Leaderboard = z.object({
  entries: z.array(LeaderboardEntry),
  total_size: z.number(),
  next_page_token: z.string(),
})

export type Leaderboard = z.infer<typeof Leaderboard>

export const useContestLeaderboard = (
  opts: {
    contestId: string
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'leaderboard', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root}/${opts.contestId}/leaderboard?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error('could not fetch leaderboard')
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )

const ContestRegistrationsView = z.object({
  registrations: z.array(ContestRegistrationView),
  next_page_token: z.string(),
  total_size: z.number(),
})

export type ContestRegistrationsView = z.infer<typeof ContestRegistrationsView>

export const useOngoingContestRegistrations = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'ongoing-contest-registrations'],
    async (): Promise<ContestRegistrationsView> => {
      const response = await fetch(`${root}/ongoing-registrations`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return ContestRegistrationsView.parse(await response.json())
    },
    options,
  )
