import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { NewLogAPISchema } from '@app/immersion/NewLogForm/domain'

const root = () => `${browserAppUrls().apiEndpoint}/immersion`

export const logExportUrl = (format: 'csv' | 'json') =>
  `${root()}/logs/export?format=${format}`

export const Language = z.object({
  code: z.string(),
  name: z.string(),
})

export type Language = z.infer<typeof Language>

export const ActivityInputType = z.enum(['amount_primary', 'time_primary'])

export type ActivityInputType = z.infer<typeof ActivityInputType>

export const Activity = z.object({
  id: z.number(),
  name: z.string(),
  default: z.boolean().nullable().optional(),
  input_type: ActivityInputType.optional(),
})

export type Activity = z.infer<typeof Activity>

const ContestRegistrationReference = z.object({
  registration_id: z.string(),
  contest_id: z.string(),
  contest_end: z.string(),
  title: z.string(),
  owner_user_display_name: z.string().optional(),
  official: z.boolean().optional(),
  score: z.number().optional(),
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
  unit_id: z.string(),
  unit_key: z.string().optional(),
  unit_name: z.string(),
  tags: z.array(z.string()),
  amount: z.number(),
  modifier: z.number(),
  score: z.number(),
  duration_seconds: z.number().optional(),
  registrations: z.array(ContestRegistrationReference).optional(),
  created_at: z.string(),
  deleted: z.boolean(),
})

export type Log = z.infer<typeof Log>

export const Logs = z.object({
  logs: z.array(Log),
  total_size: z.number(),
  next_page_token: z.string(),
})

export type Logs = z.infer<typeof Logs>

export const Unit = z.object({
  id: z.string(),
  unit_key: z.string(),
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
  user_language_codes: z.array(z.string()).default([]),
  scoring_engine_enabled: z.boolean(),
})

export type LogConfigurationOptions = z.infer<typeof LogConfigurationOptions>

export const useLogConfigurationOptions = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'log', 'configuration-options'],
    async (): Promise<LogConfigurationOptions> => {
      const response = await fetch(`${root()}/logs/configuration-options`)

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

export const fetchTagSuggestions = async (
  query: string,
): Promise<TagSuggestion[]> => {
  const response = await fetch(
    `${root()}/logs/tag-suggestions?query=${encodeURIComponent(query)}`,
  )
  if (!response.ok) return []
  const data = TagSuggestions.parse(await response.json())
  return data.suggestions
}

export const useCreateLog = (onSuccess: (id: string) => void) =>
  useMutation({
    mutationFn: async (contest: NewLogAPISchema) => {
      const response = await fetch(`${root()}/logs`, {
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
  amount?: number
  unit_id?: string
  unit_key?: string
  duration_seconds?: number
  tags: string[]
  description?: string
}

export const useCreateLogV2 = (onSuccess: (log: Log) => void) =>
  useMutation({
    mutationFn: async (payload: CreateLogV2Payload) => {
      const response = await fetch(`${root()}/logs`, {
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

export type UpdateLogPayload = {
  amount?: number
  unit_id?: string
  unit_key?: string
  duration_seconds?: number
  tags: string[]
  description?: string
}

export const useUpdateLog = (onSuccess: (log: Log) => void) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      logId,
      payload,
    }: {
      logId: string
      payload: UpdateLogPayload
    }) => {
      const response = await fetch(`${root()}/logs/${logId}`, {
        method: 'PUT',
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
      queryClient.invalidateQueries(['log', 'findByID', data.id])
      onSuccess(data)
    },
  })
}

export const useUpdateLogContestRegistrations = (
  onSuccess: (log: Log) => void,
) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      logId,
      registrationIds,
    }: {
      logId: string
      registrationIds: string[]
    }) => {
      const response = await fetch(
        `${root()}/logs/${logId}/contest-registrations`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ registration_ids: registrationIds }),
        },
      )
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
      const response = await fetch(`${root()}/logs/${id}`)

      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }

      return Log.parse(await response.json())
    },
    {
      ...options,
      retry: (failureCount, error) =>
        error instanceof Error && error.message === '404'
          ? false
          : failureCount < 3,
    },
  )

export const useDeleteLog = (onSuccess: () => void, onError: () => void) =>
  useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(`${root()}/logs/${id}`, {
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
