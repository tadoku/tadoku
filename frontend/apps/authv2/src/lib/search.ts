import { isCode, isOperation, type Code, type Operation } from './identity'

const saved = ['display-name', 'email', 'password'] as const

export type AccountSearch = {
  readonly return_to?: string
  /** Set by a native submission that the identity API rejected, with the form that sent it. */
  readonly error?: Code
  readonly form?: Operation
  /** Set by a native Settings submission that succeeded. */
  readonly saved?: (typeof saved)[number]
  /** Set by a native recovery request. */
  readonly sent?: true
}

/** Parses untrusted query parameters. The router also re-parses this function's own output, so it must accept it. */
export function accountSearch(search: Record<string, unknown>): AccountSearch {
  return {
    return_to: typeof search.return_to === 'string' ? search.return_to : undefined,
    error: isCode(search.error) ? search.error : undefined,
    form: typeof search.form === 'string' && isOperation(search.form) ? search.form : undefined,
    saved: saved.find(value => value === search.saved),
    sent: [true, 1, '1'].includes(search.sent as never) ? true : undefined,
  }
}
