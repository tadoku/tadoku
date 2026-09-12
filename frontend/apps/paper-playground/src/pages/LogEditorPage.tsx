import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { FormProvider, useController, useForm, useWatch } from 'react-hook-form'
import { CheckIcon } from 'paper-ui/icons'
import { Button, Checkbox, Flash, Input, Select, TagsInput, TextArea, buttonClassName, useToast } from 'paper-ui'
import { compatibleModifiers, contestStatus, formatDate, formatNumber, modifierRates, sampleToday, scoreLog, scoreSubmission, supportedLanguages, type Activity, type LogModifier, type LogUnit, type SampleLog } from '../data'
import { usePlayground } from '../state'
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
  if (logId && !existing) return <section className="empty-state"><h1 className="paper-type-page">Log not found</h1><Link className="text-link" to="/users/anton/activity">View your activity</Link></section>
  if (existing && existing.userId !== app.userId) return <section className="empty-state"><h1 className="paper-type-page">This is someone else’s record</h1><Link className="text-link" to={`/logs/${existing.id}`}>View the log</Link></section>
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
  const methods = useForm<LogFields>({ defaultValues: {
    title: existing?.title || '', language: existing?.language || 'Japanese', activity: existing?.activity || 'Reading',
    amount: existing?.amount ?? '', unit: existing?.unit || 'pages', modifiers: existing?.modifiers || [],
    minutes: existing?.minutes ?? '', date: existing?.date || today, note: existing?.note || '', tags: existing?.tags || [],
    media: existing?.media || 'Book', submissions: existing?.submissions.filter(s => !closedSubmissions.some(closed => closed.contestId === s.contestId)).map(s => s.contestId) || [params.get('contest')].filter((id): id is string => Boolean(id)),
    expandMetadata: metadataPreference(),
  } })
  const values = useWatch({ control: methods.control })
  const activity = existing?.activity ?? values.activity ?? 'Reading'
  const language = existing?.language ?? values.language ?? 'Japanese'
  const unit: LogUnit = activity === 'Listening' ? 'minutes' : values.unit === 'minutes' ? 'pages' : values.unit ?? 'pages'
  const amount = Number(values.amount)
  const minutes = Number(values.minutes)
  const modifiers = compatibleModifiers(activity, unit, values.modifiers || [])
  const { field: modifierField } = useController({ control: methods.control, name: 'modifiers' })
  const [step, setStep] = useState(1)
  const [visibleLogs, setVisibleLogs] = useState(5)
  const [origin, setOrigin] = useState<SampleLog | null>(null)
  const [beforePrefill, setBeforePrefill] = useState<PrefillSnapshot | null>(null)
  const [metadataOpen, setMetadataOpen] = useState(metadataPreference)
  const recentDetails = useRef<HTMLDetailsElement>(null)
  const recentList = useRef<HTMLDivElement>(null)
  const firstAddedLog = useRef<number | null>(null)
  const stepHeading = useRef<HTMLParagraphElement>(null)
  const initialStep = useRef(true)
  const recentLogs = app.logs.filter(log => log.userId === app.userId).sort((a, b) => b.date.localeCompare(a.date))
  const eligible = app.contests.filter(c => app.joinedContests.includes(c.id) && contestStatus(c, today) === 'live' && (values.date || '') >= c.start && (values.date || '') <= c.end && c.activities.includes(activity) && (c.languages.includes('All languages') || c.languages.includes(language)) && (!app.registrations[c.id] || app.registrations[c.id].includes(language)) && (c.id !== 'reading-circle' || minutes > 0))

  useEffect(() => {
    if (values.unit !== unit) methods.setValue('unit', unit)
    const current = methods.getValues('modifiers')
    const compatible = compatibleModifiers(activity, unit, current)
    if (compatible.join('|') !== current.join('|')) methods.setValue('modifiers', compatible)
  }, [activity, unit, values.unit, methods])

  useEffect(() => {
    try { localStorage.setItem(metadataPreferenceKey, String(values.expandMetadata)) } catch { /* The form still works when browser storage is unavailable. */ }
  }, [values.expandMetadata])

  useEffect(() => {
    if (firstAddedLog.current !== null) {
      recentList.current?.querySelectorAll<HTMLButtonElement>('button')[firstAddedLog.current]?.focus()
      firstAddedLog.current = null
    }
  }, [visibleLogs])

  useEffect(() => {
    if (initialStep.current) { initialStep.current = false; return }
    if (step === 1) methods.setFocus('amount')
    else stepHeading.current?.focus()
  }, [step, methods])

  function applyRecent(log: SampleLog) {
    const previous = structuredClone(methods.getValues())
    setBeforePrefill({ values: previous, origin, metadataOpen, previous: beforePrefill })
    methods.reset({ ...previous, title: log.title, language: log.language, activity: log.activity,
      unit: log.unit, modifiers: compatibleModifiers(log.activity, log.unit, log.modifiers || []),
      tags: [...log.tags], media: log.media, amount: '', minutes: '',
    })
    setOrigin(log)
    if (recentDetails.current) recentDetails.current.open = false
    requestAnimationFrame(() => methods.setFocus('amount'))
  }

  function undoRecent() {
    if (!beforePrefill) return
    methods.reset(beforePrefill.values)
    setOrigin(beforePrefill.origin)
    setMetadataOpen(beforePrefill.metadataOpen)
    setBeforePrefill(beforePrefill.previous)
    requestAnimationFrame(() => methods.setFocus('amount'))
  }

  function saveLog(fields: LogFields) {
    if (step === 1) { setStep(2); return }
    const id = existing?.id || `local-${crypto.randomUUID()}`
    const scoring = { activity, language, amount: Number(fields.amount), unit, modifiers, minutes: activity === 'Reading' && fields.minutes !== '' ? Number(fields.minutes) : undefined }
    const submissions = [...closedSubmissions, ...fields.submissions.filter(id => eligible.some(c => c.id === id)).map(contestId => scoreSubmission(scoring, contestId))]
    const log: SampleLog = { id, userId: app.userId, title: fields.title.trim(), ...scoring, date: fields.date, note: fields.note, tags: fields.tags, media: fields.media, submissions }
    app.saveLog(log)
    toast.add({ title: existing ? 'Activity updated' : 'Activity saved', description: 'Your personal record is up to date.' })
    navigate(`/logs/${id}?asOf=${today}`)
  }

  const quantityLabel = activity === 'Listening' ? 'Minutes listened' : unit === 'sentences' ? 'Sentences read' : unit === 'characters' ? 'Characters read' : 'Pages read'
  const estimate = Number.isFinite(amount) && amount > 0 ? `${formatNumber(scoreLog({ activity, language, amount, unit, modifiers }))} points` : 'Enter an amount'

  return <section className="log-editor">
    <header className="page-header log-editor__header">
      <Link className="text-link page-back-link" to={detailHref}>{existing ? 'Back to log' : 'Your activity log'}</Link>
      <h1 className="paper-type-page">{existing ? 'Edit activity' : 'Log activity'}</h1>
      <p className="log-editor__step" ref={stepHeading} tabIndex={-1}>{step === 1 ? 'Log details' : 'Choose contests'} · {step} of 2</p>
    </header>
    <FormProvider {...methods}><form className="log-editor__form" noValidate onSubmit={methods.handleSubmit(saveLog, errors => {
      const field = Object.keys(errors)[0] as keyof LogFields
      setStep(field === 'date' ? 2 : 1)
      if (['minutes', 'tags', 'note', 'media'].includes(field)) setMetadataOpen(true)
      requestAnimationFrame(() => methods.setFocus(field))
    })}>
      <div className="log-editor__step-content" hidden={step !== 1}>
        {!existing && <details className="log-editor__recent" ref={recentDetails}>
          <summary>Use a recent log</summary>
          <p className="log-editor__hint">Copy its details and enter a new amount.</p>
          <div className="log-editor__recent-list" ref={recentList}>
            {recentLogs.slice(0, visibleLogs).map(log => <div className="log-editor__recent-row" key={log.id}>
              <div className="log-editor__recent-copy"><strong>{logTitle(log)}</strong><span>{log.language} · {log.activity} · {log.unit}</span><time dateTime={log.date}>{formatDate(log.date)}</time></div>
              <Button variant="ghost" onClick={() => applyRecent(log)}>Use this log</Button>
            </div>)}
            {!recentLogs.length && <p className="log-editor__hint">Your recent entries will appear here after you save your first log.</p>}
          </div>
          {recentLogs.length > 0 && <div className="log-editor__recent-footer"><span role="status">{Math.min(visibleLogs, recentLogs.length)} recent entries shown{visibleLogs >= recentLogs.length ? ' · End of history' : ''}</span><Button variant="ghost" disabled={visibleLogs >= recentLogs.length} onClick={() => { firstAddedLog.current = visibleLogs; setVisibleLogs(count => count + 5) }}>{visibleLogs >= recentLogs.length ? 'All logs shown' : 'Show more'}</Button></div>}
        </details>}
        {origin && <div className="log-editor__reuse-notice" role="status"><div><strong>Using {logTitle(origin)}</strong><p>Details copied. Enter the amount for this new log.</p></div>{beforePrefill && <Button variant="ghost" onClick={undoRecent}>Undo</Button>}</div>}
        <fieldset className="log-editor__section"><legend>What are you logging?</legend><div className="log-editor__fields">
          <div className="log-editor__row"><Select name="language" label="Language" disabled={Boolean(existing)} value={existing?.language} required={!existing} options={supportedLanguages.map(value => ({ value, label: value }))} /><Select name="activity" label="Activity" disabled={Boolean(existing)} value={existing?.activity} options={[{ value: 'Reading', label: 'Reading' }, { value: 'Listening', label: 'Listening' }]} /></div>
          {existing && <p className="log-editor__hint">Language and activity stay fixed for existing entries.</p>}
          <Input name="title" label="What did you read or listen to?" hint="Optional. A book, episode, article, or anything you enjoyed." maxLength={180} />
        </div></fieldset>
        <fieldset className="log-editor__section"><legend>Scoring</legend><div className="log-editor__fields">
          <div className="log-editor__amount"><Input name="amount" label={quantityLabel} type="number" min="0.01" step="any" placeholder="Enter today’s amount" required rules={{ valueAsNumber: true, min: { value: .01, message: 'Enter an amount greater than zero.' }, max: { value: 1000000, message: 'Check this amount.' } }} />{activity === 'Reading' && <Select name="unit" label="Unit" options={['pages', 'sentences', 'characters'].map(value => ({ value, label: value }))} />}</div>
          {origin && <p className="log-editor__hint">Previous entry: {formatNumber(origin.amount)} {origin.unit}.</p>}
          {activity === 'Reading' && unit === 'pages' ? <fieldset className="log-editor__modifier-group"><legend>Score modifiers</legend><div className="log-editor__modifiers">{readingModifiers.map(modifier => <Button key={modifier} variant={modifiers.includes(modifier) ? 'outline' : 'ghost'} aria-pressed={modifiers.includes(modifier)} leadingIcon={modifiers.includes(modifier) ? <CheckIcon className="paper-icon-default" aria-hidden="true" /> : undefined} onClick={() => modifierField.onChange(modifiers.includes(modifier) ? [] : [modifier])}>{modifier} <span>×{modifierRates[modifier]}</span></Button>)}</div></fieldset> : activity === 'Listening' ? <Checkbox name="modifiers" value="Passive listening" label={`Passive listening ×${modifierRates['Passive listening']}`} /> : <p className="log-editor__hint">No score modifiers for this unit.</p>}
          <div className="log-editor__score" aria-live="polite"><span>Estimated score</span><strong>{estimate}</strong></div>
          <Link className="text-link log-editor__scoring-link" to="/guide/scoring">How scoring works</Link>
        </div></fieldset>
        <details className="log-editor__metadata" open={metadataOpen} onToggle={event => setMetadataOpen(event.currentTarget.open)}>
          <summary>Metadata{values.tags?.length ? <span> · {values.tags.length} {values.tags.length === 1 ? 'tag' : 'tags'}</span> : null}</summary>
          <div className="log-editor__fields">
            <TagsInput name="tags" label="Tags" options={[...new Set(['fiction', 'bookclub', 'daily', 'podcast', ...app.logs.filter(log => log.userId === app.userId).flatMap(log => log.tags)])]} hint="Optional. Choose tags to find this entry again." />
            <div className="log-editor__row">{activity === 'Reading' && <Input name="minutes" label="Time spent reading" type="number" min="0" step="1" hint="Optional, in minutes." rules={{ validate: value => value === '' || Number.isFinite(Number(value)) && Number(value) >= 0 || 'Use zero or a positive duration.' }} />}<Select name="media" label="Medium" options={['Book', 'Article', 'Manga', 'Game', 'Podcast', 'Video', 'Audiobook', 'Conversation'].map(value => ({ value, label: value }))} /></div>
            <TextArea name="note" label="Notes" rows={3} hint="Optional. Keep a thought, a new word, or where you left off." />
            <Checkbox name="expandMetadata" label="Expand section by default" />
          </div>
        </details>
      </div>
      <div className="log-editor__step-content" hidden={step !== 2}>
        <div className="log-editor__review"><strong>{values.title?.trim() || `${language} ${activity.toLocaleLowerCase()}`}</strong><p>{language} · {activity} · {formatNumber(amount)} {unit}</p><span>{estimate}</span></div>
        <div className="log-editor__section"><Input name="date" label="Activity date" type="date" max={today} required rules={{ validate: value => String(value) <= today || 'Choose today or an earlier date.' }} hint="The day you read or listened." /></div>
        <fieldset className="log-editor__section"><legend>Choose contests</legend><p className="log-editor__hint">Choose where this entry counts. It always stays in your personal record.</p><div className="log-editor__contests">{eligible.length ? eligible.map(contest => <Checkbox key={contest.id} name="submissions" value={contest.id} label={contest.title} />) : <p className="log-editor__hint">No joined contests accept this language, activity, and date. You can still save your personal record.</p>}</div>
          {closedSubmissions.length > 0 && <Flash title="Finished contests keep their saved scores">Editing this personal record leaves its finished-contest snapshots unchanged.</Flash>}
        </fieldset>
      </div>
      <div className="log-editor__actions">{step === 1 ? <><Link to={detailHref} className={buttonClassName({ variant: 'ghost' })}>Cancel</Link><Button type="submit">Next</Button></> : <><Button variant="ghost" onClick={() => setStep(1)}>Back</Button><Button type="submit">{existing ? 'Save changes' : 'Save activity'}</Button></>}</div>
    </form></FormProvider>
  </section>
}
