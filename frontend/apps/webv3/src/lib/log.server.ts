import { AsyncLocalStorage } from 'node:async_hooks'
import { randomUUID } from 'node:crypto'
import * as ApiClient from 'api-client'
import { z } from 'zod'
import { urlsForHost } from './urls.server'

/** The name of a generated API operation: a compile-time literal, so caller data cannot become one. */
type ApiOperation = {
  [Name in keyof typeof ApiClient]: (typeof ApiClient)[Name] extends (options: never) => unknown ? Name : never
}[keyof typeof ApiClient]

export type UpstreamFailure = {
  operation: ApiOperation
  durationMs: number
  failure:
    | { kind: 'http_status'; status: number }
    | { kind: 'transport'; error: 'timeout' | 'network' }
    | { kind: 'invalid_response'; status: number; issues: readonly { path: readonly PropertyKey[] }[] }
}

export type SsrError =
  | { kind: 'response'; status: number }
  | { kind: 'thrown'; errorName: string }

type RequestLog = { requestId: string; host: string | undefined }

const requestLog = new AsyncLocalStorage<RequestLog>()
const requestIdPattern = /^[A-Za-z0-9_-]{1,128}$/
const maxIssuePaths = 20
const schemaFields = new Set(
  Object.values(ApiClient).flatMap(schema => (schema instanceof z.ZodObject ? Object.keys(schema.shape) : [])),
)
const reportedErrorNames = new Set([
  'Error',
  'RangeError',
  'ReferenceError',
  'SyntaxError',
  'TimeoutError',
  'TypeError',
  'ZodError',
])

/** The API gateway host this request reads from, which only a valid branch label in the Host header can select. */
function upstreamHost(host: string | null) {
  try {
    return new URL(urlsForHost(host).apiUrl).host
  } catch {
    return undefined
  }
}

/** Runs one request with the ID and host that every record made during it carries. */
export function withRequestLog<T>(request: Request, run: () => T): T {
  const ingressId = request.headers.get('x-request-id')
  const requestId = ingressId && requestIdPattern.test(ingressId) ? ingressId : randomUUID()
  return requestLog.run({ requestId, host: upstreamHost(request.headers.get('host')) }, run)
}

/** Keeps field names in step with Tadoku API's slog JSON handler, so one Loki query covers both. */
function write(
  event: string,
  msg: string,
  fields: { operation?: ApiOperation; status?: number; latency?: number; error_type?: string; issue_paths?: string[] },
) {
  const request = requestLog.getStore()
  const entry = {
    time: new Date().toISOString(),
    level: 'ERROR',
    msg,
    event,
    correlation_id: request?.requestId,
    host: request?.host,
    ...fields,
  }
  process.stdout.write(`${JSON.stringify(entry)}\n`)
}

/** A segment Zod took from a record key is response data, so only field names the API schemas declare are kept. */
function issuePath(path: readonly PropertyKey[]) {
  return path.map(segment => {
    if (typeof segment === 'number') return String(segment)
    return typeof segment === 'string' && schemaFields.has(segment) ? segment : '*'
  }).join('.') || '$'
}

export function recordUpstreamFailure({ operation, durationMs, failure }: UpstreamFailure) {
  write('upstream_failure', 'upstream request failed', {
    operation,
    latency: Math.round(durationMs * 1_000_000),
    status: failure.kind === 'transport' ? undefined : failure.status,
    error_type: failure.kind === 'transport' ? failure.error : failure.kind,
    issue_paths: failure.kind === 'invalid_response'
      ? failure.issues.slice(0, maxIssuePaths).map(issue => issuePath(issue.path))
      : undefined,
  })
}

export function recordSsrError(error: SsrError) {
  if (error.kind === 'response') {
    write('ssr_error', 'server rendering failed', { status: error.status })
    return
  }

  write('ssr_error', 'server rendering failed', {
    error_type: reportedErrorNames.has(error.errorName) ? error.errorName : 'other',
  })
}
