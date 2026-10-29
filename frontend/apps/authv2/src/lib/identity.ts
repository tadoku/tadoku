import {
  identityAccountDelete,
  identityLoginCreate,
  identityLogout,
  identityPasswordUpdate,
  identityProfileUpdate,
  identityRecoveryCreate,
  identityRegistrationCreate,
  zIdentityErrorResponse,
  type IdentityErrorResponse,
} from 'api-client'

type Values = Record<string, string>

/**
 * Where one call reaches the identity API, and the headers or credentials its side of the boundary needs. Headers are
 * a plain record: generated operations with a JSON body spread them, which would drop a `Headers` instance.
 */
export type Transport = {
  readonly baseUrl: string
  readonly headers?: Readonly<Record<string, string>>
  readonly credentials?: RequestCredentials
}

type Request = Transport & {
  readonly cache: RequestCache
  readonly redirect: RequestRedirect
  readonly signal: AbortSignal
}
type Send = (request: Request, values: Values) => Promise<{ error?: unknown; response?: Response }>

/** The generated identity operation behind each account form, sent with only that form's fields. */
export const operations = {
  login: {
    page: '/login',
    send: (request, v) => identityLoginCreate({ ...request, body: { email: v.email, password: v.password } }),
  },
  register: {
    page: '/register',
    send: (request, v) => identityRegistrationCreate({
      ...request,
      body: { display_name: v.display_name, email: v.email, password: v.password },
    }),
  },
  recover: {
    page: '/recovery',
    send: (request, v) => identityRecoveryCreate({ ...request, body: { email: v.email, return_to: v.return_to } }),
  },
  'display-name': {
    page: '/',
    send: (request, v) => identityProfileUpdate({ ...request, body: { display_name: v.display_name } }),
  },
  email: {
    page: '/',
    send: (request, v) => identityProfileUpdate({ ...request, body: { email: v.email } }),
  },
  password: {
    page: '/',
    send: (request, v) => identityPasswordUpdate({ ...request, body: { password: v.password } }),
  },
  logout: {
    page: '/',
    send: request => identityLogout(request),
  },
  'delete-account': {
    page: '/delete-account',
    send: (request, v) => identityAccountDelete({ ...request, body: { email: v.email } }),
  },
} satisfies Record<string, { page: string; send: Send }>

export type Operation = keyof typeof operations

export function isOperation(name: string): name is Operation {
  return Object.hasOwn(operations, name)
}

/**
 * Identity API error codes, plus `unauthorized` for a 401 without one and `unchanged_email` for a Settings email
 * submission that sent nothing because the address wasn't edited.
 */
export type Code = NonNullable<IdentityErrorResponse['code']> | 'unauthorized' | 'unchanged_email'

export function isCode(value: unknown): value is Code {
  return value === 'unauthorized' || value === 'unchanged_email'
    || (value !== undefined && zIdentityErrorResponse.shape.code.safeParse(value).success)
}

/** An address as an email input shows it: browsers drop its line breaks and surrounding ASCII whitespace. */
function shownEmail(address: string) {
  return address.replace(/[\r\n]/g, '').replace(/^[\t\f ]+|[\t\f ]+$/g, '')
}

export type Result = { ok: true; response: Response } | { ok: false; code: Code }

/** Sends a form's values through its identity operation and reads the answer, failing closed on anything unexpected. */
export async function sendIdentity(operation: Operation, transport: Transport, values: Values): Promise<Result> {
  // The email form also carries the stored address, which its input can't always hold unchanged. Sending the shown
  // address would ask the API to change an address the visitor never edited.
  if (operation === 'email' && values.current_email !== undefined && values.email === shownEmail(values.current_email)) {
    return { ok: false, code: 'unchanged_email' }
  }
  const { error, response } = await operations[operation].send({
    ...transport,
    cache: 'no-store',
    redirect: 'error',
    signal: AbortSignal.timeout(15_000),
  }, values)
  if (!response || (response.ok && error !== undefined)) return { ok: false, code: 'unavailable' }
  if (response.ok) return { ok: true, response }
  const body = zIdentityErrorResponse.safeParse(error)
  if (body.success && body.data.code) return { ok: false, code: body.data.code }
  return { ok: false, code: response.status === 401 ? 'unauthorized' : 'unavailable' }
}

export type Problem = {
  readonly code: Code
  readonly message: string
  readonly field?: 'display_name' | 'email' | 'password'
  /** Settings problems that a fresh login resolves. */
  readonly reauthenticate?: boolean
}

/**
 * Describes an identity error code from the operation that returned it. Pass the submitted values to explain a
 * rejected password precisely.
 */
export function problemFor(code: Code, operation: Operation, values?: Record<string, string>): Problem {
  return { code, ...describe(code, operation, values) }
}

function describe(code: Code, operation: Operation, values?: Record<string, string>): Omit<Problem, 'code'> {
  switch (code) {
    case 'wrong_password':
      return { message: 'That email and password don’t match. Try again, or reset your password.' }
    case 'throttled':
      return operation === 'register'
        ? { message: 'Too many sign-up attempts from your network. Wait 15 minutes, then try again.' }
        : { message: 'Too many login attempts. Wait 15 minutes, then try again.' }
    case 'email_taken':
      return { field: 'email', message: 'This email is already registered. Log in or reset your password instead.' }
    case 'stale_session':
      return {
        reauthenticate: true,
        message: operation === 'delete-account'
          ? 'For your security, log in again before deleting your account.'
          : 'For your security, log in again before changing your email or password.',
      }
    case 'unauthorized':
      return { reauthenticate: true, message: 'Your session has ended. Log in again to continue.' }
    case 'invalid_display_name':
      return { field: 'display_name', message: displayNameMessage }
    case 'invalid_email':
      return { field: 'email', message: 'Enter a valid email address.' }
    case 'unchanged_email':
      return { field: 'email', message: 'This is already your email address. Enter a different one to change it.' }
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
    case 'invalid_recovery_token':
      return { message: 'This recovery link is invalid or has expired. Request a new one.' }
    case 'confirmation_mismatch':
      return { field: 'email', message: 'This doesn’t match your account’s email address. Check it and try again.' }
    case 'admin_account':
      return { message: 'Administrator accounts can’t be deleted. Ask another administrator to remove your role first.' }
    case 'banned_account':
      return { message: 'This account can’t be deleted here. Contact Tadoku support to delete it.' }
    case 'not_implemented':
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
  '^(?:(?:account|admin|authv2|next)\\.)?tadoku\\.(?:app|dev\\.lab)$',
  `^${label}(?:\\.(?:account|admin|authv2|next))?\\.(?:tadoku\\.dev\\.lab|preview\\.tadoku\\.app)$`,
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
 * host, with the longest Settings URL that carries it, recovery's `?edit=password&return_to=`, within 2,048 bytes.
 * Anything else has no continuation.
 */
export function continuation(raw: unknown, origin: string): string | undefined {
  if (typeof raw !== 'string' || /[\s\\]/u.test(raw)) return
  const authority = /^https:\/\/([^/?#]*)/iu.exec(raw)?.[1]?.toLowerCase()
  if (!authority || !allowedHost.test(authority)) return
  if (new TextEncoder().encode(`${origin}/?edit=password&return_to=`).length + queryEscapedLength(raw) > 2048) return
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

/**
 * Login that comes back to the delete page, keeping its trusted continuation. `reauth` asks a signed-in visitor for a
 * fresh login under a banner.
 */
export function deleteAccountLogin(origin: string, next: string | undefined, reauth: boolean) {
  const page = `${origin}/delete-account`
  // A continuation too long to nest in the page's own URL is dropped, not the return to the page.
  const here = [withReturnTo(page, next), page].find(url => continuation(url, origin))
  return withReturnTo(reauth ? '/login?reauth=1' : '/login', here)
}
