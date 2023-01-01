import { z } from 'zod'
import getConfig from 'next/config'
import { useMutation, useQuery, UseQueryOptions } from 'react-query'
import { ContestFormSchema } from '@app/contests/ContestForm'
import { DateTime } from 'luxon'

const { publicRuntimeConfig } = getConfig()

const root = `${publicRuntimeConfig.apiEndpoint}/immersion/contests`

const ContestConfigurationOptions = z
  .object({
    languages: z.array(
      z.object({
        code: z.string(),
        name: z.string(),
      }),
    ),
    activities: z.array(
      z.object({
        id: z.number(),
        name: z.string(),
        default: z.boolean(),
      }),
    ),
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
    registration_start: z.string(),
    registration_end: z.string(),
    description: z.string(),
    private: z.boolean(),
    official: z.boolean(),
    language_code_allow_list: z.array(z.string()),
    activity_type_id_allow_list: z.array(z.number()),
    deleted: z.boolean(),
  })
  .transform(contest => {
    const {
      contest_start: contestStart,
      contest_end: contestEnd,
      registration_start: registrationStart,
      registration_end: registrationEnd,
      language_code_allow_list: languageCodeAllowList,
      activity_type_id_allow_list: activityTypeIdAllowList,
      ...rest
    } = contest
    return {
      ...rest,
      contestStart: DateTime.fromISO(contestStart),
      contestEnd: DateTime.fromISO(contestEnd),
      registrationStart: DateTime.fromISO(registrationStart),
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
