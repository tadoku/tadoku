import { identityLogout, identitySessionGet, type IdentityAccount, type IdentitySession } from 'api-client'
import { browserAppUrls, serverApiEndpointForHost } from './app-urls'

export type Account = IdentityAccount
export type AccountApp = IdentitySession['account_app']

/** The request's identity session, or undefined when the identity API gives no valid answer. */
export async function readIdentitySession(
  host?: string,
  headers?: { cookie?: string; 'x-dev-branch'?: string | string[] },
) {
  const forwarded: Record<string, string> = {}
  if (headers?.cookie) forwarded.cookie = headers.cookie
  const branch = headers?.['x-dev-branch']
  if (typeof branch === 'string') forwarded['x-dev-branch'] = branch
  const endpoint = typeof window === 'undefined'
    ? serverApiEndpointForHost(host)
    : browserAppUrls().apiEndpoint
  const { data } = await identitySessionGet({
    baseUrl: endpoint,
    headers: forwarded,
    credentials: 'include',
    cache: 'no-store',
    redirect: 'error',
    signal: AbortSignal.timeout(3_000),
  })
  return data
}

export async function logout() {
  const { error, response } = await identityLogout({
    baseUrl: browserAppUrls().apiEndpoint,
    credentials: 'include',
    signal: AbortSignal.timeout(10_000),
  })
  if (error !== undefined) throw new Error(response ? response.status.toString() : 'unavailable')
}
