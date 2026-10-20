import { useLocation, useRouter } from '@tanstack/react-router'
import { Button, PageHeader, Surface, buttonClassName } from 'paper-ui'
import { ArrowLeftIcon, iconClassName } from 'paper-ui/icons'
import type { ReactNode } from 'react'
import { continuation, displayNameProblem, emailProblem, passwordProblem, problemFor, withReturnTo } from '../lib/identity'
import { navigate, useAccountForm } from '../lib/account-form'
import type { AccountSearch } from '../lib/search'
import { useIdentity } from '../lib/session'
import { FormProblem, ReturnTo } from './AccountForm'
import { passwordHint } from './Credentials'

type Account = { readonly email: string; readonly display_name: string }

type SectionProps = {
  readonly search: AccountSearch
  readonly account: Account
  /** Where Settings sends the visitor when they are done, if a trusted page asked for it. */
  readonly next?: string
  /** Login that returns here, for problems a fresh session resolves. */
  readonly reauthenticate: string
}

function SettingsSection({ id, title, description, children }: {
  readonly id: string
  readonly title: string
  readonly description: string
  readonly children: ReactNode
}) {
  return (
    <section className="settings-section" aria-labelledby={`${id}-heading`}>
      <div className="settings-section__intro">
        <h2 id={`${id}-heading`} className="settings-section__title">{title}</h2>
        <p>{description}</p>
      </div>
      <div className="settings-section__body">{children}</div>
    </section>
  )
}

function Saved({ message }: { readonly message?: string }) {
  return <p className="settings-saved" role="status">{message}</p>
}

function DisplayNameForm({ search, account, next, reauthenticate }: SectionProps) {
  const router = useRouter()
  const { form, formProps, fieldId, pending, problem, submitted, succeeded } = useAccountForm({
    operation: 'display-name',
    id: 'display-name',
    defaultValues: { display_name: account.display_name },
    nativeError: search.form === 'display-name' ? search.error : undefined,
    onSuccess: () => router.invalidate(),
  })
  const saved = succeeded || (search.saved === 'display-name' && !submitted)
  return (
    <SettingsSection id="profile" title="Profile" description="Your display name appears on leaderboards, in contests and on your profile.">
      <form.AppForm>
        <form.Form {...formProps} className="account-form">
          <FormProblem problem={problem} reauthenticate={reauthenticate} />
          <ReturnTo value={next} />
          <form.AppField name="display_name" validators={{ onChange: ({ value }) => displayNameProblem(value) }}>
            {field => <field.Input id={fieldId('display_name')} label="Display name" autoComplete="nickname" required />}
          </form.AppField>
          <div className="account-form__actions">
            <Button type="submit" variant="outline" loading={pending} loadingLabel="Saving">Save display name</Button>
            <Saved message={saved ? 'Display name saved.' : undefined} />
          </div>
        </form.Form>
      </form.AppForm>
    </SettingsSection>
  )
}

function EmailForm({ search, account, next, reauthenticate }: SectionProps) {
  const router = useRouter()
  const { form, formProps, fieldId, pending, problem, submitted, succeeded } = useAccountForm({
    operation: 'email',
    id: 'email',
    defaultValues: { email: account.email, current_email: account.email },
    nativeError: search.form === 'email' ? search.error : undefined,
    onSuccess: () => router.invalidate(),
  })
  const saved = succeeded || (search.saved === 'email' && !submitted)
  return (
    <SettingsSection
      id="email"
      title="Email address"
      description="You log in and recover your account with this address. Changing it needs a login within the last hour; we’ll let your old address know."
    >
      <form.AppForm>
        <form.Form {...formProps} className="account-form">
          <FormProblem problem={problem} reauthenticate={reauthenticate} />
          <ReturnTo value={next} />
          <input type="hidden" name="current_email" value={account.email} />
          <form.AppField name="email" validators={{ onChange: ({ value }) => emailProblem(value) }}>
            {field => <field.Input id={fieldId('email')} label="Email" type="email" autoComplete="email" required maxLength={320} />}
          </form.AppField>
          <div className="account-form__actions">
            <Button type="submit" variant="outline" loading={pending} loadingLabel="Saving">Change email</Button>
            <Saved message={saved ? 'Email changed. We’ve let your previous address know.' : undefined} />
          </div>
        </form.Form>
      </form.AppForm>
    </SettingsSection>
  )
}

function PasswordForm({ search, account, next, reauthenticate }: SectionProps) {
  const { form, formProps, fieldId, pending, problem, submitted, succeeded } = useAccountForm({
    operation: 'password',
    id: 'password',
    defaultValues: { password: '' },
    nativeError: search.form === 'password' ? search.error : undefined,
    onSuccess: (): Promise<never> | void => next ? navigate(next) : form.resetField('password'),
  })
  const saved = succeeded || (search.saved === 'password' && !submitted)
  return (
    <SettingsSection
      id="password"
      title="Password"
      description="Changing it needs a login within the last hour or a recovery link, and signs you out on your other devices."
    >
      <form.AppForm>
        <form.Form {...formProps} className="account-form">
          <FormProblem problem={problem} reauthenticate={reauthenticate} />
          <ReturnTo value={next} />
          <input type="email" name="username" autoComplete="username" value={account.email} hidden readOnly />
          <form.AppField name="password" validators={{ onChange: ({ value }) => passwordProblem(value) }}>
            {field => <field.Input id={fieldId('password')} label="New password" type="password" autoComplete="new-password" required hint={passwordHint} />}
          </form.AppField>
          <div className="account-form__actions">
            <Button type="submit" variant={next ? 'default' : 'outline'} loading={pending} loadingLabel="Saving">
              {next ? 'Change password and continue' : 'Change password'}
            </Button>
            <Saved message={saved ? 'Password changed. Your other devices have been signed out.' : undefined} />
          </div>
        </form.Form>
      </form.AppForm>
    </SettingsSection>
  )
}

export function Settings({ search }: { readonly search: AccountSearch }) {
  const { origin, session } = useIdentity()
  const pathname = useLocation({ select: location => location.pathname })
  const account = session.account
  if (!account) return null

  const next = continuation(search.return_to, origin)
  const here = withReturnTo(`${origin}${pathname}`, next)
  const props: SectionProps = {
    search,
    account,
    next,
    reauthenticate: continuation(here, origin) ? withReturnTo('/login', here) : '/login',
  }
  return (
    <div className="settings">
      <PageHeader
        title="Account settings"
        lead={<>Logged in as <strong>{account.email}</strong></>}
        actions={next ? (
          <a className={buttonClassName({ variant: 'outline' })} href={next}>
            <ArrowLeftIcon className={iconClassName('compact')} aria-hidden="true" />
            Back to Tadoku
          </a>
        ) : null}
      />
      {search.form === 'logout' && search.error ? <div className="settings-notice"><FormProblem problem={problemFor(search.error, 'logout')} /></div> : null}
      <Surface className="settings-list">
        <DisplayNameForm {...props} />
        <EmailForm {...props} />
        <PasswordForm {...props} />
      </Surface>
    </div>
  )
}
