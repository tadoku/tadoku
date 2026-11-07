import * as stylex from '@stylexjs/stylex'
import { Link } from '@tanstack/react-router'
import { buttonStyles, EmptyState, Flash, linkStyles } from 'paper-ui'
import { pageStyles } from './page-styles'
import { space } from 'paper-ui/tokens.stylex'
import { formatDate } from '../lib/format'
import { useLinks } from '../lib/links'
import type { ProfileActivity as Activity } from '../lib/profile'
import { LogEntries } from './LogEntries'
import { ProfileHeader } from './ProfileHeader'

const styles = stylex.create({ day: { marginBlockEnd: space[16] } })

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
    <div>
      <ProfileHeader profile={profile} owner={owner} view="activity" overviewHref={links.profile(profile.id)} />
      <section {...stylex.props(pageStyles.section)} aria-label="Activity log">
        {date ? (
          <Flash xstyle={styles.day} action={<Link {...stylex.props(linkStyles.base)} to={links.profileUpdates(profile.id)}>Show all updates</Link>}>
            Showing updates on {formatDate(date)}
          </Flash>
        ) : null}
        {!logs.ok ? <p {...stylex.props(pageStyles.muted)}>Could not load updates. Try reloading this page.</p> : (
          <LogEntries
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
                actions={owner ? <a {...stylex.props(...buttonStyles())} href={links.logs}>Log activity</a> : null}
              />
            )}
          />
        )}
      </section>
    </div>
  )
}
