import {
  immersionLogContestRegistrationUpdate,
  immersionLogCreate,
  immersionLogDeleteById,
  immersionLogFindById,
  immersionLogGetConfigurations,
  immersionLogTagSuggestions,
  immersionLogUpdate,
  type ImmersionActivity,
  type ImmersionContestRegistrationReference,
  type ImmersionLanguage,
  type ImmersionLog,
  type ImmersionLogConfigurationOptions,
  type ImmersionLogCreateData,
  type ImmersionLogs,
  type ImmersionLogUpdateData,
  type ImmersionTagSuggestion,
  type ImmersionUnit,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { browserAppUrls } from 'ui/app-urls'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

// A download link the browser navigates to, not a fetched operation.
export const logExportUrl = (format: 'csv' | 'json') =>
  `${browserAppUrls().apiEndpoint}/immersion/logs/export?format=${format}`

export type Language = ImmersionLanguage
export type ActivityInputType = NonNullable<ImmersionActivity['input_type']>
export type Activity = ImmersionActivity
export type ContestRegistrationReference = ImmersionContestRegistrationReference
export type Log = ImmersionLog
export type Logs = ImmersionLogs
export type Unit = ImmersionUnit
export type LogConfigurationOptions = ImmersionLogConfigurationOptions
export type TagSuggestion = ImmersionTagSuggestion

export const useLogConfigurationOptions = (options?: { enabled?: boolean }) =>
  useQuery(
    ['contest', 'log', 'configuration-options'],
    async (): Promise<LogConfigurationOptions> =>
      apiData(immersionLogGetConfigurations(browserApi())),
    options,
  )

export const fetchTagSuggestions = async (
  query: string,
): Promise<TagSuggestion[]> => {
  const { data } = await immersionLogTagSuggestions({
    ...browserApi(),
    query: { query },
  })
  return data?.suggestions ?? []
}

export type CreateLogV2Payload = NonNullable<ImmersionLogCreateData['body']>

export const useCreateLogV2 = (onSuccess: (log: Log) => void) =>
  useMutation({
    mutationFn: async (payload: CreateLogV2Payload) =>
      apiData(immersionLogCreate({ ...browserApi(), body: payload })),
    onSuccess(data) {
      onSuccess(data)
    },
  })

export type UpdateLogPayload = NonNullable<ImmersionLogUpdateData['body']>

export const useUpdateLog = (onSuccess: (log: Log) => void) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      logId,
      payload,
    }: {
      logId: string
      payload: UpdateLogPayload
    }) =>
      apiData(
        immersionLogUpdate({
          ...browserApi(),
          path: { id: logId },
          body: payload,
        }),
      ),
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
    }) =>
      apiData(
        immersionLogContestRegistrationUpdate({
          ...browserApi(),
          path: { id: logId },
          body: { registration_ids: registrationIds },
        }),
      ),
    onSuccess(data) {
      queryClient.invalidateQueries(['log', 'findByID', data.id])
      onSuccess(data)
    },
  })
}

export const useLog = (id: string, options?: { enabled?: boolean }) =>
  useQuery(
    ['log', 'findByID', id],
    async (): Promise<Log> =>
      apiData(immersionLogFindById({ ...browserApi(), path: { id } })),
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
      await apiData(immersionLogDeleteById({ ...browserApi(), path: { id } }))
    },
    onSuccess() {
      onSuccess()
    },
    onError() {
      onError()
    },
  })
