import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import type { ImmersionContestView } from 'api-client'
import { Button, DescriptionList, EmptyState, Flash, SectionHeading, StatusTag, buttonClassName } from 'paper-ui'
import type { ReactNode } from 'react'
import { formatDate, formatPeriod, numberFormat } from '../lib/format'
import { latestBoardQuery, leaderboardHref, standingsQuery, contestPhase, readDay, type ContestBoard, type ContestPhase } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { useLoginUrl, viewerQueryOptions } from '../lib/session'
import { LeaderboardFrame, StandingsFootnote } from './LeaderboardPage'
import { Standings } from './Standings'

export function RoundStatus({ phase }: { phase: ContestPhase }) {
  if (phase.ended) return <StatusTag>Ended</StatusTag>
  return phase.started ? <StatusTag tone="success">Live now</StatusTag> : <StatusTag tone="information">Upcoming</StatusTag>
}

/** The latest official contest's board, or why there is none. */
export function LatestLeaderboard({ page }: { page: number }) {
  const latest = useQuery(latestBoardQuery())
  const today = readDay(latest.dataUpdatedAt)
  const currentYear = Number(today.slice(0, 4))

  if (latest.data?.kind === 'board') return <LatestBoard board={latest.data.board} today={today} page={page} />
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

/** The latest contest's board with Webv2's two actions: join while registration is open, then log while it runs. */
function LatestBoard({ board, today, page }: { board: ContestBoard; today: string; page: number }) {
  const links = useLinks()
  const { contest } = board
  const phase = contestPhase(contest, today)
  const { viewer, registered } = useParticipation(contest.id)

  const actions = phase.ongoing && registered
    ? <a className={buttonClassName({ variant: 'outline' })} href={links.logs}>Log activity</a>
    : phase.registrationOpen && !registered && viewer?.status !== 'banned'
      ? <a className={buttonClassName()} href={links.joinContest(contest.id)}>Join contest</a>
      : null

  return (
    <LeaderboardFrame period="latest" currentYear={Number(today.slice(0, 4))} actions={actions}>
      <ContestBoardBody
        board={board}
        phase={phase}
        page={page}
        pageHref={leaderboardHref.latest}
        title={contest.title}
        status={<RoundStatus phase={phase} />}
        asideLink={<a className="text-link" href={links.contest(contest.id)}>Contest details</a>}
      />
    </LeaderboardFrame>
  )
}

/** A contest's notices, standings, conditions, summary and, while it runs, its ten most recent logs. */
export function ContestBoardBody({ board, phase, page, pageHref, title, status, asideLink }: {
  board: ContestBoard
  phase: ContestPhase
  page: number
  pageHref: (page: number) => string
  /** The standings sheet's heading. */
  title: string
  status?: ReactNode
  /** The link that closes the conditions section. */
  asideLink: ReactNode
}) {
  const links = useLinks()
  const loginUrl = useLoginUrl()
  const { contest, summary, recentLogs } = board
  const { data: viewer } = useQuery(viewerQueryOptions())
  const standings = useQuery(standingsQuery({ kind: 'contest', contestId: contest.id }, page))

  return (
    <>
      {viewer?.status === 'anonymous' && !phase.ended ? (
        <Flash className="standings-notice" action={<a className="text-link" href={loginUrl}>Log in</a>}>
          You need to log in to participate in this contest.
        </Flash>
      ) : null}
      {phase.ended ? (
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
              title={title}
              status={status}
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
                description: phase.started ? 'Join the round and log your first activity.' : 'Scores begin when the round starts.',
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
            <ContestFacts contest={contest} phase={phase} />
            {contest.description ? <p className="standings-description">{contest.description}</p> : null}
            {asideLink}
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
          {phase.ongoing && recentLogs.length > 0 ? (
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
    </>
  )
}

/** A contest's dates, registration, eligibility, listing and organizer. Dates are UTC calendar days. */
export function ContestFacts({ contest, phase }: { contest: ImmersionContestView; phase: ContestPhase }) {
  const links = useLinks()
  const organizer = contest.official
    ? 'Tadoku'
    : contest.owner_user_id
      ? <a className="text-link" href={links.profile(contest.owner_user_id)}>{contest.owner_user_display_name ?? 'Unknown user'}</a>
      : contest.owner_user_display_name ?? 'Unknown user'
  return (
    <DescriptionList
      items={[
        { label: phase.started ? 'Started' : 'Starts', value: formatDate(contest.contest_start) },
        { label: phase.ended ? 'Ended' : 'Ends', value: formatDate(contest.contest_end) },
        { label: 'Registration', value: phase.registrationOpen ? `Open until ${formatDate(contest.registration_end)}` : 'Closed to new participants' },
        { label: 'Languages', value: contest.allowed_languages?.length ? contest.allowed_languages.map(language => language.name).join(', ') : 'No restrictions' },
        { label: 'Activities', value: contest.allowed_activities.map(activity => activity.name).join(', ') || 'No restrictions' },
        ...(contest.private ? [{ label: 'Listing', value: 'Unlisted: left out of contest lists, open to anyone with the link' }] : []),
        { label: 'Administered by', value: organizer },
      ]}
    />
  )
}
