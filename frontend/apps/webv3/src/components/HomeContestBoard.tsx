import * as stylex from '@stylexjs/stylex'
import { useQuery } from '@tanstack/react-query'
import { Button, EmptyState, Loading, SectionHeading, StatGrid, Table, linkStyles, tableRowLink } from 'paper-ui'
import {
  border,
  breakpoints,
  color,
  font,
  fontSize,
  lineHeight,
  measure,
  size,
  space,
  type,
} from 'paper-ui/tokens.stylex'
import type { ReactNode } from 'react'
import { dayNumber, type HomeData, type Phase } from '../lib/home'
import { formatDate, formatPeriod, numberFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { viewerQueryOptions } from '../lib/session'
import { Participant, Rank } from './standings-parts'

const ledgerInset = { default: space[24], [breakpoints.xl]: space[20], [breakpoints.xs]: space[16] }

const styles = stylex.create({
  ledgerHeader: {
    margin: 0,
    paddingBlock: { default: space[22], [breakpoints.xs]: space[18] },
    paddingInline: ledgerInset,
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: border.static,
  },
  ledgerStats: { paddingBlock: space[18], paddingInline: ledgerInset },
  ledgerMessage: { paddingBlock: space[24], paddingInline: ledgerInset },
  ledgerMessageTitle: { marginInline: 0, fontSize: fontSize[17], lineHeight: lineHeight.normal, marginBlockEnd: space[12], marginBlockStart: 0, },
  ledgerMessageText: { marginInline: 0, lineHeight: lineHeight.open, marginBlockEnd: space[20], marginBlockStart: 0, maxInlineSize: measure.short, },
  firstCell: { color: color.textMuted, paddingInlineStart: ledgerInset },
  lastCell: { paddingInlineEnd: ledgerInset },
  startDate: { margin: 0, gap: space[4], display: 'grid', },
  startLabel: { color: color.textMuted, fontSize: type.metadataSize },
  startValue: { fontFamily: font.display, fontSize: fontSize[32], lineHeight: lineHeight.statValue },
})


/** The latest official contest beside the hero: its phase, headline figures and current top standings. */
export function ContestBoard({ data, phase, today, retrying, retry }: {
  data: HomeData | undefined
  phase: Phase
  today: string
  retrying: boolean
  retry: () => void
}) {
  const links = useLinks()
  const { data: viewer } = useQuery(viewerQueryOptions())
  const viewerId = viewer?.status === 'signed-in' ? viewer.userId : undefined
  if (phase === 'loading') {
    return <>
      <SectionHeading level={2} id="home-board-title" xstyle={styles.ledgerHeader} title="Current contest" />
      <Loading xstyle={styles.ledgerMessage} label="Loading the current contest" />
    </>
  }

  const unavailable = <>
    <SectionHeading level={2} id="home-board-title" xstyle={styles.ledgerHeader} title={data?.latestContest?.title ?? 'Current contest'} description="Standings temporarily unavailable" />
    <div {...stylex.props(styles.ledgerMessage)} role="status">
      <h3 {...stylex.props(styles.ledgerMessageTitle)}>The board could not be loaded</h3>
      <p {...stylex.props(styles.ledgerMessageText)}>Try again to see the current standings. You can still log your activity in the meantime.</p>
      <Button variant="outline" loading={retrying} onClick={retry}>Retry standings</Button>
    </div>
  </>
  if (phase === 'unavailable' || !data) return unavailable

  const contest = data.latestContest
  if (!contest) {
    return <>
      <SectionHeading level={2} id="home-board-title" xstyle={styles.ledgerHeader} title="Official contests" />
      <EmptyState
        xstyle={styles.ledgerMessage}
        variant="plain"
        headingLevel={3}
        title="No official round right now"
        description="Keep logging at your own pace. The next round will appear here once it’s scheduled."
        actions={<a href={links.contests} {...stylex.props(linkStyles.base)}>Browse contests</a>}
      />
    </>
  }

  const length = dayNumber(contest.contest_start, contest.contest_end)
  const header = (description: string, action: ReactNode) => (
    <SectionHeading level={2} id="home-board-title" xstyle={styles.ledgerHeader} title={contest.title} description={description} actions={action} />
  )

  if (phase === 'upcoming') {
    const activities = contest.allowed_activities.map(({ name }, index) => index ? name.toLowerCase() : name).join(' & ')
    return <>
      {header(formatPeriod(contest), <a href={links.contest(contest.id)} {...stylex.props(linkStyles.base)}>Contest details</a>)}
      <StatGrid xstyle={styles.ledgerStats} stats={[{ label: 'Length', value: `${length} days` }, ...(activities ? [{ label: 'Activities', value: activities }] : [])]} />
      <div {...stylex.props(styles.ledgerMessage)}>
        <p {...stylex.props(styles.startDate)}><span {...stylex.props(styles.startLabel)}>Starts</span><strong {...stylex.props(styles.startValue)}>{formatDate(contest.contest_start)}</strong></p>
      </div>
    </>
  }

  const { summary, standings } = data
  if (!summary || !standings) return unavailable
  const ended = phase === 'ended'
  return <>
    {header(
      `${formatPeriod(contest)} · ${ended ? 'Final results' : `Day ${dayNumber(contest.contest_start, today)} of ${length}`}`,
      <a href={links.contest(contest.id)} {...stylex.props(linkStyles.base)}>{ended ? 'See final board' : 'Full leaderboard'}</a>,
    )}
    <StatGrid xstyle={styles.ledgerStats} stats={[
      { label: 'Participants', value: numberFormat.format(summary.participant_count) },
      { label: 'Languages', value: numberFormat.format(summary.language_count) },
      { label: 'Points logged', value: numberFormat.format(summary.total_score) },
    ]} />
    {standings.entries.length ? (
      <Table
        caption={`${ended ? 'Final' : 'Current'} top ${standings.entries.length} in ${contest.title}`}
        captionVisibility="screen-reader"
        minWidth="0"
        linkedRows
        variant="sheet"
        columns={[
          { id: 'rank', header: 'Rank', width: size[72], xstyle: styles.firstCell, cell: row => <Rank rank={row.rank} /> },
          {
            id: 'person',
            header: 'Participant',
            rowHeader: true,
            cell: row => (
              <Participant
                name={row.user_display_name}
                href={links.profile(row.user_id)}
                you={row.user_id === viewerId}
                rowLink={tableRowLink}
              />
            ),
          },
          { id: 'score', header: 'Score', align: 'end', xstyle: styles.lastCell, cell: row => numberFormat.format(row.score) },
        ]}
        rows={standings.entries}
        getRowKey={row => row.user_id}
        getRowProps={row => (row.user_id === viewerId ? { 'aria-current': 'true' } : {})}
      />
    ) : (
      <EmptyState
        xstyle={styles.ledgerMessage}
        variant="plain"
        headingLevel={3}
        title="No scores yet"
        description="Standings appear here once participants log reading or listening in this round."
      />
    )}
  </>
}
