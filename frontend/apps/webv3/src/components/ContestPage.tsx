import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, PageHeader, StatusTag, Tabbar, buttonClassName } from 'paper-ui'
import { contestsHref } from '../lib/contests'
import { formatDate, formatPeriod } from '../lib/format'
import { contestBoardQuery, contestPhase, readDay, type ContestBoard } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { useParticipation } from '../lib/registration'
import { viewerKey, viewerQueryOptions } from '../lib/session'
import { ContestBoardBody, RoundStatus } from './ContestLeaderboard'
import { ContestUpdates } from './ContestUpdates'
import { renderLink } from './router-link'
import { standingsHeader } from './standings-styles'

export type ContestView = 'standings' | 'updates'

/** A contest's page: its standings or its updates, with the actions this viewer may take. */
export function ContestPage({ contestId, view, page }: { contestId: string; view: ContestView; page: number }) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const board = useQuery(contestBoardQuery(contestId, viewerKey(viewer)))

  if (board.data?.kind === 'board') {
    return <ContestBoardPage board={board.data.board} today={readDay(board.dataUpdatedAt)} view={view} page={page} />
  }
  if (board.data?.kind === 'none') return <ContestNotFound />
  return (
    <EmptyState
      headingLevel={1}
      title="This contest could not be loaded"
      description="Try loading the contest again in a moment."
      actions={<Button variant="outline" loading={board.isFetching} onClick={() => void board.refetch()}>Try again</Button>}
    />
  )
}

export function ContestNotFound() {
  return (
    <EmptyState
      headingLevel={1}
      title="Contest not found"
      description="This link doesn’t match a contest."
      actions={<Link className={buttonClassName()} to={contestsHref.official()}>Browse contests</Link>}
    />
  )
}

/**
 * Webv2's actions: join while registration is open, change languages until the contest ends, log while it runs and
 * the scoring rules for administrators. Anonymous visitors who join are sent to log in first.
 */
function ContestBoardPage({ board, today, view, page }: { board: ContestBoard; today: string; view: ContestView; page: number }) {
  const links = useLinks()
  const { contest } = board
  const phase = contestPhase(contest, today)
  const { viewer, registered } = useParticipation(contest.id)
  const banned = viewer?.status === 'banned'
  const admin = viewer?.status === 'signed-in' && viewer.role === 'admin'

  const actions = (
    <>
      {admin ? <Link className={buttonClassName({ variant: 'ghost' })} to={links.contestScoring(contest.id)}>Scoring rules</Link> : null}
      {phase.registrationOpen && !registered && !banned
        ? <a className={buttonClassName()} href={links.joinContest(contest.id)}>Join contest</a>
        : null}
      {registered && !phase.ended
        ? <a className={buttonClassName({ variant: 'outline' })} href={links.joinContest(contest.id)}>Update registration</a>
        : null}
    </>
  )

  return (
    <div className="standings-page">
      <PageHeader
        className={standingsHeader}
        backLink={contest.official
          ? { href: contestsHref.official(), label: 'Official contests' }
          : { href: contestsHref.community(), label: 'Community contests' }}
        renderLink={renderLink}
        title={contest.title}
        lead={`${formatPeriod(contest)}${phase.registrationOpen ? ` · Registration closes ${formatDate(contest.registration_end)}` : ''}`}
        classNames={{ lead: 'tablet:hidden' }}
        status={<><RoundStatus phase={phase} />{contest.private ? <StatusTag>Unlisted</StatusTag> : null}</>}
        actions={actions}
      />
      <Tabbar
        label="Contest views"
        renderLink={renderLink}
        links={[
          { id: 'standings', label: 'Leaderboard', href: links.contest(contest.id), current: view === 'standings' },
          { id: 'updates', label: 'Activity', href: links.contestUpdates(contest.id), current: view === 'updates' },
        ]}
      />
      {view === 'standings' ? (
        <ContestBoardBody
          board={board}
          phase={phase}
          page={page}
          pageHref={next => links.contest(contest.id, next)}
        />
      ) : <ContestUpdates contest={contest} page={page} />}
    </div>
  )
}
