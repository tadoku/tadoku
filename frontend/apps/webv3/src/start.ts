import { createMiddleware, createStart } from '@tanstack/react-start'
import { getRequest, setResponseHeader } from '@tanstack/react-start/server'

const privateResponses = createMiddleware().server(async ({ next }) => {
  setResponseHeader('Cache-Control', 'private, no-store')
  setResponseHeader('Vary', 'Cookie')
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
  requestMiddleware: [privateResponses], functionMiddleware: [mutations],
}))
