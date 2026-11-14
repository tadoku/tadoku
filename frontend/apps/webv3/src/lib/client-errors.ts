import { isPageUnavailable } from './page-status'

/** What `server/routes/client-error.post.ts` accepts; it logs each report as one `client_error` line. */
export const clientErrorKinds = ['error', 'unhandledrejection', 'hydration', 'chunk', 'route'] as const
export type ClientErrorKind = (typeof clientErrorKinds)[number]

export const clientErrorLimits = { message: 500, stack: 4_000, path: 500, build: 64, body: 16_384 }

/** Reports per page load, so a render loop or a burst of failing chunks cannot flood the server. */
const maxReports = 5
const sent = new Set<string>()
let serverRecorded = false

/**
 * Sends one report of an error the browser met. Each message is sent once per page load, so a chunk error reported by
 * Vite and again by the route's error view is one report. A page whose main read failed is recorded by the server, so
 * that error is not sent, nor what React reports while hydrating that page around it.
 */
export function reportClientError(kind: ClientErrorKind, error: unknown) {
  if (isPageUnavailable(error)) {
    serverRecorded = true
    return
  }
  if (serverRecorded && kind === 'hydration') return

  const message = (error instanceof Error ? error.message : String(error)).slice(0, clientErrorLimits.message)
  if (typeof navigator === 'undefined' || sent.size >= maxReports || sent.has(message)) return
  sent.add(message)

  const stack = error instanceof Error ? error.stack?.slice(0, clientErrorLimits.stack) : undefined
  // A `VITE_` variable is inlined at build time; see the Dockerfile.
  const build = String(import.meta.env.VITE_BUILD_ID || 'dev').slice(0, clientErrorLimits.build)
  const path = location.pathname.slice(0, clientErrorLimits.path)
  navigator.sendBeacon('/client-error', JSON.stringify({ kind, message, stack, path, build }))
}

/** Reports uncaught errors, unhandled rejections and chunks that fail to load, for the page's lifetime. */
export function listenForClientErrors() {
  window.addEventListener('error', event => reportClientError('error', event.error ?? event.message))
  window.addEventListener('unhandledrejection', event => reportClientError('unhandledrejection', event.reason))
  window.addEventListener('vite:preloadError', event => reportClientError('chunk', (event as Event & { payload: unknown }).payload))
}
