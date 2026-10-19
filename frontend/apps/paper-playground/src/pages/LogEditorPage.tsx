import { useStore } from '@tanstack/react-form'
import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { usePaperForm, Button, EmptyState, Flash, PageHeader, buttonClassName, useToast } from 'paper-ui'
import { ChevronDownIcon } from 'paper-ui/icons'
import { compatibleModifiers, contestStatus, formatNumber, modifierRates, personalScoreExplanation, sampleToday, scoreLog, scoreSubmission, supportedLanguages, type Activity, type LogModifier, type LogUnit, type SampleLog } from '../data'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import { RecentLogPicker } from './RecentLogPicker'
import './log-editor.css'

type LogFields = {
  title: string; language: string; activity: Activity; amount: number | string
  unit: LogUnit; modifiers: LogModifier[]; minutes: number | string; date: string
  note: string; tags: string[]; media: string; submissions: string[]; expandMetadata: boolean
}
type PrefillSnapshot = { values: LogFields; origin: SampleLog | null; metadataOpen: boolean; previous: PrefillSnapshot | null }
const metadataPreferenceKey = 'tadoku-log-metadata-expanded'
const readingModifiers: LogModifier[] = ['Manga', 'Comic', 'Two column']

function metadataPreference() {
  try { return localStorage.getItem(metadataPreferenceKey) === 'true' } catch { return false }
}

function logTitle(log: SampleLog) { return log.title || `${log.language} ${log.activity.toLocaleLowerCase()}` }

export function LogEditorPage() {
  const { logId } = useParams()
  const app = usePlayground()
  const existing = app.logs.find(log => log.id === logId)
  if (app.viewer === 'guest') return <Navigate replace to={`/sign-in?next=${encodeURIComponent(logId ? `/logs/${logId}/edit` : '/logs/new')}`} />
  if (logId && !existing) return <EmptyState headingLevel={1} title="Log not found" actions={<Link className="text-link" to="/users/anton/activity">View your activity</Link>} />
  if (existing && existing.userId !== app.userId) return <EmptyState headingLevel={1} title="This is someone else’s record" actions={<Link className="text-link" to={`/logs/${existing.id}`}>View the log</Link>} />
  return <LogForm key={logId || 'new'} existing={existing} />
}

function LogForm({ existing }: { existing?: SampleLog }) {
  const app = usePlayground()
  const navigate = useNavigate()
  const toast = useToast()
  const [params] = useSearchParams()
  const requestedDate = params.get('asOf')
  const contextDate = requestedDate && /^\d{4}-\d{2}-\d{2}$/.test(requestedDate) ? requestedDate : existing?.submissions.some(s => s.closed) ? '2026-10-03' : sampleToday
  const today = existing && existing.date > contextDate ? existing.date : contextDate
  const detailHref = existing ? `/logs/${existing.id}?asOf=${today}` : '/users/anton/activity'
  const closedSubmissions = (existing?.submissions || []).filter(submission => {
    const contest = app.contests.find(c => c.id === submission.contestId)
    return submission.closed || contest && contestStatus(contest, today) === 'ended'
  }).map(submission => ({ ...submission, closed: true }))
  const methods = usePaperForm({ defaultValues: ({
    title: existing?.title || '', language: existing?.language || 'Japanese', activity: existing?.activity || 'Reading',
    amount: existing?.amount ?? '', unit: existing?.unit || 'pages', modifiers: existing?.modifiers || [],
    minutes: existing?.minutes ?? '', date: existing?.date || today, note: existing?.note || '', tags: existing?.tags || [],
    media: existing?.media || '', submissions: existing?.submissions.filter(s => !closedSubmissions.some(closed => closed.contestId === s.contestId)).map(s => s.contestId) || [params.get('contest')].filter((id): id is string => Boolean(id)),
    expandMetadata: metadataPreference(),
  } as LogFields), onSubmit: ({ value }) => saveLog(value), onSubmitInvalid: ({ formApi }) => {
    const field = Object.entries(formApi.state.fieldMeta).find(([, meta]) => meta?.errors.length)?.[0]
    if (!field) return
    setStep(field === 'date' ? 2 : 1)
    if (['minutes', 'tags', 'note'].includes(field)) setMetadataOpen(true)
    requestAnimationFrame(() => document.getElementsByName(field)[0]?.focus())
  } })
  const values = useStore(methods.store, state => state.values)
  const activity = existing?.activity ?? values.activity ?? 'Reading'
  const language = existing?.language ?? values.language ?? 'Japanese'
  const unit: LogUnit = activity === 'Listening' ? 'minutes' : values.unit === 'minutes' ? 'pages' : values.unit ?? 'pages'
  const amount = Number(values.amount)
  const minutes = Number(values.minutes)
  const modifiers = compatibleModifiers(activity, unit, values.modifiers || [])
  const [step, setStep] = useState(1)
  const [origin, setOrigin] = useState<SampleLog | null>(null)
  const [beforePrefill, setBeforePrefill] = useState<PrefillSnapshot | null>(null)
  const [metadataOpen, setMetadataOpen] = useState(metadataPreference)
  const stepHeading = useRef<HTMLParagraphElement>(null)
  const initialStep = useRef(true)
  const eligible = app.contests.filter(c => app.joinedContests.includes(c.id) && contestStatus(c, today) === 'live' && (values.date || '') >= c.start && (values.date || '') <= c.end && c.activities.includes(activity) && (c.languages.includes('All languages') || c.languages.includes(language)) && (!app.registrations[c.id] || app.registrations[c.id].includes(language)) && (c.id !== 'reading-circle' || minutes > 0))

  useEffect(() => {
    if (values.unit !== unit) methods.setFieldValue('unit', unit)
    const current = methods.getFieldValue('modifiers')
    const compatible = compatibleModifiers(activity, unit, current)
    if (compatible.join('|') !== current.join('|')) methods.setFieldValue('modifiers', compatible)
  }, [activity, unit, values.unit, methods])

  useEffect(() => {
    try { localStorage.setItem(metadataPreferenceKey, String(values.expandMetadata)) } catch { /* The form still works when browser storage is unavailable. */ }
  }, [values.expandMetadata])

  useEffect(() => {
    if (initialStep.current) { initialStep.current = false; return }
    if (step === 1) document.getElementsByName('amount')[0]?.focus()
    else stepHeading.current?.focus()
  }, [step, methods])

  function applyRecent(log: SampleLog) {
    const previous = structuredClone(methods.state.values)
    setBeforePrefill({ values: previous, origin, metadataOpen, previous: beforePrefill })
    methods.reset({ ...previous, title: log.title, language: log.language, activity: log.activity,
      unit: log.unit, modifiers: compatibleModifiers(log.activity, log.unit, log.modifiers || []),
      tags: [...log.tags], media: log.media, amount: '', minutes: '',
    }, { keepDefaultValues: true })
    setOrigin(log)
    requestAnimationFrame(() => document.getElementsByName('amount')[0]?.focus())
  }

  function undoRecent() {
    if (!beforePrefill) return
    methods.reset(beforePrefill.values, { keepDefaultValues: true })
    setOrigin(beforePrefill.origin)
    setMetadataOpen(beforePrefill.metadataOpen)
    setBeforePrefill(beforePrefill.previous)
    requestAnimationFrame(() => document.getElementsByName('amount')[0]?.focus())
  }

  function saveLog(fields: LogFields) {
    if (step === 1) { setStep(2); return }
    const id = existing?.id || `local-${crypto.randomUUID()}`
    const scoring = { activity, language, amount: Number(fields.amount), unit, modifiers, minutes: activity === 'Reading' && fields.minutes !== '' && Number.isFinite(Number(fields.minutes)) ? Number(fields.minutes) : undefined }
    const submissions = [...closedSubmissions, ...fields.submissions.filter(id => eligible.some(c => c.id === id)).map(contestId => scoreSubmission(scoring, contestId))]
    const log: SampleLog = { id, userId: app.userId, title: fields.title.trim(), ...scoring, date: fields.date, note: fields.note, tags: fields.tags, media: fields.media, submissions }
    app.saveLog(log)
    toast.add({ title: existing ? 'Activity updated' : 'Activity saved', description: 'Your personal record is up to date.' })
    navigate(`/logs/${id}?asOf=${today}`)
  }

  const quantityLabel = activity === 'Listening' ? 'Minutes listened' : unit === 'sentences' ? 'Sentences read' : unit === 'characters' ? 'Characters read' : 'Pages read'
  const estimate = Number.isFinite(amount) && amount > 0 ? `${formatNumber(scoreLog({ activity, language, amount, unit, modifiers }))} points` : 'Enter an amount'

  return <section className="log-editor">
    <PageHeader backLink={{ href: detailHref, label: existing ? 'Back to log' : 'Your activity log' }} renderLink={renderLink} title={existing ? 'Edit activity' : 'Log activity'}>
      <p className="log-editor__step" ref={stepHeading} tabIndex={-1}>{step === 1 ? 'Log details' : 'Choose contests'} · {step} of 2</p>
    </PageHeader>
    <div className="log-editor__layout">
    <methods.AppForm><methods.Form className="log-editor__form">
      <div className="log-editor__step-content" hidden={step !== 1}>
        {!existing && <RecentLogPicker logs={app.logs.filter(log => log.userId === app.userId)} onSelect={applyRecent} />}
        {origin && <div className="log-editor__reuse-notice" role="status"><div><strong>Using {logTitle(origin)}</strong><p>Details copied. Enter the amount for this new log.</p></div>{beforePrefill && <Button variant="ghost" onClick={undoRecent}>Undo</Button>}</div>}
        <fieldset className="log-editor__section"><legend>What are you logging?</legend><div className="log-editor__fields">
          <div className="log-editor__row"><methods.AppField name="language" validators={existing ? {} : { onChange: ({ value }) => { if (!value) return 'Choose an option.'; } }}>{field => <field.Select label="Language" disabled={Boolean(existing)} required={!existing} options={supportedLanguages.map(value => ({ value, label: value }))} />}</methods.AppField><methods.AppField name="activity">{field => <field.Select label="Activity" disabled={Boolean(existing)} options={[{ value: 'Reading', label: 'Reading' }, { value: 'Listening', label: 'Listening' }]} />}</methods.AppField></div>
          {existing && <p className="log-editor__hint">Language and activity stay fixed for existing entries.</p>}
          <methods.AppField name="title">{field => <field.Input label="What did you read or listen to?" hint="Optional. A book, episode, article, or anything you enjoyed." maxLength={180} />}</methods.AppField>
        </div></fieldset>
        <fieldset className="log-editor__section"><legend>Scoring</legend><div className="log-editor__fields">
          <div className="log-editor__amount"><methods.AppField name="amount" validators={{ onChange: ({ value }) => { if (String(value) === '' || !Number.isFinite(Number(value))) return "This field is required."; if (Number(value) < .01) return 'Enter an amount greater than zero.'; if (Number(value) > 1000000) return 'Check this amount.'; } }}>{field => <field.Input label={quantityLabel} type="number" min="0.01" step="any" placeholder="0" required />}</methods.AppField>{activity === 'Reading' && <methods.AppField name="unit">{field => <field.Select label="Unit" options={['pages', 'sentences', 'characters'].map(value => ({ value, label: value }))} />}</methods.AppField>}</div>
          {origin && <p className="log-editor__hint">Previous entry: {formatNumber(origin.amount)} {origin.unit}.</p>}
          {activity === 'Reading' && unit !== 'pages' ? <p className="log-editor__hint">No score modifiers for this unit.</p> : <methods.AppField name="modifiers">{field => <field.ToggleSelect label="Score modifiers" options={(activity === 'Reading' ? readingModifiers : ['Passive listening'] as LogModifier[]).map(value => ({ value, label: value, description: `×${modifierRates[value]}` }))} />}</methods.AppField>}
          <div className="log-editor__score" aria-live="polite"><span>Estimated score</span><strong data-empty={!Number.isFinite(amount) || amount <= 0 || undefined}>{estimate}</strong></div>
          <Link className="text-link log-editor__scoring-link" to="/guide/scoring">How scoring works</Link>
        </div></fieldset>
        <details className="log-editor__disclosure log-editor__metadata" open={metadataOpen} onToggle={event => setMetadataOpen(event.currentTarget.open)}>
          <summary><span>Metadata{values.tags?.length ? <span className="log-editor__metadata-count"> · {values.tags.length} {values.tags.length === 1 ? 'tag' : 'tags'}</span> : null}</span><ChevronDownIcon className="paper-icon-compact" aria-hidden="true" /></summary>
          <div className="log-editor__fields">
            <methods.AppField name="tags">{field => <field.TagsInput label="Tags" options={[...new Set(['fiction', 'bookclub', 'daily', 'podcast', ...app.logs.filter(log => log.userId === app.userId).flatMap(log => log.tags)])]} placeholder="Type a tag…" hint="Type a tag and press Enter to add it." />}</methods.AppField>
            {activity === 'Reading' && <methods.AppField name="minutes" validators={{ onChange: ({ value }) => { const result = (value => value === '' || Number.isNaN(value) || Number.isFinite(Number(value)) && Number(value) >= 0 || 'Use zero or a positive duration.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Time spent reading" type="number" min="0" step="1" hint="Optional, in minutes." />}</methods.AppField>}
            <methods.AppField name="note">{field => <field.TextArea label="Notes" rows={3} hint="Optional. Keep a thought, a new word, or where you left off." />}</methods.AppField>
            <methods.AppField name="expandMetadata">{field => <field.Checkbox label="Expand section by default" />}</methods.AppField>
          </div>
        </details>
      </div>
      <div className="log-editor__step-content" hidden={step !== 2}>
        <div className="log-editor__review"><strong>{values.title?.trim() || `${language} ${activity.toLocaleLowerCase()}`}</strong><p>{language} · {activity} · {formatNumber(amount)} {unit}</p><span>{estimate}</span></div>
        <div className="log-editor__section"><methods.AppField name="date" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => String(value) <= today || 'Choose today or an earlier date.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label="Activity date" type="date" max={today} required hint="The day you read or listened." />}</methods.AppField></div>
        <fieldset className="log-editor__section"><legend>Choose contests</legend><p className="log-editor__hint">Choose where this entry counts. It always stays in your personal record.</p><div className="log-editor__contests">{eligible.length ? eligible.map(contest => <methods.AppField key={contest.id} name="submissions">{field => <field.Checkbox  value={contest.id} label={contest.title} />}</methods.AppField>) : <p className="log-editor__hint">No joined contests accept this language, activity, and date. You can still save your personal record.</p>}</div>
          {closedSubmissions.length > 0 && <Flash title="Finished contests keep their saved scores">Editing this personal record leaves its finished-contest snapshots unchanged.</Flash>}
        </fieldset>
      </div>
      <div className="log-editor__actions">{step === 1 ? <><Link to={detailHref} className={buttonClassName({ variant: 'ghost' })}>Cancel</Link><Button type="submit">Next</Button></> : <><Button variant="ghost" onClick={() => setStep(1)}>Back</Button><Button type="submit">{existing ? 'Save changes' : 'Save activity'}</Button></>}</div>
    </methods.Form></methods.AppForm>
    <aside className="log-editor__aside" aria-labelledby="log-score-heading">
      <h2 id="log-score-heading">Estimated score</h2>
      <p className="log-editor__aside-score" data-empty={!Number.isFinite(amount) || amount <= 0 || undefined} aria-live="polite">{estimate}</p>
      {Number.isFinite(amount) && amount > 0 && <p className="log-editor__aside-basis">{personalScoreExplanation({ activity, language, amount, unit, modifiers })}</p>}
      <p>Contest scores may use a different rule.</p>
      <Link className="text-link" to="/guide/scoring">How scoring works</Link>
    </aside>
    </div>
  </section>
}
