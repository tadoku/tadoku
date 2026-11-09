import { Link } from '@tanstack/react-router'
import { EmptyState, Flash, buttonClassName } from 'paper-ui'
import { formatDate } from '../lib/format'
import { useLinks } from '../lib/links'
import type { ProfileActivity as Activity } from '../lib/profile'
import { LogEntries } from './LogEntries'
import { ProfileHeader } from './ProfileHeader'
import { AppLink } from './router-link'

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
      <section className="mt-8" aria-label="Activity">
        {date ? (
          <Flash className="mb-4" action={<Link className="text-link hover:text-link-hover" to={links.profileUpdates(profile.id)}>Show all activity</Link>}>
            Showing activity on {formatDate(date)}
          </Flash>
        ) : null}
        {!logs.ok ? <p className="mb-0 mt-2 text-metadata leading-160 text-muted">Could not load activity. Try reloading this page.</p> : (
          <LogEntries
            logs={logs.data.logs}
            total={logs.data.total_size}
            page={page}
            pageHref={target => `/profile/${profile.id}/updates/${target}${date ? `?date=${date}` : ''}`}
            empty={(
              <EmptyState
                variant="plain"
                headingLevel={3}
                title={date ? `No activity on ${formatDate(date)}` : 'No activity yet'}
                description={owner
                  ? 'Start with what you read or listened to today. A few pages or minutes are enough.'
                  : 'Logs appear here once they have been recorded.'}
                actions={owner ? <AppLink className={buttonClassName()} href={links.logs}>Log activity</AppLink> : null}
              />
            )}
          />
        )}
      </section>
    </div>
  )
}
