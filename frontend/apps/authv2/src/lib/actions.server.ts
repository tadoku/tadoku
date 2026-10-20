import { getRequestHeader } from '@tanstack/react-start/server'
import { z } from 'zod'
import { errors, returnDestination, type Operation } from './account'
import { requestUrls } from './urls.server'

const operations = {
  login: { path: 'login', method: 'POST', fields: ['email', 'password'], page: '/login' },
  registration: { path: 'registration', method: 'POST', fields: ['email', 'display_name', 'password'], page: '/registration' },
  recovery: { path: 'recovery', method: 'POST', fields: ['email', 'return_to'], page: '/recovery' },
  profile: { path: 'settings/profile', method: 'PATCH', fields: ['email', 'display_name'], page: '/' },
  password: { path: 'settings/password', method: 'PUT', fields: ['password'], page: '/' },
  logout: { path: 'logout', method: 'POST', fields: [], page: '/' },
} as const

async function readForm(request: Request) {
  if (request.headers.get('content-type')?.split(';')[0] !== 'application/x-www-form-urlencoded') return
  const reader = request.body?.getReader()
  if (!reader) return
  const chunks: Uint8Array[] = []
  let length = 0
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    length += value.length
    if (length > 16_384) { await reader.cancel(); return }
    chunks.push(value)
  }
  const bytes = new Uint8Array(length)
  let offset = 0
  for (const chunk of chunks) { bytes.set(chunk, offset); offset += chunk.length }
  const form = new URLSearchParams(new TextDecoder().decode(bytes))
  if ([...form.keys()].some(key => form.getAll(key).length !== 1)) return
  return form
}

export async function handleAccountAction(request: Request, operation: string) {
  const responseHeaders = new Headers({ 'Cache-Control': 'private, no-store', 'Referrer-Policy': 'no-referrer' })
  if (!Object.hasOwn(operations, operation)) return new Response('Not found', { status: 404, headers: responseHeaders })
  const op = operation as Operation
  const config = operations[op]
  const urls = requestUrls()
  const expectedOrigin = `${new URL(urls.webUrl).protocol}//${getRequestHeader('host')}`
  if (request.headers.get('origin') !== expectedOrigin || !returnDestination(expectedOrigin)) {
    return Response.json({ ok: false, code: 'invalid_origin' }, { status: 403, headers: responseHeaders })
  }
  const enhanced = request.headers.get('accept') === 'application/json'
  let continuation: string | undefined
  function finish(code?: string, redirect = '/') {
    if (enhanced) return Response.json(code ? { ok: false, code } : { ok: true, redirect }, { status: code ? 400 : 200, headers: responseHeaders })
    const target = new URL(code ? config.page : redirect, expectedOrigin)
    if (code) target.searchParams.set('error', code)
    if (continuation && (code || op === 'login' || op === 'registration' || op === 'profile')) target.searchParams.set('return_to', continuation)
    responseHeaders.set('Location', target.href)
    return new Response(null, { status: 303, headers: responseHeaders })
  }

  try {
    const form = await readForm(request)
    if (!form) return finish('invalid_origin')
    const rawReturnTo = form.get('return_to')
    continuation = returnDestination(rawReturnTo)
    if (rawReturnTo && (!continuation || new TextEncoder().encode(`${expectedOrigin}/?${new URLSearchParams({ return_to: continuation })}`).length > 2048)) {
      return finish('invalid_return_to')
    }
    const body: Record<string, string> = {}
    for (const field of config.fields) {
      const value = form.get(field)
      if (value !== null) body[field] = value
    }
    if (op === 'recovery') body.return_to = continuation || urls.webUrl
    if (op === 'profile') {
      const session = await fetch(`${urls.apiUrl}/identity/session`, {
        headers: forwardedHeaders(expectedOrigin), cache: 'no-store', redirect: 'error', signal: AbortSignal.timeout(10_000),
      })
      if (!session.ok) return finish('unavailable')
      const current = z.object({ account: z.object({ email: z.string() }).nullable() }).parse(await session.json())
      if (!current.account) return finish('unauthorized')
      if (body.email === current.account.email) delete body.email
    }
    const response = await fetch(`${urls.apiUrl}/identity/${config.path}`, {
      method: config.method,
      headers: forwardedHeaders(expectedOrigin),
      body: JSON.stringify(body),
      cache: 'no-store', redirect: 'error', signal: AbortSignal.timeout(15_000),
    })
    if (!response.ok) {
      const result = z.object({ code: z.string().optional() }).safeParse(await response.json())
      const code = result.success ? result.data.code : undefined
      return finish(code && Object.hasOwn(errors, code) ? code : response.status === 401 ? 'unauthorized' : 'unavailable')
    }
    for (const cookie of response.headers.getSetCookie()) responseHeaders.append('Set-Cookie', cookie)
    if (op === 'logout') return finish(undefined, '/login')
    if (op === 'password' && continuation) return finish(undefined, continuation)
    if (op === 'login' || op === 'registration') return finish(undefined, continuation ? `/?${new URLSearchParams({ return_to: continuation })}` : '/')
    return finish(undefined, `${config.page}?${new URLSearchParams({ saved: op, ...(continuation ? { return_to: continuation } : {}) })}`)
  } catch {
    return finish('unavailable')
  }
}

function forwardedHeaders(origin: string) {
  const headers = new Headers({ Accept: 'application/json', 'Content-Type': 'application/json', Origin: origin })
  for (const name of ['cookie', 'x-dev-branch']) {
    const value = getRequestHeader(name)
    if (value) headers.set(name, value)
  }
  return headers
}
