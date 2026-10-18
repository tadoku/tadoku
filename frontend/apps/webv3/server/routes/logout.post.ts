import { defineHandler } from 'nitro'
import { z } from 'zod'
import { assertSameOrigin, sessionCookie } from '../../src/lib/session.server'
import { resolveRequestUrls } from '../../src/lib/urls.server'

export default defineHandler(async event => {
  const request = event.req
  const urls = resolveRequestUrls(
    process.env.TADOKU_WEB_URL || 'https://tadoku.dev.lab',
    process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal',
    request.headers.get('host') || undefined,
  )
  const headers = { 'Cache-Control': 'private, no-store', Vary: 'Cookie' }
  try { assertSameOrigin(request, urls.appOrigin) } catch {
    return new Response('Forbidden', { status: 403, headers })
  }
  const cookie = sessionCookie(request.headers.get('cookie') || '')
  if (!cookie) return new Response(null, { status: 303, headers: { ...headers, Location: '/' } })
  try {
    const response = await fetch(`${urls.kratosUrl}/self-service/logout/browser?return_to=${encodeURIComponent(`${urls.appOrigin}/`)}`, {
      headers: { Accept: 'application/json', Cookie: cookie }, redirect: 'error', signal: AbortSignal.timeout(10_000),
    })
    if (response.status === 401) return new Response(null, { status: 303, headers: { ...headers, Location: '/' } })
    if (!response.ok) throw new Error('Logout unavailable')
    const { logout_url } = z.object({ logout_url: z.url() }).parse(await response.json())
    const target = new URL(logout_url)
    if (target.origin !== new URL(urls.kratosUrl).origin || target.pathname !== '/kratos/self-service/logout') {
      throw new Error('Invalid logout URL')
    }
    return new Response(null, { status: 303, headers: { ...headers, Location: target.href } })
  } catch {
    return new Response('Log out is unavailable. Return to the previous page and try again.', { status: 503, headers })
  }
})
