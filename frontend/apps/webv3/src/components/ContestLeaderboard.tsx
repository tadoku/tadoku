import { useQuery } from '@tanstack/react-query'
import type { ImmersionContestView } from 'api-client'
import { Button, buttonClassName, DescriptionList, EmptyState, Flash, SectionHeading, StatusTag, typography } from 'paper-ui'
import type { ReactNode } from 'react'
import { formatDate, formatPeriod } from '../lib/format'
import { latestBoardQuery, leaderboardHref, standingsQuery, contestPhase, readDay, type ContestBoard, type ContestPhase } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { useLoginUrl, viewerQueryOptions } from '../lib/session'
import { LeaderboardFrame } from './LeaderboardPage'
import { Standings } from './Standings'
import { description, sheet, standingsAside, standingsContext, standingsContextTitle, standingsNotice, workspace } from './standings-styles'

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

  const actions = phase.registrationOpen && !registered && viewer?.status !== 'banned'
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
        asideLink={<a className="text-link hover:text-link-hover" href={links.contest(contest.id)}>Contest details</a>}
      />
    </LeaderboardFrame>
  )
}

/** A contest's notices, standings and conditions. */
export function ContestBoardBody({ board, phase, page, pageHref, title, status, asideLink }: {
  board: ContestBoard
  phase: ContestPhase
  page: number
  pageHref: (page: number) => string
  /** The standings sheet's visible heading; without one the page's tab names the standings. */
  title?: string
  status?: ReactNode
  /** The link that closes the conditions section. */
  asideLink?: ReactNode
}) {
  const links = useLinks()
  const loginUrl = useLoginUrl()
  const { contest } = board
  const { data: viewer } = useQuery(viewerQueryOptions())
  const standings = useQuery(standingsQuery({ kind: 'contest', contestId: contest.id }, page))

  return (
    <>
      {viewer?.status === 'anonymous' && !phase.ended ? (
        <Flash className={standingsNotice} action={<a className="text-link hover:text-link-hover" href={loginUrl}>Log in</a>}>
          You need to log in to participate in this contest.
        </Flash>
      ) : null}
      {phase.ended ? (
        <Flash className={standingsNotice} variant="warning">
          This contest has already ended and does not accept any new participants.
        </Flash>
      ) : null}
      <div className={workspace}>
        <div>
          <section className={sheet} aria-labelledby={title ? 'standings-title' : undefined} aria-label={title ? undefined : 'Standings'}>
            {title ? (
              <SectionHeading
                level={2}
                id="standings-title"
                className={standingsContext}
                classNames={{ title: standingsContextTitle }}
                title={title}
                status={status}
              />
            ) : null}
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
        </div>
        <aside className={standingsAside}>
          <section aria-labelledby="standings-conditions">
            <h2 id="standings-conditions" className={typography({ role: 'component' })}>About this contest</h2>
            {contest.description ? <p className={description}>{contest.description}</p> : null}
            <ContestFacts contest={contest} phase={phase} />
            {asideLink}
          </section>
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
      ? <a className="text-link hover:text-link-hover" href={links.profile(contest.owner_user_id)}>{contest.owner_user_display_name ?? 'Unknown user'}</a>
      : contest.owner_user_display_name ?? 'Unknown user'
  return (
    <DescriptionList
      items={[
        { label: 'Dates', value: formatPeriod(contest) },
        { label: 'Registration', value: phase.registrationOpen ? `Open until ${formatDate(contest.registration_end)}` : 'Closed to new participants' },
        { label: 'Languages', value: contest.allowed_languages?.length ? contest.allowed_languages.map(language => language.name).join(', ') : 'No restrictions' },
        { label: 'Activities', value: contest.allowed_activities.map(activity => activity.name).join(', ') || 'No restrictions' },
        ...(contest.private ? [{ label: 'Listing', value: 'Unlisted: left out of contest lists, open to anyone with the link' }] : []),
        { label: 'Organized by', value: organizer },
      ]}
    />
  )
}
