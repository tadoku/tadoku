import * as stylex from '@stylexjs/stylex'
import { iconSizes, linkStyles, Tabbar, typography } from 'paper-ui'
import { border, breakpoints, color, font, fontSize, fontWeight, size, space, tracking } from 'paper-ui/tokens.stylex'
import { ArrowDownTrayIcon } from 'paper-ui/icons'
import type { ImmersionUserProfile } from 'api-client'
import { renderLink } from './router-link'
import { useLinks } from '../lib/links'

const styles = stylex.create({
  header: {
    gap: space[16],
    alignItems: { default: 'center', [breakpoints.narrow]: 'start' },
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    marginBlockEnd: space[32],
    marginBlockStart: space[8],
  },
  identity: { gap: space[20], alignItems: 'center', display: 'flex', minInlineSize: 0, },
  identityText: { minInlineSize: 0 },
  avatar: {    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    placeItems: 'center',
    backgroundColor: color.surfacePaper,
    blockSize: size[64],
 boxSizing: "border-box",
    display: { default: 'grid', [breakpoints.phone]: 'none' },
    flexBasis: size[64],
    flexGrow: 0,
    flexShrink: 0,
    fontFamily: font.body,
    fontSize: fontSize[20],
    fontWeight: fontWeight.semibold,
    letterSpacing: tracking.wide,
    lineHeight: 'normal',
  },
  name: { marginInline: 0, marginBlockEnd: space[8], marginBlockStart: 0, overflowWrap: 'anywhere', },
  since: { margin: 0, color: color.textMuted, fontSize: fontSize[15], },
  export: {
    margin: 0,
    gap: space[6],
    alignItems: 'center',
    color: color.textMuted,
    display: 'flex',
    flexBasis: { default: null, [breakpoints.narrow]: '100%' },
    fontSize: fontSize[14],
  },
})

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
      <header {...stylex.props(styles.header)}>
        <div {...stylex.props(styles.identity)}>
          <span {...stylex.props(styles.avatar)} aria-hidden="true">{initials(profile.display_name)}</span>
          <div {...stylex.props(styles.identityText)}>
            <h1 {...stylex.props(typography.page, styles.name)}>{profile.display_name}</h1>
            <p {...stylex.props(styles.since)}>Tracking since {profile.created_at.slice(0, 4)}</p>
          </div>
        </div>
        {owner ? (
          <p {...stylex.props(styles.export)}>
            <ArrowDownTrayIcon {...stylex.props(iconSizes.compact)} aria-hidden="true" />
            Export logs: <a {...stylex.props(linkStyles.base)} href={links.logExport('csv')} aria-label="Download CSV">CSV</a> · <a {...stylex.props(linkStyles.base)} href={links.logExport('json')} aria-label="Download JSON">JSON</a>
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
