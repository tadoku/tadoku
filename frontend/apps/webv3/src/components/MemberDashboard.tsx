import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { EmptyState, StatGrid } from 'paper-ui'
import { contestsHref } from '../lib/contests'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { memberDashboardQueries } from '../lib/home'
import { activeDays } from '../lib/profile'
import type { Viewer } from '../lib/session'
import { ActivityCalendar } from './ActivityCalendar'
import { LogEntries } from './LogEntries'

type SignedIn = Extract<Viewer, { status: 'signed-in' }>

const recentCount = 5

/**
 * A signed-in learner's year under the home hero: totals and the calendar, their newest updates and this year's
 * contests, unfinished ones first. Each part that can't be read is left out rather than shown as an error.
 */
export function MemberDashboard({ viewer, year }: { viewer: SignedIn; year: number }) {
  const links = useLinks()
  const [statisticsQuery, activityQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  const recent = useQuery(activityQuery).data
  if (statistics?.status !== 'found') return null

  const { today, activity, scores, contests } = statistics
  const total = scores.ok ? scores.data.scores.reduce((sum, score) => sum + score.score, 0) : undefined
  const todayScore = activity.ok ? activity.data.scores.find(day => day.date === today)?.score ?? 0 : undefined
  const logs = recent?.status === 'found' && recent.logs.ok ? recent.logs.data.logs.slice(0, recentCount) : null
  const ordered = contests.ok
    ? [...contests.data.contests].sort((a, b) => Number(a.contest_end < today) - Number(b.contest_end < today)).slice(0, recentCount)
    : null

  return (
    <div className="home-dashboard">
      <section className="home-dashboard__year" aria-labelledby="home-year-title">
        <div className="home-dashboard__head">
          <h2 id="home-year-title">Your {year}</h2>
          <a className="text-link" href={links.profile(viewer.userId)}>Your profile</a>
        </div>
        <StatGrid
          className="home-dashboard__stats"
          stats={[
            ...(total === undefined ? [] : [{ label: 'Points this year', value: scoreFormat.format(total) }]),
            ...(todayScore === undefined ? [] : [{ label: 'Today', value: scoreFormat.format(todayScore), detail: pointsUnit(todayScore) }]),
            ...(activity.ok ? [
              { label: 'Updates', value: activity.data.total_updates },
              { label: 'Active days', value: activeDays(activity.data.scores) },
            ] : []),
          ]}
        />
        {activity.ok ? (
          <ActivityCalendar userId={viewer.userId} year={year} today={today} activity={activity.data} labelledBy="home-year-title" />
        ) : null}
      </section>

      <div className="home-dashboard__columns">
        {logs ? (
          <section aria-labelledby="home-recent-title">
            <div className="home-dashboard__head">
              <h2 id="home-recent-title">Recent updates</h2>
              {logs.length ? <Link className="text-link" to={links.profileUpdates(viewer.userId)}>All updates</Link> : null}
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
                  title="No updates yet"
                  actions={<a className="text-link" href={links.logs}>Log your first activity</a>}
                />
              )}
            />
          </section>
        ) : null}
        {ordered ? (
          <section aria-labelledby="home-contests-title">
            <div className="home-dashboard__head">
              <h2 id="home-contests-title">Your contests</h2>
              {ordered.length ? <Link className="text-link" to={contestsHref.joined(year)}>All joined</Link> : null}
            </div>
            {ordered.length ? (
              <ul className="profile-contests">
                {ordered.map(contest => (
                  <li key={contest.id}>
                    <a className="profile-entry-title" href={links.contestProfile(contest.id, viewer.userId)}>{contest.title}</a>
                    <span className="profile-entry-meta">
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
                actions={<Link className="text-link" to={contestsHref.official()}>Browse official contests</Link>}
              />
            )}
          </section>
        ) : null}
      </div>
    </div>
  )
}
