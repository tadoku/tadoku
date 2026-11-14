import { createStartHandler, defaultStreamHandler, defineHandlerCallback } from '@tanstack/react-start/server'
import { createServerEntry } from '@tanstack/react-start/server-entry'
import { getNormalizedURL } from '@tanstack/react-router/ssr/server'
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

const startHandler = createStartHandler(handler)

/**
 * Answers a server function call that names no function, such as `/_serverFn/`, with 404. TanStack Start answers one
 * with 500 before any request middleware in `start.ts` runs; this reads the ID from the path the way it does.
 */
export default createServerEntry({
  fetch(request, options) {
    const base = process.env.TSS_SERVER_FN_BASE
    const { pathname } = getNormalizedURL(request.url).url
    if (!base || !pathname.startsWith(base) || pathname.slice(base.length).split('/')[0]) {
      return startHandler(request, options)
    }
    return new Response('Not Found\n', { status: 404, headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'no-store' } })
  },
})
