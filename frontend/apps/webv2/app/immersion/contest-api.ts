import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ContestFormSchema } from '@app/immersion/ContestForm'
import type { ContestRegistrationFormSchema } from '@app/immersion/ContestRegistration'
import { Activity, Language, Logs } from '@app/immersion/api'
import { Score } from '@app/immersion/scoring-api'

const root = () => `${browserAppUrls().apiEndpoint}/immersion`

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
      const response = await fetch(`${root()}/contests/configuration-options`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
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

      const response = await fetch(`${root()}/contests`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Contest.parse(await response.json())
    },
    onSuccess(data) {
      onSuccess(data.id)
    },
  })

export const useContestCreatePermissionCheck = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'createPermissionCheck'],
    async (): Promise<true> => {
      const response = await fetch(`${root()}/contests/create-permissions`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return true
    },
    { ...options, retry: false },
  )

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
  deleted: z.boolean().optional(),
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
      const response = await fetch(
        `${root()}/contests?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
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
      const response = await fetch(`${root()}/contests/${id}`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestView.parse(await response.json())
    },
    options,
  )

export const useLatestOfficialContest = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findLatestOfficial'],
    async (): Promise<ContestView> => {
      const response = await fetch(`${root()}/contests/latest-official`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
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
    async (): Promise<ContestRegistrationView | null> => {
      const response = await fetch(`${root()}/contests/${id}/registration`)

      if (response.status === 204) {
        return null
      }

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestRegistrationView.parse(await response.json())
    },
    { ...options, retry: false },
  )

export const ContestSummary = z.object({
  participant_count: z.number(),
  language_count: z.number(),
  total_score: z.number(),
})

export type ContestSummary = z.infer<typeof ContestSummary>

export const useContestSummary = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestSummary', id],
    async (): Promise<ContestSummary | null> => {
      const response = await fetch(`${root()}/contests/${id}/summary`)

      if (response.status === 404) {
        return null
      }

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestSummary.parse(await response.json())
    },
    { ...options, retry: false },
  )

export const useContestRegistrationUpdate = (onSuccess: () => void) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (registration: ContestRegistrationFormSchema) => {
      const response = await fetch(
        `${root()}/contests/${registration.contest_id}/registration`,
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
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return
    },
    onSuccess() {
      queryClient.invalidateQueries([
        'contest',
        'ongoing-contest-registrations',
      ])
      onSuccess()
    },
  })
}

export const ContestRegistrationsView = z.object({
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
      const response = await fetch(`${root()}/contests/ongoing-registrations`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestRegistrationsView.parse(await response.json())
    },
    options,
  )

const ContestProfileScores = z.object({
  overall_score: z.number(),
  registration: ContestRegistrationView,
  scores: z.array(Score),
})

export type ContestProfileScores = z.infer<typeof ContestProfileScores>

export const useContestProfileScores = (
  opts: {
    userId: string
    contestId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', opts.contestId, 'profile', opts.userId, 'scores'],
    async (): Promise<ContestProfileScores> => {
      const response = await fetch(
        `${root()}/contests/${opts.contestId}/profile/${opts.userId}/scores`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestProfileScores.parse(await response.json())
    },
    options,
  )

const ContestProfileActivity = z.object({
  rows: z.array(
    z.object({
      date: z.string(),
      language_code: z.string(),
      score: z.number(),
    }),
  ),
})

export type ContestProfileActivity = z.infer<typeof ContestProfileActivity>

export const useContestProfileActivity = (
  opts: {
    userId: string
    contestId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', opts.contestId, 'profile', opts.userId, 'readingActivity'],
    async (): Promise<ContestProfileActivity> => {
      const response = await fetch(
        `${root()}/contests/${opts.contestId}/profile/${opts.userId}/activity`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestProfileActivity.parse(await response.json())
    },
    options,
  )

export const getContestLogsQueryKey = (opts: {
  pageSize: number
  page: number
  includeDeleted: boolean
  userId?: string
  contestId: string
}) => [
  'contest',
  opts.contestId,
  'profile',
  opts.userId,
  'logs',
  opts.pageSize,
  opts.page,
  opts.includeDeleted,
]

export const useContestLogs = (
  opts: {
    pageSize: number
    page: number
    includeDeleted: boolean
    userId?: string
    contestId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    getContestLogsQueryKey(opts),
    async (): Promise<Logs> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        include_deleted: opts.includeDeleted.toString(),
        ...(opts.userId ? { user_id: opts.userId } : {}),
      }
      const response = await fetch(
        `${root()}/contests/${opts.contestId}/logs?${new URLSearchParams(
          params,
        )}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Logs.parse(await response.json())
    },
    options,
  )

export const useDetachLogFromContest = (
  onSuccess: () => void,
  onError: () => void,
) =>
  useMutation({
    mutationFn: async ({
      contestId,
      logId,
      reason,
    }: {
      contestId: string
      logId: string
      reason: string
    }) => {
      const response = await fetch(
        `${root()}/contests/${contestId}/moderation/detach/${logId}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ reason }),
        },
      )
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
    },
    onSuccess,
    onError,
  })
