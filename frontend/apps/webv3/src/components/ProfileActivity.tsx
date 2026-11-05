import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { EmptyState, Flash, Pagination, SectionHeading, buttonClassName } from 'paper-ui'
import type { ImmersionLog } from 'api-client'
import { formatDate, numberFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { activityPageSize, type ProfileActivity as Activity } from '../lib/profile'
import { ProfileHeader } from './ProfileHeader'
import { renderLink } from './router-link'

type Found = Extract<Activity, { status: 'found' }>

/** A page of a profile's logs, newest first, optionally one UTC day's, each opening its detail page. */
export function ProfileActivity({ activity, page, date, owner }: {
  activity: Found
  page: number
  date: string | undefined
  owner: boolean
}) {
  const { profile, logs } = activity
  const links = useLinks()

  return (
    <div className="profile-page">
      <ProfileHeader profile={profile} owner={owner} view="activity" overviewHref={links.profile(profile.id)} />
      <section className="profile-activity" aria-labelledby="profile-activity-heading">
        <SectionHeading level={2} id="profile-activity-heading" title="Activity log" />
        {date ? (
          <Flash className="profile-day" action={<Link className="text-link" to={links.profileUpdates(profile.id)}>Show all updates</Link>}>
            Showing updates on {formatDate(date)}
          </Flash>
        ) : null}
        {!logs.ok ? <p className="profile-muted">Could not load updates. Try reloading this page.</p> : (
          <Entries
            logs={logs.data.logs}
            total={logs.data.total_size}
            page={page}
            pageHref={target => `/profile/${profile.id}/updates/${target}${date ? `?date=${date}` : ''}`}
            empty={(
              <EmptyState
                variant="plain"
                headingLevel={3}
                title={date ? `No updates on ${formatDate(date)}` : 'No activity yet'}
                description={owner
                  ? 'Start with what you read or listened to today. A few pages or minutes are enough.'
                  : 'Updates appear here once they have been recorded.'}
                actions={owner ? <a className={buttonClassName()} href={links.logs}>Log activity</a> : null}
              />
            )}
          />
        )}
      </section>
    </div>
  )
}

function Entries({ logs, total, page, pageHref, empty }: {
  logs: ImmersionLog[]
  total: number
  page: number
  pageHref: (page: number) => string
  empty: ReactNode
}) {
  if (logs.length === 0) {
    return total > 0 ? (
      <EmptyState
        variant="plain"
        headingLevel={3}
        title="This page is past the end of the updates"
        description={`This profile has ${numberFormat.format(total)} updates.`}
        actions={<Link className="text-link" to={pageHref(1)}>Go to the first page</Link>}
      />
    ) : empty
  }

  const first = (page - 1) * activityPageSize + 1
  const pages = Math.ceil(total / activityPageSize)

  return (
    <>
      <ol className="profile-entries">
        {logs.map(log => <Entry key={log.id} log={log} />)}
      </ol>
      <div className="profile-entries-footer">
        <p>{numberFormat.format(first)}–{numberFormat.format(first + logs.length - 1)} of {numberFormat.format(total)} updates</p>
        {pages > 1 ? <Pagination label="Update pages" totalPages={pages} currentPage={page} getHref={pageHref} renderLink={renderLink} /> : null}
      </div>
    </>
  )
}

/** Webv2 hides a description tagged NSFW until the reader chooses to see it. */
function Entry({ log }: { log: ImmersionLog }) {
  const day = new Date(log.created_at).toISOString().slice(0, 10)
  const subject = `${log.language.name} ${log.activity.name.toLocaleLowerCase()}`
  const nsfw = log.tags.some(tag => tag.toLowerCase() === 'nsfw')

  return (
    <li>
      <time dateTime={day} className="profile-entry-date">{formatDate(day)}</time>
      <div className="profile-entry-body">
        <Link className="profile-entry-title text-link" to="/logs/$logId" params={{ logId: log.id }}>
          {nsfw || !log.description ? subject : log.description}
        </Link>
        <p className="profile-entry-meta">{[log.language.name, log.activity.name, tracking(log)].filter(Boolean).join(' · ')}</p>
        {nsfw && log.description ? (
          <details className="profile-entry-nsfw">
            <summary>Show NSFW description</summary>
            <p>{log.description}</p>
          </details>
        ) : null}
      </div>
      <span className="profile-entry-score">{numberFormat.format(log.score)}<small>points</small></span>
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
