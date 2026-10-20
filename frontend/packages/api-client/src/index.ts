import { client } from './generated/client.gen'
import { successResponses } from './generated/responses.gen'

// Every Tadoku API answer is JSON or empty. Anything else, such as a proxy's HTML page, is a malformed answer.
client.setConfig({ parseAs: 'json' })

// A success must use a status the operation documents, with a body exactly when that status documents one. The
// generated client accepts any 2xx and skips validation of empty bodies. Errors keep their status for the caller.
client.interceptors.response.use(async (response, _request, options) => {
  if (!response.ok) return response
  const body = successResponses[`${options.method} ${options.url}`]?.[response.status]
  if (body === undefined) throw new Error(`undocumented ${response.status} answer`)
  if (body !== ((await response.clone().text()) !== '')) {
    throw new Error(body ? `empty ${response.status} answer` : `unexpected body on ${response.status} answer`)
  }
  return response
})

export * from './generated'
export * from './generated/zod.gen'
