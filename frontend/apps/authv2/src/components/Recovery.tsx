import { Button } from 'paper-ui'
import { continuation, emailProblem, withReturnTo } from '../lib/identity'
import { useAccountForm } from '../lib/account-form'
import type { AccountSearch } from '../lib/search'
import { useIdentity } from '../lib/session'
import { FormProblem, ReturnTo } from './AccountForm'
import { CardPage } from './CardPage'

export function Recovery({ search }: { readonly search: AccountSearch }) {
  const { origin, webUrl } = useIdentity()
  const next = continuation(search.return_to, origin)
  const { form, formProps, fieldId, pending, problem, submitted, succeeded } = useAccountForm({
    operation: 'recover',
    id: 'recover',
    defaultValues: { email: '' },
    nativeError: search.error,
    body: ({ email }) => ({ email, return_to: next ?? webUrl }),
    onSuccess: () => undefined,
  })
  const login = withReturnTo('/login', next)

  if (succeeded || (search.sent && !submitted)) {
    return (
      <CardPage title="Check your email" footer={<a href={login}>Back to log in</a>}>
        <div className="account-confirmation" role="status">
          <p>If an account uses that address, we’ve sent it a link to choose a new password. The link works once and expires after an hour.</p>
          <p>Nothing arrived? Check your spam folder, or <a href={withReturnTo('/recovery', next)}>send another link</a>.</p>
        </div>
      </CardPage>
    )
  }

  return (
    <CardPage
      title="Reset your password"
      lead="Enter the email you use for Tadoku. We’ll send you a link to choose a new password."
      footer={<>Remembered it? <a href={login}>Log in</a></>}
    >
      <form.AppForm>
        <form.Form {...formProps} className="account-form">
          <FormProblem problem={problem} />
          <ReturnTo value={next} />
          <form.AppField name="email" validators={{ onChange: ({ value }) => emailProblem(value) }}>
            {field => <field.Input id={fieldId('email')} label="Email" type="email" autoComplete="email" required maxLength={320} />}
          </form.AppField>
          <Button type="submit" fullWidth loading={pending} loadingLabel="Sending">Send recovery link</Button>
        </form.Form>
      </form.AppForm>
    </CardPage>
  )
}
