import { immersionFeatureFlagDecisions } from 'api-client'
import { webv3OriginForHost } from 'ui/app-urls'
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

const errorPages = ['/404', '/500', '/_error']

/**
 * Where a server-rendered page request belongs while `release-webv3` is on for its account, if anywhere else.
 * Takes the request's raw URL so that the path and query arrive unchanged. An error page stays here, because
 * a path Webv2 does not have may be a missing asset.
 */
export const webv3RedirectLocation = (
  decisions: FeatureFlagDecisions,
  pathname: string,
  url: string | undefined,
  host: string | undefined,
) => {
  if (!decisions['release-webv3'] || !url || errorPages.includes(pathname)) {
    return undefined
  }

  const origin = webv3OriginForHost(host)
  return new URL(origin).host === host?.toLowerCase() ? undefined : `${origin}${url}`
}
