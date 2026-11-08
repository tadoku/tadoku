import { Button, Flash } from 'paper-ui'
import { ExclamationTriangleIcon, InformationCircleIcon, iconClassName } from 'paper-ui/icons'
import { continuation, displayNameProblem, emailProblem, passwordProblem, withReturnTo } from '../lib/identity'
import { navigate, useAccountForm } from '../lib/account-form'
import type { AccountSearch } from '../lib/search'
import { useIdentity } from '../lib/session'
import { FormProblem, ReturnTo } from './AccountForm'
import { CardPage } from './CardPage'
import { accountForm } from '../styles'

export const passwordHint = '8 or more characters. Avoid words from your email.'

function SignedInNotice() {
  const { session } = useIdentity()
  if (!session.account) return null
  return (
    <Flash variant="information" icon={next?.includes('delete-account') ? <ExclamationTriangleIcon className={iconClassName()} /> : <InformationCircleIcon className={iconClassName()} />}>
      You’re logged in as {session.account.display_name}. <a href="/">Go to your account</a>
    </Flash>
  )
}

export function Login({ search }: { readonly search: AccountSearch }) {
  const { origin, session } = useIdentity()
  const next = continuation(search.return_to, origin)
  // A privileged page sent this signed-in visitor back for a fresh login.
  const reauth = search.reauth ? session.account : null
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'login',
    id: 'login',
    defaultValues: { email: reauth?.email ?? '', password: '' },
    nativeError: search.error,
    onSuccess: () => navigate(next ?? '/'),
  })

  return (
    <CardPage
      title={!reauth ? 'Log in' : next?.includes('delete-account') ? 'Confirm account deletion' : 'Confirm it’s you'}
      footer={reauth ? (next?.includes('delete-account') ? <a href="/settings">Keep my account</a> : undefined) : <>New to Tadoku? <a href={withReturnTo('/register', next)}>Create an account</a></>}
    >
      <form.AppForm>
        <form.Form {...formProps} className={accountForm}>
          {reauth ? (
            <Flash variant={next?.includes('delete-account') ? 'danger' : 'information'} icon={<InformationCircleIcon className={iconClassName()} />}>
              Enter your password to continue.{next?.includes('delete-account') ? ' Deleting your account can’t be undone.' : ''}
            </Flash>
          ) : <SignedInNotice />}
          <FormProblem problem={problem} />
          <ReturnTo value={next} />
          {reauth ? <input type="hidden" name="reauth" value="1" /> : null}
          <form.AppField name="email" validators={{ onChange: ({ value }) => emailProblem(value) }}>
            {field => <field.Input id={fieldId('email')} label="Email" type="email" autoComplete="username" required maxLength={320} />}
          </form.AppField>
          <form.AppField name="password" validators={{ onChange: ({ value }) => value ? undefined : 'Enter your password.' }}>
            {field => <field.Input id={fieldId('password')} label="Password" type="password" autoComplete="current-password" required />}
          </form.AppField>
          <Button type="submit" variant={reauth && next?.includes('delete-account') ? 'destructive' : undefined} fullWidth loading={pending} loadingLabel="Logging in">{reauth ? (next?.includes('delete-account') ? 'Continue to deletion' : 'Continue') : 'Log in'}</Button>
          <a className="justify-self-center text-sm" href={withReturnTo('/recovery', next)}>Forgot your password?</a>
        </form.Form>
      </form.AppForm>
    </CardPage>
  )
}

export function Register({ search }: { readonly search: AccountSearch }) {
  const { origin } = useIdentity()
  const next = continuation(search.return_to, origin)
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'register',
    id: 'register',
    defaultValues: { display_name: '', email: '', password: '' },
    nativeError: search.error,
    onSuccess: () => navigate(next ?? '/'),
  })

  return (
    <CardPage
      title="Create an account"
      footer={<>Already have an account? <a href={withReturnTo('/login', next)}>Log in</a></>}
    >
      <form.AppForm>
        <form.Form {...formProps} className={accountForm}>
          <SignedInNotice />
          <FormProblem problem={problem} />
          <ReturnTo value={next} />
          <form.AppField name="display_name" validators={{ onChange: ({ value }) => displayNameProblem(value) }}>
            {field => <field.Input id={fieldId('display_name')} label="Display name" autoComplete="nickname" required hint="Shown on leaderboards and your profile." />}
          </form.AppField>
          <form.AppField name="email" validators={{ onChange: ({ value }) => emailProblem(value) }}>
            {field => <field.Input id={fieldId('email')} label="Email" type="email" autoComplete="username" required maxLength={320} />}
          </form.AppField>
          <form.AppField name="password" validators={{ onChange: ({ value }) => passwordProblem(value) }}>
            {field => <field.Input id={fieldId('password')} label="Password" type="password" autoComplete="new-password" required hint={passwordHint} />}
          </form.AppField>
          <Button type="submit" fullWidth loading={pending} loadingLabel="Creating your account">Create account</Button>
        </form.Form>
      </form.AppForm>
    </CardPage>
  )
}
