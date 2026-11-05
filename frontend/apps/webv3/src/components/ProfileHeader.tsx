import { Breadcrumb, Tabbar, buttonClassName } from 'paper-ui'
import type { ImmersionUserProfile } from 'api-client'
import { renderLink } from './router-link'
import { useLinks } from '../lib/links'

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
      <Breadcrumb items={[{ id: 'home', label: 'Home', href: '/' }, { id: 'profile', label: profile.display_name }]} renderLink={renderLink} />
      <header className="profile-header">
        <div className="profile-identity">
          <span className="profile-avatar" aria-hidden="true">{profile.display_name.slice(0, 2).toLocaleUpperCase()}</span>
          <div>
            <h1 className="paper-type-page">{profile.display_name}</h1>
            <p>Tracking since {profile.created_at.slice(0, 4)}</p>
          </div>
        </div>
        {owner ? (
          <div className="profile-export" role="group" aria-label="Export your logs">
            <a className={buttonClassName({ variant: 'outline' })} href={links.logExport('csv')}>Download CSV</a>
            <a className={buttonClassName({ variant: 'outline' })} href={links.logExport('json')}>Download JSON</a>
          </div>
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
