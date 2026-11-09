import { useQuery } from '@tanstack/react-query'
import { Link, useHydrated, useNavigate } from '@tanstack/react-router'
import { ActionMenu, Button, EmptyState, PageHeader, StatusTag, Tabbar, buttonClassName, type ActionMenuItem } from 'paper-ui'
import { EllipsisHorizontalIcon, iconClassName } from 'paper-ui/icons'
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
 * the scoring rules for administrators. Joining is the only button; the rest sit in the Contest options menu, which is
 * left out when it would be empty. Anonymous visitors who join are sent to log in first.
 */
function ContestBoardPage({ board, today, view, page }: { board: ContestBoard; today: string; view: ContestView; page: number }) {
  const links = useLinks()
  const { contest } = board
  const phase = contestPhase(contest, today)
  const { viewer, registered } = useParticipation(contest.id)
  const banned = viewer?.status === 'banned'
  const admin = viewer?.status === 'signed-in' && viewer.role === 'admin'

  const options = [
    ...(registered && !phase.ended ? [{ id: 'registration', label: 'Update registration', to: links.joinContest(contest.id) }] : []),
    ...(admin ? [{ id: 'scoring', label: 'Scoring rules', to: links.contestScoring(contest.id) }] : []),
  ]
  const join = phase.registrationOpen && !registered && !banned
  const actions = join || options.length > 0
    ? (
      <>
        {join ? <a className={buttonClassName()} href={links.joinContest(contest.id)}>Join contest</a> : null}
        {options.length > 0 ? <ContestOptions options={options} /> : null}
      </>
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
        className={`${standingsHeader} max-md:relative`}
        classNames={{ lead: 'tablet:hidden', actions: 'max-md:contents' }}
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

/** On phones the menu moves up to the back link's row so the title keeps the full width. */
function ContestOptions({ options }: { options: { id: string; label: string; to: string }[] }) {
  const hydrated = useHydrated()
  const navigate = useNavigate()
  const items: ActionMenuItem[] = options.map(({ id, label, to }) => ({ id, label, onSelect: () => void navigate({ to }) }))
  return (
    <span className="max-md:absolute max-md:-top-1.5 max-md:right-0">
      {hydrated
        ? <ActionMenu label="Contest options" items={items} iconOnly />
        : (
          // The menu's trigger as server HTML renders it, disabled until it can open.
          <button
            type="button"
            className={buttonClassName({
              variant: 'outline',
              className: 'w-target min-h-target min-w-target p-0 disabled:cursor-progress disabled:opacity-full',
            })}
            aria-label="Contest options"
            title="Contest options"
            disabled
          >
            <EllipsisHorizontalIcon className={iconClassName()} aria-hidden="true" />
          </button>
        )}
    </span>
  )
}
