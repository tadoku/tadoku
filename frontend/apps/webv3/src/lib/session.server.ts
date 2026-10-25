import { getRequestHeader } from '@tanstack/react-start/server'
import { authzRoleGet, identitySessionGet } from 'api-client'
import type { Viewer } from './session'
import { requestUrls } from './urls.server'

/**
 * Every Tadoku session cookie in the request, empty ones included: the API selects its tenant's cookie name from
 * them and rejects that name when it occurs more than once.
 */
export function sessionCookie(cookie = getRequestHeader('cookie')) {
  const sessions = cookie?.split(';').map(value => value.trim())
    .filter(value => /^tadoku_session(?:_[a-z0-9_-]+)?=/.test(value))
  return sessions?.length ? sessions.join('; ') : undefined
}

/** Options for a generated API operation that the server sends for this request, with the visitor's session if given. */
export function sessionRequest(cookie?: string, apiUrl = requestUrls().apiUrl) {
  return {
    baseUrl: apiUrl,
    headers: { Accept: 'application/json', ...(cookie ? { Cookie: cookie } : {}) },
    cache: 'no-store' as const,
    redirect: 'error' as const,
    signal: AbortSignal.timeout(10_000),
  }
}

export async function readViewer(): Promise<Viewer> {
  const cookie = sessionCookie()
  if (!cookie) return { status: 'anonymous' }

  const session = await identitySessionGet(sessionRequest(cookie))
  if (!session.data) return { status: 'unavailable' }
  const account = session.data.account
  if (!account) return { status: 'anonymous' }

  const role = (await authzRoleGet(sessionRequest(cookie))).data?.role
  if (!role) return { status: 'unavailable' }
  if (role === 'guest') return { status: 'anonymous' }
  const user = { userId: account.id, displayName: account.display_name }
  return role === 'banned' ? { status: 'banned', ...user } : { status: 'signed-in', ...user, role }
}

export function assertSameOrigin(request: Request, appOrigin = requestUrls().appOrigin) {
  const site = request.headers.get('sec-fetch-site')
  const origin = request.headers.get('origin')
  if ((site && site !== 'same-origin') || !origin || origin !== appOrigin) {
    throw new Response('Forbidden', { status: 403, headers: { 'Cache-Control': 'private, no-store', Vary: 'Cookie' } })
  }
}
