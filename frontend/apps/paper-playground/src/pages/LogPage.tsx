import { useEffect, useRef, useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { Breadcrumb, Button, Checkbox, Flash, Modal, Surface, buttonClassName, type NavigationLinkProps } from 'paper-ui'
import { PencilSquareIcon, TrashIcon } from 'paper-ui/icons'
import { contestStatus, formatDate, formatNumber, sampleToday, scoreLog, scoreSubmission, users, type SampleContest, type SampleLog, type Submission } from '../data'
import { usePlayground, useScenario } from '../state'
import './records.css'

const renderLink = ({ href, ...props }: NavigationLinkProps) => <Link className="text-link" to={href} {...props} />
const scenarioLogs: Record<string, string> = { reading: 'reading-konbini', listening: 'listening-teppei', visitor: 'reading-mei', personal: 'personal-reading', ended: 'ended-reading' }

export function submissionEligibility(log: SampleLog, contest: SampleContest, joined: readonly string[], registrations: Record<string, string[]> = {}, asOf = sampleToday) {
  if (log.submissions.some(submission => submission.contestId === contest.id && submission.closed)) return 'This contest’s saved contribution is closed to changes.'
  if (!joined.includes(contest.id)) return 'Join this contest before submitting activity.'
  if (contestStatus(contest, asOf) === 'ended') return 'The submission window has closed.'
  if (contestStatus(contest, asOf) === 'upcoming') return 'This contest has not started yet.'
  if (log.date.slice(0, 10) < contest.start || log.date.slice(0, 10) > contest.end) return 'This activity is outside the contest dates.'
  if (!contest.languages.includes('All languages') && !contest.languages.includes(log.language)) return `${log.language} is not an allowed language.`
  if (registrations[contest.id] && !registrations[contest.id].includes(log.language)) return `You have not registered ${log.language} for this contest.`
  if (!contest.activities.includes(log.activity)) return `${log.activity} is not an allowed activity.`
  if (contest.id === 'reading-circle' && (!Number.isFinite(log.minutes) || !log.minutes || log.minutes <= 0)) return 'Record the time spent reading; this contest scores tracked minutes.'
  return undefined
}

export function updateSubmissions(log: SampleLog, selected: readonly string[], contests: readonly SampleContest[], joined: readonly string[], registrations: Record<string, string[]> = {}, asOf = sampleToday): Submission[] {
  const unchanged = log.submissions.filter(submission => {
    const contest = contests.find(item => item.id === submission.contestId)
    return !contest || submissionEligibility(log, contest, joined, registrations, asOf) !== undefined
  })
  const editable = contests.filter(contest => selected.includes(contest.id) && !submissionEligibility(log, contest, joined, registrations, asOf)).map(contest => {
    const saved = log.submissions.find(submission => submission.contestId === contest.id)
    if (saved) return saved
    return scoreSubmission(log, contest.id)
  })
  return [...unchanged, ...editable]
}

function SubmissionManager({ log, asOf, onSaved }: { log: SampleLog; asOf: string; onSaved: () => void }) {
  const { contests, joinedContests, registrations, saveLog } = usePlayground()
  const [open, setOpen] = useState(false)
  const defaults = () => ({ selected: Object.fromEntries(log.submissions.filter(submission => !submission.closed).map(submission => [submission.contestId, true])) })
  const methods = useForm<{ selected: Record<string, boolean> }>({ defaultValues: defaults() })
  const choices = contests.filter(contest => joinedContests.includes(contest.id) || log.submissions.some(submission => submission.contestId === contest.id))
  const available = choices.filter(contest => !submissionEligibility(log, contest, joinedContests, registrations, asOf))
  return <Modal title="Contest submissions" description="Choose where this activity contributes. Each contest keeps its own score." trigger={<Button variant="outline">{log.submissions.length ? 'Manage submissions' : 'Submit to a contest'}</Button>} open={open} onOpenChange={next => { if (next) methods.reset(defaults()); setOpen(next) }} footer={null}>
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(values => { saveLog({ ...log, submissions: updateSubmissions(log, Object.entries(values.selected).filter(([, selected]) => selected).map(([id]) => id), contests, joinedContests, registrations, asOf) }); setOpen(false); onSaved() })}>
        {choices.length ? <div className="records-submission-choices">{choices.map(contest => {
          const reason = submissionEligibility(log, contest, joinedContests, registrations, asOf)
          return <div key={contest.id}><Checkbox name={`selected.${contest.id}`} label={contest.title} disabled={Boolean(reason)} hint={reason ?? `${formatDate(contest.start)} – ${formatDate(contest.end)} · Eligible`} /></div>
        })}</div> : null}
        {!available.length ? <div className="records-submission-empty"><h3 className="paper-type-component">No eligible contests</h3><p>Check the activity date, language, and your registrations. This activity already counts in your personal history.</p><Link className="text-link" to={`/contests/official?asOf=${asOf}`} onClick={() => setOpen(false)}>Find a contest to join</Link></div> : <p className="muted text-sm">Removing a selection removes that contest contribution. Your personal record stays the same.</p>}
        <div className="app-form__actions"><Button type="submit" disabled={!available.length}>Save submissions</Button><Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button></div>
      </form>
    </FormProvider>
  </Modal>
}

function ScoreExplanation({ basis }: { basis: string }) {
  return <details className="records-score-explanation"><summary>How this was scored</summary><p>{basis}</p><p className="muted">Illustrative saved scoring rule.</p></details>
}

export function LogPage() {
  const { logId } = useParams()
  const [params] = useSearchParams()
  const [scenario, setScenario] = useScenario('log')
  const id = params.has('scenario') ? scenarioLogs[scenario] ?? logId : logId
  return <LogRecord key={`${id}-${scenario}`} id={id} scenario={scenario} onRetry={() => setScenario('')} />
}

function LogRecord({ id, scenario, onRetry }: { id?: string; scenario: string; onRetry: () => void }) {
  const { logs, contests, viewer, userId, user, removeLog, restoreLog } = usePlayground()
  const [params] = useSearchParams()
  const log = logs.find(item => item.id === id)
  const [deleted, setDeleted] = useState<SampleLog | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [saved, setSaved] = useState(false)
  const resultRef = useRef<HTMLHeadingElement>(null)
  const cancelRef = useRef<HTMLButtonElement>(null)
  useEffect(() => { if (deleted) resultRef.current?.focus() }, [deleted])
  const owner = log?.userId === userId && viewer !== 'guest'
  const author = users.find(item => item.id === log?.userId)
  const name = author?.id === userId && user ? user.name : author?.name ?? 'This member'
  const requestedDate = params.get('asOf')
  const contextDate = requestedDate && /^\d{4}-\d{2}-\d{2}$/.test(requestedDate) ? requestedDate : log?.submissions.some(submission => submission.closed) ? '2026-10-03' : sampleToday
  const asOf = log && log.date > contextDate ? log.date : contextDate
  const ended = log?.submissions.some(submission => {
    const contest = contests.find(item => item.id === submission.contestId)
    return submission.closed || contest && contestStatus(contest, asOf) === 'ended'
  }) ?? false

  if (deleted) return <section className="empty-state records-log-result"><h1 className="paper-type-page" ref={resultRef} tabIndex={-1}>Log deleted</h1><p>The sample entry was removed from your local activity and contest submissions.</p><div className="app-form__actions"><Link className={buttonClassName()} to={`/users/${deleted.userId}/activity`}>View your activity</Link><Button variant="outline" onClick={() => { restoreLog(deleted); setDeleted(null) }}>Restore sample</Button></div></section>
  if (!log || scenario === 'missing' || scenario === 'unavailable') {
    const unavailable = scenario === 'unavailable'
    return <><Breadcrumb items={[{ id: 'activity', label: 'Activity', href: `/users/${userId}/activity` }, { id: 'log', label: 'Log details' }]} renderLink={renderLink} /><section className="empty-state records-log-result"><h1 className="paper-type-page">{unavailable ? 'This log couldn’t be loaded' : 'Log not found'}</h1><p>{unavailable ? 'There was a problem loading this record. Try again, or return to your activity.' : 'The link may be incorrect, or the log may have been removed. Return to the activity record to find another entry.'}</p><div className="app-form__actions">{unavailable ? <Button onClick={onRetry}>Try again</Button> : null}<Link className={buttonClassName({ variant: unavailable ? 'outline' : 'default' })} to={`/users/${userId}/activity`}>View activity</Link></div></section></>
  }

  return <>
    <Breadcrumb items={[{ id: 'activity', label: owner ? 'Your activity' : `${name}’s activity`, href: `/users/${log.userId}/activity` }, { id: 'log', label: 'Log details' }]} renderLink={renderLink} />
    {ended ? <Flash className="records-log-notice" title="The contests are finished">You can still edit your personal record. Their saved points stay unchanged.</Flash> : null}
    {saved ? <Flash variant="success" className="records-log-notice">Contest submissions saved.</Flash> : null}
    <article className="records-log-sheet">
      <header className="records-log-header">
        <div className="meta-row"><strong>{log.language}</strong><span>{log.activity}</span><span>{log.media}</span></div>
        <div className="records-log-title"><div><h1 className="paper-type-page">{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</h1><p className="muted">Logged by <Link className="text-link" to={`/users/${log.userId}`}>{name}</Link> on <time dateTime={log.date}>{formatDate(log.date)}</time> · UTC</p></div>{owner ? <Link className={buttonClassName({ variant: 'outline' })} to={`/logs/${log.id}/edit?asOf=${asOf}`}><PencilSquareIcon className="paper-icon-default" aria-hidden="true" />Edit log</Link> : null}</div>
      </header>
      <div className="records-log-body"><Surface as="section" className="records-log-entry" aria-label="Recorded activity">
      <div className="records-log-measure"><p><strong>{formatNumber(log.amount)}</strong><span>{log.unit} {log.activity === 'Reading' ? 'read' : 'listened'}</span></p>{log.activity === 'Reading' ? <p className="records-log-time">{log.minutes === undefined ? 'Time not recorded' : <><strong>{formatNumber(log.minutes)} min</strong><span>time spent</span></>}</p> : null}</div>
      {log.note ? <p className="records-log-note">{log.note}</p> : null}
      {log.tags.length ? <ul className="records-tags" aria-label="Tags">{log.tags.map(tag => <li key={tag}>#{tag}</li>)}</ul> : null}
      </Surface><div className="records-log-score-grid">
        <section className="records-personal-score" aria-labelledby="personal-score-heading"><h2 className="paper-type-section" id="personal-score-heading">Personal score</h2><p className="records-score-number">{formatNumber(scoreLog(log))}<span>points</span></p><p className="muted">Counts toward {owner ? 'your' : `${name}’s`} personal activity totals.</p><ScoreExplanation basis={`${formatNumber(log.amount)} ${log.unit} × ${log.activity === 'Reading' ? '1 point per page' : '0.5 points per minute'} = ${formatNumber(scoreLog(log))} personal points.${ended ? ' The personal record was corrected after the contest finished.' : ''}`} /></section>
        <section aria-labelledby="contributions-heading"><div className="records-contribution-heading"><h2 className="paper-type-section" id="contributions-heading">Contest contributions</h2></div>
          {!owner ? <p className="records-visibility-note">Contest submissions are visible to the log’s owner.</p> : <>
            {log.submissions.length ? <ul className="records-contributions">{log.submissions.map(submission => {
              const contest = contests.find(item => item.id === submission.contestId)
              const closed = submission.closed || (contest && contestStatus(contest, asOf) === 'ended')
              return <li key={submission.contestId}><div className="records-contribution-title"><Link className="text-link" to={`/contests/${submission.contestId}?asOf=${asOf}`}>{contest?.title ?? 'Contest'}</Link><strong>{formatNumber(submission.score)}<small>points</small></strong></div><p className="records-contribution-meta">{contest?.scope === 'official' ? 'Official Tadoku contest' : 'Community contest'} · {closed ? 'Finished · submission closed' : 'In progress'}</p><ScoreExplanation basis={submission.basis + (closed ? ' This saved contribution stays unchanged after later personal-log edits.' : '')} /></li>
            })}</ul> : <div className="records-visibility-note"><h3 className="paper-type-component">No contest submissions</h3><p>This activity is in your personal history. Submit it to an eligible contest to count there too.</p></div>}
            <p className="records-scope-note">{ended ? 'These are saved contest scores, so they can differ from the current personal record.' : 'Each contest scores this activity separately. These points are not added to your personal score.'}</p>
            {!ended ? <SubmissionManager key={log.id} log={log} asOf={asOf} onSaved={() => setSaved(true)} /> : null}
          </>}
        </section>
      </div></div>
    </article>
    <footer className="records-log-footer"><div><p className="muted">Viewed as of {formatDate(asOf)}.{ended ? ' Contest submission windows are closed.' : ''}</p><Link className="text-link" to="/guide/scoring">How logging and scoring work</Link></div>{owner ? <details className="records-maintenance"><summary>More actions</summary><Modal trigger={<Button variant="ghost" leadingIcon={<TrashIcon className="paper-icon-default" />}>Delete log</Button>} title="Delete this log?" description="This removes the activity from your personal history and its contest submissions." initialFocus={cancelRef} open={deleteOpen} onOpenChange={setDeleteOpen} footer={<><Button ref={cancelRef} variant="outline" onClick={() => setDeleteOpen(false)}>Keep log</Button><Button variant="destructive" onClick={() => { setDeleteOpen(false); removeLog(log.id); setDeleted(log) }}>Delete log</Button></>}><p>“{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}” will be removed from this playground. You can restore the sample afterward.</p></Modal></details> : null}</footer>
  </>
}
