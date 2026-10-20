export type Operation = 'login' | 'registration' | 'recovery' | 'profile' | 'password' | 'logout'

export const errors: Record<string, string> = {
  wrong_password: 'The email or password is incorrect.',
  throttled: 'Too many login attempts. Wait 15 minutes, then try again.',
  email_taken: 'This email is already registered.',
  stale_session: 'Log in again before changing your email or password.',
  unauthorized: 'Your session has ended. Log in again to continue.',
  invalid_display_name: 'Use 2–32 characters, including a visible character.',
  invalid_email: 'Enter a valid email address.',
  invalid_password: 'Use at least 8 characters, at most 72 UTF-8 bytes, and a password different from your email.',
  invalid_recovery_token: 'This recovery link has expired or has already been used. Request a new link.',
  invalid_origin: 'This request could not be verified. Reload this page and try again.',
  invalid_return_to: 'That return address is not allowed.',
  not_implemented: 'This account uses the existing account app. Open it to continue.',
  unavailable: 'Account service is unavailable. Try again later.',
}

export function errorMessage(code: string) {
  return Object.hasOwn(errors, code) ? errors[code] : errors.unavailable
}

export function returnDestination(raw: unknown): string | undefined {
  if (typeof raw !== 'string' || new TextEncoder().encode(raw).length > 2048 || /[\\\s]/u.test(raw)) return
  const authority = /^https:\/\/([^/?#]+)/iu.exec(raw)?.[1]?.toLowerCase()
  if (!authority || /[:@%]/u.test(authority)) return
  const label = '[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?'
  const canonical = /^(?:(?:account|admin|authv2)\.)?tadoku\.(?:app|dev\.lab)$/u
  const branch = new RegExp(`^${label}(?:\\.(?:account|admin|authv2|next))?\\.tadoku\\.dev\\.lab$|^${label}(?:\\.(?:account|admin|authv2))?\\.preview\\.tadoku\\.app$`)
  if (!canonical.test(authority) && !branch.test(authority)) return
  try {
    const url = new URL(raw)
    if (url.hostname === authority) return raw
  } catch { /* Invalid URLs have no continuation. */ }
}

export function accountSearch(search: Record<string, unknown>) {
  return {
    return_to: typeof search.return_to === 'string' ? search.return_to : undefined,
    error: typeof search.error === 'string' && Object.hasOwn(errors, search.error) ? search.error : undefined,
    saved: search.saved === 'profile' || search.saved === 'password' || search.saved === 'recovery' ? search.saved : undefined,
  }
}

export type ActionResult = { ok: true; redirect: string } | { ok: false; code: string }

export async function submitAccount(operation: Operation, values: Record<string, string>): Promise<ActionResult> {
  try {
    const response = await fetch(`/actions/${operation}`, {
      method: 'POST',
      headers: { Accept: 'application/json', 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams(values),
      credentials: 'same-origin',
      redirect: 'error',
    })
    const body: unknown = await response.json()
    if (typeof body !== 'object' || body === null) return { ok: false, code: 'unavailable' }
    if ('ok' in body && body.ok === true && 'redirect' in body && typeof body.redirect === 'string') {
      return { ok: true, redirect: body.redirect }
    }
    return { ok: false, code: 'code' in body && typeof body.code === 'string' ? body.code : 'unavailable' }
  } catch {
    return { ok: false, code: 'unavailable' }
  }
}
