import { getRequestHeader } from '@tanstack/react-start/server'
import { z } from 'zod'
import type { Viewer } from './session'
import { requestUrls } from './urls.server'

export function sessionCookie(cookie = getRequestHeader('cookie')) {
  return cookie?.split(';').map(value => value.trim()).find(value => /^ory_kratos_session=[^;]+$/.test(value))
}

export async function readViewer(): Promise<Viewer> {
  const cookie = sessionCookie()
  if (!cookie) return { status: 'anonymous' }
  const { kratosUrl, apiUrl } = requestUrls()
  try {
    const options = {
      headers: { Accept: 'application/json', Cookie: cookie },
      redirect: 'error' as const, signal: AbortSignal.timeout(10_000),
    }
    const [session, role] = await Promise.all([
      fetch(`${kratosUrl}/sessions/whoami`, options),
      fetch(`${apiUrl}/authz/current-user/role`, options),
    ])
    if (session.status === 401) return { status: 'anonymous' }
    if (!session.ok || !role.ok) return { status: 'unavailable' }
    const parsedRole = z.object({ role: z.enum(['guest', 'user', 'admin', 'banned']) }).parse(await role.json()).role
    if (parsedRole === 'guest') return { status: 'anonymous' }
    const { identity } = z.object({ identity: z.object({
      id: z.uuid(), traits: z.object({ display_name: z.string() }),
    }) }).parse(await session.json())
    const user = { userId: identity.id, displayName: identity.traits.display_name }
    return parsedRole === 'banned' ? { status: 'banned', ...user } : { status: 'signed-in', ...user, role: parsedRole }
  } catch {
    return { status: 'unavailable' }
  }
}

export function assertSameOrigin(request: Request, appOrigin = requestUrls().appOrigin) {
  const site = request.headers.get('sec-fetch-site')
  const origin = request.headers.get('origin')
  if ((site && site !== 'same-origin') || !origin || origin !== appOrigin) {
    throw new Response('Forbidden', { status: 403, headers: { 'Cache-Control': 'private, no-store', Vary: 'Cookie' } })
  }
}
