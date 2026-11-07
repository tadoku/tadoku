import { Tabbar } from 'paper-ui'
import type { ImmersionUserProfile } from 'api-client'
import { renderLink } from './router-link'
import { useLinks } from '../lib/links'

/** The first letters of the first two words, or the first two letters of a one-word name. */
const initials = (name: string) => {
  const words = name.trim().split(/\s+/)
  return (words.length > 1 ? words[0][0] + words[1][0] : name.trim().slice(0, 2)).toLocaleUpperCase()
}

/** A profile's identity and its views. Only the profile's own account sees the export of its logs. */
export function ProfileHeader({ profile, owner, view, overviewHref }: {
  profile: ImmersionUserProfile
  owner: boolean
  view: 'overview' | 'activity'
  overviewHref: string
}) {
  const links = useLinks()

  return (
    <>
      <header className="profile-header">
        <div className="profile-identity">
          <span className="profile-avatar" aria-hidden="true">{initials(profile.display_name)}</span>
          <div>
            <h1 className="paper-type-page">{profile.display_name}</h1>
            <p>Tracking since {profile.created_at.slice(0, 4)}</p>
          </div>
        </div>
        {owner ? (
          <p className="profile-export">
            Export logs: <a className="text-link" href={links.logExport('csv')} aria-label="Download CSV">CSV</a> · <a className="text-link" href={links.logExport('json')} aria-label="Download JSON">JSON</a>
          </p>
        ) : null}
      </header>
      <Tabbar
        label="Profile views"
        renderLink={renderLink}
        links={[
          { id: 'overview', label: 'Overview', href: overviewHref, current: view === 'overview' },
          { id: 'activity', label: 'Activity', href: links.profileUpdates(profile.id), current: view === 'activity' },
        ]}
      />
    </>
  )
}
