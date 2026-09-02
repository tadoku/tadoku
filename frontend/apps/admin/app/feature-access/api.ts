import { useMutation, useQuery, useQueryClient } from 'react-query'
import getConfig from 'next/config'
import {
  FeatureAccessResult,
  FeatureFlagKey,
  featureAccessResultSchema,
} from './contracts'

const { publicRuntimeConfig } = getConfig()
const immersionRoot = `${publicRuntimeConfig.apiEndpoint}/immersion`

interface FeatureAccessVariables {
  targetUserId: string
  flagKey: FeatureFlagKey
}

interface FeatureAccessMutationVariables extends FeatureAccessVariables {
  enabled: boolean
}

const featureAccessEndpoint = ({
  targetUserId,
  flagKey,
}: FeatureAccessVariables) =>
  `${immersionRoot}/admin/feature-flags/${encodeURIComponent(
    flagKey,
  )}/users/${encodeURIComponent(targetUserId)}`

const queryKey = ({ targetUserId, flagKey }: FeatureAccessVariables) => [
  'feature-access',
  targetUserId,
  flagKey,
]

const responseError = async (response: Response) => {
  if (response.status === 401) throw new Error('401')
  if (response.status === 403) throw new Error('403')
  throw new Error('feature access unavailable')
}

export const getFeatureAccess = async ({
  targetUserId,
  flagKey,
}: FeatureAccessVariables): Promise<FeatureAccessResult> => {
  const response = await fetch(
    featureAccessEndpoint({ targetUserId, flagKey }),
    {
      credentials: 'include',
    },
  )
  if (!response.ok) await responseError(response)
  return featureAccessResultSchema.parse(await response.json())
}

export const updateFeatureAccess = async ({
  targetUserId,
  flagKey,
  enabled,
}: FeatureAccessMutationVariables): Promise<FeatureAccessResult> => {
  const response = await fetch(
    featureAccessEndpoint({ targetUserId, flagKey }),
    {
      method: enabled ? 'PUT' : 'DELETE',
      credentials: 'include',
    },
  )
  if (!response.ok) await responseError(response)
  return featureAccessResultSchema.parse(await response.json())
}

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
