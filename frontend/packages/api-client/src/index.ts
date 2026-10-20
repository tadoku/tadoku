import { client } from './generated/client.gen'

// Every Tadoku API answer is JSON or empty. Anything else, such as a proxy's HTML page, is a malformed answer.
client.setConfig({ parseAs: 'json' })

// The generated client returns empty success bodies without validating them. Validate them as absent, so an empty
// answer fails unless the operation documents a success without a body.
client.interceptors.response.use(async (response, _request, options) => {
  if (response.ok && (response.status === 204 || response.headers.get('Content-Length') === '0')) {
    await options.responseValidator?.(undefined)
  }
  return response
})

export * from './generated'
export * from './generated/zod.gen'
