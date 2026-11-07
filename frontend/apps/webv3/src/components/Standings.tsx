import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, Pagination, Table, linkStyles, tableRowLink, type TableColumn } from 'paper-ui'
import { size } from 'paper-ui/tokens.stylex'
import { Participant, Rank } from './standings-parts'
import type { ImmersionLeaderboard, ImmersionLeaderboardEntry } from 'api-client'
import { numberFormat } from '../lib/format'
import { standingsPageSize, type Read } from '../lib/leaderboard'
import { viewerQueryOptions } from '../lib/session'
import { renderLink } from './router-link'
import { standingsLayout } from './page-styles'

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
        xstyle={standingsLayout.empty}
        headingLevel={3}
        title="Standings could not be loaded"
        description="Try loading the standings again in a moment."
        actions={<Button variant="outline" loading={retrying} onClick={onRetry}>Try again</Button>}
      />
    )
  }
  if (!result) return <p {...stylex.props(standingsLayout.empty)} role="status">Loading standings…</p>

  const { entries, total_size: total } = result.data
  if (entries.length === 0) {
    return total > 0 ? (
      <EmptyState
        variant="plain"
        xstyle={standingsLayout.empty}
        headingLevel={3}
        title="This page is past the end of the standings"
        description={`The standings have ${numberFormat.format(total)} participants.`}
        actions={<Link {...stylex.props(linkStyles.base)} to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : <EmptyState variant="plain" xstyle={standingsLayout.empty} headingLevel={3} {...empty} />
  }

  const columns: TableColumn<ImmersionLeaderboardEntry>[] = [
    {
      id: 'rank',
      header: 'Rank',
      width: size[72],
      cell: row => <Rank rank={row.rank} tie={row.is_tie} />,
    },
    {
      id: 'participant',
      header: 'Participant',
      rowHeader: true,
      cell: row => (
        <Participant name={row.user_display_name} href={profileHref(row.user_id)} you={row.user_id === viewerId} rowLink={tableRowLink} />
      ),
    },
    { id: 'score', header: 'Score', align: 'end', width: size[112], cell: row => numberFormat.format(row.score) },
  ]
  const first = (page - 1) * standingsPageSize + 1
  const totalPages = Math.max(page, Math.ceil(total / standingsPageSize))

  return (
    <>
      <div data-standings="">
      <Table
        caption={caption}
        captionVisibility="screen-reader"
        columns={columns}
        rows={entries}
        getRowKey={row => row.user_id}
        getRowProps={row => (row.user_id === viewerId ? { 'aria-current': 'true' } : {})}
        minWidth={size[272]}
        linkedRows
        variant="sheet"
        cellXstyle={standingsLayout.cell}
      />
      </div>
      {totalPages > 1 ? (
        <div {...stylex.props(standingsLayout.pagination)} data-standings-pagination="">
          <p {...stylex.props(standingsLayout.paginationSummary)}>
            {numberFormat.format(first)}–{numberFormat.format(first + entries.length - 1)} of {numberFormat.format(total)} participants
          </p>
          <Pagination totalPages={totalPages} currentPage={page} getHref={pageHref} renderLink={renderLink} label="Standings pages" listXstyle={standingsLayout.paginationList} />
        </div>
      ) : null}
      {entries.some(entry => entry.is_tie) ? (
        <p {...stylex.props(standingsLayout.tieNote)}><abbr title="Tied rank" {...stylex.props(standingsLayout.plainAbbr)}>T</abbr> marks a tied rank.</p>
      ) : null}
    </>
  )
}
