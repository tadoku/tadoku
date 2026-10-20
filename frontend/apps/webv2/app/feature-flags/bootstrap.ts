import type { Account } from 'ui/identity'
import getConfig from 'next/config'
import {
  defaultFeatureFlagDecisions,
  featureFlagResponseSchema,
  featureFlagSubjectSchema,
} from './registry'
import type { FeatureFlagDecisions } from './registry'

const { publicRuntimeConfig, serverRuntimeConfig } = getConfig()
const featureFlagEndpoint = `${
  serverRuntimeConfig?.apiEndpoint ?? publicRuntimeConfig.apiEndpoint
}/immersion/feature-flags`

export const bootstrapFeatureFlagDecisions = async (
  session: Account | undefined,
  cookie: string | undefined,
  isServer = typeof window === 'undefined',
  request: typeof fetch = fetch,
  timeoutMilliseconds = 3_000,
  apiEndpoint?: string,
  branch?: string | string[],
): Promise<FeatureFlagDecisions> => {
  const subject = session?.id
  if (
    !isServer ||
    !cookie ||
    !subject ||
    !featureFlagSubjectSchema.safeParse(subject).success
  ) {
    return { ...defaultFeatureFlagDecisions }
  }

  try {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), timeoutMilliseconds)

    try {
      const response = await request(apiEndpoint ? `${apiEndpoint}/immersion/feature-flags` : featureFlagEndpoint, {
        cache: 'no-store',
        headers: {
          cookie,
          ...(typeof branch === 'string' ? { 'x-dev-branch': branch } : {}),
        },
        signal: controller.signal,
      })
      if (!response.ok) {
        return { ...defaultFeatureFlagDecisions }
      }

      return featureFlagResponseSchema.parse(await response.json()).decisions
    } finally {
      clearTimeout(timeout)
    }
  } catch {
    return { ...defaultFeatureFlagDecisions }
  }
}
