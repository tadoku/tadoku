import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { Link } from '@tanstack/react-router'
import { Button, buttonStyles } from 'paper-ui'
import { useSyncExternalStore } from 'react'
import { navigate, useAccountForm } from '../lib/account-form'
import { continuation, deleteAccountLogin } from '../lib/identity'
import type { AccountSearch } from '../lib/search'
import { useIdentity } from '../lib/session'
import { FormProblem, ReturnTo } from './AccountForm'
import { CardPage } from './CardPage'
import { formStyles, cardStyles, textLink } from './styles'

const subscribeNever = () => () => {}

/**
 * The API's trimmed, per-character lowercase: Σ is always σ, never a final ς, and İ is i, not i with a combining dot.
 */
function normalized(address: string) {
  return [...address.trim()].map(character => character === 'İ' ? 'i' : character.toLowerCase()).join('')
}

function matches(typed: string, email: string) {
  return normalized(typed) === normalized(email)
}

export function DeleteAccount({ search }: { readonly search: AccountSearch }) {
  const { origin, webUrl, session } = useIdentity()
  const next = continuation(search.return_to, origin)
  const { form, formProps, fieldId, pending, problem } = useAccountForm({
    operation: 'delete-account',
    id: 'delete-account',
    defaultValues: { email: '' },
    nativeError: search.error,
    reauthenticate: deleteAccountLogin(origin, next, true),
    onSuccess: () => navigate('/deleted'),
  })
  const typed = useStore(form.store, state => state.values.email)
  // The server-rendered button stays enabled so the form also works before hydration and without JavaScript.
  const hydrated = useSyncExternalStore(subscribeNever, () => true, () => false)
  const account = session.account
  if (!account) return null
  const matched = matches(typed, account.email)

  return (
    <CardPage title="Delete your account" lead="You can’t undo this.">
      <div {...stylex.props(cardStyles.confirmation, cardStyles.spaced)}>
        <p {...stylex.props(cardStyles.confirmationText)}>Deleting your account removes your sign-in, your logs outside finished contests and your registrations in running contests.</p>
        <p {...stylex.props(cardStyles.confirmationText)}>Finished contests keep your scores as “Deleted participant”. Contests you organize continue under “Deleted organizer”.</p>
        <p {...stylex.props(cardStyles.confirmationText)}><a {...stylex.props(textLink)} href={`${webUrl}/profile/${account.id}/updates`}>Export your logs first</a> if you want to keep a copy.</p>
      </div>
      <form.AppForm>
        <form.Form {...formProps} xstyle={formStyles.form}>
          <FormProblem problem={problem} />
          <ReturnTo value={next} />
          <form.AppField name="email" validators={{ onChange: ({ value }) => value.trim() ? undefined : 'Type your email address.' }}>
            {field => (
              <field.Input
                id={fieldId('email')}
                label="Type your email address to confirm"
                inputMode="email"
                autoComplete="off"
                autoCapitalize="none"
                spellCheck={false}
                required
                maxLength={320}
              />
            )}
          </form.AppField>
          <div {...stylex.props(formStyles.actions)}>
            <Button xstyle={formStyles.action}
              type="submit"
              variant="destructive"
              loading={pending}
              loadingLabel="Deleting"
              disabled={hydrated && !matched}
            >
              Delete my account
            </Button>
            <Link to="/" search={{ return_to: next }} {...stylex.props(...buttonStyles({ variant: 'ghost' }))}>Cancel</Link>
          </div>
        </form.Form>
      </form.AppForm>
    </CardPage>
  )
}
