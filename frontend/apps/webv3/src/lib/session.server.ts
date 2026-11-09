import { getRequest, getRequestHeader } from '@tanstack/react-start/server'
import { authzRoleGet, identitySessionGet } from 'api-client'
import { recordUpstreamFailure, type ApiOperation } from './log.server'
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

/** How long one GET request may wait on upstream calls; a slower page renders with those sections marked as failed. */
const requestDeadlineMs = 8_000

const deadlines = new WeakMap<Request, AbortSignal>()

function requestDeadline() {
  const request = getRequest()
  if (!['GET', 'HEAD'].includes(request.method)) return undefined

  let deadline = deadlines.get(request)
  if (!deadline) {
    deadline = AbortSignal.timeout(requestDeadlineMs)
    deadlines.set(request, deadline)
  }
  return deadline
}

/** Reads the whole response, so a deadline during the body is recorded too, once for each upstream call it aborted. */
function reportingDeadline(operation: ApiOperation, deadline: AbortSignal): typeof fetch {
  const started = performance.now()
  return async (input, init) => {
    try {
      const response = await fetch(input, init)
      const body = await response.arrayBuffer()
      return new Response(body.byteLength ? body : null, response)
    } catch (error) {
      if (deadline.aborted) {
        recordUpstreamFailure({
          operation,
          durationMs: performance.now() - started,
          failure: { kind: 'transport', error: 'timeout' },
        })
      }
      throw error
    }
  }
}

/**
 * Options for a generated API operation that the server sends for this request, with the visitor's session if given.
 * Pass `apiUrl` outside a Start request, where the request's deadline and URLs are unavailable.
 */
export function sessionRequest(operation: ApiOperation, cookie?: string, apiUrl?: string) {
  const deadline = apiUrl ? undefined : requestDeadline()
  const timeout = AbortSignal.timeout(10_000)
  return {
    baseUrl: apiUrl ?? requestUrls().apiUrl,
    headers: { Accept: 'application/json', ...(cookie ? { Cookie: cookie } : {}) },
    cache: 'no-store' as const,
    redirect: 'error' as const,
    signal: deadline ? AbortSignal.any([timeout, deadline]) : timeout,
    ...(deadline ? { fetch: reportingDeadline(operation, deadline) } : {}),
  }
}

/** Reads with the visitor's session, or signed out when the API rejects that session. */
export async function readAsVisitor<R extends { response?: Response }>(read: (cookie?: string) => Promise<R>): Promise<R> {
  const cookie = sessionCookie()
  const answer = await read(cookie)
  return cookie && answer.response?.status === 401 ? read() : answer
}

export type SessionAccount =
  | Extract<Viewer, { status: 'anonymous' | 'unavailable' }>
  | { status: 'account'; userId: string; displayName: string }

/** The session's account, without the role read that admin-only access needs. */
export async function readAccount(cookie: string): Promise<SessionAccount> {
  const session = await identitySessionGet(sessionRequest('identitySessionGet', cookie))
  if (!session.data) return { status: 'unavailable' }
  const account = session.data.account
  if (!account) return { status: 'anonymous' }
  return { status: 'account', userId: account.id, displayName: account.display_name }
}

export async function readViewer(): Promise<Viewer> {
  const cookie = sessionCookie()
  if (!cookie) return { status: 'anonymous' }

  const account = await readAccount(cookie)
  if (account.status !== 'account') return account

  const role = (await authzRoleGet(sessionRequest('authzRoleGet', cookie))).data?.role
  if (!role) return { status: 'unavailable' }
  if (role === 'guest') return { status: 'anonymous' }
  const user = { userId: account.userId, displayName: account.displayName }
  return role === 'banned' ? { status: 'banned', ...user } : { status: 'signed-in', ...user, role }
}

export function assertSameOrigin(request: Request, appOrigin = requestUrls().appOrigin) {
  const site = request.headers.get('sec-fetch-site')
  const origin = request.headers.get('origin')
  if ((site && site !== 'same-origin') || !origin || origin !== appOrigin) {
    throw new Response('Forbidden', { status: 403, headers: { 'Cache-Control': 'private, no-store', Vary: 'Cookie' } })
  }
}
