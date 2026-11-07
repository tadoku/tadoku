import { useEffect, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { Breadcrumb, Button, buttonClassName, EmptyState, Flash, Modal, PageHeader, Surface, typography, usePaperForm } from 'paper-ui'
import { iconClassName, PencilSquareIcon, TrashIcon } from 'paper-ui/icons'
import { contestStatus, formatDate, formatNumber, personalScoreExplanation, sampleToday, scoreLog, scoreSubmission, users, type SampleContest, type SampleLog, type Submission } from '../data'
import { contestEnd, contestStart, formatDateRange } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground, useScenario } from '../state'

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
  const methods = usePaperForm({ defaultValues: (defaults() as { selected: Record<string, boolean> }), onSubmit: ({ value: values }) => { saveLog({ ...log, submissions: updateSubmissions(log, Object.entries(values.selected).filter(([, selected]) => selected).map(([id]) => id), contests, joinedContests, registrations, asOf) }); setOpen(false); onSaved() } })
  const choices = contests.filter(contest => joinedContests.includes(contest.id) || log.submissions.some(submission => submission.contestId === contest.id))
  const available = choices.filter(contest => !submissionEligibility(log, contest, joinedContests, registrations, asOf))
  return <Modal title="Contest submissions" description="Choose where this activity contributes. Each contest keeps its own score." trigger={<Button variant="outline">{log.submissions.length ? 'Manage submissions' : 'Submit to a contest'}</Button>} open={open} onOpenChange={next => { if (next) methods.reset(defaults(), { keepDefaultValues: true }); setOpen(next) }} footer={null}>
    <methods.AppForm>
      <methods.Form >
        {choices.length ? <div className="grid gap-5">{choices.map(contest => {
          const reason = submissionEligibility(log, contest, joinedContests, registrations, asOf)
          return <div key={contest.id}><methods.AppField name={`selected.${contest.id}`}>{field => <field.Checkbox label={contest.title} disabled={Boolean(reason)} hint={reason ?? `${formatDateRange(contestStart(contest.start), contestEnd(contest.end))} · Eligible`} />}</methods.AppField></div>
        })}</div> : null}
        {!available.length ? <EmptyState variant="plain" headingLevel={3} title="No eligible contests" description="Check the activity date, language, and your registrations. This activity already counts in your personal history." actions={<Link className="text-link" to={`/contests/official?asOf=${asOf}`} onClick={() => setOpen(false)}>Find a contest to join</Link>} /> : <p className="text-muted text-sm">Removing a selection removes that contest contribution. Your personal record stays the same.</p>}
        <div className="mt-7 flex flex-wrap gap-3"><Button type="submit" disabled={!available.length}>Save submissions</Button><Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button></div>
      </methods.Form>
    </methods.AppForm>
  </Modal>
}

function ScoreExplanation({ basis }: { basis: string }) {
  return <details className="mt-3 text-fine leading-170 [&_summary]:cursor-pointer [&_summary]:text-link [&_summary]:underline-offset-link [&[open]_summary]:mb-3 [&_p]:my-2 [&_p]:max-w-measure-description"><summary>How this was scored</summary><p>{basis}</p><p className="text-muted">Illustrative saved scoring rule.</p></details>
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

  if (deleted) return <EmptyState className="mt-6" headingLevel={1} headingRef={resultRef} title="Log deleted" description="The sample entry was removed from your local activity and contest submissions." actions={<><Link className={buttonClassName()} to={`/users/${deleted.userId}/activity`}>View your activity</Link><Button variant="outline" onClick={() => { restoreLog(deleted); setDeleted(null) }}>Restore sample</Button></>} />
  if (!log || scenario === 'missing' || scenario === 'unavailable') {
    const unavailable = scenario === 'unavailable'
    return <><Breadcrumb items={[{ id: 'activity', label: 'Activity', href: `/users/${userId}/activity` }, { id: 'log', label: 'Log details' }]} renderLink={renderLink} /><EmptyState className="mt-6" headingLevel={1} title={unavailable ? 'This log couldn’t be loaded' : 'Log not found'} description={unavailable ? 'There was a problem loading this record. Try again, or return to your activity.' : 'The link may be incorrect, or the log may have been removed. Return to the activity record to find another entry.'} actions={<>{unavailable ? <Button onClick={onRetry}>Try again</Button> : null}<Link className={buttonClassName({ variant: unavailable ? 'outline' : 'default' })} to={`/users/${userId}/activity`}>View activity</Link></>} /></>
  }

  return <>
    <Breadcrumb items={[{ id: 'activity', label: owner ? 'Your activity' : `${name}’s activity`, href: `/users/${log.userId}/activity` }, { id: 'log', label: 'Log details' }]} renderLink={renderLink} />
    {ended ? <Flash className="my-6" title="The contests are finished">You can still edit your personal record. Their saved points stay unchanged.</Flash> : null}
    {saved ? <Flash variant="success" className="my-6">Contest submissions saved.</Flash> : null}
    <article className="mt-6">
      <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-sm text-muted"><strong>{log.language}</strong><span>{log.activity}</span><span>{log.media}</span></div>
      <PageHeader
        className="mb-0 mt-4 border-b border-rule pb-8 max-sm:pb-6"
        title={log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}
        lead={<>Logged by <Link className="text-link" to={`/users/${log.userId}`}>{name}</Link> on <time dateTime={log.date}>{formatDate(log.date)}</time></>}
        actions={owner ? <Link className={buttonClassName({ variant: 'outline' })} to={`/logs/${log.id}/edit?asOf=${asOf}`}><PencilSquareIcon className={iconClassName()} aria-hidden="true" />Edit log</Link> : null}
      />
      <div className="grid grid-cols-log-body items-start gap-12 pt-8 *:min-w-0 max-lg:grid-cols-log-body-narrow max-lg:gap-8 max-tablet:grid-cols-single max-sm:gap-6 max-sm:pt-6"><Surface as="section" className="p-8 max-sm:p-5" aria-label="Recorded activity">
      <div className="mb-7 mt-0 flex flex-wrap items-baseline gap-x-8 gap-y-5 max-sm:mb-6 [&_p]:m-0 [&_p]:flex [&_p]:flex-wrap [&_p]:items-baseline [&_p]:gap-2 [&_p>strong]:text-measure [&_p>strong]:leading-120 [&_p>strong]:tabular-nums [&_p>span]:text-ui [&_p>span]:text-muted [&_.records-log-time>strong]:text-xl"><p><strong>{formatNumber(log.amount)}</strong><span>{log.unit} {log.activity === 'Reading' ? 'read' : 'listened'}</span></p>{log.activity === 'Reading' ? <p className="records-log-time text-ui text-muted">{log.minutes === undefined ? 'Time not recorded' : <><strong>{formatNumber(log.minutes)} min</strong><span>time spent</span></>}</p> : null}</div>
      {log.note ? <p className="mb-6 mt-0 max-w-measure-description leading-185">{log.note}</p> : null}
      {log.tags.length ? <ul className="m-0 flex list-none flex-wrap gap-x-4 gap-y-2 p-0 text-fine text-muted" aria-label="Tags">{log.tags.map(tag => <li key={tag}>#{tag}</li>)}</ul> : null}
      </Surface><div className="grid content-start gap-7 [&_h2]:mb-4 [&_h2]:mt-0 [&_h2]:font-body [&_h2]:text-base [&_h2]:leading-150 [&_h2]:font-semibold [&>section+section]:border-t [&>section+section]:border-rule [&>section+section]:pt-6 max-tablet:grid-cols-halves max-tablet:gap-8 max-tablet:border-s-0 max-tablet:border-t max-tablet:border-rule max-tablet:ps-0 max-tablet:pt-6 max-tablet:[&>section+section]:border-0 max-tablet:[&>section+section]:p-0 max-sm:grid-cols-single max-sm:gap-6 max-sm:[&>section+section]:border-t max-sm:[&>section+section]:pt-6">
        <section className="[&>.text-muted]:my-3 [&>.text-muted]:text-sm [&>.text-muted]:leading-170" aria-labelledby="personal-score-heading"><h2 className={typography({ role: 'section' })} id="personal-score-heading">Personal score</h2><p className="m-0 flex items-baseline gap-2 text-score leading-125 font-bold tabular-nums [&>span]:text-sm [&>span]:font-regular [&>span]:text-muted">{formatNumber(scoreLog(log))}<span>points</span></p><p className="text-muted">Counts toward {owner ? 'your' : `${name}’s`} personal activity totals.</p><ScoreExplanation basis={`${personalScoreExplanation(log)}${ended ? ' The personal record was corrected after the contest finished.' : ''}`} /></section>
        <section aria-labelledby="contributions-heading"><div className=""><h2 className={typography({ role: 'section' })} id="contributions-heading">Contest contributions</h2></div>
          {!owner ? <p className="m-0 text-sm leading-170 text-muted [&_h3]:mb-2 [&_h3]:mt-0 [&_h3]:font-body [&_h3]:text-ui [&_h3]:leading-150 [&_h3]:font-semibold [&_h3]:text-ink [&_p]:m-0">Contest submissions are visible to the log’s owner.</p> : <>
            {log.submissions.length ? <ul className="m-0 list-none p-0 [&>li]:border-b [&>li]:border-rule [&>li]:py-5 [&>li:first-child]:pt-0">{log.submissions.map(submission => {
              const contest = contests.find(item => item.id === submission.contestId)
              const closed = submission.closed || (contest && contestStatus(contest, asOf) === 'ended')
              return <li key={submission.contestId}><div className="flex justify-between gap-4 text-sm font-semibold [&_strong]:whitespace-nowrap [&_strong]:text-end [&_strong]:tabular-nums [&_small]:block [&_small]:text-xs [&_small]:font-regular [&_small]:text-muted"><Link className="text-link" to={`/contests/${submission.contestId}?asOf=${asOf}`}>{contest?.title ?? 'Contest'}</Link><strong>{formatNumber(submission.score)}<small>points</small></strong></div><p className="mb-0 mt-2 text-xs leading-160 text-muted">{contest?.scope === 'official' ? 'Official Tadoku contest' : 'Community contest'} · {closed ? 'Finished · submission closed' : 'In progress'}</p><ScoreExplanation basis={submission.basis + (closed ? ' This saved contribution stays unchanged after later personal-log edits.' : '')} /></li>
            })}</ul> : <div className="m-0 text-sm leading-170 text-muted [&_h3]:mb-2 [&_h3]:mt-0 [&_h3]:font-body [&_h3]:text-ui [&_h3]:leading-150 [&_h3]:font-semibold [&_h3]:text-ink [&_p]:m-0"><h3 className={typography({ role: 'component' })}>No contest submissions</h3><p>This activity is in your personal history. Submit it to an eligible contest to count there too.</p></div>}
            <p className="mb-5 mt-4 text-fine leading-170 text-muted">{ended ? 'These are saved contest scores, so they can differ from the current personal record.' : 'Each contest scores this activity separately. These points are not added to your personal score.'}</p>
            {!ended ? <SubmissionManager key={log.id} log={log} asOf={asOf} onSaved={() => setSaved(true)} /> : null}
          </>}
        </section>
      </div></div>
    </article>
    <footer className="mt-6 flex items-start justify-between gap-6 text-fine max-sm:flex-col [&_p]:mb-2 [&_p]:mt-0"><div><p className="text-muted">Viewed as of {formatDate(asOf)}.{ended ? ' Contest submission windows are closed.' : ''}</p><Link className="text-link" to="/guide/scoring">How logging and scoring work</Link></div>{owner ? <details className="shrink-0 text-end max-sm:text-start [&_summary]:cursor-pointer [&_summary]:py-2"><summary>More actions</summary><Modal trigger={<Button variant="ghost" leadingIcon={<TrashIcon className={iconClassName()} />}>Delete log</Button>} title="Delete this log?" description="This removes the activity from your personal history and its contest submissions." initialFocus={cancelRef} open={deleteOpen} onOpenChange={setDeleteOpen} footer={<><Button ref={cancelRef} variant="outline" onClick={() => setDeleteOpen(false)}>Keep log</Button><Button variant="destructive" onClick={() => { setDeleteOpen(false); removeLog(log.id); setDeleted(log) }}>Delete log</Button></>}><p>“{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}” will be removed from this playground. You can restore the sample afterward.</p></Modal></details> : null}</footer>
  </>
}
