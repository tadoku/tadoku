import { ErrorStrategy, FliptClient } from '@flipt-io/flipt-client-js'
import type {
  Hook,
  IFetcher,
  NodeFliptClient,
} from '@flipt-io/flipt-client-js'
import { readFile } from 'node:fs/promises'
import { z } from 'zod'
import type { Session } from '@ory/client'
import {
  defaultFeatureFlagDecisions,
  featureFlagKeys,
  featureFlagSubjectSchema,
} from './registry'
import type { FeatureFlagDecisions } from './registry'

if (typeof window !== 'undefined') {
  throw new Error('The Flipt evaluator is server-only')
}

const fliptURL =
  process.env.FLIPT_URL ?? 'http://oathkeeper-proxy.default:4455/flipt'
const fliptEnvironment = process.env.FLIPT_ENVIRONMENT ?? 'local'
const fliptNamespace = process.env.FLIPT_NAMESPACE ?? 'default'
const oathkeeperURL =
  process.env.OATHKEEPER_URL ?? 'http://oathkeeper-proxy.default:4455'
const serviceAccountTokenPath =
  process.env.SERVICE_ACCOUNT_TOKEN_PATH ?? '/var/run/secrets/tokens/token'
const positiveInteger = (value: string | undefined, fallback: number) => {
  const result = z.coerce.number().int().positive().finite().safeParse(value)
  return result.success ? result.data : fallback
}

const updateIntervalSeconds = positiveInteger(
  process.env.FLIPT_UPDATE_INTERVAL,
  30,
)
const startupTimeoutMilliseconds = positiveInteger(
  process.env.FLIPT_STARTUP_TIMEOUT,
  3_000,
)
const retryIntervalMilliseconds = positiveInteger(
  process.env.FLIPT_RETRY_INTERVAL,
  30_000,
)

const tokenResponseSchema = z.object({ access_token: z.string().min(1) })

const withTimeout = async <T>(
  operation: (signal: AbortSignal) => Promise<T>,
) => {
  const controller = new AbortController()
  const timeout = setTimeout(
    () => controller.abort(),
    startupTimeoutMilliseconds,
  )

  try {
    return await operation(controller.signal)
  } finally {
    clearTimeout(timeout)
  }
}

const exchangeServiceAccountToken = async (signal: AbortSignal) => {
  const serviceAccountToken = (
    await readFile(serviceAccountTokenPath, 'utf8')
  ).trim()

  const response = await fetch(
    `${oathkeeperURL}/token-exchange/flipt-evaluation/frontend-webv2`,
    {
      headers: { Authorization: `Bearer ${serviceAccountToken}` },
      signal,
    },
  )

  if (!response.ok) {
    throw new Error(`token exchange returned ${response.status}`)
  }

  return tokenResponseSchema.parse(await response.json()).access_token
}

export const createFliptFetcher = (): IFetcher => async options =>
  withTimeout(async signal => {
    const token = await exchangeServiceAccountToken(signal)
    const headers: Record<string, string> = {
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
      'x-flipt-accept-server-version': '1.47.0',
      'x-flipt-environment': fliptEnvironment,
    }

    if (options?.etag) {
      headers['If-None-Match'] = options.etag
    }

    return fetch(
      `${fliptURL}/internal/v1/evaluation/snapshot/namespace/${fliptNamespace}`,
      { headers, signal },
    )
  })

const evaluationHook: Hook = {
  before: () => {},
  after: ({ flagKey, value, reason }) => {
    console.info('feature flag evaluated', { flagKey, value, reason })
  },
}

type Client = NodeFliptClient
let client: Client | undefined
let initialization: Promise<Client | undefined> | undefined
let retryTimer: ReturnType<typeof setTimeout> | undefined
let lifecycleGeneration = 0

export const closeFeatureFlagClient = () => {
  lifecycleGeneration += 1
  if (retryTimer) {
    clearTimeout(retryTimer)
    retryTimer = undefined
  }
  client?.close()
  client = undefined
}

if (process.env.NODE_ENV !== 'test') {
  process.once('exit', closeFeatureFlagClient)
  const hotModule = module as NodeModule & {
    hot?: { dispose(callback: () => void): void }
  }
  hotModule.hot?.dispose(closeFeatureFlagClient)
}

const scheduleRetry = () => {
  if (retryTimer) {
    return
  }

  retryTimer = setTimeout(() => {
    retryTimer = undefined
    void getFliptClient()
  }, retryIntervalMilliseconds)
  retryTimer.unref?.()
}

const getFliptClient = async (): Promise<Client | undefined> => {
  if (client) {
    return client
  }

  if (retryTimer) {
    return undefined
  }

  if (!initialization) {
    const generation = lifecycleGeneration
    initialization = FliptClient.init({
      environment: fliptEnvironment,
      namespace: fliptNamespace,
      url: fliptURL,
      updateInterval: updateIntervalSeconds,
      errorStrategy: ErrorStrategy.Fallback,
      fetcher: createFliptFetcher(),
      hook: evaluationHook,
    })
      .then(initializedClient => {
        const nodeClient = initializedClient as Client
        if (generation !== lifecycleGeneration) {
          nodeClient.close()
          return undefined
        }
        client = nodeClient
        return nodeClient
      })
      .catch(() => {
        console.warn('feature flag provider unavailable; using safe defaults', {
          kind: 'initialization',
        })
        scheduleRetry()
        return undefined
      })
      .finally(() => {
        initialization = undefined
      })
  }

  return initialization
}

export const decisionsForSession = async (
  session: Session | undefined,
): Promise<FeatureFlagDecisions> => {
  const subject = session?.identity?.id
  if (!subject || !featureFlagSubjectSchema.safeParse(subject).success) {
    return { ...defaultFeatureFlagDecisions }
  }

  const fliptClient = await getFliptClient()
  if (!fliptClient) {
    return { ...defaultFeatureFlagDecisions }
  }

  return featureFlagKeys.reduce<FeatureFlagDecisions>(
    (decisions, flagKey) => {
      try {
        decisions[flagKey] = fliptClient.evaluateBoolean({
          flagKey,
          entityId: subject,
          context: { authenticated: 'true' },
        }).enabled
      } catch {
        console.warn('feature flag evaluation failed; using safe default', {
          flagKey,
          kind: 'provider_error',
        })
      }

      return decisions
    },
    { ...defaultFeatureFlagDecisions },
  )
}
