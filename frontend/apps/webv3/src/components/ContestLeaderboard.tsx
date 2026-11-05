import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, DescriptionList, EmptyState, Flash, SectionHeading, StatusTag, buttonClassName } from 'paper-ui'
import { formatDate, formatPeriod, numberFormat } from '../lib/format'
import { latestBoardQuery, leaderboardHref, registeredQuery, standingsQuery, type ContestBoard } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { useLoginUrl, viewerKey, viewerQueryOptions } from '../lib/session'
import { LeaderboardFrame, StandingsFootnote } from './LeaderboardPage'
import { Standings } from './Standings'

/** The latest official contest's board, or why there is none. */
export function LatestLeaderboard({ page }: { page: number }) {
  const latest = useQuery(latestBoardQuery())
  // The route loader read the board; its read time decides the contest day on the server and after hydration.
  const today = new Date(latest.dataUpdatedAt).toISOString().slice(0, 10)
  const currentYear = Number(today.slice(0, 4))

  if (latest.data?.kind === 'board') {
    return <ContestLeaderboard board={latest.data.board} today={today} page={page} pageHref={leaderboardHref.latest} />
  }
  return (
    <LeaderboardFrame period="latest" currentYear={currentYear}>
      {latest.data?.kind === 'none' ? (
        <EmptyState
          headingLevel={2}
          title="No official contest yet"
          description="The latest official round appears here once it’s scheduled. Yearly and all-time standings are still available."
        />
      ) : (
        <EmptyState
          headingLevel={2}
          title="The latest contest could not be loaded"
          description="Try loading the leaderboard again in a moment."
          actions={<Button variant="outline" loading={latest.isFetching} onClick={() => void latest.refetch()}>Try again</Button>}
        />
      )}
    </LeaderboardFrame>
  )
}

/** One contest's standings with its join and logging actions, conditions, summary and recent activity. */
function ContestLeaderboard({ board, today, page, pageHref }: {
  board: ContestBoard
  /** The UTC calendar day the board was read, which decides the contest's phase. */
  today: string
  page: number
  pageHref: (page: number) => string
}) {
  const links = useLinks()
  const loginUrl = useLoginUrl()
  const { contest, summary, recentLogs } = board
  const { data: viewer } = useQuery(viewerQueryOptions())
  const signedIn = viewer?.status === 'signed-in'
  const registered = useQuery({ ...registeredQuery(contest.id, viewerKey(viewer)), enabled: signedIn }).data ?? null
  const standings = useQuery(standingsQuery({ kind: 'contest', contestId: contest.id }, page))

  const started = today >= contest.contest_start
  const ended = today > contest.contest_end
  const ongoing = started && !ended
  const registrationOpen = !ended && today <= contest.registration_end
  const status = ended ? { tone: 'neutral' as const, label: 'Complete' }
    : started ? { tone: 'success' as const, label: 'Live now' }
    : { tone: 'information' as const, label: 'Upcoming' }

  const actions = ongoing && signedIn && registered === true
    ? <a className={buttonClassName({ variant: 'outline' })} href={links.logs}>Log activity</a>
    : registrationOpen && registered !== true
      ? <a className={buttonClassName()} href={links.joinContest(contest.id)}>Join contest</a>
      : null

  return (
    <LeaderboardFrame period="latest" currentYear={Number(today.slice(0, 4))} actions={actions}>
      {viewer?.status === 'anonymous' && !ended ? (
        <Flash className="standings-notice" action={<a className="text-link" href={loginUrl}>Log in</a>}>
          You need to log in to participate in this contest.
        </Flash>
      ) : null}
      {ended ? (
        <Flash className="standings-notice" variant="warning">
          This contest has already ended and does not accept any new participants.
        </Flash>
      ) : null}
      <div className="standings-workspace">
        <div className="standings-main">
          <section className="standings-sheet" aria-labelledby="standings-title">
            <SectionHeading
              level={2}
              id="standings-title"
              className="standings-context"
              title={contest.title}
              status={<StatusTag tone={status.tone}>{status.label}</StatusTag>}
              description={`${formatPeriod(contest)} · ${contest.official ? 'Official contest' : 'Community contest'}`}
            />
            <Standings
              result={standings.data}
              page={page}
              caption={`Standings in ${contest.title}`}
              pageHref={pageHref}
              profileHref={userId => links.contestProfile(contest.id, userId)}
              empty={{
                title: 'No participants yet',
                description: started ? 'Join the round and log your first activity.' : 'Scores begin when the round starts.',
              }}
              retrying={standings.isFetching}
              onRetry={() => void standings.refetch()}
            />
          </section>
          <StandingsFootnote />
        </div>
        <aside className="standings-aside">
          <section aria-labelledby="standings-conditions">
            <h2 id="standings-conditions" className="paper-type-component">Contest conditions</h2>
            <DescriptionList
              items={[
                { label: started ? 'Started' : 'Starts', value: formatDate(contest.contest_start) },
                { label: ended ? 'Ended' : 'Ends', value: formatDate(contest.contest_end) },
                { label: 'Registration', value: registrationOpen ? `Open until ${formatDate(contest.registration_end)}` : 'Closed to new participants' },
                { label: 'Languages', value: contest.allowed_languages?.length ? contest.allowed_languages.map(language => language.name).join(', ') : 'No restrictions' },
                { label: 'Activities', value: contest.allowed_activities.map(activity => activity.name).join(', ') || 'No restrictions' },
                { label: 'Administered by', value: contest.official ? 'Tadoku' : contest.owner_user_display_name ?? 'Unknown user' },
              ]}
            />
            {contest.description ? <p className="standings-description">{contest.description}</p> : null}
            <a className="text-link" href={links.contest(contest.id)}>Contest details</a>
          </section>
          {summary ? (
            <section aria-labelledby="standings-summary">
              <h2 id="standings-summary" className="paper-type-component">Summary</h2>
              <p>
                <strong>{numberFormat.format(summary.participant_count)}</strong> {summary.participant_count === 1 ? 'participant' : 'participants'} immersing
                in <strong>{numberFormat.format(summary.language_count)}</strong> {summary.language_count === 1 ? 'language' : 'languages'} for
                a total score of <strong>{numberFormat.format(summary.total_score)}</strong>.
              </p>
            </section>
          ) : null}
          {ongoing && recentLogs.length > 0 ? (
            <section aria-labelledby="standings-recent">
              <h2 id="standings-recent" className="paper-type-component">Recent updates</h2>
              <ul className="standings-activity">
                {recentLogs.map(log => (
                  <li key={log.id}>
                    <Link className="text-link" to="/logs/$logId" params={{ logId: log.id }}>
                      <span>{log.user_display_name ?? 'Unknown user'}</span>
                      <span>+{numberFormat.format(log.score)}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </section>
          ) : null}
        </aside>
      </div>
    </LeaderboardFrame>
  )
}
