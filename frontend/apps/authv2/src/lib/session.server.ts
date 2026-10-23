import { getRequestHeader, setResponseHeader } from '@tanstack/react-start/server'
import { identitySessionGet } from 'api-client'
import { zSession } from './identity'
import { requestUrls } from './urls.server'

export async function readSession() {
  setResponseHeader('Cache-Control', 'private, no-store')
  const { apiUrl, ...origins } = requestUrls()
  // Browsers send identity forms to the public API on the tenant's main host; apiUrl may be private to the server.
  const urls = { ...origins, api: `${origins.webUrl}/api/internal` }
  const headers: Record<string, string> = { Accept: 'application/json' }
  for (const name of ['cookie', 'x-dev-branch']) {
    const value = getRequestHeader(name)
    if (value) headers[name] = value
  }

  const { data } = await identitySessionGet({
    baseUrl: apiUrl,
    headers,
    cache: 'no-store',
    redirect: 'error',
    signal: AbortSignal.timeout(10_000),
    responseValidator: body => zSession.parseAsync(body),
  })
  return data
    ? { status: 'ready' as const, ...urls, session: { account: data.account } }
    : { status: 'unavailable' as const, ...urls }
}
