import { AsyncLocalStorage } from 'node:async_hooks'
import { randomUUID } from 'node:crypto'
import * as ApiClient from 'api-client'
import { z } from 'zod'
import type { ClientErrorKind } from './client-errors'
import { urlsForHost } from './urls.server'

/** The name of a generated API operation: a compile-time literal, so caller data cannot become one. */
export type ApiOperation = {
  [Name in keyof typeof ApiClient]: (typeof ApiClient)[Name] extends (options: never) => unknown ? Name : never
}[keyof typeof ApiClient]

type UpstreamFailure = {
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

type RequestLog = { requestId: string; host: string | undefined; ssrErrorRecorded: boolean }

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

/**
 * TanStack Router reports an error React recovered from while rendering, inside a Suspense boundary, only as
 * `console.error('Error in renderTo…:', error)`, and the response keeps its status. Records it as an SSR error.
 */
function recordRenderErrors() {
  const frameworkError = console.error
  console.error = (...args: unknown[]) => {
    if (requestLog.getStore() && typeof args[0] === 'string' && /^Error in renderTo[A-Za-z]+:$/.test(args[0])) {
      recordSsrError({ kind: 'thrown', errorName: args[1] instanceof Error ? args[1].name : 'unknown' })
    }
    frameworkError(...args)
  }
}

let renderErrorsRecorded = false

/** Runs one request with the ID and host that every record made during it carries. */
export function withRequestLog<T>(request: Request, run: () => T): T {
  if (!renderErrorsRecorded) {
    renderErrorsRecorded = true
    recordRenderErrors()
  }

  const ingressId = request.headers.get('x-request-id')
  const requestId = ingressId && requestIdPattern.test(ingressId) ? ingressId : randomUUID()
  return requestLog.run({ requestId, host: upstreamHost(request.headers.get('host')), ssrErrorRecorded: false }, run)
}

/** Keeps field names in step with Tadoku API's slog JSON handler, so one Loki query covers both. */
function write(
  event: string,
  msg: string,
  fields: {
    operation?: ApiOperation
    status?: number
    latency?: number
    error_type?: string
    issue_paths?: string[]
    error?: string
    stack?: string
    path?: string
    build_id?: string
    user_agent?: string
  },
  level: 'ERROR' | 'INFO' = 'ERROR',
) {
  const request = requestLog.getStore()
  const entry = {
    time: new Date().toISOString(),
    level,
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

function recordUpstreamFailure({ operation, durationMs, failure }: UpstreamFailure) {
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

/** The statuses adapters turn into an answer the page explains, so they are refusals rather than upstream failures. */
const answeredStatuses = new Set([400, 401, 403, 404, 409])

type UpstreamCall = { operation: ApiOperation; startedAt: number }

/** Request options that name the operation a failed call is recorded under. */
/**
 * The header that sends this request's ID to Tadoku API, which logs it as `correlation_id`, so both services' lines
 * join. The ID is the ingress one only when it passed `requestIdPattern`; outside a request there is none.
 */
export function correlationHeader(): { 'X-Request-Id'?: string } {
  const request = requestLog.getStore()
  return request ? { 'X-Request-Id': request.requestId } : {}
}

export const upstreamCall = (operation: ApiOperation): { meta: UpstreamCall } => ({
  meta: { operation, startedAt: performance.now() },
})

function upstreamFailure(error: unknown, response: Response | undefined): UpstreamFailure['failure'] | undefined {
  if (error instanceof Error && error.name === 'TimeoutError') return { kind: 'transport', error: 'timeout' }
  if (!response || error instanceof TypeError) return { kind: 'transport', error: 'network' }
  if (!response.ok) return answeredStatuses.has(response.status) ? undefined : { kind: 'http_status', status: response.status }
  return { kind: 'invalid_response', status: response.status, issues: error instanceof z.ZodError ? error.issues : [] }
}

// The generated client runs this once per failed call, for every operation. Development reloads this module, so the
// previous registration goes first or each failure would be recorded twice.
ApiClient.client.interceptors.error.clear()
ApiClient.client.interceptors.error.use((error, response, _request, options) => {
  const call = (options as { meta?: UpstreamCall }).meta
  const failure = call && upstreamFailure(error, response)
  if (failure) recordUpstreamFailure({ operation: call.operation, durationMs: performance.now() - call.startedAt, failure })
  return error
})

/** Records at most one SSR error per request: a render error and the 500 it causes are one failure. */
export function recordSsrError(error: SsrError) {
  const request = requestLog.getStore()
  if (request?.ssrErrorRecorded) return
  if (request) request.ssrErrorRecorded = true

  if (error.kind === 'response') {
    write('ssr_error', 'server rendering failed', { status: error.status })
    return
  }

  write('ssr_error', 'server rendering failed', {
    error_type: reportedErrorNames.has(error.errorName) ? error.errorName : 'other',
  })
}

type ClientError = {
  kind: ClientErrorKind
  message: string
  stack?: string
  path: string
  build: string
  userAgent: string | undefined
}

/**
 * A query or fragment, of an absolute or relative URL, runs to the end of its word, dropping any URL nested in it. A
 * stack frame's `:line:column` and the `)` that closes it stay. Characters a URL percent-encodes end a word.
 */
const urlQuery = /(?:(?<=\/[^\s?#'"<>]*)[?#]|[?#](?=[\w.%-]+=))[^\s'"<>]*?(?=(?::\d+){0,2}\)?(?:[\s'"<>]|$))/g
const emailAddress = /(?<![\w.+-])[\w.+-]+@[\w-]+(?:\.[\w-]+)+/g

/** Browser text can quote a URL that carries a token, or an email address, so both are redacted from it. */
const redact = (text: string) => text.replace(urlQuery, '').replace(emailAddress, '[email]')

/** Records an error a browser reported, from a page anyone can load, so its free text is redacted. */
export function recordClientError({ kind, message, stack, path, build, userAgent }: ClientError) {
  write('client_error', 'browser error', {
    error_type: kind,
    error: redact(message),
    stack: stack && redact(stack),
    path: redact(path.split(/[?#]/)[0]),
    build_id: build,
    user_agent: userAgent?.slice(0, 256),
  })
}

/** The one INFO line: the server passed its startup checks and is about to listen, which marks every rollout in the logs. */
export function recordStartup() {
  write('startup', 'server started', {}, 'INFO')
}
