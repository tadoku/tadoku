import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery, UseQueryOptions } from 'react-query'
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

const ContestConfigurationOptions = z
  .object({
    languages: z.array(Language),
    activities: z.array(Activity),
    can_create_official_round: z.boolean(),
  })
  .transform(data => {
    const { can_create_official_round: canCreateOfficialRound, ...rest } = data

    return {
      ...rest,
      canCreateOfficialRound,
    }
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
      const res = await fetch(root, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(contest),
      })
      return await res.json()
    },
    onSuccess(data) {
      onSuccess(data.id)
    },
  })

const Contest = z
  .object({
    id: z.string(),
    contest_start: z.string(),
    contest_end: z.string(),
    registration_end: z.string(),
    description: z.string(),
    private: z.boolean(),
    official: z.boolean(),
    language_code_allow_list: z.array(z.string()).nullable(),
    activity_type_id_allow_list: z.array(z.number()),
    deleted: z.boolean(),
  })
  .transform(contest => {
    const {
      contest_start: contestStart,
      contest_end: contestEnd,
      registration_end: registrationEnd,
      language_code_allow_list: languageCodeAllowList,
      activity_type_id_allow_list: activityTypeIdAllowList,
      ...rest
    } = contest
    return {
      ...rest,
      contestStart: DateTime.fromISO(contestStart),
      contestEnd: DateTime.fromISO(contestEnd),
      registrationEnd: DateTime.fromISO(registrationEnd),
      languageCodeAllowList,
      activityTypeIdAllowList,
    }
  })

export type Contest = z.infer<typeof Contest>

const Contests = z
  .object({
    contests: z.array(Contest),
    total_size: z.number(),
    next_page_token: z.string(),
  })
  .transform(post => {
    const {
      next_page_token: nextPageToken,
      total_size: totalSize,
      ...rest
    } = post
    return {
      ...rest,
      nextPageToken,
      totalSize,
    }
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

const ContestView = z
  .object({
    id: z.string().nullable(),
    contest_start: z.string(),
    contest_end: z.string(),
    registration_end: z.string(),
    description: z.string(),
    private: z.boolean(),
    official: z.boolean(),
    owner_user_id: z.string().nullable().optional(),
    owner_user_display_name: z.string().nullable().optional(),
    allowed_languages: z.array(Language),
    allowed_activities: z.array(Activity),
    deleted: z.boolean().nullable().optional(),
  })
  .transform(contest => {
    const {
      contest_start: contestStart,
      contest_end: contestEnd,
      registration_end: registrationEnd,
      allowed_languages: allowedLanguages,
      allowed_activities: allowedActivities,
      owner_user_id: ownerUserId,
      owner_user_display_name: ownerUserDisplayName,
      ...rest
    } = contest
    return {
      ...rest,
      contestStart: DateTime.fromISO(contestStart),
      contestEnd: DateTime.fromISO(contestEnd),
      registrationEnd: DateTime.fromISO(registrationEnd),
      allowedLanguages,
      allowedActivities,
      ownerUserId,
      ownerUserDisplayName,
    }
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

export const ContestRegistrationView = z
  .object({
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
    contest: ContestView.nullable(),
  })
  .transform(reg => {
    const {
      contest_id: contestId,
      user_id: userId,
      user_display_name: userDisplayName,
      ...rest
    } = reg
    return {
      ...rest,
      contestId,
      userId,
      userDisplayName,
    }
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
        `${root}/${registration.contestId}/registration`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            language_codes: registration.newLanguages.map(it => it.code),
          }),
        },
      )
      return
    },
    onSuccess() {
      onSuccess()
    },
  })

const LeaderboardEntry = z
  .object({
    rank: z.number(),
    user_id: z.string(),
    user_display_name: z.string(),
    score: z.number(),
    is_tie: z.boolean(),
  })
  .transform(entry => {
    const {
      user_id: userId,
      user_display_name: userDisplayName,
      is_tie: isTie,
      ...rest
    } = entry
    return {
      ...rest,
      userId,
      userDisplayName,
      isTie,
    }
  })

export type LeaderboardEntry = z.infer<typeof LeaderboardEntry>

const Leaderboard = z
  .object({
    entries: z.array(LeaderboardEntry),
    total_size: z.number(),
    next_page_token: z.string(),
  })
  .transform(post => {
    const {
      next_page_token: nextPageToken,
      total_size: totalSize,
      ...rest
    } = post
    return {
      ...rest,
      nextPageToken,
      totalSize,
    }
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

export const Unit = z
  .object({
    id: z.string(),
    log_activity_id: z.number(),
    name: z.string(),
    modifier: z.number(),
    language_code: z.string().nullable().optional(),
  })
  .transform(it => {
    const {
      log_activity_id: logActivityId,
      language_code: languageCode,
      ...rest
    } = it
    return {
      languageCode,
      logActivityId,
      ...rest,
    }
  })

export type Unit = z.infer<typeof Unit>

export const Tag = z
  .object({
    id: z.string(),
    log_activity_id: z.number(),
    name: z.string(),
  })
  .transform(it => {
    const { log_activity_id: logActivityId, ...rest } = it
    return {
      logActivityId,
      ...rest,
    }
  })

export type Tag = z.infer<typeof Tag>

const LogConfigurationOptions = z.object({
  languages: z.array(Language),
  activities: z.array(Activity),
  units: z.array(Unit),
  tags: z.array(Tag),
})

export type LogConfigurationOptions = z.infer<typeof LogConfigurationOptions>

export const useLogConfigurationOptions = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'log', 'configuration-options'],
    async (): Promise<LogConfigurationOptions> => {
      const response = await fetch(`${root}/log/configuration-options`)

      if (response.status !== 200) {
        throw new Error('could not fetch page')
      }

      return LogConfigurationOptions.parse(await response.json())
    },
    options,
  )

const ContestRegistrationsView = z
  .object({
    registrations: z.array(ContestRegistrationView),
    next_page_token: z.string(),
    total_size: z.number(),
  })
  .transform(regs => {
    const {
      next_page_token: nextPageToken,
      total_size: totalSize,
      ...rest
    } = regs
    return {
      ...rest,
      nextPageToken,
      totalSize,
    }
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
