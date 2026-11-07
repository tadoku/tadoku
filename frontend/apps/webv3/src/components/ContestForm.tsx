import { useStore } from '@tanstack/react-form'
import { Link, useHydrated, useNavigate } from '@tanstack/react-router'
import type { ImmersionContestConfigurationOptions } from 'api-client'
import { Button, PageHeader, Surface, usePaperForm } from 'paper-ui'
import { useMemo, useState } from 'react'
import { contestDraftSchema, firstContestDraft, isOfficial, toNewContest } from '../lib/contest-form'
import { contestsHref, useCreateContest } from '../lib/contests'
import { useLinks } from '../lib/links'
import type { LogFailure } from '../lib/logs'
import { useLoginUrl } from '../lib/session'
import { renderLink } from './router-link'

const searchable = (name: string, query: string) =>
  name.toLowerCase().replace(/[^a-z0-9]/g, '').includes(query.toLowerCase())

/**
 * The contest form for the account that started the draft, which the server refuses to create for any other account.
 * A created contest replaces this page with its contest page.
 */
export function ContestForm({ options, viewerId, admin }: {
  options: ImmersionContestConfigurationOptions
  viewerId: string
  admin: boolean
}) {
  const hydrated = useHydrated()
  const links = useLinks()
  const navigate = useNavigate()
  const createContest = useCreateContest()
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const schema = useMemo(
    () => contestDraftSchema(options, admin, new Date().toISOString().slice(0, 10)),
    [options, admin],
  )
  const [defaultValues] = useState(() => firstContestDraft(options))
  const form = usePaperForm({
    defaultValues,
    validators: { onChange: schema, onSubmit: schema },
    onSubmit: async ({ value }) => {
      setFailure(null)
      const result = await createContest(toNewContest(value, options, viewerId))
      if (result.ok) await navigate({ to: links.contest(result.data), replace: true })
      else setFailure(result.reason)
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const official = useStore(form.store, state => isOfficial(state.values, options))
  const unlisted = useStore(form.store, state => state.values.unlisted)
  const disabled = !hydrated || pending

  return (
    <div className="goal-page">
      <PageHeader
        title="Create new contest"
        backLink={{ href: contestsHref.official(), label: 'Contests' }}
        renderLink={renderLink}
      />
      <Surface className="log-surface">
        <form.AppForm>
          <form.Form className="goal-form" aria-busy={pending}>
            <fieldset className="goal-form__controls" disabled={disabled}>
              <form.AppField name="title">
                {field => <field.Input label="Title" required />}
              </form.AppField>
              <form.AppField name="description">
                {field => (
                  <field.TextArea
                    label="Description (optional)"
                    rows={3}
                  />
                )}
              </form.AppField>

              <fieldset className="contest-form-group">
                <legend>Schedule</legend>
                <div className="goal-form__row">
                  <form.AppField name="contestStart">
                    {field => <field.Input label="Start date (UTC)" type="date" required />}
                  </form.AppField>
                  <form.AppField name="contestEnd">
                    {field => <field.Input label="End date (UTC)" type="date" required />}
                  </form.AppField>
                </div>
                <form.AppField name="registrationEnd">
                  {field => <field.Input label="Sign-up deadline (UTC)" type="date" required />}
                </form.AppField>
              </fieldset>

              <fieldset className="contest-form-group">
                <legend>Eligibility</legend>
                <form.AppField name="languages">
                  {field => (
                    <field.AutocompleteMultiInput
                      label="Allowed languages"
                      placeholder="All languages"
                      options={options.languages}
                      match={(language, query) => searchable(language.name, query)}
                      getId={language => language.code}
                      format={language => language.name}
                    />
                  )}
                </form.AppField>
                <form.AppField name="activities">
                  {field => (
                    <field.AutocompleteMultiInput
                      label="Allowed activities"
                      required
                      options={options.activities}
                      match={(activity, query) => searchable(activity.name, query)}
                      getId={activity => String(activity.id)}
                      format={activity => activity.name}
                    />
                  )}
                </form.AppField>
              </fieldset>
              <fieldset className="contest-form-group">
                <legend>Visibility</legend>
                {options.can_create_official_round && !unlisted ? (
                  <form.AppField name="official">
                    {field => <field.Checkbox label="Official round" />}
                  </form.AppField>
                ) : null}
                {official ? null : (
                  <form.AppField name="unlisted">
                    {field => <field.Checkbox label="Only those with the link can sign up" />}
                  </form.AppField>
                )}
              </fieldset>
            </fieldset>

            {failure ? <CreateFailure failure={failure} /> : null}
            <div className="paper-form-footer">
              <div className="paper-form-footer__actions">
                <Button type="submit" disabled={disabled}>{pending ? 'Creating…' : 'Create contest'}</Button>
                <Link className="text-link" to="/contests/official/$" params={{ _splat: '1' }}>Cancel</Link>
              </div>
            </div>
          </form.Form>
        </form.AppForm>
      </Surface>
    </div>
  )
}

const failureMessages: Partial<Record<LogFailure, string>> = {
  'account-unavailable': 'Account service unavailable. The contest was not created. Your entries are kept here.',
  switched: 'You are now signed in as a different account. The contest was not created for either account.',
  invalid: 'The contest was not created. Check the dates, languages and activities, then try again.',
  'signed-out': 'Your session ended. Log in in a new tab, return here and create the contest again.',
  forbidden: 'Your account cannot create a contest now. Accounts can create contests from one month after joining, up to 12 a year.',
  deleting: 'Your account is being deleted. Contests cannot be created.',
}

/** Explains why the contest was not created, or that it may have been. The draft stays in the form. */
function CreateFailure({ failure }: { failure: LogFailure }) {
  const loginUrl = useLoginUrl()

  return (
    <div role="alert">
      <p>
        {failureMessages[failure] ??
          'We could not confirm whether the contest was created. Check My contests before creating it again.'}
      </p>
      {failure === 'signed-out' ? (
        <a className="text-link" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
      ) : failure === 'switched' ? (
        <a className="text-link" href="/contests/new">Start a new contest for this account</a>
      ) : failure === 'unavailable' || failure === 'not-found' ? (
        <a className="text-link" href={contestsHref.mine()} target="_blank" rel="noopener noreferrer">Check My contests</a>
      ) : null}
    </div>
  )
}
