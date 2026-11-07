import { Link, useLocation, useRouter } from '@tanstack/react-router'
import { Button, PageHeader, Surface, buttonClassName } from 'paper-ui'
import { ArrowLeftIcon, iconClassName } from 'paper-ui/icons'
import { useEffect, useRef, type ReactNode } from 'react'
import {
  continuation, displayNameProblem, emailProblem, passwordProblem, problemFor, withReturnTo, type Code,
} from '../lib/identity'
import { navigate, useAccountForm } from '../lib/account-form'
import { settingsView, type AccountSearch, type SettingsEditor, type SettingsView } from '../lib/search'
import { useIdentity } from '../lib/session'
import { FormProblem, ReturnTo } from './AccountForm'
import { passwordHint } from './Credentials'
import { accountForm, accountFormActions } from '../styles'

type Account = { readonly email: string; readonly display_name: string }

type EditorProps = {
  readonly account: Account
  /** Where Settings sends the visitor when they are done, if a trusted page asked for it. */
  readonly next?: string
  /** The error a native submission from this editor returned. */
  readonly error?: Code
  /** Login that returns to this editor, for problems a fresh session resolves. */
  readonly reauthenticate: string
}

const confirmations: Record<SettingsEditor, string> = {
  'display-name': 'Display name saved.',
  email: 'Email changed. We’ve let your previous address know.',
  password: 'Password changed. Your other devices have been signed out.',
}

/**
 * Collapses the editor onto the URL a native save redirects to, so both paths end in the same state. The navigation
 * reads the session again, so the row shows the saved value.
 */
function useShowSaved(next: string | undefined) {
  const router = useRouter()
  return (saved: SettingsEditor) => router.navigate({
    to: '.',
    search: { return_to: next, saved },
    replace: true,
    resetScroll: false,
  })
}

const settingRow = [
  'grid grid-cols-settings-row items-center gap-x-6 gap-y-1 px-5 py-4 [&+&]:border-t [&+&]:border-rule-subtle',
  'max-tablet:grid-cols-content-end',
].join(' ')
const settingRowOpen = 'grid-cols-settings-row-open items-start bg-raised shadow-rail-focus max-tablet:grid-cols-content-end'
// The email hint holds the stored address, whose raw spaces would otherwise hang past the row.
const settingEditor = [
  'min-w-0 max-w-settings-editor wrap-anywhere max-tablet:col-span-full max-tablet:max-w-none',
  '[&_[id$=-hint]]:whitespace-break-spaces [&_form]:gap-4',
].join(' ')
const settingNote = 'm-0 text-metadata leading-metadata font-regular text-success'

function SettingsGroup({ title, children }: { readonly title: string; readonly children: ReactNode }) {
  return (
    <div className="[&+&]:mt-8">
      <h2 className="mb-3 mt-0 font-display text-component leading-component">{title}</h2>
      <Surface className="p-0">{children}</Surface>
    </div>
  )
}

/** One setting: its value and a Change link, or its editor while the URL opens it. */
function SettingRow({ editor, label, value, view, next, children }: {
  readonly editor: SettingsEditor
  readonly label: string
  readonly value: ReactNode
  readonly view: SettingsView
  readonly next?: string
  readonly children: ReactNode
}) {
  const open = view.state === 'editing' && view.editor === editor
  const saved = view.state === 'summary' && view.saved === editor
  const change = useRef<HTMLAnchorElement>(null)
  const wasOpen = useRef(open)
  useEffect(() => {
    // Saving or cancelling removes the focused control along with the editor; keep keyboard users on this row.
    if (wasOpen.current && !open && document.activeElement === document.body) change.current?.focus()
    wasOpen.current = open
  }, [open])

  return (
    <section className={open ? `${settingRow} ${settingRowOpen}` : settingRow} aria-labelledby={`${editor}-label`}>
      <h3
        id={`${editor}-label`}
        className="m-0 font-body text-metadata leading-metadata font-regular text-muted max-tablet:col-span-full"
      >
        {label}
      </h3>
      {open ? <div className={settingEditor}>{children}</div> : (
        <>
          <div className="min-w-0">
            <p className="settings-row__value m-0 text-body leading-body text-ink wrap-anywhere whitespace-break-spaces">
              {value}
            </p>
            {saved ? <p id={`${editor}-saved`} className={settingNote} role="status">{confirmations[editor]}</p> : null}
          </div>
          <Link
            ref={change}
            to="."
            search={{ edit: editor, return_to: next }}
            resetScroll={false}
            className={buttonClassName({ variant: 'outline' })}
            aria-describedby={saved ? `${editor}-saved` : undefined}
          >
            Change<span className="sr-only"> {label.toLowerCase()}</span>
          </Link>
        </>
      )}
    </section>
  )
}

function Cancel({ next }: { readonly next?: string }) {
  // Its search is a subset of the open editor's, which would otherwise mark it as the current page.
  return (
    <Link
      to="."
      search={{ return_to: next }}
      resetScroll={false}
      activeOptions={{ exact: true }}
      className={buttonClassName({ variant: 'ghost' })}
    >
      Cancel
    </Link>
  )
}

function DisplayNameForm({ account, next, error, reauthenticate }: EditorProps) {
  const showSaved = useShowSaved(next)
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'display-name',
    id: 'display-name',
    defaultValues: { display_name: account.display_name },
    nativeError: error,
    onSuccess: () => showSaved('display-name'),
  })
  return (
    <form.AppForm>
      <form.Form {...formProps} className={accountForm}>
        <FormProblem problem={problem} reauthenticate={reauthenticate} />
        <ReturnTo value={next} />
        <form.AppField name="display_name" validators={{ onChange: ({ value }) => displayNameProblem(value) }}>
          {field => <field.Input id={fieldId('display_name')} label="Display name" autoComplete="nickname" required autoFocus />}
        </form.AppField>
        <div className={accountFormActions}>
          <Cancel next={next} />
          <Button type="submit" loading={pending} loadingLabel="Saving">Save display name</Button>
        </div>
      </form.Form>
    </form.AppForm>
  )
}

function EmailForm({ account, next, error, reauthenticate }: EditorProps) {
  const showSaved = useShowSaved(next)
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'email',
    id: 'email',
    defaultValues: { email: account.email, current_email: account.email },
    nativeError: error,
    onSuccess: () => showSaved('email'),
  })
  return (
    <form.AppForm>
      <form.Form {...formProps} className={accountForm}>
        <FormProblem problem={problem} reauthenticate={reauthenticate} />
        <ReturnTo value={next} />
        <input type="hidden" name="current_email" value={account.email} />
        <form.AppField name="email" validators={{ onChange: ({ value }) => emailProblem(value) }}>
          {field => (
            <field.Input
              id={fieldId('email')}
              label="New email"
              type="email"
              autoComplete="email"
              required
              maxLength={320}
              autoFocus
              hint={`Needs a login within the last hour. We’ll send a notice to ${account.email}.`}
            />
          )}
        </form.AppField>
        <div className={accountFormActions}>
          <Cancel next={next} />
          <Button type="submit" loading={pending} loadingLabel="Saving">Save email</Button>
        </div>
      </form.Form>
    </form.AppForm>
  )
}

function PasswordForm({ account, next, error, reauthenticate }: EditorProps) {
  const showSaved = useShowSaved(next)
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'password',
    id: 'password',
    defaultValues: { password: '' },
    nativeError: error,
    onSuccess: () => next ? navigate(next) : showSaved('password'),
  })
  return (
    <form.AppForm>
      <form.Form {...formProps} className={accountForm}>
        <FormProblem problem={problem} reauthenticate={reauthenticate} />
        <ReturnTo value={next} />
        <input type="email" name="username" autoComplete="username" value={account.email} hidden readOnly />
        <form.AppField name="password" validators={{ onChange: ({ value }) => passwordProblem(value) }}>
          {field => (
            <field.Input
              id={fieldId('password')}
              label="New password"
              type="password"
              autoComplete="new-password"
              required
              autoFocus
              hint={passwordHint}
              aria-describedby="password-policy"
            />
          )}
        </form.AppField>
        <p id="password-policy" className="m-0 text-metadata leading-metadata text-muted">
          Needs a login within the last hour or a recovery link. Signs you out on your other devices.
        </p>
        <div className={accountFormActions}>
          <Cancel next={next} />
          <Button type="submit" loading={pending} loadingLabel="Saving">
            {next ? 'Save password and continue' : 'Save password'}
          </Button>
        </div>
      </form.Form>
    </form.AppForm>
  )
}

export function Settings({ search }: { readonly search: AccountSearch }) {
  const { origin, session } = useIdentity()
  const pathname = useLocation({ select: location => location.pathname })
  const account = session.account
  if (!account) return null

  const view = settingsView(search)
  const next = continuation(search.return_to, origin)
  const here = withReturnTo(`${origin}${pathname}${view.state === 'editing' ? `?edit=${view.editor}` : ''}`, next)
  const editor: EditorProps = {
    account,
    next,
    error: view.state === 'editing' ? view.error : undefined,
    reauthenticate: continuation(here, origin) ? withReturnTo('/login', here) : '/login',
  }
  const row = { view, next }
  return (
    <div className="mx-auto w-settings">
      <PageHeader
        title="Account settings"
        className="mb-8"
        actions={next ? (
          <a className={buttonClassName({ variant: 'outline' })} href={next}>
            <ArrowLeftIcon className={iconClassName('compact')} aria-hidden="true" />
            Back to Tadoku
          </a>
        ) : null}
      />
      {search.form === 'logout' && search.error ? <div className="mb-5"><FormProblem problem={problemFor(search.error, 'logout')} /></div> : null}
      <SettingsGroup title="Profile">
        <SettingRow
          {...row}
          editor="display-name"
          label="Display name"
          value={account.display_name}
        >
          <DisplayNameForm {...editor} />
        </SettingRow>
      </SettingsGroup>
      <SettingsGroup title="Sign-in and security">
        <SettingRow
          {...row}
          editor="email"
          label="Email"
          value={account.email}
        >
          <EmailForm {...editor} />
        </SettingRow>
        <SettingRow
          {...row}
          editor="password"
          label="Password"
          value={<><span aria-hidden="true">••••••••••</span><span className="sr-only">Hidden</span></>}
        >
          <PasswordForm {...editor} />
        </SettingRow>
      </SettingsGroup>
      <p className="mb-0 mt-10">
        <Link to="/delete-account" search={{ return_to: next }} className="text-danger">Delete account</Link>
      </p>
    </div>
  )
}
