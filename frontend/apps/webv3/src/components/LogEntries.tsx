import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { EmptyState, Pagination, entryLink, entryList, listFooter, metaLine } from 'paper-ui'
import type { ImmersionLog } from 'api-client'
import { formatDate, formatTracking, numberFormat, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { activityPageSize } from '../lib/profile'
import { ActivityName } from './ActivityName'
import { renderLink } from './router-link'

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
        title="This page is past the end of the activity"
        description={`There are ${numberFormat.format(total)} logs.`}
        actions={<Link className="text-link hover:text-link-hover" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : empty
  }

  const first = (page - 1) * activityPageSize + 1
  const pages = Math.ceil(total / activityPageSize)

  return (
    <>
      {/* Long pages repeat their page links above the list, so the next page is reachable without scrolling 50 rows. */}
      {pages > 1 ? (
        <div className="mb-1 mt-2 flex justify-end">
          <Pagination label="Update pages, top" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} classNames={{ list: 'flex-wrap' }} />
        </div>
      ) : null}
      <ol className={entries.root({ className: 'profile-entries' })}>
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
        <div className={footer.root({ className: 'profile-entries-footer border-t border-rule py-4' })}>
          <p className={footer.summary()}>{numberFormat.format(first)}–{numberFormat.format(first + logs.length - 1)} of {numberFormat.format(total)} logs</p>
          <Pagination label="Update pages" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} classNames={{ list: 'flex-wrap' }} />
        </div>
      ) : null}
    </>
  )
}

const entries = entryList()
const footer = listFooter()

/** Webv2 hides a description tagged NSFW until the reader chooses to see it. */
const logDay = (log: ImmersionLog) => new Date(log.created_at).toISOString().slice(0, 10)

const participantLink = 'text-ink no-underline hover:text-link hover:underline'

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
  const links = useLinks()
  const tracking = formatTracking(log)
  const participant = showParticipant
    ? log.user_display_name
      ? <a className={participantLink} href={links.profile(log.user_id)}>{log.user_display_name}</a>
      : 'Unknown user'
    : null
  const meta = [
    log.language.name,
    <ActivityName activity={log.activity} />,
    tracking || null,
  ].filter(part => part !== null)

  return (
    <li className={entries.item({ className: firstOfDay ? undefined : entries.sameDay() })}>
      <time dateTime={day} className={entries.date({ className: firstOfDay ? undefined : 'sr-only' })}>{formatDate(day)}</time>
      <div className={entries.body()}>
        {participant ? <p className="mb-0.5 mt-0 text-sm font-semibold">{participant}</p> : null}
        <Link className={entryLink({ clamp: true, className: `profile-entry-title ${participant ? 'font-regular' : ''}` })} to="/logs/$logId" params={{ logId: log.id }}>
          {nsfw || !log.description ? subject : log.description}
        </Link>
        <p className={metaLine({ className: 'profile-entry-meta flex flex-wrap gap-x-3' })}>
          {meta.map((part, index) => <span key={index} className="whitespace-nowrap">{part}</span>)}
        </p>
        {nsfw && log.description ? (
          <details className="mt-2.5 text-sm">
            <summary className="cursor-pointer text-muted">Show NSFW description</summary>
            <p className="mb-0 mt-2 max-w-measure-note leading-170 wrap-anywhere">{log.description}</p>
          </details>
        ) : null}
      </div>
      {/* Actions sit under the score, so they add no row of their own. */}
      <div className={entries.score({ className: 'grid content-start justify-items-end gap-1' })}>
        <span>{scoreFormat.format(log.score)} <small>{pointsUnit(log.score)}</small></span>
        {actions}
      </div>
    </li>
  )
}
