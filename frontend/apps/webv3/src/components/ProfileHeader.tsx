import { Tabbar, typography } from 'paper-ui'
import { ArrowDownTrayIcon, iconClassName } from 'paper-ui/icons'
import type { ImmersionUserProfile } from 'api-client'
import { renderLink } from './router-link'
import { useLinks } from '../lib/links'

/** A profile's identity and its views. Only the profile's own account sees the export of its logs. */
export function ProfileHeader({ profile, owner, view, overviewHref, exportable = true }: {
  profile: ImmersionUserProfile
  owner: boolean
  view: 'overview' | 'activity'
  overviewHref: string
  exportable?: boolean
}) {
  const links = useLinks()

  return (
    <>
      <header className="mb-5 mt-1 flex flex-wrap items-center justify-between gap-4 max-md:items-start">
        <div className="flex min-w-0 items-center gap-5">
          <div className="min-w-0">
            <h1 className={typography({ role: 'page', className: 'mb-2 mt-0 wrap-anywhere' })}>{profile.display_name}</h1>
          </div>
        </div>
        {owner && exportable ? (
          <p className="m-0 flex items-center gap-1.5 text-sm text-muted max-md:basis-full">
            <ArrowDownTrayIcon className={iconClassName('compact')} aria-hidden="true" />
            Export logs: <a className="inline-block px-1 py-2 text-link hover:text-link-hover" href={links.logExport('csv')} aria-label="Download CSV">CSV</a><a className="inline-block px-1 py-2 text-link hover:text-link-hover" href={links.logExport('json')} aria-label="Download JSON">JSON</a>
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
