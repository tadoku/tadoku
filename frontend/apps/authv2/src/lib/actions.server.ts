import { getRequestHeader } from '@tanstack/react-start/server'
import { continuation, isOperation, operations, sendIdentity, withReturnTo, type Operation } from './identity'
import { requestUrls } from './urls.server'

async function readForm(request: Request) {
  if (request.headers.get('content-type')?.split(';')[0]?.trim() !== 'application/x-www-form-urlencoded') return
  const reader = request.body?.getReader()
  if (!reader) return
  const chunks: Uint8Array[] = []
  let length = 0
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    length += value.length
    if (length > 16_384) {
      await reader.cancel()
      return
    }
    chunks.push(value)
  }
  const form = new URLSearchParams(new TextDecoder().decode(Buffer.concat(chunks)))
  if ([...form.keys()].some(key => form.getAll(key).length !== 1)) return
  return form
}

function success(operation: Operation, next: string | undefined) {
  switch (operation) {
    case 'login':
    case 'register':
      return next ?? '/'
    case 'recover':
      return withReturnTo('/recovery?sent=1', next)
    case 'display-name':
    case 'email':
      return withReturnTo(`/?saved=${operation}`, next)
    case 'password':
      return next ?? '/?saved=password'
    case 'logout':
      return '/login'
  }
}

/**
 * Completes a form submitted before hydration or without JavaScript: the same operation the enhanced form sends
 * as JSON, answered with a redirect. Requests must come from this host's own pages.
 */
export async function handleNativeAction(request: Request, name: string) {
  const headers = new Headers({ 'Cache-Control': 'private, no-store', 'Referrer-Policy': 'no-referrer' })
  if (!isOperation(name)) return new Response('Not found', { status: 404, headers })
  const urls = requestUrls()
  // The account host must itself be an allowed Tadoku app, and the form must come from one of its pages.
  if (request.headers.get('origin') !== urls.origin || !continuation(urls.origin, urls.origin)) {
    return new Response('This request did not come from a Tadoku account page.', { status: 403, headers })
  }

  const { page } = operations[name]
  const form = await readForm(request)
  const next = continuation(form?.get('return_to'), urls.origin)
  const redirect = (location: string) => {
    headers.set('Location', new URL(location, urls.origin).href)
    return new Response(null, { status: 303, headers })
  }
  const failure = (code: string) => redirect(withReturnTo(`${page}?error=${code}&form=${name}`, next))
  if (!form) return failure('invalid_origin')

  const values = Object.fromEntries(form)
  if (name === 'recover') values.return_to = next ?? urls.webUrl
  const forwarded = new Headers({ Accept: 'application/json', Origin: urls.origin })
  for (const header of ['cookie', 'x-dev-branch']) {
    const value = getRequestHeader(header)
    if (value) forwarded.set(header, value)
  }
  const result = await sendIdentity(name, { baseUrl: urls.apiUrl, headers: forwarded }, values)
  if (!result.ok) return name === 'logout' && result.code === 'unauthorized' ? redirect('/login') : failure(result.code)
  for (const cookie of result.response.headers.getSetCookie()) headers.append('Set-Cookie', cookie)
  return redirect(success(name, next))
}
