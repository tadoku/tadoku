import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { EmptyState, entryLink, listRows, metaLine, StatGrid } from 'paper-ui'
import { contestsHref } from '../lib/contests'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { memberDashboardQueries } from '../lib/home'
import type { Viewer } from '../lib/session'
import { ActivityCalendar } from './ActivityCalendar'
import { LogEntries } from './LogEntries'

type SignedIn = Extract<Viewer, { status: 'signed-in' }>

const recentCount = 5

/**
 * A signed-in learner's year under the home hero: the calendar, their newest updates and this year's contests,
 * unfinished ones first. Each part that can't be read is left out rather than shown as an error.
 */
const band = 'border-t border-rule pt-7'
const head = 'mb-4 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1'
const headTitle = 'm-0 font-display text-heading leading-section'
const rows = listRows()

export function MemberDashboard({ viewer, year, className }: { viewer: SignedIn; year: number; className?: string }) {
  const links = useLinks()
  const [statisticsQuery, activityQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  const recent = useQuery(activityQuery).data
  if (statistics?.status !== 'found') return null

  const { today, activity } = statistics
  const fresh = activity.ok && activity.data.total_updates === 0
  const logs = recent?.status === 'found' && recent.logs.ok ? recent.logs.data.logs.slice(0, recentCount) : null

  return (
    <div className={`grid gap-10 *:min-w-0 ${className ?? ''}`}>
      {fresh ? null : <section className={band} aria-labelledby="home-year-title">
        <div className={head}>
          <h2 id="home-year-title" className={headTitle}>Daily activity</h2>
          <a className="text-link hover:text-link-hover" href={links.profile(viewer.userId)}>Your profile</a>
        </div>
        {activity.ok ? (
          <ActivityCalendar userId={viewer.userId} year={year} today={today} activity={activity.data} labelledBy="home-year-title" surface="canvas" />
        ) : null}
      </section>}

      <div className={`${band} *:min-w-0`}>
        {logs ? (
          <section aria-labelledby="home-recent-title">
            <div className={head}>
              <h2 id="home-recent-title" className={headTitle}>Recent activity</h2>
              {logs.length ? <Link className="text-link hover:text-link-hover" to={links.profileUpdates(viewer.userId)}>All activity</Link> : null}
            </div>
            <LogEntries
              logs={logs}
              total={logs.length}
              page={1}
              pageHref={() => links.profileUpdates(viewer.userId)}
              empty={(
                <EmptyState
                  variant="plain"
                  headingLevel={3}
                  title="No activity yet"
                  description="A few pages or minutes are enough to start."
                />
              )}
            />
          </section>
        ) : null}
      </div>
    </div>
  )
}

/** The year's totals beside the home hero's headline. */
export function MemberYearTotals({ viewer, year }: { viewer: SignedIn; year: number }) {
  const [statisticsQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  if (statistics?.status !== 'found') return null

  const { today, activity, scores } = statistics
  if (activity.ok && activity.data.total_updates === 0) return null
  const total = scores.ok ? scores.data.scores.reduce((sum, score) => sum + score.score, 0) : undefined
  const todayScore = activity.ok ? activity.data.scores.find(day => day.date === today)?.score ?? 0 : undefined
  return (
    <section className="mt-9 border-t border-rule pt-6 max-tablet:mt-0" aria-labelledby="home-totals-title">
      <h2 id="home-totals-title" className="mb-4 mt-0 text-ui font-semibold leading-150 text-muted">Your {year}</h2>
      <StatGrid
        className="grid-cols-halves"
        stats={[
          ...(total === undefined ? [] : [{ label: 'Overall score', value: scoreFormat.format(total) }]),
          ...(todayScore === undefined ? [] : [{ label: 'Today', value: scoreFormat.format(todayScore), unit: pointsUnit(todayScore) }]),
        ]}
      />
    </section>
  )
}

/** This year's contests, unfinished ones first, for the home hero's left column. */
export function MemberContests({ viewer, year }: { viewer: SignedIn; year: number }) {
  const links = useLinks()
  const [statisticsQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  if (statistics?.status !== 'found') return null

  const { today, contests } = statistics
  const ordered = contests.ok
    ? [...contests.data.contests].sort((a, b) => Number(a.contest_end < today) - Number(b.contest_end < today)).slice(0, recentCount)
    : null
  if (!ordered) return null

  return (
    <section aria-labelledby="home-contests-title" className="mt-8">

            <div className={head}>
              <h2 id="home-contests-title" className={headTitle}>Your contests</h2>
              {ordered.length ? <Link className="text-link hover:text-link-hover" to={contestsHref.joined(year)}>All joined</Link> : null}
            </div>
            {ordered.length ? (
              <ul className={rows.root()}>
                {ordered.map(contest => (
                  <li key={contest.id} className={rows.item()}>
                    <a className={entryLink({ className: 'profile-entry-title' })} href={links.contestProfile(contest.id, viewer.userId)}>{contest.title}</a>
                    <span className={metaLine({ className: 'profile-entry-meta' })}>
                      {[formatPeriod(contest), ...contest.languages, contest.contest_end < today ? 'Ended' : null].filter(Boolean).join(' · ')}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <EmptyState
                variant="plain"
                headingLevel={3}
                title={`You haven’t joined a contest in ${year}`}
                actions={<Link className="text-link hover:text-link-hover" to={contestsHref.official()}>Browse official contests</Link>}
              />
            )}
    </section>
  )
}
