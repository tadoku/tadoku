import { Button, Surface, Table, buttonClassName } from 'paper-ui'
import { Link } from 'react-router-dom'
import { formatNumber, initialLogs, leaderboardPeople, users } from '../data'
import { usePlayground, useScenario } from '../state'
import './editorial.css'

const homeContent = {
  'guest-live': {
    status: 'Round 5 is live',
    headline: 'Read more. Listen more. Keep going together.',
    description: 'Turn time with your languages into a habit. Join a friendly immersion contest, log what you read and listen to, and follow the community’s progress.',
    primary: 'Join Round 5',
    primaryHref: '/contests/round5/registration',
    secondary: 'How Tadoku works',
    secondaryHref: '/guide',
    note: 'Free to join. Registration closes 23 September.',
  },
  'member-open': {
    status: 'Registration closes 23 September',
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
    headline: 'The next round starts 1 November.',
    description: 'Choose your languages now and find something you want to read or listen to. Start the next two-week immersion sprint with the community.',
    primary: 'Join Round 6',
    primaryHref: '/contests/round6/registration',
    secondary: 'Track without a contest',
    secondaryHref: '/logs/new',
    note: 'Registration closes 7 November.',
  },
  between: {
    status: 'The next round starts 1 November',
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
    secondary: 'Try leaderboard again',
    secondaryHref: '/contests/round5/leaderboard',
    note: 'Only the contest board is unavailable right now.',
  },
}

export function HomePage() {
  const [scenario, setScenario] = useScenario('home')
  const { user, userId, joinedContests, logs } = usePlayground()
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
  const alreadyJoined = user && joinedContests.includes(selectedContest)
  const retryScenario = !user ? 'guest-live' : joinedContests.includes('round5') ? 'participant-live' : 'member-open'
  const primaryHref = between ? `/logs/new?asOf=${asOf}` : upcoming && alreadyJoined ? '/logs/new' : content.primaryHref
  const primaryLabel = upcoming && alreadyJoined ? 'Log activity' : content.primary

  return (
    <div className="home-page">
      <section className="home-hero" aria-labelledby="home-title">
        <div className="home-promise">
          <p className="home-status" data-tone={unavailable ? 'warning' : between || upcoming ? 'quiet' : 'live'}>{upcoming && alreadyJoined ? 'You’re registered for Round 6' : content.status}</p>
          <h1 id="home-title">{content.headline}</h1>
          <p className="home-lead">{participant && myStanding ? `You’re ${myStanding.rank}${myStanding.rank === 1 ? 'st' : myStanding.rank === 2 ? 'nd' : myStanding.rank === 3 ? 'rd' : 'th'} with ${formatNumber(myStanding.score)} points. ` : ''}{content.description}</p>
          <div className="home-actions flex flex-wrap items-center gap-4">
            <Link to={primaryHref} className={buttonClassName()}>{primaryLabel}</Link>
            {unavailable ? <Button variant="link" onClick={() => setScenario(retryScenario)}>{content.secondary}</Button> : <Link to={content.secondaryHref} className="text-link">{content.secondary}</Link>}
          </div>
          <p className="home-fineprint">{content.note}</p>
        </div>

        <Surface as="section" elevation="showcase" accent className="home-ledger p-0" aria-labelledby="home-board-title">
          <header className="home-ledger-header">
            <div>
              <h2 id="home-board-title">{upcoming ? '2026 Round 6' : between ? '2026 Round 5 recap' : unavailable ? 'Current contest' : '2026 Round 5'}</h2>
              <p>{upcoming ? '1–14 November 2026' : between ? '1–30 September 2026 · Final results' : unavailable ? 'Standings temporarily unavailable' : '1–30 September 2026 · Day 5 of 30'}</p>
            </div>
            {!unavailable && <Link to={`/contests/${selectedContest}${upcoming ? '' : '/leaderboard'}${between ? '?scenario=ended' : ''}`} className="text-link">{upcoming ? 'Contest details' : between ? 'See final board' : 'Full leaderboard'}</Link>}
          </header>
          {unavailable ? (
            <div className="home-ledger-message" role="status">
              <h3>The board could not be loaded</h3>
              <p>Try again to see the current standings. You can still open your personal activity log.</p>
              <Button variant="outline" onClick={() => setScenario(retryScenario)}>Retry standings</Button>
            </div>
          ) : upcoming ? (
            <>
              <dl className="home-ledger-stats">
                <div><dt>Length</dt><dd>2 weeks</dd></div>
                <div><dt>Activities</dt><dd>Reading & listening</dd></div>
              </dl>
              <div className="home-ledger-message">
                <p className="home-start-date"><span>Starts</span><strong>1 November</strong></p>
                <h3>Make space for immersion</h3>
                <p>Choose up to three languages. Contest scoring begins when the round starts.</p>
                <p className="home-fineprint">Registration closes 7 November. All dates use UTC.</p>
              </div>
            </>
          ) : (
            <>
              <dl className="home-ledger-stats">
                <div><dt>Participants</dt><dd>{ranked.length}</dd></div>
                <div><dt>Languages</dt><dd>{languageCount}</dd></div>
                <div><dt>Points logged</dt><dd>{formatNumber(points)}</dd></div>
              </dl>
              {between && <p className="home-round-leader">Round leader <Link to={`/users/${ranked[0]?.userId}?contest=round5`} className="text-link">{users.find(person => person.id === ranked[0]?.userId)?.name}</Link></p>}
              <Table
                caption={`${between ? 'Final' : 'Current'} Round 5 top ${between ? 'three' : 'five'}`}
                captionVisibility="screen-reader"
                minWidth="0"
                columns={[
                  { id: 'rank', header: 'Rank', width: '3.75rem', cell: row => row.rank },
                  { id: 'person', header: 'Participant', rowHeader: true, cell: row => <Link className="text-link" to={`/users/${row.userId}?contest=round5`}>{row.name}</Link> },
                  { id: 'score', header: 'Score', align: 'end', cell: row => formatNumber(row.score) },
                ]}
                rows={rows}
                getRowKey={row => row.userId}
              />
              {participant && myStanding && (
                <div className="home-your-standing" aria-label="Your standing in Round 5">
                  <span>{myStanding.rank}</span>
                  <Link className="text-link" to={`/users/${userId}?contest=round5`}>{user?.name ?? 'Anton'} <small>You</small></Link>
                  <strong>{formatNumber(myStanding.score)}</strong>
                </div>
              )}
              <p className="home-ledger-note">{between ? 'Thanks for a month of reading and listening together.' : 'Reading and listening, across all languages in this round.'}</p>
            </>
          )}
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

      <section className="home-schedule" aria-labelledby="home-next-title">
        <div><h2 id="home-next-title">Next official round</h2><p>1–14 November 2026</p></div>
        <p>Track your immersion anytime. Registration for Round 6 is open until 7 November.</p>
        <Link className="text-link" to="/contests/round6">See the schedule</Link>
      </section>
    </div>
  )
}
