import type { Session } from '@ory/client'
import {
  defaultFeatureFlagDecisions,
  featureFlagResponseSchema,
  featureFlagSubjectSchema,
} from './registry'
import type { FeatureFlagDecisions } from './registry'

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
    const configuredPort = Number(process.env.PORT ?? '3000')
    const port =
      Number.isInteger(configuredPort) && configuredPort > 0
        ? configuredPort
        : 3000
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), timeoutMilliseconds)

    try {
      const response = await request(
        `http://127.0.0.1:${port}/api/feature-flags`,
        {
          cache: 'no-store',
          headers: { cookie },
          signal: controller.signal,
        },
      )
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
