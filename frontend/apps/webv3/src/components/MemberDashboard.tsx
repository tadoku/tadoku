import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { EmptyState, entryLink, listRows, metaLine, RecentHeatmap, StatGrid } from 'paper-ui'
import { CheckCircleIcon, iconClassName } from 'paper-ui/icons'
import type { ReactNode } from 'react'
import { contestsHref } from '../lib/contests'
import { dayAndMonth, formatDay, pointsUnit, scoreFormat, utcToday } from '../lib/format'
import { useLinks } from '../lib/links'
import { lastYearStatisticsQuery, memberDashboardQueries } from '../lib/home'
import { activeDaysGoal, activityWeeks, homeActivity, readsLastYear } from '../lib/recent-activity'
import type { Viewer } from '../lib/session'
import { LogEntries } from './LogEntries'
import { AppLink } from './router-link'

type SignedIn = Extract<Viewer, { status: 'signed-in' }>

const runningCount = 5

const band = 'border-t border-rule pt-7'
const head = 'mb-4 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1'
const headTitle = 'm-0 font-display text-heading leading-section'
const rows = listRows()

/**
 * Beside the home page's contest board: the learner's past 30 days over their activity map, or the getting-going
 * checklist until they've cleared it. Left out when the year's activity can't be read.
 */
export function MemberActivity({ viewer, year }: { viewer: SignedIn; year: number }) {
  const links = useLinks()
  const statistics = useQuery(memberDashboardQueries(viewer, year)[0]).data
  const today = statistics?.status === 'found' ? statistics.today : utcToday()
  const lastYear = useQuery({ ...lastYearStatisticsQuery(viewer, year), enabled: readsLastYear(today) }).data
  if (statistics?.status !== 'found' || !statistics.activity.ok) return null

  // Last year's activity is a bonus: when it can't be read, the window and history simply start on 1 January.
  const previous = lastYear?.status === 'found' && readsLastYear(today) ? lastYear : undefined
  const days = [...(previous?.activity.ok ? previous.activity.data.scores : []), ...statistics.activity.data.scores]
  const joinedContest = [statistics, previous].some(read => read?.contests.ok && read.contests.data.total > 0)
  const activity = homeActivity(new Map(days.map(day => [day.date, day.score])), today, joinedContest)
  const title = activity.kind === 'active' ? 'Past 30 days' : 'Getting going'

  return (
    <>
      <div className="mb-2 flex flex-wrap items-baseline justify-between gap-x-4">
        <h1 id="home-title" className="m-0 text-ui font-semibold leading-150 text-muted">{title}</h1>
        <AppLink className="text-link hover:text-link-hover" href={links.profile(viewer.userId)}>Your profile</AppLink>
      </div>
      {activity.kind === 'active' ? (
        <>
          <p className={bigLine}>
            <span className={bigNumber}>{scoreFormat.format(activity.total)}</span><span className={bigUnit}>{pointsUnit(activity.total)}</span>
          </p>
          <StatGrid
            className="mt-6 grid-cols-thirds"
            stats={[
              { label: 'Today', value: scoreFormat.format(activity.today), unit: pointsUnit(activity.today) },
              { label: 'Streak', value: activity.streak, unit: activity.streak === 1 ? 'day' : 'days' },
              { label: 'Active days', value: activity.activeDays, unit: 'of 30' },
            ]}
          />
          <RecentHeatmap
            className="mt-7"
            weeks={activityWeeks}
            today={today}
            getHref={date => links.profileUpdates(viewer.userId, date)}
            data={days.map(day => ({
              date: day.date,
              value: day.score,
              tooltip: `${scoreFormat.format(day.score)} ${pointsUnit(day.score)} on ${formatDay(day.date)}`,
            }))}
          />
        </>
      ) : (
        <>
          <p className={bigLine}>
            <span className={bigNumber}>{activity.activeDays}</span><span className={bigUnit}>of {activeDaysGoal} active days</span>
          </p>
          <ul className={rows.root({ className: 'mt-6' })}>
            <Step
              done={activity.steps.logged}
              title="Log your first activity"
              detail="A few pages or minutes of anything you read or listen to."
              action={<AppLink className={stepLink} href={links.logs}>Log activity</AppLink>}
            />
            <Step
              done={activity.steps.loggedToday}
              title="Log something today"
              detail="Short daily sessions add up faster than long weekly ones."
              action={<AppLink className={stepLink} href={links.logs}>Log activity</AppLink>}
            />
            <Step
              done={activity.steps.joinedContest}
              title="Join a contest"
              detail="A round gives you a finish line and people to read with."
              action={<Link className={stepLink} to={contestsHref.official()}>Browse contests</Link>}
            />
            <Step
              done={activity.steps.twoDaysInARow}
              title="Log two days in a row"
              detail={activity.steps.loggedToday ? 'Today counts. Log again tomorrow.' : 'Two days back to back start a streak.'}
              action={null}
            />
          </ul>
        </>
      )}
    </>
  )
}

const bigLine = 'm-0 flex flex-wrap items-baseline gap-x-2'
const bigNumber = 'font-display text-hero leading-120 tracking-hero tabular-nums'
const bigUnit = 'text-lead text-muted'
const stepLink = 'text-link hover:text-link-hover'

function Step({ done, title, detail, action }: { done: boolean; title: string; detail: string; action: ReactNode }) {
  return (
    <li className={rows.item({ className: 'flex items-start gap-3' })}>
      {done
        ? <CheckCircleIcon className={iconClassName('default', 'mt-0.5 text-success')} aria-hidden="true" />
        : <span className={iconClassName('default', 'mt-0.5 rounded-pill border-2 border-rule-strong')} aria-hidden="true" />}
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-baseline justify-between gap-x-4">
          <span className={done ? 'text-muted line-through' : 'font-semibold text-ink'}>
            <span className="sr-only">{done ? 'Done: ' : 'Not done: '}</span>
            {title}
          </span>
          {done ? null : action}
        </div>
        {done ? null : <span className={metaLine()}>{detail}</span>}
      </div>
    </li>
  )
}

/**
 * Under the home page's top: this year's running contests beside the newest updates. Each part that can't be read is
 * left out rather than shown as an error.
 */
export function MemberDashboard({ viewer, year, boardContestId, className }: {
  viewer: SignedIn
  year: number
  /** The contest the board above shows, which the contest list leaves out. */
  boardContestId?: string
  className?: string
}) {
  const links = useLinks()
  const [statisticsQuery, activityQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  const recent = useQuery(activityQuery).data
  if (statistics?.status !== 'found') return null

  const logs = recent?.ok ? recent.data.logs : null

  return (
    <div className={`grid gap-10 *:min-w-0 ${className ?? ''}`}>
      <div className={`${band} grid grid-cols-halves items-start gap-x-16 gap-y-10 max-lg:gap-x-8 max-tablet:grid-cols-single *:min-w-0`}>
        <MemberContests viewer={viewer} year={year} boardContestId={boardContestId} />
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

/**
 * This year's running and upcoming contests, ending soonest first, without the one on the board. Only the year's
 * first page of registrations is read, so a learner with more may miss a running contest.
 */
function MemberContests({ viewer, year, boardContestId }: { viewer: SignedIn; year: number; boardContestId?: string }) {
  const links = useLinks()
  const [statisticsQuery] = memberDashboardQueries(viewer, year)
  const statistics = useQuery(statisticsQuery).data
  if (statistics?.status !== 'found' || !statistics.contests.ok) return null

  const { today } = statistics
  const joined = statistics.contests.data.contests
  const running = joined
    .filter(contest => contest.contest_end >= today && contest.id !== boardContestId)
    .sort((a, b) => a.contest_end.localeCompare(b.contest_end))
  const onBoard = joined.some(contest => contest.id === boardContestId && contest.contest_end >= today)

  return (
    <section aria-labelledby="home-contests-title">
      <div className={head}>
        <h2 id="home-contests-title" className={headTitle}>Your contests</h2>
        {joined.length ? <Link className="text-link hover:text-link-hover" to={contestsHref.joined()}>All joined</Link> : null}
      </div>
      {running.length ? (
        <ul className={rows.root()}>
          {running.slice(0, runningCount).map(contest => (
            <li key={contest.id} className={rows.item({ className: 'flex flex-wrap items-baseline justify-between gap-x-4' })}>
              <AppLink className={entryLink({ className: 'profile-entry-title min-w-0' })} href={links.contestProfile(contest.id, viewer.userId)}>
                {contest.title}
              </AppLink>
              <span className={metaLine({ className: 'mt-0' })}>
                {contest.contest_start > today ? `Starts ${dayAndMonth(contest.contest_start)}` : `Ends ${dayAndMonth(contest.contest_end)}`}
              </span>
            </li>
          ))}
          {running.length > runningCount ? (
            <li className={rows.item({ className: 'text-metadata text-muted' })}>{running.length - runningCount} more running</li>
          ) : null}
        </ul>
      ) : (
        <EmptyState
          variant="plain"
          headingLevel={3}
          title={onBoard ? 'No other contests running' : joined.length ? 'No contests running right now' : `You haven’t joined a contest in ${year}`}
          actions={<Link className="text-link hover:text-link-hover" to={contestsHref.community()}>Browse contests</Link>}
        />
      )}
    </section>
  )
}
