import { Button, SectionHeading, StatGrid, Surface, Table, buttonClassName } from 'paper-ui'
import { Link } from 'react-router-dom'
import { formatNumber, initialContests, initialLogs, leaderboardPeople, users } from '../data'
import { contestEnd, contestStart, formatDate, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'

const homeContent = {
  'guest-live': {
    status: 'Round 5 is live',
    headline: 'Read more. Listen more. Keep going together.',
    description: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.',
    primary: 'Join Round 5',
    primaryHref: '/contests/round5/registration',
    secondary: 'How Tadoku works',
    secondaryHref: '/guide',
    note: 'Free to join.',
  },
  'member-open': {
    status: 'Registration is open',
    headline: 'Round 5 is moving. There’s still time to join.',
    description: 'Choose the languages you want to immerse in, join the official round, and let the community’s progress give your daily habit a little extra pull.',
    primary: 'Join Round 5',
    primaryHref: '/contests/round5/registration',
    secondary: 'View contest details',
    secondaryHref: '/contests/round5',
    note: 'Joining a contest and logging activity are separate steps.',
  },
  'participant-live': {
    status: 'You’re participating in Round 5',
    headline: 'Keep your immersion moving.',
    description: 'Log today’s reading or listening and follow the friendly pace of everyone immersing alongside you.',
    primary: 'Log activity',
    primaryHref: '/logs/new',
    secondary: 'View my progress',
    secondaryHref: '/users/anton?contest=round5',
    note: 'Your personal record and contest contributions stay easy to follow.',
  },
  upcoming: {
    status: 'Registration is open',
    headline: 'Make time for the next round.',
    description: 'Choose your languages now and find something you want to read or listen to. Start the next two-week immersion sprint with the community.',
    primary: 'Join Round 6',
    primaryHref: '/contests/round6/registration',
    secondary: 'Track without a contest',
    secondaryHref: '/logs/new',
    note: 'Registration is open.',
  },
  between: {
    status: 'Registration for Round 6 is open',
    headline: 'Keep immersing between rounds.',
    description: 'Contests add momentum, but the habit is yours year-round. Track what you read and listen to today, then bring that rhythm into the next official round.',
    primary: 'Log activity',
    primaryHref: '/logs/new',
    secondary: 'Get ready for Round 6',
    secondaryHref: '/contests/round6',
    note: 'You can track progress without joining a contest.',
  },
  unavailable: {
    status: 'Contest updates are delayed',
    headline: 'Your immersion can keep going.',
    description: 'The current contest board is temporarily unavailable. Your personal activity log is available, so you can keep a record of today’s reading and listening.',
    primary: 'Log activity',
    primaryHref: '/logs/new',
    secondary: 'View my progress',
    secondaryHref: '/users/anton',
    note: 'Only the contest board is unavailable right now.',
  },
}

const howNumber = 'absolute start-0 top-px font-display font-bold text-muted'

export function HomePage() {
  const [scenario, setScenario] = useScenario('home')
  const { user, userId, joinedContests, logs, contests } = usePlayground()
  const content = homeContent[scenario as keyof typeof homeContent] ?? homeContent['guest-live']
  const participant = scenario === 'participant-live'
  const upcoming = scenario === 'upcoming'
  const between = scenario === 'between'
  const unavailable = scenario === 'unavailable'
  const asOf = between ? '2026-10-01' : '2026-09-05'
  const scores = new Map(leaderboardPeople.map(person => [person.userId, person.scores.reduce((total, score) => total + score, 0)]))
  // Initial contest totals already include the initial logs; only local changes alter them.
  for (const [entries, direction] of [[initialLogs, -1], [logs, 1]] as const) {
    for (const log of entries) {
      if (log.date > asOf) continue
      const score = log.submissions.filter(submission => submission.contestId === 'round5').reduce((total, submission) => total + submission.score, 0)
      if (score) scores.set(log.userId, (scores.get(log.userId) ?? 0) + direction * score)
    }
  }
  const ranked = Array.from(scores, ([id, score]) => ({ userId: id, name: users.find(person => person.id === id)?.name ?? id, score: Math.max(0, Math.round(score * 10) / 10) }))
    .sort((a, b) => b.score - a.score || a.name.localeCompare(b.name))
    .map((person, _index, all) => ({ ...person, rank: all.findIndex(other => other.score === person.score) + 1 }))
  const myStanding = ranked.find(person => person.userId === userId)
  const rows = ranked.slice(0, between ? 3 : 5)
  const points = ranked.reduce((total, person) => total + person.score, 0)
  const languageCount = new Set(['Japanese', 'Spanish', ...logs.filter(log => log.date <= asOf && log.submissions.some(submission => submission.contestId === 'round5')).map(log => log.language)]).size
  const selectedContest = upcoming ? 'round6' : 'round5'
  const contest = contests.find(item => item.id === selectedContest) ?? initialContests[upcoming ? 1 : 0]
  const period = formatDateRange(contestStart(contest.start), contestEnd(contest.end))
  const nextContest = contests.find(item => item.id === 'round6') ?? initialContests[1]
  const alreadyJoined = user && joinedContests.includes(selectedContest)
  const retryScenario = !user ? 'guest-live' : joinedContests.includes('round5') ? 'participant-live' : 'member-open'
  const primaryHref = between ? `/logs/new?asOf=${asOf}` : upcoming && alreadyJoined ? '/logs/new' : content.primaryHref
  const primaryLabel = upcoming && alreadyJoined ? 'Log activity' : content.primary

  return (
    <div className="">
      <section className="grid grid-cols-halves items-center gap-16 py-2 pe-1.25 max-lg:gap-8 max-tablet:grid-cols-single max-tablet:items-start max-tablet:gap-8 max-tablet:pt-0" aria-labelledby="home-title">
        <div className="[&_h1]:m-0 [&_h1]:max-w-measure-hero [&_h1]:text-balance [&_h1]:font-display [&_h1]:text-hero [&_h1]:leading-120 [&_h1]:tracking-hero">
          <p className="mb-4 mt-0 text-sm font-semibold text-success data-[tone=quiet]:text-muted data-[tone=warning]:text-warning" data-tone={unavailable ? 'warning' : between || upcoming ? 'quiet' : 'live'}>{upcoming && alreadyJoined ? 'You’re registered for Round 6' : content.status}</p>
          <h1 id="home-title">{scenario === 'guest-live' ? <>Read more.<br />{' '}Listen more.<br />{' '}Keep going together.</> : content.headline}</h1>
          <p className="my-6 max-w-measure-sm text-lead leading-175 text-muted max-tablet:my-5 max-tablet:text-base">{participant && myStanding ? `You’re ${myStanding.rank}${myStanding.rank === 1 ? 'st' : myStanding.rank === 2 ? 'nd' : myStanding.rank === 3 ? 'rd' : 'th'} with ${formatNumber(myStanding.score)} points. ` : ''}{content.description}</p>
          <div className="flex flex-wrap items-center gap-x-5 gap-y-3 [&>.text-link]:text-ui">
            <Link to={primaryHref} className={buttonClassName()}>{primaryLabel}</Link>
            <Link to={content.secondaryHref} className="text-link">{content.secondary}</Link>
          </div>
          <p className="mb-0 mt-4 text-fine leading-160 text-muted">{upcoming || scenario === 'guest-live' ? `${scenario === 'guest-live' ? 'Free to join. ' : ''}Registration closes ${formatDateTime(contestEnd(contest.registrationDeadline))}.` : content.note}</p>
        </div>

        <Surface as="section" elevation="showcase" accent className="min-w-0 [&_th:first-child]:ps-ledger-inset [&_td:first-child]:ps-ledger-inset [&_th:last-child]:pe-ledger-inset [&_td:last-child]:pe-ledger-inset [&_td:first-child]:text-muted p-0" aria-labelledby="home-board-title">
          <SectionHeading
            level={2}
            id="home-board-title"
            className="m-0 border-b border-rule px-ledger-inset py-5.5 max-xs:py-4.5"
            title={upcoming ? '2026 Round 6' : between ? '2026 Round 5 recap' : unavailable ? 'Current contest' : '2026 Round 5'}
            description={unavailable ? 'Standings temporarily unavailable' : <>{period}{between ? ' · Final results' : upcoming ? '' : ' · Day 5 of 30'}</>}
            actions={unavailable ? null : <Link to={`/contests/${selectedContest}${upcoming ? '' : '/leaderboard'}${between ? '?scenario=ended' : ''}`} className="text-link">{upcoming ? 'Contest details' : between ? 'See final board' : 'Full leaderboard'}</Link>}
          />
          {unavailable ? (
            <div className="px-ledger-inset py-6 [&_h3]:mb-3 [&_h3]:mt-0 [&_h3]:text-lead [&_h3]:leading-150 [&>p]:mb-5 [&>p]:mt-0 [&>p]:max-w-measure-message [&>p]:leading-175 [&>p:last-child]:mb-0" role="status">
              <h3>The board could not be loaded</h3>
              <p>Try again to see the current standings. You can still open your personal activity log.</p>
              <Button variant="outline" onClick={() => setScenario(retryScenario)}>Retry standings</Button>
            </div>
          ) : upcoming ? (
            <>
              <StatGrid className="px-ledger-inset py-4.5" stats={[{ label: 'Length', value: '2 weeks' }, { label: 'Activities', value: 'Reading & listening' }]} />
              <div className="px-ledger-inset py-6 [&_h3]:mb-3 [&_h3]:mt-0 [&_h3]:text-lead [&_h3]:leading-150 [&>p]:mb-5 [&>p]:mt-0 [&>p]:max-w-measure-message [&>p]:leading-175 [&>p:last-child]:mb-0">
                <p className="grid gap-1 [&_span]:text-fine [&_span]:text-muted [&_strong]:font-display [&_strong]:text-3xl [&_strong]:leading-130"><span>Starts</span><strong>{formatDate(contestStart(contest.start))}</strong></p>
                <h3>Make space for immersion</h3>
                <p>Choose up to three languages. Contest scoring begins when the round starts.</p>
                <p className="mb-0 mt-4 text-fine leading-160 text-muted">{formatDateTime(contestStart(contest.start))}. Times are local to you.</p>
              </div>
            </>
          ) : (
            <>
              <StatGrid className="px-ledger-inset py-4.5" stats={[{ label: 'Participants', value: ranked.length }, { label: 'Languages', value: languageCount }, { label: 'Points logged', value: formatNumber(points) }]} />
              {between && <p className="m-0 px-ledger-inset pb-4.5 pt-0 text-sm text-muted [&_a]:ms-2 [&_a]:font-semibold">Round leader <Link to={`/users/${ranked[0]?.userId}?contest=round5`} className="text-link">{users.find(person => person.id === ranked[0]?.userId)?.name}</Link></p>}
              <Table
                caption={`${between ? 'Final' : 'Current'} Round 5 top ${between ? 'three' : 'five'}`}
                captionVisibility="screen-reader"
                minWidth="0"
                columns={[
                  { id: 'rank', header: 'Rank', width: '4.5rem', cell: row => row.rank },
                  { id: 'person', header: 'Participant', rowHeader: true, cell: row => <Link className="home-person-link text-ink decoration-rule hover:text-link hover:decoration-current" to={`/users/${row.userId}?contest=round5`}>{row.name}</Link> },
                  { id: 'score', header: 'Score', align: 'end', cell: row => formatNumber(row.score) },
                ]}
                rows={rows}
                getRowKey={row => row.userId}
              />
              {participant && myStanding && (
                <div className="grid grid-cols-your-standing items-baseline border-b border-rule bg-action-soft px-ledger-inset py-3 tabular-nums [&_small]:ms-1.5 [&_small]:text-xs [&_small]:font-regular [&_small]:text-muted" aria-label="Your standing in Round 5">
                  <span>{myStanding.rank}</span>
                  <Link className="home-person-link text-ink decoration-rule hover:text-link hover:decoration-current" to={`/users/${userId}?contest=round5`}>{user?.name ?? 'Anton'} <small>You</small></Link>
                  <strong>{formatNumber(myStanding.score)}</strong>
                </div>
              )}
              <p className="m-0 px-ledger-inset py-3.5 text-xs leading-160 text-muted">{between ? 'Thanks for a month of reading and listening together.' : 'Reading and listening, across all languages in this round.'}</p>
            </>
          )}
        </Surface>
      </section>

      <section className="mt-12 border-t border-rule pt-7 max-tablet:mt-9 max-tablet:pt-6 [&_h2]:mb-6 [&_h2]:mt-0 [&_h2]:font-display [&_h2]:text-heading [&_ol]:m-0 [&_ol]:grid [&_ol]:list-none [&_ol]:grid-cols-thirds [&_ol]:gap-10 [&_ol]:p-0 max-tablet:[&_ol]:grid-cols-single max-tablet:[&_ol]:gap-6 [&_li]:relative [&_li]:ps-7 [&_h3]:mb-2 [&_h3]:mt-0 [&_h3]:text-base [&_h3]:leading-150 [&_p]:m-0 [&_p]:max-w-measure-how [&_p]:text-ui [&_p]:leading-170 [&_p]:text-muted" aria-labelledby="home-how-title">
        <h2 id="home-how-title">How Tadoku works</h2>
        <ol>
          <li><span aria-hidden="true" className={howNumber}>1</span><h3>Choose your language</h3><p>Join an official round or track immersion at your own pace between contests.</p></li>
          <li><span aria-hidden="true" className={howNumber}>2</span><h3>Read and listen</h3><p>Books, manga, games, podcasts, video. Find something that keeps you engaged.</p></li>
          <li><span aria-hidden="true" className={howNumber}>3</span><h3>Log your progress</h3><p>Record what you did, see where it counts, and let a little company keep you going.</p></li>
        </ol>
      </section>

      {!upcoming && <section className="mt-8 grid grid-cols-strip items-center gap-8 border-t border-rule pt-6 max-lg:grid-cols-strip-narrow max-lg:gap-x-6 max-lg:gap-y-4 max-tablet:mt-7 max-tablet:grid-cols-single max-tablet:gap-3 [&_h2]:mb-1 [&_h2]:mt-0 [&_h2]:font-body [&_h2]:text-sm [&_p]:m-0 [&_p]:max-w-measure-strip [&_p]:text-sm [&_p]:leading-170 [&_p]:text-muted [&>div_p]:text-lead [&>div_p]:font-semibold [&>div_p]:text-ink [&>a]:text-sm max-lg:[&>a]:col-start-2 max-tablet:[&>a]:col-auto" aria-labelledby="home-next-title">
        <div><h2 id="home-next-title">Next official round</h2><p>{formatDateRange(contestStart(nextContest.start), contestEnd(nextContest.end))}</p></div>
        <p>Registration closes {formatDateTime(contestEnd(nextContest.registrationDeadline))}.</p>
        <Link className="text-link" to="/contests/round6">See the schedule</Link>
      </section>}
    </div>
  )
}
