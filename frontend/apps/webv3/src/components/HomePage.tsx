import { useQuery } from '@tanstack/react-query'
import { BookOpenIcon, LanguageIcon, TrophyIcon, iconClassName } from 'paper-ui/icons'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, Loading, SectionHeading, StatGrid, Surface, Table, buttonClassName } from 'paper-ui'
import type { ReactNode } from 'react'
import { homeQueryOptions, type HomeData } from '../lib/home'
import { formatDate, formatPeriod, numberFormat, pointsUnit } from '../lib/format'
import { describeGoal, useStoredGoal } from '../lib/goal'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { MemberDashboard } from './MemberDashboard'
import { viewerQueryOptions, type Viewer } from '../lib/session'

const dayNumber = (from: string, to: string) => Math.round((Date.parse(to) - Date.parse(from)) / 86_400_000) + 1

type Phase = 'open' | 'live' | 'upcoming' | 'ended' | 'none' | 'unavailable' | 'loading'

function phaseOf(data: HomeData | undefined, failed: boolean, today: string): Phase {
  if (failed) return 'unavailable'
  if (!data) return 'loading'
  const contest = data.latestContest
  if (!contest) return 'none'
  if (today < contest.contest_start) return 'upcoming'
  if (today > contest.contest_end) return 'ended'
  return today <= contest.registration_end ? 'open' : 'live'
}

const guestHeadline = <>Read more.<br /> Listen more.<br /> Keep going together.</>

const promise: Record<Phase, { headline: ReactNode; lead: string }> = {
  open: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.' },
  live: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Log what you read and listen to, and follow the community’s progress in the current round.' },
  upcoming: { headline: 'Make time for the next round.', lead: 'Choose your languages now and find something you want to read or listen to. Start the next immersion round with the community.' },
  ended: { headline: 'Keep immersing between rounds.', lead: 'Contests add momentum, but the habit is yours year-round. Track what you read and listen to today, then bring that rhythm into the next official round.' },
  none: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Log what you read and listen to, and join an immersion contest when one suits you.' },
  unavailable: { headline: 'Your immersion can keep going.', lead: 'The contest board is temporarily unavailable. You can still log what you read and listen to today.' },
  loading: { headline: guestHeadline, lead: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.' },
}

export function HomePage() {
  const links = useLinks()
  const { data, dataUpdatedAt, isError, refetch, isFetching } = useQuery(homeQueryOptions())
  // The fetch time travels with the dehydrated data, so server and client agree on the contest day.
  const today = new Date(dataUpdatedAt).toISOString().slice(0, 10)
  const phase = phaseOf(data, isError, today)
  const contest = data?.latestContest ?? null
  const joinable = contest && (phase === 'open' || phase === 'upcoming')
  const { data: viewer } = useQuery(viewerQueryOptions())
  const member = viewer?.status === 'signed-in' ? viewer : null

  return (
    <div className="home-page">
      <section className="home-hero" aria-labelledby="home-title">
        {member
          ? <MemberPromise member={member} data={data} phase={phase} />
          : (
            <div className="home-promise">
              {phase === 'unavailable' && <p className="home-status" data-tone="warning">Contest updates are delayed</p>}
              <h1 id="home-title">{promise[phase].headline}</h1>
              <p className="home-lead">{promise[phase].lead}</p>
              <div className="home-actions">
                {joinable
                  ? <a href={links.joinContest(contest.id)} className={buttonClassName()}>Join {contest.title}</a>
                  : <a href={links.logs} className={buttonClassName()}>Log activity</a>}
                <SetGoalLink />
              </div>
              {joinable && <p className="home-fineprint">Free to join. Registration closes {formatDate(contest.registration_end)}.</p>}
              <DailyGoal />
            </div>
          )}

        <Surface as="section" elevation="showcase" className="home-ledger" aria-labelledby="home-board-title">
          <ContestBoard data={data} phase={phase} today={today} retrying={isFetching} retry={() => void refetch()} />
        </Surface>
      </section>

      {member ? <MemberDashboard viewer={member} year={Number(today.slice(0, 4))} /> : (
        <section className="home-how" aria-labelledby="home-how-title">
          <h2 id="home-how-title">How Tadoku works</h2>
          <ol>
            <li>
              <span className="home-how__icon" aria-hidden="true"><LanguageIcon className={iconClassName()} /></span>
              <h3>Choose your language</h3><p>Join an official round or track immersion at your own pace between contests.</p>
            </li>
            <li>
              <span className="home-how__icon" aria-hidden="true"><BookOpenIcon className={iconClassName()} /></span>
              <h3>Read and listen</h3><p>Books, manga, games, podcasts, video: whatever keeps you engaged.</p>
            </li>
            <li>
              <span className="home-how__icon" aria-hidden="true"><TrophyIcon className={iconClassName()} /></span>
              <h3>Log your progress</h3><p>Record what you did and see where it counts.</p>
            </li>
          </ol>
        </section>
      )}

      {/* While a round is running the card covers it; the next round earns a row once it is scheduled. */}
      {phase !== 'loading' && phase !== 'unavailable' && (data?.nextContest || (phase !== 'open' && phase !== 'live')) && (
        <section className="home-strip" aria-labelledby="home-next-title">
          <div>
            <h2 id="home-next-title">Next official round</h2>
            <p>{data?.nextContest ? formatPeriod(data.nextContest) : 'Not announced yet'}</p>
          </div>
          {data?.nextContest && <p>Registration closes {formatDate(data.nextContest.registration_end)}.</p>}
          <a className="text-link" href={links.contests}>See all contests</a>
        </section>
      )}

    </div>
  )
}

/** The goal is kept in the browser, so guests can set one too. Until one is saved, the hero's actions offer to set it. */
function SetGoalLink() {
  return useStoredGoal() ? null : <Link className="text-link" to="/plan">Set a daily goal</Link>
}

function DailyGoal() {
  const goal = useStoredGoal()
  if (!goal) return null

  return (
    <section className="home-goal" aria-labelledby="home-goal-title">
      <h2 id="home-goal-title">Daily goal</h2>
      <p>{describeGoal(goal)} <Link className="text-link" to="/plan">Change goal</Link></p>
    </section>
  )
}

const ordinal = new Intl.PluralRules('en', { type: 'ordinal' })
const suffix: Record<string, string> = { one: 'st', two: 'nd', few: 'rd', other: 'th' }
const place = (rank: number) => `${rank}${suffix[ordinal.select(rank)]}`

/** A signed-in learner's hero: where they stand in the current round and the one action that moves it. */
function MemberPromise({ member, data, phase }: {
  member: Extract<Viewer, { status: 'signed-in' }>
  data: HomeData | undefined
  phase: Phase
}) {
  const links = useLinks()
  const contest = data?.latestContest
  const running = contest && (phase === 'open' || phase === 'live' || phase === 'upcoming')
  const entry = running ? data?.standings?.entries.find(row => row.user_id === member.userId) : undefined
  const { registered } = useParticipation(running ? contest.id : '')
  const joinable = running && !registered && (phase === 'open' || phase === 'upcoming')

  let headline: string
  let lead: string
  if (running && entry) {
    headline = `${place(entry.rank)} place`
    lead = `${numberFormat.format(entry.score)} ${pointsUnit(entry.score)} in ${contest.title}.`
  } else if (running && registered) {
    headline = `You’re in ${contest.title}`
    lead = phase === 'upcoming' ? `The round starts ${formatDate(contest.contest_start)}.` : 'Log what you read and listen to; every entry counts towards the round.'
  } else if (joinable) {
    headline = `${contest.title} is open`
    lead = `Registration closes ${formatDate(contest.registration_end)}.`
  } else {
    headline = `Welcome back, ${member.displayName}`
    lead = 'Log what you read and listen to today.'
  }

  return (
    <div className="home-promise">
      {phase === 'unavailable' && <p className="home-status" data-tone="warning">Contest updates are delayed</p>}
      <h1 id="home-title">{headline}</h1>
      <p className="home-lead">{lead}</p>
      {/* Log activity is already the header's action; the hero adds Join only while the member can still join. */}
      <div className="home-actions">
        {joinable ? <a href={links.joinContest(contest.id)} className={buttonClassName()}>Join {contest.title}</a> : null}
        <SetGoalLink />
      </div>
      <DailyGoal />
    </div>
  )
}

function ContestBoard({ data, phase, today, retrying, retry }: {
  data: HomeData | undefined
  phase: Phase
  today: string
  retrying: boolean
  retry: () => void
}) {
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const viewerId = viewer?.status === 'signed-in' ? viewer.userId : undefined
  if (phase === 'loading') {
    return <>
      <SectionHeading level={2} id="home-board-title" className="home-ledger-header" title="Current contest" />
      <Loading className="home-ledger-message" label="Loading the current contest" />
    </>
  }

  const unavailable = <>
    <SectionHeading level={2} id="home-board-title" className="home-ledger-header" title={data?.latestContest?.title ?? 'Current contest'} description="Standings temporarily unavailable" />
    <div className="home-ledger-message" role="status">
      <h3>The board could not be loaded</h3>
      <p>Try again to see the current standings. You can still log your activity in the meantime.</p>
      <Button variant="outline" loading={retrying} onClick={retry}>Retry standings</Button>
    </div>
  </>
  if (phase === 'unavailable' || !data) return unavailable

  const contest = data.latestContest
  if (!contest) {
    return <>
      <SectionHeading level={2} id="home-board-title" className="home-ledger-header" title="Official contests" />
      <EmptyState
        className="home-ledger-message"
        variant="plain"
        headingLevel={3}
        title="No official round right now"
        description="Keep logging at your own pace. The next round will appear here once it’s scheduled."
        actions={<a href={links.contests} className="text-link">Browse contests</a>}
      />
    </>
  }

  const length = dayNumber(contest.contest_start, contest.contest_end)
  const header = (description: string, action: ReactNode) => (
    <SectionHeading level={2} id="home-board-title" className="home-ledger-header" title={contest.title} description={description} actions={action} />
  )

  if (phase === 'upcoming') {
    const activities = contest.allowed_activities.map(({ name }, index) => index ? name.toLowerCase() : name).join(' & ')
    return <>
      {header(formatPeriod(contest), <a href={links.contest(contest.id)} className="text-link">Contest details</a>)}
      <StatGrid className="home-ledger-stats" stats={[{ label: 'Length', value: `${length} days` }, ...(activities ? [{ label: 'Activities', value: activities }] : [])]} />
      <div className="home-ledger-message">
        <p className="home-start-date"><span>Starts</span><strong>{formatDate(contest.contest_start)}</strong></p>
      </div>
    </>
  }

  const { summary, standings } = data
  if (!summary || !standings) return unavailable
  const ended = phase === 'ended'
  return <>
    {header(
      `${formatPeriod(contest)} · ${ended ? 'Final results' : `Day ${dayNumber(contest.contest_start, today)} of ${length}`}`,
      <a href={links.contest(contest.id)} className="text-link">{ended ? 'See final board' : 'Full leaderboard'}</a>,
    )}
    <StatGrid className="home-ledger-stats" stats={[
      { label: 'Participants', value: numberFormat.format(summary.participant_count) },
      { label: 'Languages', value: numberFormat.format(summary.language_count) },
      { label: 'Points logged', value: numberFormat.format(summary.total_score) },
    ]} />
    {standings.entries.length ? (
      <Table
        caption={`${ended ? 'Final' : 'Current'} top ${standings.entries.length} in ${contest.title}`}
        captionVisibility="screen-reader"
        minWidth="0"
        linkedRows
        columns={[
          { id: 'rank', header: 'Rank', width: '4.5rem', cell: row => <span className="standings-rank" data-podium={row.rank <= 3 ? row.rank : undefined}>{row.rank}</span> },
          {
            id: 'person',
            header: 'Participant',
            rowHeader: true,
            cell: row => (
              <span className="standings-person">
                <a className="home-person-link" href={links.profile(row.user_id)}>{row.user_display_name}</a>
                {row.user_id === viewerId ? <span className="standings-you">You</span> : null}
              </span>
            ),
          },
          { id: 'score', header: 'Score', align: 'end', cell: row => numberFormat.format(row.score) },
        ]}
        rows={standings.entries}
        getRowKey={row => row.user_id}
        getRowProps={row => (row.user_id === viewerId ? { 'aria-current': 'true' } : {})}
      />
    ) : (
      <EmptyState
        className="home-ledger-message"
        variant="plain"
        headingLevel={3}
        title="No scores yet"
        description="Standings appear here once participants log reading or listening in this round."
      />
    )}
  </>
}
