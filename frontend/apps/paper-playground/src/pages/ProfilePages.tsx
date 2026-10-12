import { useRef, useState } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { AppliedFilters, AutocompleteMultiInput, Breadcrumb, Button, DescriptionList, EmptyState, HeatmapChart, Input, SectionHeading, Select, StatGrid, Tabbar, buttonClassName, chartPalette } from 'paper-ui'
import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon, PlusIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, scoreLog, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground, useScenario } from '../state'
import './records.css'

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
  return <ol className={`records-entries${compact ? ' records-entries--compact' : ''}`}>
    {logs.map(log => <li key={log.id}>
      <time dateTime={log.date} className="records-entry-date">{formatDate(log.date)}</time>
      <div className="min-w-0">
        <Link className="records-entry-title text-link" to={`/logs/${log.id}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</Link>
        <p className="records-entry-meta">{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
        {!compact && log.note ? <p className="records-entry-note">{log.note}</p> : null}
      </div>
      <span className="records-entry-score" aria-label={`${formatNumber(scoreLog(log))} personal points`}>{formatNumber(scoreLog(log))}<small>points</small></span>
    </li>)}
  </ol>
}

function ActivityLog({ logs, owner }: { logs: readonly SampleLog[]; owner: boolean }) {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const filtersToggle = useRef<HTMLButtonElement>(null)
  const methods = useForm<ActivityFilters>({ defaultValues: emptyFilters })
  const values = useWatch({ control: methods.control })
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
    methods.reset(emptyFilters)
    filtersToggle.current?.focus()
  }

  return <section className="page-section records-activity" aria-labelledby="activity-heading">
    <SectionHeading level={2} id="activity-heading" title="Activity log" actions={logs.length ? <><Button ref={filtersToggle} variant="ghost" aria-expanded={filtersOpen} aria-controls="activity-filters" onClick={() => setFiltersOpen(!filtersOpen)} trailingIcon={<ChevronDownIcon className={`paper-icon-compact${filtersOpen ? ' rotate-180' : ''}`} />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button>{owner ? <Link className={buttonClassName()} to="/logs/new"><PlusIcon className="paper-icon-default" aria-hidden="true" />Log activity</Link> : null}</> : null} />
    {logs.length ? <>
    <div id="activity-filters" hidden={!filtersOpen}><FormProvider {...methods}>
      <form onSubmit={event => event.preventDefault()} className="records-filter-panel">
        <Select name="period" label="Period" options={[{ value: 'all', label: 'All time' }, ...years.map(year => ({ value: String(year), label: String(year) }))]} />
        <AutocompleteMultiInput name="languages" label="Languages" placeholder="All languages" options={languages} format={value => value} getId={value => value} />
        <AutocompleteMultiInput name="activities" label="Activities" placeholder="All activities" options={['Reading', 'Listening']} format={value => value} getId={value => value} />
        <Input name="query" label="Search activity" type="search" placeholder="Title, note or tag" />
      </form>
    </FormProvider></div>
    <div className="records-filter-summary"><AppliedFilters filters={appliedFilters} onClear={reset} />
    <p className="records-result-count" role="status">{filtered.length} of {logs.length} entries</p></div>
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
    <header className="records-profile-header">
      <div className="records-identity"><span className="records-avatar" aria-hidden="true">{profile.name.slice(0, 2).toLocaleUpperCase()}</span><div><h1 className="paper-type-page">{profile.name}</h1><p className="muted">Tracking since {String(profile.joined).slice(0, 4)}</p></div></div>
      {owner ? <Link className={buttonClassName({ variant: 'outline' })} to="/settings">Edit profile</Link> : null}
    </header>
    {contestContext ? <p className="records-profile-context">Viewing {profile.name}’s personal record from <Link className="text-link" to={`/contests/${contestContext.id}/leaderboard`}>{contestContext.title} standings</Link>. Contest scores are shown on that board.</p> : null}
    <Tabbar label="Profile views" renderLink={renderLink} links={[{ id: 'overview', label: 'Overview', href: `/users/${profile.id}${viewQuery}`, current: !activity }, { id: 'activity', label: 'Activity log', href: `/users/${profile.id}/activity${viewQuery}`, current: activity }]} />
    {activity ? <ActivityLog key={profile.id} logs={ownLogs} owner={owner} /> : <>
      <section className="page-section records-year" aria-labelledby="profile-year-heading">
        <SectionHeading level={2} id="profile-year-heading" title={`${profile.name}’s activity`} actions={<div className="records-year-control" aria-label="Profile year"><Button variant="ghost" aria-label="Previous year" disabled={yearIndex === years.length - 1} onClick={() => changeYear(years[yearIndex + 1])}><ChevronLeftIcon className="paper-icon-default" /></Button><strong aria-live="polite">{year}</strong><Button variant="ghost" aria-label="Next year" disabled={yearIndex === 0} onClick={() => changeYear(years[yearIndex - 1])}><ChevronRightIcon className="paper-icon-default" /></Button></div>} />
        {summary.entries ? <><StatGrid className="records-profile-stats" stats={[
          { label: 'Personal score', value: formatNumber(summary.total) },
          { label: 'Entries', value: summary.entries, detail: `${summary.days} active days` },
          { label: 'Languages', value: summary.languages.length, detail: summary.languages[0] ? `${summary.languages[0][0]} leads` : 'A new language awaits' },
          { label: 'Contests', value: summary.contestIds.length },
        ]} />
      <div className="records-rhythm" aria-labelledby="rhythm-heading">
        <SectionHeading level={3} id="rhythm-heading" title="Immersion rhythm" actions={<Link className="text-link" to="/guide/scoring">How scoring works</Link>} />
        <div className="records-heatmap-scroll paper-focus-ring" role="region" aria-label={`Daily immersion calendar for ${year}; scroll for all months`} tabIndex={0}><HeatmapChart id={`profile-${profile.id}-${year}`} year={year} data={summary.daily} /></div>
        <p className="records-chart-caption">{summary.days} active days · Longest streak: {summary.longestStreak} {summary.longestStreak === 1 ? 'day' : 'days'}<span>Stronger color means more activity.</span></p>
      </div></> : <EmptyState variant="plain" className="records-empty-year" headingLevel={3} title={`No activity recorded in ${year}`} description={owner ? 'A few pages or minutes are a good place to start. Your personal record grows with every entry, whether you join a contest or read on your own.' : 'Personal scores, the activity calendar and contest history will appear after the first entry for this year.'} actions={owner ? <Link className={buttonClassName()} to="/logs/new">Log activity</Link> : null} />}</section>
      {summary.entries ? <div className="records-overview-columns">
          <section className="page-section" aria-labelledby="language-heading"><SectionHeading level={2} id="language-heading" title="Score by language" />
            {summary.languages.length ? <ol className="records-breakdown">{summary.languages.map(([language, score], index) => <li key={language}><div><span>{language}</span><strong>{formatNumber(score)} <small>points</small></strong></div><span className="records-bar-track" aria-hidden="true"><span style={{ width: `${summary.total ? score / summary.total * 100 : 0}%`, backgroundColor: chartPalette[index % chartPalette.length] }} /></span></li>)}</ol> : <p className="muted">No activity recorded in {year}.</p>}
          </section>
          <section className="page-section records-mix" aria-labelledby="mix-heading"><SectionHeading level={2} id="mix-heading" title="Score by activity" /><DescriptionList layout="inline" items={summary.activities.map(([name, score]) => ({ label: name, value: `${summary.total ? Math.round(score / summary.total * 100) : 0}%` }))} />{!summary.entries ? <p className="muted">Reading and listening will appear after a first log.</p> : null}</section>
          <section className="page-section records-history" aria-labelledby="history-heading"><SectionHeading level={2} id="history-heading" title="Contest history" actions={<Link className="text-link" to="/contests/official">All contests</Link>} />
            {summary.contestIds.length ? <ul className="records-contest-history">{summary.contestIds.map(id => { const contest = contests.find(item => item.id === id); const score = yearLogs.flatMap(log => log.submissions).filter(submission => submission.contestId === id).reduce((sum, item) => sum + item.score, 0); return <li key={id}><div><Link className="text-link" to={`/contests/${id}`}>{contest?.title ?? 'Contest'}</Link><p className="muted">{contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : 'Submitted activity'}</p></div><strong>{formatNumber(score)}<small>contest points</small></strong></li> })}</ul> : <p className="muted">No contest submissions in {year}. Personal activity still counts here.</p>}
          </section>
          <section className="page-section" aria-labelledby="recent-heading"><SectionHeading level={2} id="recent-heading" title="Recent activity" actions={<Link className="text-link" to={`/users/${profile.id}/activity${viewQuery}`}>View all</Link>} />{yearLogs.length ? <ActivityRows logs={yearLogs.slice(0, 3)} compact /> : <p className="muted">No entries for this year.</p>}</section>
      </div> : null}
    </>}
  </>
}
