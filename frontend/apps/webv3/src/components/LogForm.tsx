import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { Link, useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, PageHeader, Surface, usePaperForm, withPaperForm, type Option, type OptionGroup } from 'paper-ui'
import { useEffect, useMemo, useState } from 'react'
import {
  dependentFields,
  draftFromLog,
  eligibleRegistrations,
  firstDraft,
  isTimed,
  logDraftSchema,
  toLogUpdate,
  toNewLog,
  toScorePreview,
  unitsFor,
  type LogDraft,
} from '../lib/log-draft'
import { useLinks } from '../lib/links'
import {
  scorePreviewQuery,
  useSaveLog,
  useUpdateLog,
  type ImmersionLog,
  type LogFailure,
  type LogOptions,
} from '../lib/logs'
import { useLoginUrl } from '../lib/session'

/** A new activity, or a change to a saved one, whose language, activity and contest submissions stay fixed. */
export type LogFormTarget = { kind: 'new' } | { kind: 'edit'; log: ImmersionLog }

/** The activity form for the account that started the draft. The server refuses to save it for any other account. */
export function LogForm({ options, viewerId, target }: { options: LogOptions; viewerId: string; target: LogFormTarget }) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const saveLog = useSaveLog()
  const updateLog = useUpdateLog()
  const [failure, setFailure] = useState<LogFailure | null>(null)
  const schema = useMemo(() => logDraftSchema(options), [options])
  const [defaultValues] = useState(() => target.kind === 'edit' ? draftFromLog(options, target.log) : firstDraft(options))
  const form = usePaperForm({
    defaultValues,
    validators: { onChange: schema, onSubmit: schema },
    onSubmit: async ({ value }) => {
      setFailure(null)
      const result = target.kind === 'edit'
        ? await updateLog(toLogUpdate(value, options, viewerId, target.log))
        : await saveLog(toNewLog(value, options, viewerId))
      if (result.ok) {
        await navigate({ to: '/logs/$logId', params: { logId: result.data.id }, replace: target.kind === 'edit' })
      } else {
        setFailure(result.reason)
      }
    },
  })
  const pending = useStore(form.store, state => state.isSubmitting)
  const languageCode = useStore(form.store, state => state.values.languageCode)
  const activityId = useStore(form.store, state => state.values.activityId)

  const units = unitsFor(options, languageCode, activityId).map(unit => ({ value: unit.unit_key, label: unit.name }))
  const eligible = eligibleRegistrations(options, languageCode, activityId)
  const disabled = !hydrated || pending

  const editing = target.kind === 'edit' ? target.log : null

  return (
    <div className="goal-page">
      {editing ? (
        <PageHeader title="Edit activity" backLink={{ href: `/logs/${editing.id}`, label: 'Back to log' }} />
      ) : (
        <PageHeader title="Log activity" lead="Keep a record of what you read, listen to or practise." />
      )}
      <Surface className="log-surface">
        <form.AppForm>
          <form.Form className="goal-form" aria-busy={pending}>
            <fieldset className="goal-form__controls" disabled={disabled}>
              {editing ? <FixedFields form={form} log={editing} /> : <ChoiceFields form={form} options={options} />}

              <MeasureField form={form} timed={isTimed(options, activityId)} units={units} />

              <form.AppField name="description">
                {field => <field.TextArea label="Description" rows={3} />}
              </form.AppField>
              <form.AppField name="tags">
                {field => <field.Input label="Tags" hint="Separate tags with commas. Up to 10 tags." />}
              </form.AppField>

              {editing ? null : (
                <fieldset className="paper-choice-field">
                  <legend className="paper-field__label">Contests</legend>
                  {eligible.map(registration => (
                    <form.AppField key={registration.id} name="registrationIds">
                      {field => <field.Checkbox label={registration.contest.title} value={registration.id} />}
                    </form.AppField>
                  ))}
                  {eligible.length === 0 ? <p>Saved to your personal record only.</p> : null}
                </fieldset>
              )}
              <ScoreEstimate form={form} options={options} />
            </fieldset>

            {failure ? <SaveFailure failure={failure} viewerId={viewerId} editing={editing} /> : null}
            <div className="goal-actions">
              <Button type="submit" disabled={disabled}>
                {pending ? 'Saving…' : editing ? 'Save changes' : 'Save activity'}
              </Button>
              {editing ? <Link className="text-link" to="/logs/$logId" params={{ logId: editing.id }}>Cancel</Link> : null}
            </div>
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
      <>
        <form.AppField name="unitKey">
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
        <form.AppField name="minutes">
          {field => (
            <field.Input label="Time spent" hint="Optional, in minutes." type="number" inputMode="decimal" step="any" min={0} />
          )}
        </form.AppField>
      </>
    )
  },
})

/** A new log's language and activity. Changing either selects their unit and eligible contests again. */
const ChoiceFields = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { options: {} as LogOptions },
  render: function ChoiceFields({ form, options }) {
    const selectDependents = (language: string, activity: string) => {
      const { unitKey, registrationIds } = dependentFields(options, language, activity)
      form.setFieldValue('unitKey', unitKey)
      form.setFieldValue('registrationIds', registrationIds)
    }

    return (
      <>
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
      </>
    )
  },
})

/** A saved log's language and activity, which an edit cannot change. */
const FixedFields = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { log: {} as ImmersionLog },
  render: function FixedFields({ form, log }) {
    return (
      <>
        <form.AppField name="languageCode">
          {field => (
            <field.Select label="Language" disabled options={[{ value: log.language.code, label: log.language.name }]} />
          )}
        </form.AppField>
        <form.AppField name="activityId">
          {field => (
            <field.Select
              label="Activity"
              hint="Language and activity stay fixed for existing entries."
              disabled
              options={[{ value: String(log.activity.id), label: log.activity.name }]}
            />
          )}
        </form.AppField>
      </>
    )
  },
})

const estimateFormat = new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 })

/**
 * The API's estimate of the draft's score, requested once the draft is valid and has stopped changing for 300 ms. An
 * incomplete draft, a pending first estimate and a failed one show a dash, and a zero estimate shows N/A.
 */
const ScoreEstimate = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { options: {} as LogOptions },
  render: function ScoreEstimate({ form, options }) {
    const schema = useMemo(() => logDraftSchema(options), [options])
    const draft = useStore(form.store, state => state.values)
    const input = schema.safeParse(draft).success ? JSON.stringify(toScorePreview(draft, options)) : undefined
    const [settled, setSettled] = useState(input)
    useEffect(() => {
      const timer = setTimeout(() => setSettled(input), 300)
      return () => clearTimeout(timer)
    }, [input])
    const { data } = useQuery(scorePreviewQuery(settled === undefined ? undefined : JSON.parse(settled)))

    const score = input !== undefined && data?.ok ? data.data.platform.score : undefined
    return (
      <p className="log-estimate" aria-live="polite">
        <span>Estimated score</span>
        <strong>{score === undefined ? '-' : score ? estimateFormat.format(score) : 'N/A'}</strong>
      </p>
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

/** Where an edit's outcome differs: retrying an update cannot duplicate it, and its log can disappear. */
const editFailureMessages: Partial<Record<LogFailure, string>> = {
  invalid: 'Your changes were not saved. Check the amount and unit, then try again.',
  forbidden: 'Your account cannot change this activity. Your entries are kept here.',
  unavailable: 'We could not confirm whether your changes were saved. Saving again is safe.',
  'not-found': 'This activity no longer exists. Nothing was saved.',
}

/** Explains why nothing was saved. The draft stays in the form. */
function SaveFailure({ failure, viewerId, editing }: { failure: LogFailure; viewerId: string; editing: ImmersionLog | null }) {
  return (
    <div role="alert">
      <p>{(editing && editFailureMessages[failure]) ?? failureMessages[failure]}</p>
      {editing ? <EditFailureAction failure={failure} /> : <FailureAction failure={failure} viewerId={viewerId} />}
    </div>
  )
}

function EditFailureAction({ failure }: { failure: LogFailure }) {
  const loginUrl = useLoginUrl()
  return failure === 'signed-out'
    ? <a className="text-link" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
    : null
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
