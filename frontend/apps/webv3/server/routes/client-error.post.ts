import { defineHandler } from 'nitro'
import { z } from 'zod'
import { clientErrorKinds, clientErrorLimits } from '../../src/lib/client-errors'
import { recordClientError } from '../../src/lib/log.server'

const report = z.strictObject({
  kind: z.enum(clientErrorKinds),
  message: z.string().max(clientErrorLimits.message),
  stack: z.string().max(clientErrorLimits.stack).optional(),
  path: z.string().max(clientErrorLimits.path).startsWith('/'),
  build: z.string().max(clientErrorLimits.build),
})

// ponytail: one fixed window per process, so anyone posting reports can add at most this many lines a minute per
// replica; key it by client if real reports start being dropped.
const linesPerMinute = 120
let budget = { startedAt: 0, lines: 0 }

const answer = (status: number) => new Response(null, { status, headers: { 'Cache-Control': 'no-store' } })

/** The body as text, or undefined once it passes `limit` bytes, which stops reading it. */
async function readCapped(request: Request, limit: number) {
  const reader = request.body?.getReader()
  if (!reader) return ''
  const chunks: Uint8Array[] = []
  let size = 0
  for (let chunk = await reader.read(); !chunk.done; chunk = await reader.read()) {
    size += chunk.value.byteLength
    if (size > limit) {
      await reader.cancel()
      return undefined
    }
    chunks.push(chunk.value)
  }
  return Buffer.concat(chunks).toString('utf8')
}

/**
 * Logs one error a browser reported through `navigator.sendBeacon` as a `client_error` line. Anyone can call it, so a
 * body that is too large, not JSON or not a report answers 400, and nothing it is sent can make it fail with a 5xx.
 */
export default defineHandler(async event => {
  try {
    const body = await readCapped(event.req, clientErrorLimits.body)
    const parsed = body === undefined ? undefined : report.safeParse(JSON.parse(body))
    if (!parsed?.success) return answer(400)

    const now = Date.now()
    if (now - budget.startedAt > 60_000) budget = { startedAt: now, lines: 0 }
    if (++budget.lines > linesPerMinute) return answer(429)

    recordClientError({ ...parsed.data, userAgent: event.req.headers.get('user-agent') ?? undefined })
    return answer(204)
  } catch {
    return answer(400)
  }
})
