import { z } from 'zod'
import { browserAppUrls, serverApiEndpointForHost } from './app-urls'

const accountSchema = z.object({
  id: z.string().uuid(),
  email: z.string(),
  display_name: z.string(),
  created_at: z.string(),
  authenticated_at: z.string(),
})

const sessionSchema = z.object({
  account_app: z.enum(['v1', 'v2']),
  account: accountSchema.nullable(),
})

export type Account = z.infer<typeof accountSchema>
export type AccountApp = z.infer<typeof sessionSchema>['account_app']

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
  const response = await fetch(`${endpoint}/identity/session`, {
    headers: forwarded,
    credentials: 'include',
    cache: 'no-store',
    redirect: 'error',
    signal: AbortSignal.timeout(10_000),
  })
  if (!response.ok) throw new Error(response.status.toString())
  return sessionSchema.parse(await response.json())
}

export async function logout() {
  const response = await fetch(`${browserAppUrls().apiEndpoint}/identity/logout`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: '{}',
  })
  if (!response.ok) throw new Error(response.status.toString())
}
