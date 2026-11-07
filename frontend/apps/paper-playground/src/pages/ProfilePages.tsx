import { useStore } from '@tanstack/react-form'
import { useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { AppliedFilters, Breadcrumb, Button, buttonClassName, chartPalette, cn, DescriptionList, EmptyState, focusRing, HeatmapChart, SectionHeading, StatGrid, Tabbar, typography, usePaperForm } from 'paper-ui'
import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon, iconClassName, PlusIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, scoreLog, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground, useScenario } from '../state'

const years = [2026, 2025, 2024, 2023]

type ActivityFilters = { period: string; languages: string[]; activities: string[]; query: string }
const emptyFilters: ActivityFilters = { period: 'all', languages: [], activities: [], query: '' }

export function filterActivityLogs(logs: readonly SampleLog[], filters: ActivityFilters) {
  const query = filters.query.trim().toLocaleLowerCase()
  return logs.filter(log =>
    (filters.period === 'all' || log.date.startsWith(filters.period)) &&
    (!filters.languages.length || filters.languages.includes(log.language)) &&
    (!filters.activities.length || filters.activities.includes(log.activity)) &&
    (!query || [log.title, log.language, log.activity, log.note, log.media, ...log.tags].join(' ').toLocaleLowerCase().includes(query)),
  ).sort((left, right) => right.date.localeCompare(left.date))
}

export function summarizeActivity(logs: readonly SampleLog[]) {
  const daily = new Map<string, number>()
  const languages = new Map<string, number>()
  const activities = new Map<string, number>()
  for (const log of logs) {
    const score = scoreLog(log)
    daily.set(log.date.slice(0, 10), (daily.get(log.date.slice(0, 10)) ?? 0) + score)
    languages.set(log.language, (languages.get(log.language) ?? 0) + score)
    activities.set(log.activity, (activities.get(log.activity) ?? 0) + score)
  }
  const days = [...daily.keys()].sort()
  let longestStreak = 0
  let streak = 0
  days.forEach((day, index) => {
    streak = index > 0 && Date.parse(day) - Date.parse(days[index - 1]) === 86_400_000 ? streak + 1 : 1
    longestStreak = Math.max(longestStreak, streak)
  })
  return {
    total: [...daily.values()].reduce((sum, value) => sum + value, 0),
    entries: logs.length,
    days: daily.size,
    longestStreak,
    daily: [...daily].map(([date, value]) => ({ date, value, tooltip: `${formatDate(date)}: ${formatNumber(value)} personal points` })),
    languages: [...languages].sort((left, right) => right[1] - left[1]),
    activities: [...activities].sort((left, right) => right[1] - left[1]),
    contestIds: [...new Set(logs.flatMap(log => log.submissions.map(submission => submission.contestId)))],
  }
}

function ActivityRows({ logs, compact = false }: { logs: readonly SampleLog[]; compact?: boolean }) {
  return <ol className={`records-entries m-0 list-none p-0 [&>li]:grid [&>li]:grid-cols-record-entry [&>li]:items-start [&>li]:gap-6 [&>li]:border-b [&>li]:border-rule [&>li]:py-6 max-sm:[&>li]:grid-cols-record-entry-narrow max-sm:[&>li]:gap-x-4 max-sm:[&>li]:gap-y-1.5 max-sm:[&_.records-entry-date]:col-span-full max-sm:[&_.records-entry-date]:p-0 ${compact ? ' [&&>li]:grid-cols-content-end [&&>li]:gap-x-4 [&&>li]:gap-y-1 [&&>li]:py-4 [&&>li:first-child]:pt-0 max-sm:[&&>li]:grid-cols-record-entry-narrow max-sm:[&&>li]:gap-y-1.5 [&_.records-entry-date]:col-span-full [&_.records-entry-date]:p-0 [&_.records-entry-title]:text-ui' : ''}`}>
    {logs.map(log => <li key={log.id}>
      <time dateTime={log.date} className="records-entry-date pt-0.75 text-fine text-muted">{formatDate(log.date)}</time>
      <div className="min-w-0">
        <Link className="records-entry-title font-semibold leading-150 decoration-hairline underline-offset-link wrap-anywhere text-link" to={`/logs/${log.id}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</Link>
        <p className="mb-0 mt-1.5 text-fine text-muted">{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
        {!compact && log.note ? <p className="mb-0 mt-2.5 max-w-measure-note text-sm leading-170">{log.note}</p> : null}
      </div>
      <span className="text-end font-bold tabular-nums [&_small]:block [&_small]:text-xs [&_small]:font-regular [&_small]:text-muted" aria-label={`${formatNumber(scoreLog(log))} personal points`}>{formatNumber(scoreLog(log))}<small>points</small></span>
    </li>)}
  </ol>
}

function ActivityLog({ logs, owner }: { logs: readonly SampleLog[]; owner: boolean }) {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const filtersToggle = useRef<HTMLButtonElement>(null)
  const methods = usePaperForm({ defaultValues: (emptyFilters as ActivityFilters) })
  const values = useStore(methods.store, state => state.values)
  const applied = { ...emptyFilters, ...values }
  const filtered = filterActivityLogs(logs, applied)
  const languages = [...new Set(logs.map(log => log.language))].sort()
  const appliedFilters = [
    ...(applied.period !== 'all' ? [applied.period] : []),
    ...applied.languages,
    ...applied.activities,
    ...(applied.query.trim() ? [`Search: ${applied.query.trim()}`] : []),
  ]
  const reset = () => {
    methods.reset(emptyFilters, { keepDefaultValues: true })
    filtersToggle.current?.focus()
  }

  return <section className="mt-9" aria-labelledby="activity-heading">
    <SectionHeading level={2} id="activity-heading" title="Activity log" actions={logs.length ? <><Button ref={filtersToggle} variant="ghost" aria-expanded={filtersOpen} aria-controls="activity-filters" onClick={() => setFiltersOpen(!filtersOpen)} trailingIcon={<ChevronDownIcon className={iconClassName('compact', filtersOpen ? 'rotate-flip' : undefined)} />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button>{owner ? <Link className={buttonClassName()} to="/logs/new"><PlusIcon className={iconClassName()} aria-hidden="true" />Log activity</Link> : null}</> : null} />
    {logs.length ? <>
    <div id="activity-filters" hidden={!filtersOpen}><methods.AppForm>
      <form onSubmit={event => event.preventDefault()} className="grid grid-cols-standings-filters-yearly gap-3 pb-5 max-tablet:grid-cols-halves max-sm:grid-cols-single">
        <methods.AppField name="period">{field => <field.Select label="Period" options={[{ value: 'all', label: 'All time' }, ...years.map(year => ({ value: String(year), label: String(year) }))]} />}</methods.AppField>
        <methods.AppField name="languages">{field => <field.AutocompleteMultiInput label="Languages" placeholder="All languages" options={languages} format={value => value} getId={value => value} />}</methods.AppField>
        <methods.AppField name="activities">{field => <field.AutocompleteMultiInput label="Activities" placeholder="All activities" options={['Reading', 'Listening']} format={value => value} getId={value => value} />}</methods.AppField>
        <methods.AppField name="query">{field => <field.Input label="Search activity" type="search" placeholder="Title, note or tag" />}</methods.AppField>
      </form>
    </methods.AppForm></div>
    <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-rule pb-4"><AppliedFilters filters={appliedFilters} onClear={reset} />
    <p className="m-0 text-fine text-muted" role="status">{filtered.length} of {logs.length} entries</p></div>
    {filtered.length ? <ActivityRows logs={filtered} /> : <EmptyState headingLevel={3} title="No activity matches these filters." description="Try another title or include more dates and languages." actions={<Button variant="outline" onClick={reset}>Clear search and filters</Button>} />}
    </> : <EmptyState headingLevel={3} title="No activity yet" description={owner ? 'Start with what you read or listened to today. A title and a few pages or minutes are enough.' : 'Reading and listening entries will appear here once they have been recorded.'} actions={owner ? <Link className={buttonClassName()} to="/logs/new">Log activity</Link> : null} />}
  </section>
}

export function ProfilePage({ activity = false }: { activity?: boolean }) {
  const { userId } = useParams()
  const [params, setParams] = useSearchParams()
  const { logs, contests, user, viewer, userId: viewerId, displayName } = usePlayground()
  const [scenario] = useScenario('profile')
  const foundProfile = users.find(candidate => candidate.id === userId)
  const profile = foundProfile?.id === viewerId ? { ...foundProfile, name: displayName } : foundProfile
  const selected = Number(params.get('year') ?? 2026)
  const year = years.includes(selected) ? selected : 2026
  const contestContext = contests.find(contest => contest.id === params.get('contest'))
  const viewParams = new URLSearchParams({ year: String(year) })
  if (contestContext) viewParams.set('contest', contestContext.id)
  if (params.has('scenario')) viewParams.set('scenario', scenario)
  const viewQuery = `?${viewParams}`
  const ownLogs = scenario === 'empty' ? [] : logs.filter(log => log.userId === userId)
  const yearLogs = filterActivityLogs(ownLogs, { ...emptyFilters, period: String(year) })
  const summary = summarizeActivity(yearLogs)
  const owner = viewer !== 'guest' && user?.id === userId
  const yearIndex = years.indexOf(year)
  const changeYear = (next: number) => { const values = new URLSearchParams(params); values.set('year', String(next)); setParams(values) }

  if (!profile) return <EmptyState headingLevel={1} title="Profile not found" description="This account is not in the playground’s sample community." actions={<Link className={buttonClassName()} to="/leaderboard/latest">Explore the leaderboard</Link>} />

  return <>
    <Breadcrumb items={[{ id: 'home', label: 'Home', href: '/' }, { id: 'profile', label: profile.name }]} renderLink={renderLink} />
    <header className="mb-8 mt-6 flex flex-wrap items-center justify-between gap-4">
      <div className="flex min-w-0 items-center gap-5 max-sm:gap-4 [&>div]:min-w-0 [&_h1]:mb-2 [&_h1]:wrap-anywhere [&_p]:m-0 [&_p]:text-ui max-sm:[&_p]:text-fine"><span className="grid h-16 flex-none basis-16 place-items-center border border-rule bg-paper font-display text-avatar font-bold text-ink max-sm:hidden" aria-hidden="true">{profile.name.slice(0, 2).toLocaleUpperCase()}</span><div><h1 className={typography({ role: 'page' })}>{profile.name}</h1><p className="text-muted">Tracking since {String(profile.joined).slice(0, 4)}</p></div></div>
      {owner ? <Link className={buttonClassName({ variant: 'outline' })} to="/settings">Edit profile</Link> : null}
    </header>
    {contestContext ? <p className="-mt-4 mb-6 max-w-measure-context text-sm text-muted">Viewing {profile.name}’s personal record from <Link className="text-link" to={`/contests/${contestContext.id}/leaderboard`}>{contestContext.title} standings</Link>. Contest scores are shown on that board.</p> : null}
    <Tabbar label="Profile views" renderLink={renderLink} links={[{ id: 'overview', label: 'Overview', href: `/users/${profile.id}${viewQuery}`, current: !activity }, { id: 'activity', label: 'Activity log', href: `/users/${profile.id}/activity${viewQuery}`, current: activity }]} />
    {activity ? <ActivityLog key={profile.id} logs={ownLogs} owner={owner} /> : <>
      <section className="mt-9 border border-rule bg-paper p-8 max-tablet:p-6 max-sm:p-5" aria-labelledby="profile-year-heading">
        <SectionHeading level={2} id="profile-year-heading" title={`${profile.name}’s activity`} actions={<div className="flex items-center gap-1 [&_strong]:min-w-16 [&_strong]:text-center [&_strong]:tabular-nums" aria-label="Profile year"><Button variant="ghost" aria-label="Previous year" disabled={yearIndex === years.length - 1} onClick={() => changeYear(years[yearIndex + 1])}><ChevronLeftIcon className={iconClassName()} /></Button><strong aria-live="polite">{year}</strong><Button variant="ghost" aria-label="Next year" disabled={yearIndex === 0} onClick={() => changeYear(years[yearIndex - 1])}><ChevronRightIcon className={iconClassName()} /></Button></div>} />
        {summary.entries ? <><StatGrid className="mb-8 mt-7" stats={[
          { label: 'Personal score', value: formatNumber(summary.total) },
          { label: 'Entries', value: summary.entries, detail: `${summary.days} active days` },
          { label: 'Languages', value: summary.languages.length, detail: summary.languages[0] ? `${summary.languages[0][0]} leads` : 'A new language awaits' },
          { label: 'Contests', value: summary.contestIds.length },
        ]} />
      <div className="border-t border-rule pt-6" aria-labelledby="rhythm-heading">
        <SectionHeading level={3} id="rhythm-heading" title="Immersion rhythm" actions={<Link className="text-link" to="/guide/scoring">How scoring works</Link>} />
        <div className={cn(focusRing(), 'max-w-full overflow-x-auto pb-3 pt-6 [&_.paper-heatmap]:block [&_.paper-heatmap]:h-auto [&_.paper-heatmap]:w-full [&_.paper-heatmap]:min-w-heatmap')} role="region" aria-label={`Daily immersion calendar for ${year}; scroll for all months`} tabIndex={0}><HeatmapChart id={`profile-${profile.id}-${year}`} year={year} data={summary.daily} /></div>
        <p className="mb-0 mt-2 flex flex-wrap justify-between gap-x-6 gap-y-2 text-fine text-muted">{summary.days} active days · Longest streak: {summary.longestStreak} {summary.longestStreak === 1 ? 'day' : 'days'}<span>Stronger color means more activity.</span></p>
      </div></> : <EmptyState variant="plain" className="border-t border-rule" headingLevel={3} title={`No activity recorded in ${year}`} description={owner ? 'A few pages or minutes are a good place to start. Your personal record grows with every entry, whether you join a contest or read on your own.' : 'Personal scores, the activity calendar and contest history will appear after the first entry for this year.'} actions={owner ? <Link className={buttonClassName()} to="/logs/new">Log activity</Link> : null} />}</section>
      {summary.entries ? <div className="grid grid-cols-overview-columns gap-x-14 gap-y-0 *:min-w-0 *:border-t *:border-rule *:pt-7 max-tablet:grid-cols-single max-tablet:gap-0">
          <section className="mt-9" aria-labelledby="language-heading"><SectionHeading level={2} id="language-heading" title="Score by language" />
            {summary.languages.length ? <ol className="m-0 list-none p-0 [&_li+li]:mt-5 [&_li>div]:mb-2 [&_li>div]:flex [&_li>div]:justify-between [&_li>div]:gap-4 [&_li>div]:text-ui [&_strong]:tabular-nums [&_small]:font-regular [&_small]:text-muted">{summary.languages.map(([language, score], index) => <li key={language}><div><span>{language}</span><strong>{formatNumber(score)} <small>points</small></strong></div><span className="block h-1.5 bg-rule-subtle [&>span]:block [&>span]:h-full [&>span]:min-w-0.5" aria-hidden="true"><span style={{ width: `${summary.total ? score / summary.total * 100 : 0}%`, backgroundColor: chartPalette[index % chartPalette.length] }} /></span></li>)}</ol> : <p className="text-muted">No activity recorded in {year}.</p>}
          </section>
          <section className="mt-9" aria-labelledby="mix-heading"><SectionHeading level={2} id="mix-heading" title="Score by activity" /><DescriptionList layout="inline" items={summary.activities.map(([name, score]) => ({ label: name, value: `${summary.total ? Math.round(score / summary.total * 100) : 0}%` }))} />{!summary.entries ? <p className="text-muted">Reading and listening will appear after a first log.</p> : null}</section>
          <section className="mt-9" aria-labelledby="history-heading"><SectionHeading level={2} id="history-heading" title="Contest history" actions={<Link className="text-link" to="/contests/official">All contests</Link>} />
            {summary.contestIds.length ? <ul className="m-0 list-none p-0 [&_li]:flex [&_li]:justify-between [&_li]:gap-4 [&_li]:border-b [&_li]:border-rule [&_li]:py-4.5 [&_li:first-child]:pt-0 [&_p]:mb-0 [&_p]:mt-1.5 [&_p]:text-fine [&_strong]:whitespace-nowrap [&_strong]:text-end [&_strong]:tabular-nums [&_strong_small]:block [&_strong_small]:text-xs [&_strong_small]:font-regular [&_strong_small]:text-muted">{summary.contestIds.map(id => { const contest = contests.find(item => item.id === id); const score = yearLogs.flatMap(log => log.submissions).filter(submission => submission.contestId === id).reduce((sum, item) => sum + item.score, 0); return <li key={id}><div><Link className="text-link" to={`/contests/${id}`}>{contest?.title ?? 'Contest'}</Link><p className="text-muted">{contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : 'Submitted activity'}</p></div><strong>{formatNumber(score)}<small>contest points</small></strong></li> })}</ul> : <p className="text-muted">No contest submissions in {year}. Personal activity still counts here.</p>}
          </section>
          <section className="mt-9" aria-labelledby="recent-heading"><SectionHeading level={2} id="recent-heading" title="Recent activity" actions={<Link className="text-link" to={`/users/${profile.id}/activity${viewQuery}`}>View all</Link>} />{yearLogs.length ? <ActivityRows logs={yearLogs.slice(0, 3)} compact /> : <p className="text-muted">No entries for this year.</p>}</section>
      </div> : null}
    </>}
  </>
}
