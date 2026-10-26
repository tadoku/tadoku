import { useStore } from '@tanstack/react-form'
import { useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, PageHeader, Surface, usePaperForm, withPaperForm, type Option, type OptionGroup } from 'paper-ui'
import { useMemo, useState } from 'react'
import {
  dependentFields,
  eligibleRegistrations,
  firstDraft,
  isTimed,
  logDraftSchema,
  toNewLog,
  unitsFor,
  type LogDraft,
} from '../lib/log-draft'
import { useLinks } from '../lib/links'
import { useSaveLog, type LogFailure, type LogOptions } from '../lib/logs'
import { useLoginUrl } from '../lib/session'

/** The activity form for the account that started the draft. The server refuses to save it for any other account. */
export function LogForm({ options, viewerId }: { options: LogOptions; viewerId: string }) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const saveLog = useSaveLog()
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const schema = useMemo(() => logDraftSchema(options), [options])
  const form = usePaperForm({
    defaultValues: useMemo(() => firstDraft(options), [options]),
    validators: { onChange: schema, onSubmit: schema },
    onSubmit: async ({ value }) => {
      setFailure(null)
      const result = await saveLog(toNewLog(value, options, viewerId))
      if (result.ok) await navigate({ to: '/logs/$logId', params: { logId: result.data.id } })
      else setFailure(result.reason)
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const languageCode = useStore(form.store, state => state.values.languageCode)
  const activityId = useStore(form.store, state => state.values.activityId)

  const selectDependents = (language: string, activity: string) => {
    const { unitId, registrationIds } = dependentFields(options, language, activity)
    form.setFieldValue('unitId', unitId)
    form.setFieldValue('registrationIds', registrationIds)
  }
  const units = unitsFor(options, languageCode, activityId).map(unit => ({ value: unit.id, label: unit.name }))
  const eligible = eligibleRegistrations(options, languageCode, activityId)
  const disabled = !hydrated || pending

  return (
    <div className="goal-page">
      <PageHeader title="Log activity" lead="Keep a record of what you read, listen to or practise." />
      <Surface className="log-surface">
        <form.AppForm>
          <form.Form className="goal-form" aria-busy={pending}>
            <fieldset className="goal-form__controls" disabled={disabled}>
              <form.AppField
                name="languageCode"
                listeners={{ onChange: ({ value }) => selectDependents(value, form.getFieldValue('activityId')) }}
              >
                {field => <field.Select label="Language" options={[]} groups={languageGroups(options)} />}
              </form.AppField>
              <form.AppField
                name="activityId"
                listeners={{
                  onChange: ({ value }) => {
                    selectDependents(form.getFieldValue('languageCode'), value)
                    // Activities count in different units: a quantity for the previous one doesn't carry over.
                    form.setFieldValue('amount', Number.NaN)
                    form.setFieldValue('minutes', Number.NaN)
                  },
                }}
              >
                {field => (
                  <field.Select
                    label="Activity"
                    options={options.configuration.activities.map(activity => ({
                      value: String(activity.id),
                      label: activity.name,
                    }))}
                  />
                )}
              </form.AppField>

              <MeasureField form={form} timed={isTimed(options, activityId)} units={units} />

              <form.AppField name="description">
                {field => <field.TextArea label="Description" rows={3} />}
              </form.AppField>
              <form.AppField name="tags">
                {field => <field.Input label="Tags" hint="Separate tags with commas. Up to 10 tags." />}
              </form.AppField>

              <fieldset className="paper-choice-field">
                <legend className="paper-field__label">Contests</legend>
                {eligible.map(registration => (
                  <form.AppField key={registration.id} name="registrationIds">
                    {field => <field.Checkbox label={registration.contest.title} value={registration.id} />}
                  </form.AppField>
                ))}
                {eligible.length === 0 ? <p>Saved to your personal record only.</p> : null}
              </fieldset>
            </fieldset>

            {failure ? <SaveFailure failure={failure} viewerId={viewerId} /> : null}
            <Button type="submit" disabled={disabled}>{pending ? 'Saving…' : 'Save activity'}</Button>
          </form.Form>
        </form.AppForm>
      </Surface>
    </div>
  )
}

/** Minutes for a timed activity; otherwise an amount in one of the activity's units. */
const MeasureField = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { timed: false, units: [] as Option[] },
  render: function MeasureField({ form, timed, units }) {
    if (timed) {
      return (
        <form.AppField name="minutes">
          {field => <field.Input label="Minutes" type="number" inputMode="decimal" step="any" min={0} />}
        </form.AppField>
      )
    }
    // The amount field must be the innermost, because AmountWithUnit binds the field it is rendered in.
    return (
      <form.AppField name="unitId">
        {unitField => (
          <form.AppField name="amount">
            {field => (
              <field.AmountWithUnit
                label="Amount"
                unitsLabel="Unit"
                units={units.length ? units : [{ value: '', label: 'No units available' }]}
                unitField={unitField}
                inputMode="decimal"
                step="any"
                min={0}
              />
            )}
          </form.AppField>
        )}
      </form.AppField>
    )
  },
})

/** The viewer's own languages first, when they have chosen any. */
function languageGroups({ configuration }: LogOptions): OptionGroup[] {
  const own = new Set(configuration.user_language_codes ?? [])
  const options = (mine: boolean) => configuration.languages
    .filter(language => own.has(language.code) === mine)
    .map(language => ({ value: language.code, label: language.name }))

  const all = { label: 'All languages', options: options(false) }
  return own.size ? [{ label: 'Your languages', options: options(true) }, all] : [all]
}

const failureMessages: Record<LogFailure, string> = {
  'account-unavailable':
    'Account service unavailable. Nothing was saved. Your entries are kept here; try again when the service recovers.',
  switched:
    'You are now signed in as a different account. Nothing was saved. This draft is kept for the account that started it.',
  invalid: 'The activity was not saved. Check the amount, unit and contests, then try again.',
  'signed-out': 'Your session ended. Log in in a new tab, return here and save again. Your entries are kept here.',
  forbidden: 'Your account cannot save activity. Your entries are kept here.',
  deleting: 'Your account is being deleted. Activity cannot be saved.',
  unavailable:
    'We could not confirm whether your activity was saved. Check your activity before saving again to avoid a duplicate.',
  'not-found': 'The activity could not be saved. Review the contests and try again.',
}

/** Explains why nothing was saved. The draft stays in the form. */
function SaveFailure({ failure, viewerId }: { failure: LogFailure; viewerId: string }) {
  return (
    <div role="alert">
      <p>{failureMessages[failure]}</p>
      <FailureAction failure={failure} viewerId={viewerId} />
    </div>
  )
}

/** The way forward after a failure, where there is one. */
function FailureAction({ failure, viewerId }: { failure: LogFailure; viewerId: string }) {
  const loginUrl = useLoginUrl()
  const links = useLinks()

  switch (failure) {
    case 'switched':
      return <a className="text-link" href="/logs/new">Start a new activity for this account</a>
    case 'signed-out':
      return <a className="text-link" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
    case 'unavailable':
      return (
        <a
          className="text-link"
          href={links.legacyUrl(`/profile/${viewerId}/updates`)}
          target="_blank"
          rel="noopener noreferrer"
        >
          Check your activity
        </a>
      )
    default:
      return null
  }
}
