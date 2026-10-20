import {
  immersionContestCreate,
  immersionContestCreatePermissionCheck,
  immersionContestFetchSummary,
  immersionContestFindById,
  immersionContestFindLatestOfficial,
  immersionContestFindOngoingRegistrations,
  immersionContestFindRegistration,
  immersionContestGetConfigurations,
  immersionContestList,
  immersionContestListLogs,
  immersionContestModerationDetachLog,
  immersionContestProfileFetchActivity,
  immersionContestProfileFetchScores,
  immersionContestRegistrationUpsert,
  type ImmersionContest,
  type ImmersionContestConfigurationOptions,
  type ImmersionContestProfileActivity,
  type ImmersionContestProfileScores,
  type ImmersionContestRegistration,
  type ImmersionContestRegistrations,
  type ImmersionContests,
  type ImmersionContestSummary,
  type ImmersionContestView,
} from 'api-client'
import { apiData, browserApi, requireFields, type WithRequired } from 'ui/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ContestFormSchema } from '@app/immersion/ContestForm'
import type { ContestRegistrationFormSchema } from '@app/immersion/ContestRegistration'
import type { Logs } from '@app/immersion/api'

export type ContestConfigurationOptions = ImmersionContestConfigurationOptions

export const useContestConfigurationOptions = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'configuration-options'],
    async (): Promise<ContestConfigurationOptions> =>
      apiData(immersionContestGetConfigurations(browserApi())),
    options,
  )

export const useCreateContest = (onSuccess: (id: string) => void) =>
  useMutation({
    mutationFn: async (contest: ContestFormSchema) =>
      apiData(
        immersionContestCreate({
          ...browserApi(),
          body: {
            ...contest,
            activity_type_id_allow_list: contest.activity_type_id_allow_list.map(
              it => it.id,
            ),
            language_code_allow_list: contest.language_code_allow_list.map(
              it => it.code,
            ),
          },
        }),
      ),
    onSuccess(data) {
      if (data.id) onSuccess(data.id)
    },
  })

export const useContestCreatePermissionCheck = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'createPermissionCheck'],
    async (): Promise<true> => {
      await apiData(immersionContestCreatePermissionCheck(browserApi()))
      return true
    },
    { ...options, retry: false },
  )

export type Contest = WithRequired<ImmersionContest, 'id'>
export type Contests = Omit<ImmersionContests, 'contests'> & { contests: Contest[] }

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
      const data = await apiData(
        immersionContestList({
          ...browserApi(),
          query: {
            page_size: opts.pageSize,
            page: opts.page - 1,
            official: opts.official,
            include_deleted: opts.includeDeleted,
            ...(opts.userId ? { user_id: opts.userId } : {}),
          },
        }),
      )
      return {
        ...data,
        contests: data.contests.map(contest => requireFields(contest, 'id')),
      }
    },
    options,
  )

export type ContestView = ImmersionContestView

export const useContest = (id: string, options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findByID', id],
    async (): Promise<ContestView> =>
      apiData(immersionContestFindById({ ...browserApi(), path: { id } })),
    options,
  )

export const useLatestOfficialContest = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'findLatestOfficial'],
    async (): Promise<ContestView> =>
      apiData(immersionContestFindLatestOfficial(browserApi())),
    options,
  )

export type ContestRegistrationView = WithRequired<
  ImmersionContestRegistration,
  'id'
>

export const useContestRegistration = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestRegistrationForUser', id],
    async (): Promise<ContestRegistrationView | null> => {
      const result = await immersionContestFindRegistration({
        ...browserApi(),
        path: { id },
      })
      // 204: the user has not registered for this contest.
      if (result.response?.status === 204) return null
      const registration = (await apiData(result)) as ImmersionContestRegistration
      return requireFields(registration, 'id')
    },
    { ...options, retry: false },
  )

export type ContestSummary = ImmersionContestSummary

export const useContestSummary = (
  id: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', 'findContestSummary', id],
    async (): Promise<ContestSummary | null> => {
      const result = await immersionContestFetchSummary({
        ...browserApi(),
        path: { id },
      })
      return result.response?.status === 404 ? null : apiData(result)
    },
    { ...options, retry: false },
  )

export const useContestRegistrationUpdate = (onSuccess: () => void) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (registration: ContestRegistrationFormSchema) => {
      await apiData(
        immersionContestRegistrationUpsert({
          ...browserApi(),
          path: { id: registration.contest_id },
          body: {
            language_codes: registration.new_languages.map(it => it.code),
          },
        }),
      )
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

export type ContestRegistrationsView = Omit<
  ImmersionContestRegistrations,
  'registrations'
> & { registrations: ContestRegistrationView[] }

/** Registrations with their IDs, which the log forms submit. */
export const withRegistrationIds = (
  data: ImmersionContestRegistrations,
): ContestRegistrationsView => ({
  ...data,
  registrations: data.registrations.map(registration =>
    requireFields(registration, 'id'),
  ),
})

export const useOngoingContestRegistrations = (options?: {
  enabled?: boolean
}) =>
  useQuery(
    ['contest', 'ongoing-contest-registrations'],
    async (): Promise<ContestRegistrationsView> =>
      withRegistrationIds(
        await apiData(immersionContestFindOngoingRegistrations(browserApi())),
      ),
    options,
  )

export type ContestProfileScores = Omit<
  ImmersionContestProfileScores,
  'registration'
> & { registration: ContestRegistrationView }

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
      const data = await apiData(
        immersionContestProfileFetchScores({
          ...browserApi(),
          path: { id: opts.contestId, user_id: opts.userId },
        }),
      )
      return { ...data, registration: requireFields(data.registration, 'id') }
    },
    options,
  )

export type ContestProfileActivity = ImmersionContestProfileActivity

export const useContestProfileActivity = (
  opts: {
    userId: string
    contestId: string
  },
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['contest', opts.contestId, 'profile', opts.userId, 'readingActivity'],
    async (): Promise<ContestProfileActivity> =>
      apiData(
        immersionContestProfileFetchActivity({
          ...browserApi(),
          path: { id: opts.contestId, user_id: opts.userId },
        }),
      ),
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
    async (): Promise<Logs> =>
      apiData(
        immersionContestListLogs({
          ...browserApi(),
          path: { id: opts.contestId },
          query: {
            page_size: opts.pageSize,
            page: opts.page - 1,
            include_deleted: opts.includeDeleted,
            ...(opts.userId ? { user_id: opts.userId } : {}),
          },
        }),
      ),
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
      await apiData(
        immersionContestModerationDetachLog({
          ...browserApi(),
          path: { id: contestId, log_id: logId },
          body: { reason },
        }),
      )
    },
    onSuccess,
    onError,
  })
