import * as stylex from '@stylexjs/stylex'
import { useStore } from '@tanstack/react-form'
import { useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { a11y, AppliedFilters, Breadcrumb, Button, buttonStyles, chartPalette, DescriptionList, EmptyState, HeatmapChart, iconSizes, linkStyles, SectionHeading, StatGrid, Tabbar, typography, usePaperForm } from 'paper-ui'
import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon, PlusIcon } from 'paper-ui/icons'
import { formatDate, formatNumber, scoreLog, users, type SampleLog } from '../data'
import { contestEnd, contestStart, formatDateRange } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground, useScenario } from '../state'
import { recordsStyles } from './records.styles'
import { utilities } from '../utilities.styles'
import { appStyles } from '../app.styles'
import { recordsDescendants as descendantStyles } from './records-descendants.styles'

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
  return <ol {...stylex.props(recordsStyles.recordsEntries, compact && recordsStyles.recordsEntries_compact)} data-entries="">
    {logs.map(log => <li {...stylex.props(descendantStyles.recordsEntriesU_uLi)} key={log.id}>
      <time dateTime={log.date} {...stylex.props(recordsStyles.recordsEntryDate)}>{formatDate(log.date)}</time>
      <div {...stylex.props(utilities.tw_min_w_0)}>
        <Link {...stylex.props(recordsStyles.recordsEntryTitle, linkStyles.base)} to={`/logs/${log.id}`}>{log.title || `${log.language} ${log.activity.toLocaleLowerCase()}`}</Link>
        <p {...stylex.props(recordsStyles.recordsEntryMeta)}>{log.language} · {log.activity} · {formatNumber(log.amount)} {log.unit}</p>
        {!compact && log.note ? <p {...stylex.props(recordsStyles.recordsEntryNote)}>{log.note}</p> : null}
      </div>
      <span {...stylex.props(recordsStyles.recordsEntryScore)} aria-label={`${formatNumber(scoreLog(log))} personal points`}>{formatNumber(scoreLog(log))}<small {...stylex.props(descendantStyles.recordsEntryScoreU_uSmall)}>points</small></span>
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

  return <section {...stylex.props(appStyles.pageSection)} aria-labelledby="activity-heading">
    <SectionHeading level={2} id="activity-heading" title="Activity log" actions={logs.length ? <><Button ref={filtersToggle} variant="ghost" aria-expanded={filtersOpen} aria-controls="activity-filters" onClick={() => setFiltersOpen(!filtersOpen)} trailingIcon={<ChevronDownIcon {...stylex.props(iconSizes.compact, filtersOpen && utilities.tw_rotate_180)} />}>Filters{appliedFilters.length ? ` (${appliedFilters.length})` : ''}</Button>{owner ? <Link {...stylex.props(...buttonStyles())} to="/logs/new"><PlusIcon {...stylex.props(iconSizes.default)} aria-hidden="true" />Log activity</Link> : null}</> : null} />
    {logs.length ? <>
    <div id="activity-filters" hidden={!filtersOpen}><methods.AppForm>
      <form onSubmit={event => event.preventDefault()} {...stylex.props(recordsStyles.recordsFilterPanel)}>
        <methods.AppField name="period">{field => <field.Select label="Period" options={[{ value: 'all', label: 'All time' }, ...years.map(year => ({ value: String(year), label: String(year) }))]} />}</methods.AppField>
        <methods.AppField name="languages">{field => <field.AutocompleteMultiInput label="Languages" placeholder="All languages" options={languages} format={value => value} getId={value => value} />}</methods.AppField>
        <methods.AppField name="activities">{field => <field.AutocompleteMultiInput label="Activities" placeholder="All activities" options={['Reading', 'Listening']} format={value => value} getId={value => value} />}</methods.AppField>
        <methods.AppField name="query">{field => <field.Input label="Search activity" type="search" placeholder="Title, note or tag" />}</methods.AppField>
      </form>
    </methods.AppForm></div>
    <div {...stylex.props(recordsStyles.recordsFilterSummary)}><AppliedFilters filters={appliedFilters} onClear={reset} />
    <p {...stylex.props(recordsStyles.recordsResultCount)} role="status">{filtered.length} of {logs.length} entries</p></div>
    {filtered.length ? <ActivityRows logs={filtered} /> : <EmptyState headingLevel={3} title="No activity matches these filters." description="Try another title or include more dates and languages." actions={<Button variant="outline" onClick={reset}>Clear search and filters</Button>} />}
    </> : <EmptyState headingLevel={3} title="No activity yet" description={owner ? 'Start with what you read or listened to today. A title and a few pages or minutes are enough.' : 'Reading and listening entries will appear here once they have been recorded.'} actions={owner ? <Link {...stylex.props(...buttonStyles())} to="/logs/new">Log activity</Link> : null} />}
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

  if (!profile) return <EmptyState headingLevel={1} title="Profile not found" description="This account is not in the playground’s sample community." actions={<Link {...stylex.props(...buttonStyles())} to="/leaderboard/latest">Explore the leaderboard</Link>} />

  return <>
    <Breadcrumb items={[{ id: 'home', label: 'Home', href: '/' }, { id: 'profile', label: profile.name }]} renderLink={renderLink} />
    <header {...stylex.props(recordsStyles.recordsProfileHeader)}>
      <div {...stylex.props(recordsStyles.recordsIdentity)}><span {...stylex.props(recordsStyles.recordsAvatar)} aria-hidden="true">{profile.name.slice(0, 2).toLocaleUpperCase()}</span><div {...stylex.props(descendantStyles.recordsIdentityU_uDiv)}><h1 {...stylex.props(typography.page, descendantStyles.recordsIdentityU_uH1)}>{profile.name}</h1><p {...stylex.props(appStyles.muted, descendantStyles.recordsIdentityU_uP)}>Tracking since {String(profile.joined).slice(0, 4)}</p></div></div>
      {owner ? <Link {...stylex.props(...buttonStyles({ variant: 'outline' }))} to="/settings">Edit profile</Link> : null}
    </header>
    {contestContext ? <p {...stylex.props(recordsStyles.recordsProfileContext)}>Viewing {profile.name}’s personal record from <Link {...stylex.props(linkStyles.base)} to={`/contests/${contestContext.id}/leaderboard`}>{contestContext.title} standings</Link>. Contest scores are shown on that board.</p> : null}
    <Tabbar label="Profile views" renderLink={renderLink} links={[{ id: 'overview', label: 'Overview', href: `/users/${profile.id}${viewQuery}`, current: !activity }, { id: 'activity', label: 'Activity log', href: `/users/${profile.id}/activity${viewQuery}`, current: activity }]} />
    {activity ? <ActivityLog key={profile.id} logs={ownLogs} owner={owner} /> : <>
      <section {...stylex.props(appStyles.pageSection, recordsStyles.recordsYear)} aria-labelledby="profile-year-heading">
        <SectionHeading level={2} id="profile-year-heading" title={`${profile.name}’s activity`} actions={<div {...stylex.props(recordsStyles.recordsYearControl)} aria-label="Profile year"><Button variant="ghost" aria-label="Previous year" disabled={yearIndex === years.length - 1} onClick={() => changeYear(years[yearIndex + 1])}><ChevronLeftIcon {...stylex.props(iconSizes.default)} /></Button><strong {...stylex.props(descendantStyles.recordsYearControlU_uStrong)} aria-live="polite">{year}</strong><Button variant="ghost" aria-label="Next year" disabled={yearIndex === 0} onClick={() => changeYear(years[yearIndex - 1])}><ChevronRightIcon {...stylex.props(iconSizes.default)} /></Button></div>} />
        {summary.entries ? <><StatGrid xstyle={[recordsStyles.recordsProfileStats]} stats={[
          { label: 'Personal score', value: formatNumber(summary.total) },
          { label: 'Entries', value: summary.entries, detail: `${summary.days} active days` },
          { label: 'Languages', value: summary.languages.length, detail: summary.languages[0] ? `${summary.languages[0][0]} leads` : 'A new language awaits' },
          { label: 'Contests', value: summary.contestIds.length },
        ]} />
      <div {...stylex.props(recordsStyles.recordsRhythm)} aria-labelledby="rhythm-heading">
        <SectionHeading level={3} id="rhythm-heading" title="Immersion rhythm" actions={<Link {...stylex.props(linkStyles.base)} to="/guide/scoring">How scoring works</Link>} />
        <div {...stylex.props(recordsStyles.recordsHeatmapScroll, a11y.focusRing)} role="region" aria-label={`Daily immersion calendar for ${year}; scroll for all months`} tabIndex={0}><HeatmapChart id={`profile-${profile.id}-${year}`} year={year} data={summary.daily} /></div>
        <p {...stylex.props(recordsStyles.recordsChartCaption)}>{summary.days} active days · Longest streak: {summary.longestStreak} {summary.longestStreak === 1 ? 'day' : 'days'}<span>Stronger color means more activity.</span></p>
      </div></> : <EmptyState variant="plain" xstyle={[recordsStyles.recordsEmptyYear]} headingLevel={3} title={`No activity recorded in ${year}`} description={owner ? 'A few pages or minutes are a good place to start. Your personal record grows with every entry, whether you join a contest or read on your own.' : 'Personal scores, the activity calendar and contest history will appear after the first entry for this year.'} actions={owner ? <Link {...stylex.props(...buttonStyles())} to="/logs/new">Log activity</Link> : null} />}</section>
      {summary.entries ? <div {...stylex.props(recordsStyles.recordsOverviewColumns)}>
          <section {...stylex.props(appStyles.pageSection, descendantStyles.recordsOverviewColumnsU_uAny)} aria-labelledby="language-heading"><SectionHeading level={2} id="language-heading" title="Score by language" />
            {summary.languages.length ? <ol {...stylex.props(recordsStyles.recordsBreakdown)}>{summary.languages.map(([language, score], index) => <li {...stylex.props(descendantStyles.recordsBreakdownU_uLi)} key={language}><div {...stylex.props(descendantStyles.recordsBreakdownU_uLiU_uDiv)}><span>{language}</span><strong {...stylex.props(descendantStyles.recordsBreakdownU_uStrong)}>{formatNumber(score)} <small {...stylex.props(descendantStyles.recordsBreakdownU_uSmall)}>points</small></strong></div><span {...stylex.props(recordsStyles.recordsBarTrack)} aria-hidden="true"><span {...stylex.props(descendantStyles.recordsBarTrackU_uSpan, recordsStyles.barFill(`${summary.total ? score / summary.total * 100 : 0}%`, chartPalette[index % chartPalette.length]))} /></span></li>)}</ol> : <p {...stylex.props(appStyles.muted)}>No activity recorded in {year}.</p>}
          </section>
          <section {...stylex.props(appStyles.pageSection, descendantStyles.recordsOverviewColumnsU_uAny)} aria-labelledby="mix-heading"><SectionHeading level={2} id="mix-heading" title="Score by activity" /><DescriptionList layout="inline" items={summary.activities.map(([name, score]) => ({ label: name, value: `${summary.total ? Math.round(score / summary.total * 100) : 0}%` }))} />{!summary.entries ? <p {...stylex.props(appStyles.muted)}>Reading and listening will appear after a first log.</p> : null}</section>
          <section {...stylex.props(appStyles.pageSection, descendantStyles.recordsOverviewColumnsU_uAny)} aria-labelledby="history-heading"><SectionHeading level={2} id="history-heading" title="Contest history" actions={<Link {...stylex.props(linkStyles.base)} to="/contests/official">All contests</Link>} />
            {summary.contestIds.length ? <ul {...stylex.props(recordsStyles.recordsContestHistory)}>{summary.contestIds.map(id => { const contest = contests.find(item => item.id === id); const score = yearLogs.flatMap(log => log.submissions).filter(submission => submission.contestId === id).reduce((sum, item) => sum + item.score, 0); return <li {...stylex.props(descendantStyles.recordsContestHistoryU_uLi)} key={id}><div><Link {...stylex.props(linkStyles.base)} to={`/contests/${id}`}>{contest?.title ?? 'Contest'}</Link><p {...stylex.props(appStyles.muted, descendantStyles.recordsContestHistoryU_uP)}>{contest ? formatDateRange(contestStart(contest.start), contestEnd(contest.end)) : 'Submitted activity'}</p></div><strong {...stylex.props(descendantStyles.recordsContestHistoryU_uStrong)}>{formatNumber(score)}<small {...stylex.props(descendantStyles.recordsContestHistoryU_uStrongU_uSmall)}>contest points</small></strong></li> })}</ul> : <p {...stylex.props(appStyles.muted)}>No contest submissions in {year}. Personal activity still counts here.</p>}
          </section>
          <section {...stylex.props(appStyles.pageSection, descendantStyles.recordsOverviewColumnsU_uAny)} aria-labelledby="recent-heading"><SectionHeading level={2} id="recent-heading" title="Recent activity" actions={<Link {...stylex.props(linkStyles.base)} to={`/users/${profile.id}/activity${viewQuery}`}>View all</Link>} />{yearLogs.length ? <ActivityRows logs={yearLogs.slice(0, 3)} compact /> : <p {...stylex.props(appStyles.muted)}>No entries for this year.</p>}</section>
      </div> : null}
    </>}
  </>
}
