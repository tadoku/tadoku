import { immersionFeatureFlagDecisions } from 'api-client'
import type { Account } from 'ui/identity'
import getConfig from 'next/config'
import {
  defaultFeatureFlagDecisions,
  featureFlagSubjectSchema,
  registeredDecisions,
} from './registry'
import type { FeatureFlagDecisions } from './registry'

const { publicRuntimeConfig, serverRuntimeConfig } = getConfig()
const defaultApiEndpoint =
  serverRuntimeConfig?.apiEndpoint ?? publicRuntimeConfig.apiEndpoint

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

  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), timeoutMilliseconds)
  try {
    const { data } = await immersionFeatureFlagDecisions({
      baseUrl: apiEndpoint ?? defaultApiEndpoint,
      fetch: request,
      cache: 'no-store',
      headers: {
        cookie,
        ...(typeof branch === 'string' ? { 'x-dev-branch': branch } : {}),
      },
      signal: controller.signal,
    })
    return registeredDecisions(data?.decisions)
  } finally {
    clearTimeout(timeout)
  }
}
