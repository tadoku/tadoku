import { identityLogout } from 'api-client'
import { defineHandler } from 'nitro'
import { assertSameOrigin, sessionCookie, sessionRequest } from '../../src/lib/session.server'
import { urlsForHost } from '../../src/lib/urls.server'

export default defineHandler(async event => {
  const request = event.req
  const urls = urlsForHost(request.headers.get('host'))
  const headers = new Headers({ 'Cache-Control': 'private, no-store', Vary: 'Cookie' })
  try {
    assertSameOrigin(request, urls.appOrigin)
  } catch {
    return new Response('Forbidden', { status: 403, headers })
  }

  const cookie = sessionCookie(request.headers.get('cookie') || '')
  if (cookie) {
    // The identity API accepts logout only from an app Origin on its allowlist; send the origin checked above.
    const options = sessionRequest(cookie, urls.apiUrl)
    const { error, response } = await identityLogout({
      ...options,
      headers: { ...options.headers, Origin: urls.appOrigin },
    })
    if (!response?.ok || error !== undefined) {
      return new Response('Log out is unavailable. Return to the previous page and try again.', { status: 503, headers })
    }
    for (const value of response.headers.getSetCookie()) headers.append('Set-Cookie', value)
  }

  headers.set('Location', '/')
  return new Response(null, { status: 303, headers })
})
