import { createMiddleware, createStart } from '@tanstack/react-start'
import { getRequest, getRequestHeader, setResponseHeader } from '@tanstack/react-start/server'

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

const mutations = createMiddleware({ type: 'function' }).server(async ({ next }) => {
  const request = getRequest()
  if (!['GET', 'HEAD'].includes(request.method)) {
    const { assertSameOrigin } = await import('./lib/session.server')
    assertSameOrigin(request)
  }
  return next()
})

export const startInstance = createStart(() => ({
  requestMiddleware: [requestLogging, securityHeaders, privateResponses], functionMiddleware: [mutations],
}))
