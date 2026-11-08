import { useQuery } from '@tanstack/react-query'
import { useStore } from '@tanstack/react-form'
import { Link, useHydrated, useNavigate } from '@tanstack/react-router'
import { Button, buttonClassName, cn, formFooter, type Option, type OptionGroup, PageHeader, Surface, usePaperForm, withPaperForm } from 'paper-ui'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
  draftFromLog,
  firstDraft,
  isTimed,
  logDraftSchema,
  toLogUpdate,
  toNewLog,
  toScorePreview,
  unitFor,
  unitsFor,
  type LogDraft,
} from '../lib/log-draft'
import { readLastLogged, rememberLastLogged } from '../lib/log-preferences'
import { useLinks } from '../lib/links'
import {
  scorePreviewQuery,
  tagSuggestionsQuery,
  useSaveLog,
  useUpdateLog,
  type ImmersionLog,
  type LogFailure,
  type LogOptions,
} from '../lib/logs'
import { useLoginUrl } from '../lib/session'
import { scoreFormat } from '../lib/format'
import { ChevronRightIcon } from 'paper-ui/icons'
import { TagsSidebar, maxTags } from './TagsSidebar'
import { formControls } from './form-styles'

/** A new activity, or a change to a saved one, whose language, activity and contest submissions stay fixed. */
export type LogFormTarget = { kind: 'new'; amount?: number } | { kind: 'edit'; log: ImmersionLog }

/**
 * The activity form for the account that started the draft. The server refuses to save it for any other account. A new
 * log starts from the browser's last logged language and activity once hydrated, and after saving opens its contest
 * step when the viewer has ongoing registrations.
 */
const card = [
  'min-w-0 overflow-hidden p-0 wrap-anywhere',
  '[&>.paper-flash]:mx-record-inset [&>.paper-flash]:mb-0 [&>.paper-flash]:mt-4',
  '[&>[role=alert]]:mx-record-inset [&>[role=alert]]:mb-0 [&>[role=alert]]:mt-4',
].join(' ')

export function LogForm({ options, viewerId, target }: { options: LogOptions; viewerId: string; target: LogFormTarget }) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const saveLog = useSaveLog()
  const updateLog = useUpdateLog()
  const [failure, setFailure] = useState<LogFailure | null>(null)
  // Details stay open on wide screens and while editing; on a phone they fold away behind one tap.
  const [detailsToggled, setDetailsOpen] = useState<boolean | null>(null)
  const detailsOpen = detailsToggled ?? (target.kind === 'edit' || !hydrated || matchMedia('(min-width: 64rem)').matches)
  const schema = useMemo(() => logDraftSchema(options), [options])
  const [defaultValues] = useState(() => target.kind === 'edit' ? draftFromLog(options, target.log) : firstDraft(options, target.amount))
  const form = usePaperForm({
    defaultValues,
    validators: { onChange: schema, onSubmit: schema },
    onSubmit: async ({ value }) => {
      setFailure(null)
      if (target.kind === 'edit') {
        const result = await updateLog(toLogUpdate(value, options, viewerId, target.log))
        if (!result.ok) return setFailure(result.reason)
        return navigate({ to: '/logs/$logId', params: { logId: result.data.id }, replace: true })
      }

      const result = await saveLog(toNewLog(value, options, viewerId))
      if (!result.ok) return setFailure(result.reason)
      const { id, language, activity } = result.data
      rememberLastLogged(language.code, activity.id)
      // As in Webv2, contests are chosen in a separate step after saving.
      return options.registrations.length
        ? navigate({ to: '/logs/$logId/contests', params: { logId: id }, search: { preselect: 1 }, replace: true })
        : navigate({ to: '/logs/$logId', params: { logId: id }, replace: true })
    },
  })
  useRestoredChoices(form, options, target)
  const pending = useStore(form.store, state => state.isSubmitting)
  const languageCode = useStore(form.store, state => state.values.languageCode)
  const activityId = useStore(form.store, state => state.values.activityId)

  const units = unitsFor(options, languageCode, activityId).map(unit => ({ value: unit.unit_key, label: unit.name }))
  const disabled = !hydrated || pending
  const editing = target.kind === 'edit' ? target.log : null

  return (
    <div className="log-form-page">
      {editing ? (
        <PageHeader
          title="Edit log"
          lead={`${editing.language.name} · ${editing.activity.name}`}
          backLink={{ href: `/logs/${editing.id}`, label: editing.description || `${editing.language.name} ${editing.activity.name}` }}
        />
      ) : (
        <PageHeader title="Log activity" />
      )}
      <form.AppForm>
        <div className="min-w-0">
          <form.Form className="grid min-w-0 max-w-panel gap-4" aria-busy={pending}>
            <Surface className={card}>
              <fieldset className={`${formControls} p-record-inset`} disabled={disabled}>
                {/* An edit can't change language or activity; the header states them. */}
                {editing ? null : <ChoiceFields form={form} options={options} />}
                <MeasureField form={form} timed={isTimed(options, activityId)} units={units} />
                <details className="group/details" open={detailsOpen} onToggle={event => setDetailsOpen(event.currentTarget.open)}>
                  <summary className="inline-flex cursor-pointer list-none items-center gap-1.5 text-sm font-semibold text-muted [&::-webkit-details-marker]:hidden">
                    <ChevronRightIcon className={`size-2.5 transition-transform duration-quick ${detailsOpen ? 'rotate-quarter' : ''}`} aria-hidden="true" />
                    Details
                  </summary>
                  <div className={`${formControls} mt-4`}>
                  <form.AppField name="description">
                    {field => <field.TextArea label="Description (optional)" placeholder="e.g. One Piece volume 45" />}
                  </form.AppField>
                  <TagsField form={form} viewerId={viewerId} />
                  <div className="empty:hidden">
                    <TagsSidebar form={form} activityId={activityId} disabled={disabled} compact />
                  </div>
                  </div>
                </details>
              </fieldset>
              {failure ? <SaveFailure failure={failure} viewerId={viewerId} editing={editing} /> : null}
              <div className={cn(formFooter().root(), 'mx-record-inset mt-0 pb-record-inset')}>
                <div className={formFooter().actions()}>
                  {editing
                    ? <Link className={buttonClassName({ variant: 'ghost' })} to="/logs/$logId" params={{ logId: editing.id }}>Cancel</Link>
                    : <Link className={buttonClassName({ variant: 'ghost' })} to="/">Cancel</Link>}
                  <Button type="submit" disabled={disabled} loading={pending} loadingLabel="Saving">
                    {editing ? 'Save' : 'Log activity'}
                  </Button>
                </div>
                <ScoreEstimate form={form} options={options} />
              </div>
            </Surface>
          </form.Form>
        </div>
      </form.AppForm>
    </div>
  )
}

/**
 * Replaces a new draft's language and activity with the browser's last logged ones once, after hydration, so the
 * server and the first client render agree. Controls stay disabled until then, so no entry is overwritten.
 */
function useRestoredChoices(form: { reset: (values: LogDraft, options: { keepDefaultValues: true }) => void }, options: LogOptions, target: LogFormTarget) {
  const restored = useRef(false)
  useEffect(() => {
    if (restored.current || target.kind !== 'new') return
    restored.current = true
    const last = readLastLogged()
    if (last.languageCode !== null || last.activityId !== null) {
      form.reset(firstDraft(options, target.amount, last), { keepDefaultValues: true })
    }
  }, [form, options, target])
}

/** Tag chips with the viewer's tags containing the typed text, fetched once typing settles for 300 ms. */
const TagsField = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { viewerId: '' },
  render: function TagsField({ form, viewerId }) {
    const hydrated = useHydrated()
    const [query, setQuery] = useState('')
    const [settled, setSettled] = useState(query)
    useEffect(() => {
      const timer = setTimeout(() => setSettled(query.trim()), 300)
      return () => clearTimeout(timer)
    }, [query])
    const { data: suggestions = [] } = useQuery({ ...tagSuggestionsQuery(viewerId, settled), enabled: hydrated })
    const counts = new Map(suggestions.map(suggestion => [suggestion.tag, suggestion.count]))

    return (
      <form.AppField name="tags">
        {field => (
          <field.TagsInput
            label={`Tags (optional, up to ${maxTags})`}
            placeholder="Add tags..."
            maxSelections={maxTags}
            options={suggestions.map(suggestion => suggestion.tag)}
            onQueryChange={setQuery}
            formatOption={tag => (counts.get(tag) ?? 0) > 0 ? `${tag} (${counts.get(tag)}×)` : tag}
          />
        )}
      </form.AppField>
    )
  },
})

/** An amount in one of the activity's units, unless the activity is timed, and the time spent, which timed ones require. */
const MeasureField = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { timed: false, units: [] as Option[] },
  render: function MeasureField({ form, timed, units }) {
    const timeSpent = (
      <form.AppField name="minutes">
        {field => (
          <field.Input label={timed ? 'Time spent (minutes)' : 'Time spent (minutes, optional)'} required={timed} type="number" inputMode="decimal" step="any" min={0} />
        )}
      </form.AppField>
    )
    if (timed) return timeSpent

    // The amount field must be the innermost, because AmountWithUnit binds the field it is rendered in.
    return (
      <>
        <form.AppField name="unitKey">
          {unitField => (
            <form.AppField name="amount">
              {field => (
                <field.AmountWithUnit
                  label="Amount"
                  required
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
        {timeSpent}
      </>
    )
  },
})

/** A new log's language and activity. Changing either selects their first unit again. */
const ChoiceFields = withPaperForm({
  defaultValues: {} as LogDraft,
  props: { options: {} as LogOptions },
  render: function ChoiceFields({ form, options }) {
    const selectDependents = (language: string, activity: string) => {
      form.setFieldValue('unitKey', unitFor(options, language, activity))
    }

    return (
      <>
        <form.AppField
          name="languageCode"
          listeners={{ onChange: ({ value }) => selectDependents(value, form.getFieldValue('activityId')) }}
        >
          {field => <field.Select label="Language" placeholder="Choose a language" options={[]} groups={languageGroups(options)} />}
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


/**
 * The API's estimate of the draft's score, requested once the draft is valid and has stopped changing for 300 ms. An
 * incomplete draft, a pending first estimate and a failed one show an em dash, and a zero estimate shows N/A.
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
      <p className="log-form__estimate m-0 text-lead tabular-nums max-sm:basis-full" aria-live="polite">
        Estimated score: <strong className="data-pending:font-regular data-pending:text-muted" data-pending={score === undefined ? '' : undefined}>{score === undefined ? '—' : score ? scoreFormat.format(score) : 'N/A'}</strong>
      </p>
    )
  },
})

/** As in Webv2, the languages the viewer used before, then every language. */
function languageGroups({ configuration }: LogOptions): OptionGroup[] {
  const own = new Set(configuration.user_language_codes ?? [])
  const all = configuration.languages.map(language => ({ value: language.code, label: language.name }))
  const used = all.filter(language => own.has(language.value))

  return [...(used.length ? [{ label: 'Previously used', options: used }] : []), { label: 'All languages', options: all }]
}

const failureMessages: Record<LogFailure, string> = {
  'account-unavailable':
    'Account service unavailable. Nothing was saved. Your entries are kept here; try again when the service recovers.',
  switched:
    'You are now signed in as a different account. Nothing was saved. This draft is kept for the account that started it.',
  invalid: 'The activity was not saved. Check the amount, unit and tags, then try again.',
  'signed-out': 'Your session ended. Log in in a new tab, return here and save again. Your entries are kept here.',
  forbidden: 'Your account cannot save activity. Your entries are kept here.',
  deleting: 'Your account is being deleted. Activity cannot be saved.',
  unavailable:
    'We could not confirm whether your activity was saved. Check your activity before saving again to avoid a duplicate.',
  'not-found': 'The activity could not be saved. Review the language and activity, then try again.',
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
    ? <a className="text-link hover:text-link-hover" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
    : null
}

/** The way forward after a failure, where there is one. */
function FailureAction({ failure, viewerId }: { failure: LogFailure; viewerId: string }) {
  const loginUrl = useLoginUrl()
  const links = useLinks()

  switch (failure) {
    case 'switched':
      return <a className="text-link hover:text-link-hover" href="/logs/new">Start a new activity for this account</a>
    case 'signed-out':
      return <a className="text-link hover:text-link-hover" href={loginUrl} target="_blank" rel="noopener noreferrer">Log in in a new tab</a>
    case 'unavailable':
      return (
        <a
          className="text-link hover:text-link-hover"
          href={links.profileUpdates(viewerId)}
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
