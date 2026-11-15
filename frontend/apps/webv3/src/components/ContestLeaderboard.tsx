import { useQuery } from '@tanstack/react-query'
import type { ImmersionContestView } from 'api-client'
import { Button, buttonClassName, DescriptionList, EmptyState, Flash, SectionHeading, StatusTag, typography, type DescriptionListItem } from 'paper-ui'
import { useEffect, useState, type ReactNode } from 'react'
import { countLabel, formatDate, formatPeriod, numberFormat } from '../lib/format'
import { latestBoardQuery, leaderboardHref, standingsQuery, contestPhase, readDay, type ContestBoard, type ContestPhase } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { Avatar } from './Avatar'
import { LeaderboardFrame } from './LeaderboardPage'
import { Standings } from './Standings'
import { description, sheet, standingsAside, standingsContext, standingsContextTitle, standingsNotice, workspace } from './standings-styles'
import { AppLink } from './router-link'

export function RoundStatus({ phase }: { phase: ContestPhase }) {
  if (phase.ended) return <StatusTag>Ended</StatusTag>
  return phase.started ? <StatusTag tone="success">Live now</StatusTag> : <StatusTag tone="information">Upcoming</StatusTag>
}

/** The latest official contest's board, or why there is none. */
export function LatestLeaderboard({ page }: { page: number }) {
  const latest = useQuery(latestBoardQuery())
  const today = readDay(latest.dataUpdatedAt)
  const currentYear = Number(today.slice(0, 4))

  if (latest.data?.kind === 'board') return <LatestBoard board={latest.data.board} readAt={latest.dataUpdatedAt} page={page} />
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
function LatestBoard({ board, readAt, page }: { board: ContestBoard; readAt: number; page: number }) {
  const links = useLinks()
  const { contest } = board
  const today = readDay(readAt)
  const phase = contestPhase(contest, today)
  const { viewer, registered } = useParticipation(contest.id)

  const actions = phase.registrationOpen && !registered && viewer?.status !== 'banned'
    ? <AppLink className={buttonClassName()} href={links.joinContest(contest.id)}>Join contest</AppLink>
    : null

  return (
    <LeaderboardFrame period="latest" currentYear={Number(today.slice(0, 4))} actions={actions}>
      <ContestBoardBody
        board={board}
        phase={phase}
        readAt={readAt}
        page={page}
        pageHref={leaderboardHref.latest}
        title={contest.title}
        status={<RoundStatus phase={phase} />}
        asideLink={<AppLink className="text-link hover:text-link-hover" href={links.contest(contest.id)}>Contest details</AppLink>}
      />
    </LeaderboardFrame>
  )
}

/** A contest's notices, standings and conditions. */
export function ContestBoardBody({ board, phase, readAt, page, pageHref, title, status, asideLink }: {
  board: ContestBoard
  phase: ContestPhase
  /** When the board was read, which starts the countdown so the server and hydration render the same time. */
  readAt: number
  page: number
  pageHref: (page: number) => string
  /** The standings sheet's visible heading; without one the page's tab names the standings. */
  title?: string
  status?: ReactNode
  /** The link that closes the conditions section. */
  asideLink?: ReactNode
}) {
  const links = useLinks()
  const { contest } = board
  const standings = useQuery(standingsQuery({ kind: 'contest', contestId: contest.id }, page))

  return (
    <>
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
            <LiveContestFacts
              key={contest.id}
              contest={contest}
              phase={phase}
              readAt={readAt}
              participants={standings.data?.ok ? standings.data.data.total_size : undefined}
            />
            {asideLink}
          </section>
        </aside>
      </div>
    </>
  )
}

/**
 * The contest's facts with Webv2's countdown to its start or end, the UTC clock and the standings' participant count.
 * The clock starts at `readAt` and ticks each second after hydration until the contest's last UTC second has passed.
 */
function LiveContestFacts({ contest, phase, readAt, participants }: {
  contest: ImmersionContestView
  phase: ContestPhase
  readAt: number
  participants: number | undefined
}) {
  const start = Date.parse(`${contest.contest_start}T00:00:00Z`)
  const end = Date.parse(`${contest.contest_end}T00:00:00Z`) + 86_400_000
  const now = useClock(readAt, end)

  const clock = now < end
    ? [
      { label: now < start ? 'Starting in' : 'Ending in', value: countdown((now < start ? start : end) - now) },
      { label: 'Tadoku time', value: utcTime(now) },
    ]
    : []
  const count = participants === undefined ? [] : [{ label: 'Participants', value: numberFormat.format(participants) }]
  return <ContestFacts contest={contest} phase={phase} live={[...clock, ...count]} />
}

/**
 * The time from `readAt`, then the browser's clock: once right after mounting, which also replaces a cached read's
 * older time, and each second after that until `until`.
 */
function useClock(readAt: number, until: number) {
  const [now, setNow] = useState(readAt)
  const ended = now >= until
  useEffect(() => {
    if (ended) return
    let timer: ReturnType<typeof setTimeout>
    const tick = () => {
      const time = Date.now()
      setNow(time)
      timer = setTimeout(tick, 1000 - time % 1000)
    }
    timer = setTimeout(tick, 0)
    return () => clearTimeout(timer)
  }, [ended])
  return now
}

/** Whole days, hours and minutes as Webv2 counts down, for example `3 days, 4 hr, 12 min`. */
function countdown(milliseconds: number) {
  const minutes = Math.round(milliseconds / 60_000)
  const days = Math.floor(minutes / 1440)
  return `${countLabel(days, 'day')}, ${Math.floor(minutes % 1440 / 60)} hr, ${minutes % 60} min`
}

/** For example `15 Nov 2026, 14:03:09 UTC`. */
function utcTime(milliseconds: number) {
  const instant = new Date(milliseconds).toISOString()
  return `${formatDate(instant.slice(0, 10))}, ${instant.slice(11, 19)} UTC`
}

/**
 * A contest's dates, registration, eligibility, listing and organizer. Dates are UTC calendar days. `live` facts follow
 * the registration.
 */
export function ContestFacts({ contest, phase, live = [] }: { contest: ImmersionContestView; phase: ContestPhase; live?: DescriptionListItem[] }) {
  const links = useLinks()
  const organizer = contest.official
    ? 'Tadoku'
    : contest.owner_user_id
      ? (
        <span>
          <Avatar userId={contest.owner_user_id} name={contest.owner_user_display_name ?? '?'} size="sm" className="me-1.5" />
          <AppLink className="text-link hover:text-link-hover" href={links.profile(contest.owner_user_id)}>{contest.owner_user_display_name ?? 'Unknown user'}</AppLink>
        </span>
      )
      : contest.owner_user_display_name ?? 'Unknown user'
  return (
    <DescriptionList
      className="max-tablet:[&>.paper-description-list__item:nth-child(-n+2)]:hidden"
      items={[
        { label: 'Dates', value: formatPeriod(contest) },
        { label: 'Registration', value: phase.registrationOpen ? `Open until ${formatDate(contest.registration_end)}` : 'Closed to new participants' },
        ...live,
        { label: 'Languages', value: contest.allowed_languages?.length ? contest.allowed_languages.map(language => language.name).join(', ') : 'No restrictions' },
        { label: 'Activities', value: contest.allowed_activities.map(activity => activity.name).join(', ') || 'No restrictions' },
        ...(contest.private ? [{ label: 'Listing', value: 'Unlisted: left out of contest lists, open to anyone with the link' }] : []),
        { label: 'Organized by', value: organizer },
      ]}
    />
  )
}
