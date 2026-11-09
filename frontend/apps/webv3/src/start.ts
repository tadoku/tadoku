import { createMiddleware, createStart } from '@tanstack/react-start'
import { getRequest, setResponseHeader } from '@tanstack/react-start/server'

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
  requestMiddleware: [requestLogging, privateResponses], functionMiddleware: [mutations],
}))
