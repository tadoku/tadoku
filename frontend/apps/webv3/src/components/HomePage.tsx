import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, Loading, SectionHeading, StatGrid, Surface, Table, buttonClassName } from 'paper-ui'
import type { ReactNode } from 'react'
import { homeQueryOptions, type Contest, type HomeData } from '../lib/home'
import { links } from '../lib/links'

// Contest dates are UTC calendar days. Formatting them by hand keeps server (Node ICU) and browser markup identical.
const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sept', 'Oct', 'Nov', 'Dec']
const numberFormat = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 })
const dayAndMonth = (date: string) => `${Number(date.slice(8, 10))} ${months[Number(date.slice(5, 7)) - 1]}`
const formatDate = (date: string) => `${dayAndMonth(date)} ${date.slice(0, 4)}`
const formatPeriod = ({ contest_start: start, contest_end: end }: Pick<Contest, 'contest_start' | 'contest_end'>) =>
  `${start.slice(0, 4) === end.slice(0, 4) ? dayAndMonth(start) : formatDate(start)} – ${formatDate(end)}`
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
  const { data, dataUpdatedAt, isError, refetch, isFetching } = useQuery(homeQueryOptions())
  // The fetch time travels with the dehydrated data, so server and client agree on the contest day.
  const today = new Date(dataUpdatedAt).toISOString().slice(0, 10)
  const phase = phaseOf(data, isError, today)
  const contest = data?.latestContest ?? null
  const joinable = contest && (phase === 'open' || phase === 'upcoming')

  return (
    <div className="home-page">
      <section className="home-hero" aria-labelledby="home-title">
        <div className="home-promise">
          {contest && phase !== 'unavailable' ? (
            <p className="home-status" data-tone={phase === 'open' || phase === 'live' ? 'live' : 'quiet'}>
              {phase === 'upcoming' ? 'Registration is open' : phase === 'ended' ? `${contest.title} has ended` : `${contest.title} is live`}
            </p>
          ) : phase === 'unavailable' ? <p className="home-status" data-tone="warning">Contest updates are delayed</p> : null}
          <h1 id="home-title">{promise[phase].headline}</h1>
          <p className="home-lead">{promise[phase].lead}</p>
          <div className="home-actions">
            {joinable
              ? <a href={links.joinContest(contest.id)} className={buttonClassName()}>Join {contest.title}</a>
              : <a href={links.logs} className={buttonClassName()}>Log activity</a>}
            <a href={links.guide} className="text-link">How Tadoku works</a>
          </div>
          {joinable
            ? <p className="home-fineprint">Free to join. Registration closes {formatDate(contest.registration_end)}.</p>
            : <p className="home-fineprint">You can track your progress without joining a contest.</p>}
        </div>

        <Surface as="section" elevation="showcase" accent className="home-ledger" aria-labelledby="home-board-title">
          <ContestBoard data={data} phase={phase} today={today} retrying={isFetching} retry={() => void refetch()} />
        </Surface>
      </section>

      <section className="home-how" aria-labelledby="home-how-title">
        <h2 id="home-how-title">How Tadoku works</h2>
        <ol>
          <li><h3>Choose your language</h3><p>Join an official round or track immersion at your own pace between contests.</p></li>
          <li><h3>Read and listen</h3><p>Books, manga, games, podcasts, video. Find something that keeps you engaged.</p></li>
          <li><h3>Log your progress</h3><p>Record what you did, see where it counts, and let a little company keep you going.</p></li>
        </ol>
      </section>

      {phase !== 'loading' && phase !== 'unavailable' && (
        <section className="home-strip" aria-labelledby="home-next-title">
          <div>
            <h2 id="home-next-title">Next official round</h2>
            <p>{data?.nextContest ? formatPeriod(data.nextContest) : 'Not announced yet'}</p>
          </div>
          <p>{data?.nextContest
            ? `Registration closes ${formatDate(data.nextContest.registration_end)}.`
            : 'New official rounds appear on the contests page as soon as they’re scheduled.'}</p>
          <a className="text-link" href={links.contests}>See all contests</a>
        </section>
      )}

      <section className="home-strip" aria-labelledby="home-goal-title">
        <div>
          <h2 id="home-goal-title">Daily goal</h2>
          <p>Plan your sessions</p>
        </div>
        <p>Decide how long you want to read or listen each day, and keep that goal in view.</p>
        <Link className="text-link" to="/plan">Set a daily goal</Link>
      </section>
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
        <h3>Make space for immersion</h3>
        <p>Choose your languages when you join. Contest scoring begins when the round starts.</p>
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
        columns={[
          { id: 'rank', header: 'Rank', width: '4.5rem', cell: row => row.rank },
          { id: 'person', header: 'Participant', rowHeader: true, cell: row => <a className="home-person-link" href={links.profile(row.user_id)}>{row.user_display_name}</a> },
          { id: 'score', header: 'Score', align: 'end', cell: row => numberFormat.format(row.score) },
        ]}
        rows={standings.entries}
        getRowKey={row => row.user_id}
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
    <p className="home-ledger-note">{ended ? 'Thanks for reading and listening together.' : 'Reading and listening, across all languages in this round.'}</p>
  </>
}
