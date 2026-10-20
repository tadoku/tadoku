export type Operation = 'login' | 'register' | 'recover' | 'display-name' | 'email' | 'password' | 'logout'

/** Identity API operations behind each account form. Fields are the only request properties a form may send. */
export const operations = {
  login: { path: 'login', method: 'POST', fields: ['email', 'password'], page: '/login' },
  register: { path: 'registration', method: 'POST', fields: ['display_name', 'email', 'password'], page: '/register' },
  recover: { path: 'recovery', method: 'POST', fields: ['email'], page: '/recovery' },
  'display-name': { path: 'settings/profile', method: 'PATCH', fields: ['display_name'], page: '/' },
  email: { path: 'settings/profile', method: 'PATCH', fields: ['email'], page: '/' },
  password: { path: 'settings/password', method: 'PUT', fields: ['password'], page: '/' },
  logout: { path: 'logout', method: 'POST', fields: [], page: '/' },
} as const satisfies Record<Operation, { path: string; method: string; fields: readonly string[]; page: string }>

export function isOperation(name: string): name is Operation {
  return Object.hasOwn(operations, name)
}

const codes = [
  'wrong_password', 'throttled', 'email_taken', 'stale_session', 'invalid_display_name', 'invalid_email',
  'invalid_password', 'invalid_origin', 'invalid_return_to', 'not_implemented', 'unauthorized', 'unavailable',
] as const

/** Identity API error codes a form can receive, plus `unauthorized` for a 401 without one. */
export type Code = (typeof codes)[number]

export function isCode(value: unknown): value is Code {
  return codes.includes(value as Code)
}

export type Result = { ok: true } | { ok: false; code: Code }

export async function readResult(response: Response): Promise<Result> {
  if (response.ok) return { ok: true }
  const body: unknown = await response.json().catch(() => undefined)
  const code = typeof body === 'object' && body !== null && 'code' in body ? body.code : undefined
  if (isCode(code)) return { ok: false, code }
  return { ok: false, code: response.status === 401 ? 'unauthorized' : 'unavailable' }
}

/**
 * Sends a hydrated form from the browser to the tenant's identity API, so login throttling sees the visitor's own
 * address. `api` is the public API root on the tenant's main host.
 */
export async function submitIdentity(api: string, operation: Operation, body: Record<string, string>): Promise<Result> {
  const { path, method } = operations[operation]
  try {
    const response = await fetch(`${api}/identity/${path}`, {
      method,
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      credentials: 'include',
      cache: 'no-store',
      redirect: 'error',
      signal: AbortSignal.timeout(15_000),
    })
    return await readResult(response)
  } catch {
    return { ok: false, code: 'unavailable' }
  }
}

export type Problem = {
  readonly code: Code
  readonly message: string
  readonly field?: 'display_name' | 'email' | 'password'
  /** Settings problems that a fresh login resolves. */
  readonly reauthenticate?: boolean
}

/** Describes an identity error code. Pass the submitted values to explain a rejected password precisely. */
export function problemFor(code: Code, values?: Record<string, string>): Problem {
  return { code, ...describe(code, values) }
}

function describe(code: Code, values?: Record<string, string>): Omit<Problem, 'code'> {
  switch (code) {
    case 'wrong_password':
      return { message: 'That email and password don’t match. Try again, or reset your password.' }
    case 'throttled':
      return { message: 'Too many login attempts. Wait 15 minutes, then try again.' }
    case 'email_taken':
      return { field: 'email', message: 'This email is already registered. Log in or reset your password instead.' }
    case 'stale_session':
      return { reauthenticate: true, message: 'For your security, log in again before changing your email or password.' }
    case 'unauthorized':
      return { reauthenticate: true, message: 'Your session has ended. Log in again to continue.' }
    case 'invalid_display_name':
      return { field: 'display_name', message: displayNameMessage }
    case 'invalid_email':
      return { field: 'email', message: 'Enter a valid email address.' }
    case 'invalid_password':
      return {
        field: 'password',
        message: values?.password === undefined
          ? 'Use 8 or more characters, at most 72 bytes, that aren’t too similar to your email.'
          : passwordProblem(values.password) ?? 'This password is too similar to your email. Choose something else.',
      }
    case 'invalid_origin':
      return { message: 'This request didn’t come from a Tadoku page. Reload the page and try again.' }
    case 'invalid_return_to':
      return { message: 'The link you followed has a return address Tadoku doesn’t allow.' }
    case 'not_implemented':
      return { message: 'This account still uses the previous account pages. Reload the page to continue.' }
    case 'unavailable':
      return { message: 'Something went wrong on our side. Try again in a moment.' }
  }
}

export function emailProblem(email: string) {
  if (!email.trim()) return 'Enter your email address.'
  if (!/^[^\s@]+@[^\s@]+$/u.test(email.trim())) return 'Enter a valid email address.'
}

const displayNameMessage = 'Use 2 to 32 characters, including one that isn’t a space.'

export function displayNameProblem(name: string) {
  const length = [...name].length
  if (length < 2 || length > 32 || !name.trim()) return displayNameMessage
}

export function passwordProblem(password: string) {
  if ([...password].length < 8) return 'Use at least 8 characters.'
  if (new TextEncoder().encode(password).length > 72) {
    return 'Use at most 72 bytes. Shorten it, or use fewer accented or non-Latin characters.'
  }
}

const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
const allowedHost = new RegExp([
  '^(?:(?:account|admin|authv2)\\.)?tadoku\\.(?:app|dev\\.lab)$',
  `^${label}(?:\\.(?:account|admin|authv2|next))?\\.tadoku\\.dev\\.lab$`,
  `^${label}(?:\\.(?:account|admin|authv2))?\\.preview\\.tadoku\\.app$`,
].join('|'))

/** Bytes Go's url.QueryEscape produces, which the identity API uses to build the Settings URL. */
function queryEscapedLength(value: string) {
  let length = 0
  for (const byte of new TextEncoder().encode(value)) {
    const unreserved = (byte >= 0x30 && byte <= 0x39) || (byte >= 0x41 && byte <= 0x5a) || (byte >= 0x61 && byte <= 0x7a)
      || byte === 0x2d || byte === 0x2e || byte === 0x5f || byte === 0x7e || byte === 0x20
    length += unreserved ? 1 : 3
  }
  return length
}

/**
 * Parses an untrusted return_to into a destination the identity API would also accept: HTTPS on an allowed Tadoku
 * host, with the Settings URL that carries it within 2,048 bytes. Anything else has no continuation.
 */
export function continuation(raw: unknown, origin: string): string | undefined {
  if (typeof raw !== 'string' || /[\s\\]/u.test(raw)) return
  const authority = /^https:\/\/([^/?#]*)/iu.exec(raw)?.[1]?.toLowerCase()
  if (!authority || !allowedHost.test(authority)) return
  if (new TextEncoder().encode(`${origin}/?return_to=`).length + queryEscapedLength(raw) > 2048) return
  try {
    if (new URL(raw).hostname === authority) return raw
  } catch {
    return
  }
}

export function withReturnTo(path: string, destination: string | undefined) {
  if (!destination) return path
  return `${path}${path.includes('?') ? '&' : '?'}return_to=${encodeURIComponent(destination)}`
}
