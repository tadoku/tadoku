import * as stylex from '@stylexjs/stylex'
import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { EmptyState, Pagination, a11y, linkStyles } from 'paper-ui'
import { border, breakpoints, color, fontSize, fontWeight, lineHeight, measure, size, space, type } from 'paper-ui/tokens.stylex'
import type { ImmersionLog } from 'api-client'
import { formatDate, formatTracking, numberFormat, pointsUnit } from '../lib/format'
import { activityPageSize } from '../lib/profile'
import { renderLink } from './router-link'
import { entryStyles, entryTitle } from './page-styles'

const dateGutter = size[112]



const styles = stylex.create({
  paginationTop: { display: 'flex', justifyContent: 'flex-end', marginBlockEnd: space[4], marginBlockStart: space[8] },
  wrap: { flexWrap: 'wrap' },
  list: {
    margin: 0,
    padding: 0,
    listStyle: 'none',
    borderBlockEndColor: color.ruleDefault,
    borderBlockEndStyle: 'solid',
    borderBlockEndWidth: border.static,
  },
  entry: {
    paddingBlock: space[16],
    alignItems: 'start',
    borderBlockStartColor: { default: color.ruleDefault, ':first-child': 'transparent' },
    borderBlockStartStyle: 'solid',
    borderBlockStartWidth: border.static,
    columnGap: { default: space[24], [breakpoints.phone]: space[16] },
    display: 'grid',
    gridTemplateColumns: { default: `${dateGutter} minmax(0, 1fr) auto`, [breakpoints.phone]: 'minmax(0, 1fr) auto' },
    position: 'relative',
    rowGap: { default: space[24], [breakpoints.phone]: space[6] },
  },
  /** Within a day the separator starts after the date gutter, so each day reads as one group. */
  sameDay: {
    borderBlockStartColor: 'transparent',
    '::before': {
      borderBlockStartColor: color.ruleSubtle,
      borderBlockStartStyle: 'solid',
      borderBlockStartWidth: border.static,
      content: '""',
      insetBlockStart: `calc(-1 * ${border.static})`,
      insetInlineEnd: 0,
      insetInlineStart: { default: `calc(${dateGutter} + ${space[24]})`, [breakpoints.phone]: 0 },
      position: 'absolute',
    },
  },
  date: {
    gridColumn: { default: null, [breakpoints.phone]: '1 / -1' },
    color: color.textMuted,
    fontSize: type.metadataSize,
    paddingBlockStart: { default: space[3], [breakpoints.phone]: 0 },
  },
  body: { gridColumn: { default: 2, [breakpoints.phone]: 1 }, minInlineSize: 0 },
  /** A log's description can run long; the list previews two lines and the log page shows it whole. */
  clamp: { overflow: 'hidden', WebkitBoxOrient: 'vertical', WebkitLineClamp: 2, display: '-webkit-box', },
  nsfw: { fontSize: fontSize[14], marginBlockStart: space[10] },
  nsfwSummary: { color: color.textMuted, cursor: 'pointer' },
  nsfwText: { marginInline: 0, lineHeight: lineHeight.airy, marginBlockEnd: 0, marginBlockStart: space[8], maxInlineSize: measure.text, overflowWrap: 'anywhere', },
  actions: { marginBlockStart: space[8], marginInlineStart: `calc(-1 * ${space[12]})` },
  score: {
    gridColumn: { default: 3, [breakpoints.phone]: 2 },
    gridRow: { default: 1, [breakpoints.phone]: 'auto' },
    fontVariantNumeric: 'tabular-nums',
    fontWeight: fontWeight.bold,
    textAlign: 'end',
    whiteSpace: 'nowrap',
  },
  unit: { color: color.textMuted, fontSize: type.metadataSize, fontWeight: fontWeight.regular },
  footer: {
    paddingBlock: space[16],
    alignItems: 'center',
    columnGap: space[16],
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    rowGap: space[8],
  },
  footerText: { margin: 0, color: color.textMuted, fontSize: type.metadataSize, },
})

/** A page of logs, newest first, each opening its detail page, with its count, pagination and empty states. */
export function LogEntries({ logs, total, page, pageHref, empty, showParticipant = false, actions }: {
  logs: ImmersionLog[]
  total: number
  page: number
  pageHref: (page: number) => string
  empty: ReactNode
  /** Names each log's learner, for lists that mix learners. */
  showParticipant?: boolean
  /** Controls shown beside each log, such as a moderator's. */
  actions?: (log: ImmersionLog) => ReactNode
}) {
  if (logs.length === 0) {
    return total > 0 ? (
      <EmptyState
        variant="plain"
        headingLevel={3}
        title="This page is past the end of the updates"
        description={`There are ${numberFormat.format(total)} updates.`}
        actions={<Link {...stylex.props(linkStyles.base)} to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : empty
  }

  const first = (page - 1) * activityPageSize + 1
  const pages = Math.ceil(total / activityPageSize)

  return (
    <>
      {/* Long pages repeat their page links above the list, so the next page is reachable without scrolling 50 rows. */}
      {pages > 1 ? (
        <div {...stylex.props(styles.paginationTop)}>
          <Pagination label="Update pages, top" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} listXstyle={styles.wrap} />
        </div>
      ) : null}
      <ol {...stylex.props(styles.list)} data-entries="">
        {logs.map((log, index) => (
          <Entry
            key={log.id}
            log={log}
            showParticipant={showParticipant}
            actions={actions?.(log)}
            firstOfDay={index === 0 || logDay(logs[index - 1]) !== logDay(log)}
          />
        ))}
      </ol>
      {pages > 1 ? (
        <div {...stylex.props(styles.footer)} data-entries-footer="">
          <p {...stylex.props(styles.footerText)}>{numberFormat.format(first)}–{numberFormat.format(first + logs.length - 1)} of {numberFormat.format(total)} updates</p>
          <Pagination label="Update pages" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} listXstyle={styles.wrap} />
        </div>
      ) : null}
    </>
  )
}

/** Webv2 hides a description tagged NSFW until the reader chooses to see it. */
const logDay = (log: ImmersionLog) => new Date(log.created_at).toISOString().slice(0, 10)

/** A day's first entry shows its date; later entries that day keep it for assistive technology only. */
function Entry({ log, showParticipant, actions, firstOfDay }: {
  log: ImmersionLog
  showParticipant: boolean
  actions: ReactNode
  firstOfDay: boolean
}) {
  const day = logDay(log)
  const subject = `${log.language.name} ${log.activity.name.toLocaleLowerCase()}`
  const nsfw = log.tags.some(tag => tag.toLowerCase() === 'nsfw')
  const meta = [showParticipant ? log.user_display_name ?? 'Unknown user' : null, log.language.name, log.activity.name, formatTracking(log)]

  return (
    <li {...stylex.props(styles.entry, !firstOfDay && styles.sameDay)}>
      <time dateTime={day} {...stylex.props(styles.date, !firstOfDay && a11y.visuallyHidden)}>{formatDate(day)}</time>
      <div {...stylex.props(styles.body)}>
        <Link {...stylex.props(...entryTitle, styles.clamp)} to="/logs/$logId" params={{ logId: log.id }} data-entry-title="">
          {nsfw || !log.description ? subject : log.description}
        </Link>
        <p {...stylex.props(entryStyles.meta)} data-entry-meta="">{meta.filter(Boolean).join(' · ')}</p>
        {nsfw && log.description ? (
          <details {...stylex.props(styles.nsfw)}>
            <summary {...stylex.props(styles.nsfwSummary)}>Show NSFW description</summary>
            <p {...stylex.props(styles.nsfwText)}>{log.description}</p>
          </details>
        ) : null}
        {actions ? <div {...stylex.props(styles.actions)}>{actions}</div> : null}
      </div>
      <span {...stylex.props(styles.score)}>{numberFormat.format(log.score)} <small {...stylex.props(styles.unit)}>{pointsUnit(log.score)}</small></span>
    </li>
  )
}
