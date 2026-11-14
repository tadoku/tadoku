import { createMiddleware, createStart } from '@tanstack/react-start'
import { getRequest, getRequestHeader, setResponseHeader } from '@tanstack/react-start/server'
import { outdatedBuildHeader, serverFunctionFetch } from './lib/outdated-build'

const requestLogging = createMiddleware().server(async ({ next }) => {
  const request = getRequest()
  const { recordSsrError, withRequestLog } = await import('./lib/log.server')
  return withRequestLog(request, async () => {
    try {
      const result = await next()
      if (result.response.status >= 500) recordSsrError({ kind: 'response', status: result.response.status })
      return result
    } catch (error) {
      if (!request.signal.aborted) {
        recordSsrError({ kind: 'thrown', errorName: error instanceof Error ? error.name : 'unknown' })
      }
      throw error
    }
  })
})

/** The CSP to try out before enforcing it; browsers report violations in the console, as there is no report endpoint. */
const reportOnlyPolicy = (nonce: string) => [
  "default-src 'self'",
  `script-src 'self' 'nonce-${nonce}'`,
  "style-src 'self' 'unsafe-inline'",
  "img-src 'self' data: https:",
  "font-src 'self'",
  "connect-src 'self'",
  "form-action 'self'",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "object-src 'none'",
].join('; ')

/** Gives every Start response the security headers and a new CSP nonce, which `router.tsx` passes to the page's scripts. */
const securityHeaders = createMiddleware().server(async ({ next }) => {
  const cspNonce = btoa(String.fromCharCode(...crypto.getRandomValues(new Uint8Array(16))))
  const result = await next({ context: { cspNonce } })
  const { isWebHost } = await import('./lib/urls.server')
  const headers = result.response.headers
  headers.set('Strict-Transport-Security', 'max-age=31536000')
  headers.set('X-Content-Type-Options', 'nosniff')
  headers.set('Referrer-Policy', 'strict-origin-when-cross-origin')
  headers.set('Content-Security-Policy', "frame-ancestors 'none'")
  headers.set('Content-Security-Policy-Report-Only', reportOnlyPolicy(cspNonce))
  if (!isWebHost(getRequestHeader('host'))) headers.set('X-Robots-Tag', 'noindex')
  return result
})

const privateResponses = createMiddleware().server(async ({ next }) => {
  setResponseHeader('Cache-Control', 'private, no-store')
  const result = await next()
  result.response.headers.set('Cache-Control', 'private, no-store')
  result.response.headers.append('Vary', 'Cookie')
  return result
})

/** Answers a path with a trailing slash with 308 to it without, as Webv2 did, before the router's 307 does the same. */
const trailingSlash = createMiddleware().server(({ next }) => {
  const url = new URL(getRequest().url)
  if (url.pathname === '/' || !url.pathname.endsWith('/')) return next()
  // A leading `//` would make the location another host.
  const path = url.pathname.slice(0, -1).replace(/^\/+/, '/')
  return new Response(null, { status: 308, headers: { Location: path + url.search } })
})

/** What TanStack Start throws for a server function ID this build doesn't have: built, then on the Vite server. */
const unknownServerFunction = /^(Server function info not found for |Invalid server function ID: )/

/**
 * Answers a call to a server function this build doesn't have, as from a tab loaded before a release, with 409 rather
 * than a 500, so it never counts as a server error; `serverFunctionFetch` then reloads that tab.
 */
const outdatedServerFunctions = createMiddleware().server(async ({ next, handlerType }) => {
  if (handlerType !== 'serverFn') return next()
  try {
    return await next()
  } catch (error) {
    if (!(error instanceof Error) || !unknownServerFunction.test(error.message)) throw error
    return new Response('This page is from an earlier version of Tadoku. Reload it to continue.', {
      status: 409, headers: { 'Content-Type': 'text/plain; charset=utf-8', [outdatedBuildHeader]: 'true' },
    })
  }
})

/** Server function calls whose payload TanStack Start decoded, which it does before any function middleware runs. */
const decodedCalls = new WeakSet<Request>()

const decodedPayloads = createMiddleware({ type: 'function' }).server(({ next }) => {
  decodedCalls.add(getRequest())
  return next()
})

/**
 * Answers a server function call whose payload does not decode with 400 rather than TanStack Start's 500, which would
 * count as a server error and return the decoder's message. A call that did decode keeps its status.
 */
const undecodableServerFunctions = createMiddleware().server(async ({ next, handlerType }) => {
  if (handlerType !== 'serverFn') return next()
  const result = await next()
  if (result.response.status !== 500 || decodedCalls.has(getRequest())) return result
  return new Response('Bad Request\n', { status: 400, headers: { 'Content-Type': 'text/plain; charset=utf-8' } })
})

const mutations = createMiddleware({ type: 'function' }).server(async ({ next }) => {
  const request = getRequest()
  if (!['GET', 'HEAD'].includes(request.method)) {
    const { assertSameOrigin } = await import('./lib/session.server')
    assertSameOrigin(request)
  }
  return next()
})

export const startInstance = createStart(() => ({
  requestMiddleware: [
    requestLogging,
    securityHeaders,
    privateResponses,
    trailingSlash,
    outdatedServerFunctions,
    undecodableServerFunctions,
  ],
  functionMiddleware: [decodedPayloads, mutations],
  serverFns: { fetch: serverFunctionFetch },
}))
