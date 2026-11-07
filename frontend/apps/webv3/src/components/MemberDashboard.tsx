import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { EmptyState, StatGrid, linkStyles } from 'paper-ui'
import { border, breakpoints, color, font, fontSize, lineHeight, space } from 'paper-ui/tokens.stylex'
import { contestsHref } from '../lib/contests'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { memberDashboardQueries } from '../lib/home'
import { activeDays } from '../lib/profile'
import type { Viewer } from '../lib/session'
import { ActivityCalendar } from './ActivityCalendar'
import { LogEntries } from './LogEntries'
import { entryStyles, entryTitle } from './page-styles'

const ruled = {
  borderBlockStartColor: color.ruleDefault,
  borderBlockStartStyle: 'solid',
  borderBlockStartWidth: border.static,
  minInlineSize: 0,
  paddingBlockStart: space[28],
} as const

const styles = stylex.create({
  dashboard: { gap: space[40], display: 'grid', marginBlockStart: space[48], },
  year: ruled,
  columns: {
    ...ruled,
    columnGap: space[56],
    display: 'grid',
    gridTemplateColumns: { default: 'repeat(2, minmax(0, 1fr))', [breakpoints.lg]: 'minmax(0, 1fr)' },
    rowGap: space[32],
  },
  column: { minInlineSize: 0 },
  head: {
    alignItems: 'baseline',
    columnGap: space[16],
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    marginBlockEnd: space[16],
    rowGap: space[4],
  },
  title: { margin: 0, fontFamily: font.display, fontSize: fontSize[22], lineHeight: lineHeight[30], },
})

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
    <div {...stylex.props(styles.dashboard)}>
      <section {...stylex.props(styles.year)} aria-labelledby="home-year-title">
        <div {...stylex.props(styles.head)}>
          <h2 id="home-year-title" {...stylex.props(styles.title)}>Your {year}</h2>
          <a {...stylex.props(linkStyles.base)} href={links.profile(viewer.userId)}>Your profile</a>
        </div>
        <StatGrid
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
          <ActivityCalendar userId={viewer.userId} year={year} today={today} activity={activity.data} labelledBy="home-year-title" surface="canvas" />
        ) : null}
      </section>

      <div {...stylex.props(styles.columns)}>
        {logs ? (
          <section {...stylex.props(styles.column)} aria-labelledby="home-recent-title">
            <div {...stylex.props(styles.head)}>
              <h2 id="home-recent-title" {...stylex.props(styles.title)}>Recent updates</h2>
              {logs.length ? <Link {...stylex.props(linkStyles.base)} to={links.profileUpdates(viewer.userId)}>All updates</Link> : null}
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
                  actions={<a {...stylex.props(linkStyles.base)} href={links.logs}>Log your first activity</a>}
                />
              )}
            />
          </section>
        ) : null}
        {ordered ? (
          <section {...stylex.props(styles.column)} aria-labelledby="home-contests-title">
            <div {...stylex.props(styles.head)}>
              <h2 id="home-contests-title" {...stylex.props(styles.title)}>Your contests</h2>
              {ordered.length ? <Link {...stylex.props(linkStyles.base)} to={contestsHref.joined(year)}>All joined</Link> : null}
            </div>
            {ordered.length ? (
              <ul {...stylex.props(entryStyles.contestList)}>
                {ordered.map(contest => (
                  <li key={contest.id} {...stylex.props(entryStyles.contestItem)}>
                    <a {...stylex.props(...entryTitle)} href={links.contestProfile(contest.id, viewer.userId)}>{contest.title}</a>
                    <span {...stylex.props(entryStyles.meta)}>
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
                actions={<Link {...stylex.props(linkStyles.base)} to={contestsHref.official()}>Browse official contests</Link>}
              />
            )}
          </section>
        ) : null}
      </div>
    </div>
  )
}
