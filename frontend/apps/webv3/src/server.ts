import { createStartHandler, defaultStreamHandler, defineHandlerCallback } from '@tanstack/react-start/server'
import { createServerEntry } from '@tanstack/react-start/server-entry'
import { PageUnavailableError } from './lib/page-status'

/** Answers a page that failed with `PageUnavailableError` with 503 rather than the router's 500, keeping its body. */
const handler = defineHandlerCallback(async context => {
  const result = await defaultStreamHandler(context)
  const unavailable = context.router.state.matches
    .some(match => match.status === 'error' && match.error instanceof PageUnavailableError)
  if (!unavailable) return result

  const response = result instanceof Response ? result : result.response
  const replaced = new Response(response.body, { status: 503, headers: response.headers })
  return result instanceof Response ? replaced : { ...result, response: replaced }
})

export default createServerEntry({ fetch: createStartHandler(handler) })
