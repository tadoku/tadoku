import type { Session } from '@ory/client'
import getConfig from 'next/config'
import {
  defaultFeatureFlagDecisions,
  featureFlagResponseSchema,
  featureFlagSubjectSchema,
} from './registry'
import type { FeatureFlagDecisions } from './registry'

const { publicRuntimeConfig } = getConfig()
const featureFlagEndpoint = `${publicRuntimeConfig.apiEndpoint}/immersion/feature-flags`

export const bootstrapFeatureFlagDecisions = async (
  session: Session | undefined,
  cookie: string | undefined,
  isServer = typeof window === 'undefined',
  request: typeof fetch = fetch,
  timeoutMilliseconds = 3_000,
): Promise<FeatureFlagDecisions> => {
  const subject = session?.identity?.id
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
      const response = await request(featureFlagEndpoint, {
        cache: 'no-store',
        headers: { cookie },
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
