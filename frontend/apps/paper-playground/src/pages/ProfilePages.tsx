import { useRef, useState } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { AutocompleteMultiInput, Breadcrumb, Button, HeatmapChart, Input, Select, Tabbar, buttonClassName, chartPalette, type NavigationLinkProps } from 'paper-ui'
import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon, PlusIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, scoreLog, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange } from '../dates'
import { usePlayground, useScenario } from '../state'
import './records.css'

const renderLink = ({ href, ...props }: NavigationLinkProps) => <Link className="text-link" to={href} {...props} />
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
    <div className="section-heading records-section-heading">
      <div><h2 className="paper-type-section" id="activity-heading">Activity log</h2></div>
      {logs.length ? <div className="records-activity-actions"><Button ref={filtersToggle} variant="ghost" className="records-filter-toggle" aria-expanded={filtersOpen} aria-controls="activity-filters" onClick={() => setFiltersOpen(!filtersOpen)} trailingIcon={<ChevronDownIcon className="paper-icon-compact" />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button>{owner ? <Link className={buttonClassName()} to="/logs/new"><PlusIcon className="paper-icon-default" aria-hidden="true" />Log activity</Link> : null}</div> : null}
    </div>
    {logs.length ? <>
    <div id="activity-filters" hidden={!filtersOpen}><FormProvider {...methods}>
      <form onSubmit={event => event.preventDefault()} className="records-filter-panel">
        <Select name="period" label="Period" options={[{ value: 'all', label: 'All time' }, ...years.map(year => ({ value: String(year), label: String(year) }))]} />
        <AutocompleteMultiInput name="languages" label="Languages" placeholder="All languages" options={languages} format={value => value} getId={value => value} />
        <AutocompleteMultiInput name="activities" label="Activities" placeholder="All activities" options={['Reading', 'Listening']} format={value => value} getId={value => value} />
        <Input name="query" label="Search activity" type="search" placeholder="Title, note or tag" />
      </form>
    </FormProvider></div>
    <div className="records-filter-summary">{appliedFilters.length ? <div className="records-applied-filters" aria-label="Applied filters">
      <p>{appliedFilters.join(' · ')}</p>
      <Button variant="link" onClick={reset}>Clear all</Button>
    </div> : null}
    <p className="records-result-count" role="status">{filtered.length} of {logs.length} entries</p></div>
    {filtered.length ? <ActivityRows logs={filtered} /> : <div className="empty-state"><h3 className="paper-type-component">No activity matches these filters.</h3><p>Try another title or include more dates and languages.</p><Button variant="outline" onClick={reset}>Clear search and filters</Button></div>}
    </> : <div className="empty-state"><h3 className="paper-type-component">No activity yet</h3><p>{owner ? 'Start with what you read or listened to today. A title and a few pages or minutes are enough.' : 'Reading and listening entries will appear here once they have been recorded.'}</p>{owner ? <Link className={buttonClassName()} to="/logs/new">Log activity</Link> : null}</div>}
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

  if (!profile) return <section className="empty-state"><h1 className="paper-type-page">Profile not found</h1><p>This account is not in the playground’s sample community.</p><Link className={buttonClassName()} to="/leaderboard/latest">Explore the leaderboard</Link></section>

  return <>
    <Breadcrumb items={[{ id: 'home', label: 'Home', href: '/' }, { id: 'profile', label: profile.name }]} renderLink={renderLink} />
    <header className="page-header records-profile-header">
      <div className="records-identity"><span className="records-avatar" aria-hidden="true">{profile.name.slice(0, 2).toLocaleUpperCase()}</span><div><h1 className="paper-type-page">{profile.name}</h1><p className="muted">Tracking since {String(profile.joined).slice(0, 4)}</p></div></div>
      {owner ? <Link className={buttonClassName({ variant: 'outline' })} to="/settings">Edit profile</Link> : null}
    </header>
    {contestContext ? <p className="records-profile-context">Viewing {profile.name}’s personal record from <Link className="text-link" to={`/contests/${contestContext.id}/leaderboard`}>{contestContext.title} standings</Link>. Contest scores are shown on that board.</p> : null}
    <Tabbar label="Profile views" renderLink={renderLink} links={[{ id: 'overview', label: 'Overview', href: `/users/${profile.id}${viewQuery}`, current: !activity }, { id: 'activity', label: 'Activity log', href: `/users/${profile.id}/activity${viewQuery}`, current: activity }]} />
    {activity ? <ActivityLog key={profile.id} logs={ownLogs} owner={owner} /> : <>
      <section className="page-section records-year" aria-labelledby="profile-year-heading">
        <div className="section-heading records-section-heading records-year-heading"><div><h2 className="paper-type-section" id="profile-year-heading">{profile.name}’s activity</h2></div>
          <div className="records-year-control" aria-label="Profile year"><Button variant="ghost" aria-label="Previous year" disabled={yearIndex === years.length - 1} onClick={() => changeYear(years[yearIndex + 1])}><ChevronLeftIcon className="paper-icon-default" /></Button><strong aria-live="polite">{year}</strong><Button variant="ghost" aria-label="Next year" disabled={yearIndex === 0} onClick={() => changeYear(years[yearIndex - 1])}><ChevronRightIcon className="paper-icon-default" /></Button></div>
        </div>
        {summary.entries ? <><dl className="stat-grid records-profile-stats"><div className="stat"><dt>Personal score</dt><dd>{formatNumber(summary.total)}</dd></div><div className="stat"><dt>Entries</dt><dd>{summary.entries}<small>{summary.days} active days</small></dd></div><div className="stat"><dt>Languages</dt><dd>{summary.languages.length}<small>{summary.languages[0] ? `${summary.languages[0][0]} leads` : 'A new language awaits'}</small></dd></div><div className="stat"><dt>Contests</dt><dd>{summary.contestIds.length}</dd></div></dl>
      <div className="records-rhythm" aria-labelledby="rhythm-heading">
        <div className="records-calendar-heading"><h3 id="rhythm-heading">Immersion rhythm</h3><Link className="text-link" to="/guide/scoring">How scoring works</Link></div>
        <div className="records-heatmap-scroll paper-focus-ring" role="region" aria-label={`Daily immersion calendar for ${year}; scroll for all months`} tabIndex={0}><HeatmapChart id={`profile-${profile.id}-${year}`} year={year} data={summary.daily} /></div>
        <p className="records-chart-caption">{summary.days} active days · Longest streak: {summary.longestStreak} {summary.longestStreak === 1 ? 'day' : 'days'}<span>Stronger color means more activity.</span></p>
      </div></> : <div className="records-empty-year"><h3>No activity recorded in {year}</h3><p>{owner ? 'A few pages or minutes are a good place to start. Your personal record grows with every entry, whether you join a contest or read on your own.' : 'Personal scores, the activity calendar and contest history will appear after the first entry for this year.'}</p>{owner ? <Link className={buttonClassName()} to="/logs/new">Log activity</Link> : null}</div>}</section>
      {summary.entries ? <div className="records-overview-columns">
          <section className="page-section" aria-labelledby="language-heading"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section" id="language-heading">Score by language</h2></div></div>
            {summary.languages.length ? <ol className="records-breakdown">{summary.languages.map(([language, score], index) => <li key={language}><div><span>{language}</span><strong>{formatNumber(score)} <small>points</small></strong></div><span className="records-bar-track" aria-hidden="true"><span style={{ width: `${summary.total ? score / summary.total * 100 : 0}%`, backgroundColor: chartPalette[index % chartPalette.length] }} /></span></li>)}</ol> : <p className="muted">No activity recorded in {year}.</p>}
          </section>
          <section className="page-section records-mix" aria-labelledby="mix-heading"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section" id="mix-heading">Score by activity</h2></div></div><dl className="records-activity-mix">{summary.activities.map(([name, score]) => <div key={name}><dt>{name}</dt><dd>{summary.total ? Math.round(score / summary.total * 100) : 0}%</dd></div>)}</dl>{!summary.entries ? <p className="muted">Reading and listening will appear after a first log.</p> : null}</section>
          <section className="page-section records-history" aria-labelledby="history-heading"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section" id="history-heading">Contest history</h2></div><Link className="text-link" to="/contests/official">All contests</Link></div>
            {summary.contestIds.length ? <ul className="records-contest-history">{summary.contestIds.map(id => { const contest = contests.find(item => item.id === id); const score = yearLogs.flatMap(log => log.submissions).filter(submission => submission.contestId === id).reduce((sum, item) => sum + item.score, 0); return <li key={id}><div><Link className="text-link" to={`/contests/${id}`}>{contest?.title ?? 'Contest'}</Link><p className="muted">{contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : 'Submitted activity'}</p></div><strong>{formatNumber(score)}<small>contest points</small></strong></li> })}</ul> : <p className="muted">No contest submissions in {year}. Personal activity still counts here.</p>}
          </section>
          <section className="page-section" aria-labelledby="recent-heading"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section" id="recent-heading">Recent activity</h2></div><Link className="text-link" to={`/users/${profile.id}/activity${viewQuery}`}>View all</Link></div>{yearLogs.length ? <ActivityRows logs={yearLogs.slice(0, 3)} compact /> : <p className="muted">No entries for this year.</p>}</section>
      </div> : null}
    </>}
  </>
}
