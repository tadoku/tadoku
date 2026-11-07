import * as stylex from '@stylexjs/stylex'
import { size } from 'paper-ui/tokens.stylex'
import { Button, buttonStyles, linkStyles, SectionHeading, StatGrid, Surface, Table } from 'paper-ui'
import { Link } from 'react-router-dom'
import { formatNumber, initialContests, initialLogs, leaderboardPeople, users } from '../data'
import { contestEnd, contestStart, formatDate, formatDateRange, formatDateTime } from '../dates'
import { usePlayground, useScenario } from '../state'
import { editorialStyles } from './editorial.styles'
import { utilities } from '../utilities.styles'
import { descendantStyles } from '../descendants.styles'

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
    <div>
      <section {...stylex.props(editorialStyles.homeHero)} aria-labelledby="home-title">
        <div {...stylex.props(editorialStyles.homePromise)}>
          <p {...stylex.props(editorialStyles.homeStatus)} data-tone={unavailable ? 'warning' : between || upcoming ? 'quiet' : 'live'}>{upcoming && alreadyJoined ? 'You’re registered for Round 6' : content.status}</p>
          <h1 {...stylex.props(descendantStyles.homePromiseU_uH1)} id="home-title">{scenario === 'guest-live' ? <>Read more.<br />{' '}Listen more.<br />{' '}Keep going together.</> : content.headline}</h1>
          <p {...stylex.props(editorialStyles.homeLead)}>{participant && myStanding ? `You’re ${myStanding.rank}${myStanding.rank === 1 ? 'st' : myStanding.rank === 2 ? 'nd' : myStanding.rank === 3 ? 'rd' : 'th'} with ${formatNumber(myStanding.score)} points. ` : ''}{content.description}</p>
          <div {...stylex.props(editorialStyles.homeActions)}>
            <Link to={primaryHref} {...stylex.props(...buttonStyles())}>{primaryLabel}</Link>
            <Link to={content.secondaryHref} {...stylex.props(linkStyles.base)}>{content.secondary}</Link>
          </div>
          <p {...stylex.props(editorialStyles.homeFineprint)}>{upcoming || scenario === 'guest-live' ? `${scenario === 'guest-live' ? 'Free to join. ' : ''}Registration closes ${formatDateTime(contestEnd(contest.registrationDeadline))}.` : content.note}</p>
        </div>

        <Surface as="section" elevation="showcase" accent xstyle={[editorialStyles.homeLedger, utilities.tw_p_0]} aria-labelledby="home-board-title">
          <SectionHeading
            level={2}
            id="home-board-title"
            xstyle={[editorialStyles.homeLedgerHeader]}
            title={upcoming ? '2026 Round 6' : between ? '2026 Round 5 recap' : unavailable ? 'Current contest' : '2026 Round 5'}
            description={unavailable ? 'Standings temporarily unavailable' : <>{period}{between ? ' · Final results' : upcoming ? '' : ' · Day 5 of 30'}</>}
            actions={unavailable ? null : <Link to={`/contests/${selectedContest}${upcoming ? '' : '/leaderboard'}${between ? '?scenario=ended' : ''}`} {...stylex.props(linkStyles.base)}>{upcoming ? 'Contest details' : between ? 'See final board' : 'Full leaderboard'}</Link>}
          />
          {unavailable ? (
            <div {...stylex.props(editorialStyles.homeLedgerMessage)} role="status">
              <h3 {...stylex.props(descendantStyles.homeLedgerMessageU_uH3)}>The board could not be loaded</h3>
              <p {...stylex.props(descendantStyles.homeLedgerMessageU_uP)}>Try again to see the current standings. You can still open your personal activity log.</p>
              <Button variant="outline" onClick={() => setScenario(retryScenario)}>Retry standings</Button>
            </div>
          ) : upcoming ? (
            <>
              <StatGrid xstyle={[editorialStyles.homeLedgerStats]} stats={[{ label: 'Length', value: '2 weeks' }, { label: 'Activities', value: 'Reading & listening' }]} />
              <div {...stylex.props(editorialStyles.homeLedgerMessage)}>
                <p {...stylex.props(editorialStyles.homeStartDate, descendantStyles.homeLedgerMessageU_uP)}><span {...stylex.props(descendantStyles.homeStartDateU_uSpan)}>Starts</span><strong {...stylex.props(descendantStyles.homeStartDateU_uStrong)}>{formatDate(contestStart(contest.start))}</strong></p>
                <h3 {...stylex.props(descendantStyles.homeLedgerMessageU_uH3)}>Make space for immersion</h3>
                <p {...stylex.props(descendantStyles.homeLedgerMessageU_uP)}>Choose up to three languages. Contest scoring begins when the round starts.</p>
                <p {...stylex.props(editorialStyles.homeFineprint, descendantStyles.homeLedgerMessageU_uP)}>{formatDateTime(contestStart(contest.start))}. Times are local to you.</p>
              </div>
            </>
          ) : (
            <>
              <StatGrid xstyle={[editorialStyles.homeLedgerStats]} stats={[{ label: 'Participants', value: ranked.length }, { label: 'Languages', value: languageCount }, { label: 'Points logged', value: formatNumber(points) }]} />
              {between && <p {...stylex.props(editorialStyles.homeRoundLeader)}>Round leader <Link to={`/users/${ranked[0]?.userId}?contest=round5`} {...stylex.props(linkStyles.base)}>{users.find(person => person.id === ranked[0]?.userId)?.name}</Link></p>}
              <Table
                caption={`${between ? 'Final' : 'Current'} Round 5 top ${between ? 'three' : 'five'}`}
                captionVisibility="screen-reader"
                minWidth="0"
                columns={[
                  { id: 'rank', header: 'Rank', width: size[72], xstyle: editorialStyles.homeLedgerFirstCell, cell: row => row.rank },
                  { id: 'person', header: 'Participant', rowHeader: true, cell: row => <Link {...stylex.props(editorialStyles.homePersonLink)} to={`/users/${row.userId}?contest=round5`}>{row.name}</Link> },
                  { id: 'score', header: 'Score', align: 'end', xstyle: editorialStyles.homeLedgerLastCell, cell: row => formatNumber(row.score) },
                ]}
                rows={rows}
                getRowKey={row => row.userId}
              />
              {participant && myStanding && (
                <div {...stylex.props(editorialStyles.homeYourStanding)} aria-label="Your standing in Round 5">
                  <span>{myStanding.rank}</span>
                  <Link {...stylex.props(editorialStyles.homePersonLink)} to={`/users/${userId}?contest=round5`}>{user?.name ?? 'Anton'} <small {...stylex.props(descendantStyles.homeYourStandingU_uSmall)}>You</small></Link>
                  <strong>{formatNumber(myStanding.score)}</strong>
                </div>
              )}
              <p {...stylex.props(editorialStyles.homeLedgerNote)}>{between ? 'Thanks for a month of reading and listening together.' : 'Reading and listening, across all languages in this round.'}</p>
            </>
          )}
        </Surface>
      </section>

      <section {...stylex.props(editorialStyles.homeHow)} aria-labelledby="home-how-title">
        <h2 {...stylex.props(descendantStyles.homeHowU_uH2)} id="home-how-title">How Tadoku works</h2>
        <ol {...stylex.props(descendantStyles.homeHowU_uOl)}>
          <li {...stylex.props(descendantStyles.homeHowU_uLi)}><h3 {...stylex.props(descendantStyles.homeHowU_uH3)}>Choose your language</h3><p {...stylex.props(descendantStyles.homeHowU_uP)}>Join an official round or track immersion at your own pace between contests.</p></li>
          <li {...stylex.props(descendantStyles.homeHowU_uLi)}><h3 {...stylex.props(descendantStyles.homeHowU_uH3)}>Read and listen</h3><p {...stylex.props(descendantStyles.homeHowU_uP)}>Books, manga, games, podcasts, video. Find something that keeps you engaged.</p></li>
          <li {...stylex.props(descendantStyles.homeHowU_uLi)}><h3 {...stylex.props(descendantStyles.homeHowU_uH3)}>Log your progress</h3><p {...stylex.props(descendantStyles.homeHowU_uP)}>Record what you did, see where it counts, and let a little company keep you going.</p></li>
        </ol>
      </section>

      {!upcoming && <section {...stylex.props(editorialStyles.homeSchedule)} aria-labelledby="home-next-title">
        <div><h2 {...stylex.props(descendantStyles.homeScheduleU_uH2)} id="home-next-title">Next official round</h2><p {...stylex.props(descendantStyles.homeScheduleU_uP, descendantStyles.homeScheduleU_uDivU_uP)}>{formatDateRange(contestStart(nextContest.start), contestEnd(nextContest.end))}</p></div>
        <p {...stylex.props(descendantStyles.homeScheduleU_uP)}>Registration closes {formatDateTime(contestEnd(nextContest.registrationDeadline))}.</p>
        <Link {...stylex.props(linkStyles.base)} to="/contests/round6">See the schedule</Link>
      </section>}
    </div>
  )
}
