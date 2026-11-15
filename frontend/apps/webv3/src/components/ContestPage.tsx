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
import { PageActions } from './PageActions'
import { pageActionsHeader } from './page-actions-styles'
import { PendingView } from './StatusViews'
import { renderLink } from './router-link'
import { standingsHeader } from './standings-styles'

export type ContestView = 'standings' | 'updates'

/** A contest's page: its standings or its updates, with the actions this viewer may take. */
export function ContestPage({ contestId, view, page }: { contestId: string; view: ContestView; page: number }) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const board = useQuery(contestBoardQuery(contestId, viewerKey(viewer)))

  if (board.isPending) return <PendingView />
  if (board.data?.kind === 'board') {
    return <ContestBoardPage board={board.data.board} readAt={board.dataUpdatedAt} view={view} page={page} />
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
 * the scoring rules for administrators. Joining is the primary action; `PageActions` moves the rest into the Contest
 * options menu when the header is too narrow for them. Anonymous visitors who join are sent to log in first.
 */
function ContestBoardPage({ board, readAt, view, page }: { board: ContestBoard; readAt: number; view: ContestView; page: number }) {
  const links = useLinks()
  const { contest } = board
  const phase = contestPhase(contest, readDay(readAt))
  const { viewer, registered } = useParticipation(contest.id)
  const banned = viewer?.status === 'banned'
  const admin = viewer?.status === 'signed-in' && viewer.role === 'admin'

  const options = [
    ...(registered && !phase.ended ? [{ id: 'registration', label: 'Update registration', href: links.joinContest(contest.id) }] : []),
    ...(admin ? [{ id: 'scoring', label: 'Scoring rules', href: links.contestScoring(contest.id) }] : []),
  ]
  const join = phase.registrationOpen && !registered && !banned
  const actions = join || options.length > 0
    ? (
      <PageActions
        label="Contest options"
        primary={join ? [{ id: 'join', label: 'Join contest', variant: 'default', href: links.joinContest(contest.id) }] : []}
        secondary={options}
      />
    )
    : null

  return (
    <div className="standings-page">
      <PageHeader
        backLink={contest.official
          ? { href: contestsHref.official(), label: 'Official contests' }
          : { href: contestsHref.community(), label: 'Community contests' }}
        renderLink={renderLink}
        title={contest.title}
        lead={`${formatPeriod(contest)}${phase.registrationOpen ? ` · Registration closes ${formatDate(contest.registration_end)}` : ''}`}
        status={<><RoundStatus phase={phase} />{contest.private ? <StatusTag>Unlisted</StatusTag> : null}</>}
        className={standingsHeader}
        classNames={{ ...pageActionsHeader, lead: 'tablet:hidden' }}
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
          readAt={readAt}
          page={page}
          pageHref={next => links.contest(contest.id, next)}
        />
      ) : <ContestUpdates contest={contest} page={page} />}
    </div>
  )
}

