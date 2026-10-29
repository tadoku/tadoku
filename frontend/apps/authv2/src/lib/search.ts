import { isCode, isOperation, type Code, type Operation } from './identity'

const settingsEditors = ['display-name', 'email', 'password'] as const

export type SettingsEditor = (typeof settingsEditors)[number]

export type AccountSearch = {
  readonly return_to?: string
  /** Set by a native submission that the identity API rejected, with the form that sent it. */
  readonly error?: Code
  readonly form?: Operation
  /** The Settings editor a Change link opened. */
  readonly edit?: SettingsEditor
  /** Set by a Settings save that succeeded. */
  readonly saved?: SettingsEditor
  /** Set by a native recovery request. */
  readonly sent?: true
  /** Asks a signed-in visitor at /login for a fresh login before a privileged action; 1 keeps the URL `reauth=1`. */
  readonly reauth?: 1
}

/** Parses untrusted query parameters. The router also re-parses this function's own output, so it must accept it. */
export function accountSearch(search: Record<string, unknown>): AccountSearch {
  return {
    return_to: typeof search.return_to === 'string' ? search.return_to : undefined,
    error: isCode(search.error) ? search.error : undefined,
    form: typeof search.form === 'string' && isOperation(search.form) ? search.form : undefined,
    edit: settingsEditors.find(value => value === search.edit),
    saved: settingsEditors.find(value => value === search.saved),
    sent: [true, 1, '1'].includes(search.sent as never) ? true : undefined,
    reauth: [true, 1, '1'].includes(search.reauth as never) ? 1 : undefined,
  }
}

export type SettingsView =
  | { readonly state: 'summary'; readonly saved?: SettingsEditor }
  | { readonly state: 'editing'; readonly editor: SettingsEditor; readonly error?: Code }

/** The error's form wins over `edit`; `saved` only applies when nothing is open. */
export function settingsView(search: AccountSearch): SettingsView {
  const failed = settingsEditors.find(editor => editor === search.form)
  if (failed && search.error) return { state: 'editing', editor: failed, error: search.error }
  if (search.edit) return { state: 'editing', editor: search.edit }
  return { state: 'summary', saved: search.saved }
}
