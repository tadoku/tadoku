import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { ContestFormSchema } from '@app/immersion/ContestForm'
import { ContestRegistrationFormSchema } from '@app/immersion/ContestRegistration'
import { NewLogAPISchema } from '@app/immersion/NewLogForm/domain'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion`
const authzRoot = `${publicRuntimeConfig.apiEndpoint}/authz`

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
      const response = await fetch(`${root}/contests/configuration-options`)

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

      const response = await fetch(`${root}/contests`, {
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
    async (): Promise<void | undefined> => {
      const response = await fetch(`${root}/contests/create-permissions`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
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
        `${root}/contests?${new URLSearchParams(params)}`,
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
      const response = await fetch(`${root}/contests/${id}`)

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
      const response = await fetch(`${root}/contests/latest-official`)

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
    async (): Promise<ContestRegistrationView | undefined> => {
      const response = await fetch(`${root}/contests/${id}/registration`)

      if (response.status === 204) {
        return undefined
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
    async (): Promise<ContestSummary | undefined> => {
      const response = await fetch(`${root}/contests/${id}/summary`)

      if (response.status === 404) {
        return undefined
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
        `${root}/contests/${registration.contest_id}/registration`,
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
      queryClient.invalidateQueries(['contest', 'ongoing-contest-registrations'])
      onSuccess()
    },
  })
}

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
        `${root}/contests/${opts.contestId}/leaderboard?${new URLSearchParams(
          params,
        )}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )

export const useYearlyLeaderboard = (
  opts: {
    year: number
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'yearly', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root}/leaderboard/yearly/${opts.year}?${new URLSearchParams(
          params,
        )}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Leaderboard.parse(await response.json())
    },
    options,
  )

export const useAllTimeLeaderboard = (
  opts: {
    pageSize: number
    page: number
    languageCode?: string
    activityId?: number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['leaderboard', 'all-time', opts],
    async (): Promise<Leaderboard> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        ...(opts.languageCode ? { language_code: opts.languageCode } : {}),
        ...(opts.activityId ? { activity_id: opts.activityId.toString() } : {}),
      }
      const response = await fetch(
        `${root}/leaderboard/global?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
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
      const response = await fetch(`${root}/contests/ongoing-registrations`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestRegistrationsView.parse(await response.json())
    },
    options,
  )

export const Score = z.object({
  language_code: z.string(),
  language_name: z.string().optional(),
  score: z.number(),
})

export type Score = z.infer<typeof Score>

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
        `${root}/contests/${opts.contestId}/profile/${opts.userId}/scores`,
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
        `${root}/contests/${opts.contestId}/profile/${opts.userId}/activity`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestProfileActivity.parse(await response.json())
    },
    options,
  )

const ContestRegistrationReference = z.object({
  registration_id: z.string(),
  contest_id: z.string(),
  contest_end: z.string(),
  title: z.string(),
  owner_user_display_name: z.string().optional(),
  official: z.boolean().optional(),
})

export type ContestRegistrationReference = z.infer<
  typeof ContestRegistrationReference
>

export const Log = z.object({
  id: z.string(),
  user_id: z.string(),
  user_display_name: z.string().optional(),
  description: z.string().optional(),
  language: Language,
  activity: Activity,
  unit_name: z.string(),
  tags: z.array(z.string()),
  amount: z.number(),
  modifier: z.number(),
  score: z.number(),
  registrations: z.array(ContestRegistrationReference).optional(),
  created_at: z.string(),
  deleted: z.boolean(),
})

export type Log = z.infer<typeof Log>

const Logs = z.object({
  logs: z.array(Log),
  total_size: z.number(),
  next_page_token: z.string(),
})

export type Logs = z.infer<typeof Logs>

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
        `${root}/contests/${opts.contestId}/logs?${new URLSearchParams(
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

export const useProfileLogs = (
  opts: {
    pageSize: number
    page: number
    includeDeleted: boolean
    userId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    [
      'profile',
      opts.userId,
      'logs',
      opts.pageSize,
      opts.page,
      opts.includeDeleted,
    ],
    async (): Promise<Logs> => {
      const params = {
        page_size: opts.pageSize.toString(),
        page: (opts.page - 1).toString(),
        include_deleted: opts.includeDeleted.toString(),
      }
      const response = await fetch(
        `${root}/users/${opts.userId}/logs?${new URLSearchParams(params)}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Logs.parse(await response.json())
    },
    options,
  )

const UserProfile = z.object({
  id: z.string(),
  display_name: z.string(),
  created_at: z.string(),
})

export type UserProfile = z.infer<typeof UserProfile>

export const useUserProfile = (
  opts: {
    userId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'profile'],
    async (): Promise<UserProfile> => {
      const response = await fetch(`${root}/users/${opts.userId}/profile`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserProfile.parse(await response.json())
    },
    options,
  )

export const Unit = z.object({
  id: z.string(),
  log_activity_id: z.number(),
  name: z.string(),
  modifier: z.number(),
  language_code: z.string().nullable().optional(),
})

export type Unit = z.infer<typeof Unit>

const LogConfigurationOptions = z.object({
  languages: z.array(Language),
  activities: z.array(Activity),
  units: z.array(Unit),
})

export type LogConfigurationOptions = z.infer<typeof LogConfigurationOptions>

export const useLogConfigurationOptions = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'log', 'configuration-options'],
    async (): Promise<LogConfigurationOptions> => {
      const response = await fetch(`${root}/logs/configuration-options`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return LogConfigurationOptions.parse(await response.json())
    },
    options,
  )

const TagSuggestion = z.object({
  tag: z.string(),
  count: z.number(),
})

export type TagSuggestion = z.infer<typeof TagSuggestion>

const TagSuggestions = z.object({
  suggestions: z.array(TagSuggestion),
})

export const fetchTagSuggestions = async (query: string): Promise<TagSuggestion[]> => {
  const response = await fetch(
    `${root}/logs/tag-suggestions?query=${encodeURIComponent(query)}`,
  )
  if (!response.ok) return []
  const data = TagSuggestions.parse(await response.json())
  return data.suggestions
}

export const useCreateLog = (onSuccess: (id: string) => void) =>
  useMutation({
    mutationFn: async (contest: NewLogAPISchema) => {
      const response = await fetch(`${root}/logs`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(contest),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Log.parse(await response.json())
    },
    onSuccess(data) {
      onSuccess(data.id)
    },
  })

export type CreateLogV2Payload = {
  language_code: string
  activity_id: number
  amount: number
  unit_id: string
  tags: string[]
  description?: string
}

export const useCreateLogV2 = (onSuccess: (log: Log) => void) =>
  useMutation({
    mutationFn: async (payload: CreateLogV2Payload) => {
      const response = await fetch(`${root}/logs`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Log.parse(await response.json())
    },
    onSuccess(data) {
      onSuccess(data)
    },
  })

export const useUpdateLogContestRegistrations = (onSuccess: (log: Log) => void) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      logId,
      registrationIds,
    }: {
      logId: string
      registrationIds: string[]
    }) => {
      const response = await fetch(`${root}/logs/${logId}/contest-registrations`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ registration_ids: registrationIds }),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Log.parse(await response.json())
    },
    onSuccess(data) {
      queryClient.invalidateQueries(['log', 'findByID', data.id])
      onSuccess(data)
    },
  })
}

export const useLog = (id: string, options?: { enabled?: boolean }) =>
  useQuery(
    ['log', 'findByID', id],
    async (): Promise<Log> => {
      const response = await fetch(`${root}/logs/${id}`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Log.parse(await response.json())
    },
    options,
  )

const UserActivityScore = z.object({
  date: z.string(),
  score: z.number(),
})

export type UserActivityScore = z.infer<typeof UserActivityScore>

const UserActivity = z.object({
  total_updates: z.number(),
  scores: z.array(UserActivityScore),
})

export type UserActivity = z.infer<typeof UserActivity>

export const useUserYearlyActivity = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'yearly-activity', opts.year],
    async (): Promise<UserActivity> => {
      const response = await fetch(
        `${root}/users/${opts.userId}/activity/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserActivity.parse(await response.json())
    },
    options,
  )

const ProfileScores = z.object({
  overall_score: z.number(),
  scores: z.array(Score),
})

export type ProfileScores = z.infer<typeof ProfileScores>

export const useProfileScores = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'scores', opts.year],
    async (): Promise<ProfileScores> => {
      const response = await fetch(
        `${root}/users/${opts.userId}/scores/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ProfileScores.parse(await response.json())
    },
    options,
  )

export const useYearlyContestRegistrations = (
  opts: {
    userId: string
    year: string | number
  },
  options?: {
    enabled?: boolean
  },
) =>
  useQuery(
    ['users', opts.userId, 'contest-registrations', opts.year],
    async (): Promise<ContestRegistrationsView> => {
      const response = await fetch(
        `${root}/users/${opts.userId}/contest-registrations/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ContestRegistrationsView.parse(await response.json())
    },
    options,
  )

const ActivitySplitScore = z.object({
  activity_id: z.number(),
  activity_name: z.string(),
  score: z.number(),
})

export type ActivitySplitScore = z.infer<typeof ActivitySplitScore>

const ActivitySplit = z.object({
  activities: z.array(ActivitySplitScore),
})

export type ActivitySplit = z.infer<typeof ActivitySplit>

export const useUserYearlyActivitySplit = (
  opts: {
    userId: string
    year: string | number
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', opts.userId, 'activity-split', opts.year],
    async (): Promise<ActivitySplit> => {
      const response = await fetch(
        `${root}/users/${opts.userId}/activity-split/${opts.year}`,
      )

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return ActivitySplit.parse(await response.json())
    },
    options,
  )

export const useDeleteLog = (onSuccess: () => void, onError: () => void) =>
  useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(`${root}/logs/${id}`, {
        method: 'DELETE',
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
    },
    onSuccess() {
      onSuccess()
    },
    onError() {
      onError()
    },
  })

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
        `${root}/contests/${contestId}/moderation/detach/${logId}`,
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

// Admin

const UserListEntry = z.object({
  id: z.string(),
  display_name: z.string(),
  email: z.string(),
  created_at: z.string(),
  role: z.string().optional(),
})

export type UserListEntry = z.infer<typeof UserListEntry>

const UserList = z.object({
  users: z.array(UserListEntry),
  total_size: z.number(),
})

export type UserList = z.infer<typeof UserList>

export const useUserList = (
  opts: {
    pageSize: number
    page: number
    query?: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['users', 'list', opts],
    async (): Promise<UserList> => {
      const params: Record<string, string> = {
        page_size: opts.pageSize.toString(),
        page: opts.page.toString(),
      }
      if (opts.query) {
        params.query = opts.query
      }
      const response = await fetch(
        `${root}/users?${new URLSearchParams(params)}`,
      )

      if (response.status === 401) {
        throw new Error('401')
      }
      if (response.status === 403) {
        throw new Error('403')
      }
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return UserList.parse(await response.json())
    },
    { ...options, retry: false },
  )

export const useUpdateUserRole = (
  onSuccess: () => void,
  onError: () => void,
) =>
  useMutation({
    mutationFn: async ({
      userId,
      role,
      reason,
    }: {
      userId: string
      role: 'user' | 'banned'
      reason: string
    }) => {
      const response = await fetch(`${authzRoot}/users/${userId}/role`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role, reason }),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
    },
    onSuccess,
    onError,
  })
