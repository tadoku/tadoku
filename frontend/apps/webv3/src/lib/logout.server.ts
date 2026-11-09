import { identityLogout } from 'api-client'
import { assertSameOrigin, sessionCookie, sessionRequest } from './session.server'
import { requestUrls } from './urls.server'

/**
 * Logs the request's session out with the API and answers with the redirect home, or 403 for a cross-origin request.
 * Returns undefined when the API did not confirm the logout; the browser keeps its cookie, though the session may be gone.
 */
export async function logOut(request: Request): Promise<Response | undefined> {
  const urls = requestUrls()
  const headers = new Headers({ 'Cache-Control': 'private, no-store', Vary: 'Cookie' })
  try {
    assertSameOrigin(request, urls.appOrigin)
  } catch {
    return new Response('Forbidden', { status: 403, headers })
  }

  const cookie = sessionCookie(request.headers.get('cookie') || '')
  if (cookie) {
    // The identity API accepts logout only from an app Origin on its allowlist; send the origin checked above.
    const options = sessionRequest('identityLogout', cookie)
    const { error, response } = await identityLogout({
      ...options,
      headers: { ...options.headers, Origin: urls.appOrigin },
    })
    if (!response?.ok || error !== undefined) return undefined
    for (const value of response.headers.getSetCookie()) headers.append('Set-Cookie', value)
  }

  headers.set('Location', '/')
  return new Response(null, { status: 303, headers })
}
