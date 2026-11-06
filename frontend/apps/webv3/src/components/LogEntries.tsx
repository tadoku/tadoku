import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { EmptyState, Pagination } from 'paper-ui'
import type { ImmersionLog } from 'api-client'
import { formatDate, numberFormat, pointsUnit } from '../lib/format'
import { activityPageSize } from '../lib/profile'
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
        title="This page is past the end of the updates"
        description={`There are ${numberFormat.format(total)} updates.`}
        actions={<Link className="text-link" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : empty
  }

  const first = (page - 1) * activityPageSize + 1
  const pages = Math.ceil(total / activityPageSize)

  return (
    <>
      <ol className="profile-entries">
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
        <div className="profile-entries-footer">
          <p>{numberFormat.format(first)}–{numberFormat.format(first + logs.length - 1)} of {numberFormat.format(total)} updates</p>
          <Pagination label="Update pages" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} />
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
  const meta = [showParticipant ? log.user_display_name ?? 'Unknown user' : null, log.language.name, log.activity.name, tracking(log)]

  return (
    <li className={firstOfDay ? undefined : 'profile-entry--same-day'}>
      <time dateTime={day} className={firstOfDay ? 'profile-entry-date' : 'profile-entry-date paper-visually-hidden'}>{formatDate(day)}</time>
      <div className="profile-entry-body">
        <Link className="profile-entry-title" to="/logs/$logId" params={{ logId: log.id }}>
          {nsfw || !log.description ? subject : log.description}
        </Link>
        <p className="profile-entry-meta">{meta.filter(Boolean).join(' · ')}</p>
        {nsfw && log.description ? (
          <details className="profile-entry-nsfw">
            <summary>Show NSFW description</summary>
            <p>{log.description}</p>
          </details>
        ) : null}
        {actions ? <div className="profile-entry-actions">{actions}</div> : null}
      </div>
      <span className="profile-entry-score">{numberFormat.format(log.score)} <small>{pointsUnit(log.score)}</small></span>
    </li>
  )
}

/** The amount in its unit and the duration, as Webv2 lists them, for example `12 pages · 30 min`. */
function tracking(log: ImmersionLog) {
  const values = []
  if (log.unit_name) values.push(`${numberFormat.format(log.amount)} ${log.unit_name.toLowerCase()}${log.amount === 1 ? '' : 's'}`)
  if (log.duration_seconds !== undefined) values.push(`${numberFormat.format(log.duration_seconds / 60)} min`)
  return values.join(' · ')
}
