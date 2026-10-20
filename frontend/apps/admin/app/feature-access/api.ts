import {
  immersionFeatureAccessGet,
  immersionFeatureAccessGrant,
  immersionFeatureAccessRevoke,
  type ImmersionFeatureAccessResponse,
  type ImmersionManagedFeatureFlagKey,
} from 'api-client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { browserApi } from 'ui/api'

export type FeatureFlagKey = ImmersionManagedFeatureFlagKey
export type FeatureAccessResult = ImmersionFeatureAccessResponse

interface FeatureAccessVariables {
  targetUserId: string
  flagKey: FeatureFlagKey
}

interface FeatureAccessMutationVariables extends FeatureAccessVariables {
  enabled: boolean
}

const queryKey = ({ targetUserId, flagKey }: FeatureAccessVariables) => [
  'feature-access',
  targetUserId,
  flagKey,
]

const resultOf = async (
  request: Promise<{ data?: FeatureAccessResult; error?: unknown; response?: Response }>,
) => {
  const { data, error, response } = await request
  if (response?.ok && error === undefined && data) return data
  if (response?.status === 401) throw new Error('401')
  if (response?.status === 403) throw new Error('403')
  throw new Error('feature access unavailable')
}

const options = ({ targetUserId, flagKey }: FeatureAccessVariables) => ({
  ...browserApi(),
  path: { flagKey, userId: targetUserId },
})

export const getFeatureAccess = async (
  variables: FeatureAccessVariables,
): Promise<FeatureAccessResult> =>
  resultOf(immersionFeatureAccessGet(options(variables)))

export const updateFeatureAccess = async ({
  enabled,
  ...variables
}: FeatureAccessMutationVariables): Promise<FeatureAccessResult> =>
  resultOf(
    enabled
      ? immersionFeatureAccessGrant(options(variables))
      : immersionFeatureAccessRevoke(options(variables)),
  )

export const useFeatureAccess = (
  variables: FeatureAccessVariables,
  options?: { enabled?: boolean },
) =>
  useQuery(queryKey(variables), () => getFeatureAccess(variables), {
    ...options,
    retry: false,
  })

export const useUpdateFeatureAccess = () => {
  const queryClient = useQueryClient()
  return useMutation(updateFeatureAccess, {
    onSuccess: (result, variables) => {
      queryClient.setQueryData(queryKey(variables), result)
      return queryClient.invalidateQueries(queryKey(variables))
    },
  })
}
