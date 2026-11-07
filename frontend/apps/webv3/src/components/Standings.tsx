import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, Pagination, Table, type TableColumn } from 'paper-ui'
import type { ImmersionLeaderboard, ImmersionLeaderboardEntry } from 'api-client'
import { numberFormat } from '../lib/format'
import { standingsPageSize, type Read } from '../lib/leaderboard'
import { viewerQueryOptions } from '../lib/session'
import { renderLink } from './router-link'

interface StandingsProps {
  /** One page of standings; undefined while it loads after a client navigation. */
  result: Read<ImmersionLeaderboard> | undefined
  page: number
  /** Names the scope for assistive technology, for example `Official contests in 2026`. */
  caption: string
  pageHref: (page: number) => string
  profileHref: (userId: string) => string
  empty: { title: string; description: string }
  retrying: boolean
  onRetry: () => void
}

/** A page of ranked participants with ties, the viewer's own row, pagination and its empty and failed states. */
export function Standings({ result, page, caption, pageHref, profileHref, empty, retrying, onRetry }: StandingsProps) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const viewerId = viewer?.status === 'signed-in' ? viewer.userId : undefined

  if (result && !result.ok) {
    return (
      <EmptyState
        variant="plain"
        className="standings-empty"
        headingLevel={3}
        title="Standings could not be loaded"
        description="Try loading the standings again in a moment."
        actions={<Button variant="outline" loading={retrying} onClick={onRetry}>Try again</Button>}
      />
    )
  }
  if (!result) return <p className="standings-empty" role="status">Loading standings…</p>

  const { entries, total_size: total } = result.data
  if (entries.length === 0) {
    return total > 0 ? (
      <EmptyState
        variant="plain"
        className="standings-empty"
        headingLevel={3}
        title="This page is past the end of the standings"
        description={`The standings have ${numberFormat.format(total)} participants.`}
        actions={<Link className="text-link" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : <EmptyState variant="plain" className="standings-empty" headingLevel={3} {...empty} />
  }

  const columns: TableColumn<ImmersionLeaderboardEntry>[] = [
    {
      id: 'rank',
      header: 'Rank',
      width: '4.5rem',
      cell: row => (
        <span className="standings-rank" data-podium={row.rank <= 3 ? row.rank : undefined}>
          {row.is_tie ? <abbr title="Tied rank">T</abbr> : null}{row.rank}
        </span>
      ),
    },
    {
      id: 'participant',
      header: 'Participant',
      rowHeader: true,
      cell: row => (
        <span className="standings-person">
          <a href={profileHref(row.user_id)}>{row.user_display_name}</a>
          {row.user_id === viewerId ? <span className="standings-you">You</span> : null}
        </span>
      ),
    },
    { id: 'score', header: 'Score', align: 'end', width: '7rem', cell: row => numberFormat.format(row.score) },
  ]
  const first = (page - 1) * standingsPageSize + 1
  const totalPages = Math.max(page, Math.ceil(total / standingsPageSize))

  return (
    <>
      <Table
        caption={caption}
        captionVisibility="screen-reader"
        columns={columns}
        rows={entries}
        getRowKey={row => row.user_id}
        getRowProps={row => (row.user_id === viewerId ? { 'aria-current': 'true' } : {})}
        minWidth="17rem"
        linkedRows
        tableClassName="standings-table"
      />
      {totalPages > 1 ? (
        <div className="standings-pagination">
          <p>
            {numberFormat.format(first)}–{numberFormat.format(first + entries.length - 1)} of {numberFormat.format(total)} participants
          </p>
          <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Standings pages" />
        </div>
      ) : null}
      {entries.some(entry => entry.is_tie) ? <p className="standings-tie-note"><abbr title="Tied rank">T</abbr> marks a tied rank.</p> : null}
    </>
  )
}
