import { isCode, isOperation, type Code, type Operation } from './identity'

const settingsEditors = ['display-name', 'email', 'password'] as const

export type SettingsEditor = (typeof settingsEditors)[number]

/** A Settings row a save can confirm: an editor, or the avatar's fetch. */
export type SettingsSaved = SettingsEditor | 'avatar'

const settingsTabs = ['profile', 'appearance', 'security'] as const

export type SettingsTab = (typeof settingsTabs)[number]

const rowTabs: Record<SettingsSaved, SettingsTab> = {
  avatar: 'profile',
  'display-name': 'profile',
  email: 'security',
  password: 'security',
}

export type AccountSearch = {
  readonly return_to?: string
  /** Set by a native submission that the identity API rejected, with the form that sent it. */
  readonly error?: Code
  readonly form?: Operation
  /** The Settings tab a tab link opened. */
  readonly tab?: SettingsTab
  /** The Settings editor a Change link opened. */
  readonly edit?: SettingsEditor
  /** Set by a Settings save that succeeded. */
  readonly saved?: SettingsSaved
  /** Set by a native avatar fetch: a fresh query for the avatar image, so the browser can't show a cached one. */
  readonly fetched?: string
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
    tab: settingsTabs.find(value => value === search.tab),
    edit: settingsEditors.find(value => value === search.edit),
    saved: search.saved === 'avatar' ? 'avatar' : settingsEditors.find(value => value === search.saved),
    fetched: /^\d+$/.test(String(search.fetched)) ? String(search.fetched) : undefined,
    sent: [true, 1, '1'].includes(search.sent as never) ? true : undefined,
    reauth: [true, 1, '1'].includes(search.reauth as never) ? 1 : undefined,
  }
}

export type SettingsView =
  | { readonly state: 'summary'; readonly tab: SettingsTab; readonly saved?: SettingsSaved }
  | { readonly state: 'editing'; readonly tab: SettingsTab; readonly editor: SettingsEditor; readonly error?: Code }

/**
 * The error's form wins over `edit`; `saved` only applies when nothing is open. An open, failed or saved row shows its
 * own tab, so native redirects that only name the row land on it.
 */
export function settingsView(search: AccountSearch): SettingsView {
  const failed = settingsEditors.find(editor => editor === search.form)
  if (failed && search.error) return { state: 'editing', tab: rowTabs[failed], editor: failed, error: search.error }
  if (search.edit) return { state: 'editing', tab: rowTabs[search.edit], editor: search.edit }
  const avatarFailed = search.form === 'avatar' && search.error
  const tab = search.saved ? rowTabs[search.saved] : avatarFailed ? 'profile' : search.tab ?? 'profile'
  return { state: 'summary', tab, saved: search.saved }
}
