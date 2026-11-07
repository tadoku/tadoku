import * as stylex from '@stylexjs/stylex'
import { a11y, buttonStyles, EmptyState, Pagination, SectionHeading, StatGrid, StatusTag } from 'paper-ui'
import { entryStyles, entryTitle, pageStyles } from './page-styles'
import { border, breakpoints, color, font, fontSize, fontWeight, lineHeight, size, space } from 'paper-ui/tokens.stylex'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { activeDays, contestPageSize, type ContestHistory, type ProfileStatistics as Statistics, type Section } from '../lib/profile'
import { ActivityCalendar } from './ActivityCalendar'
import { ProfileHeader } from './ProfileHeader'
import { YearStepper } from './YearStepper'

const ruledBlock = { borderBlockStartColor: color.ruleDefault, borderBlockStartStyle: 'solid', borderBlockStartWidth: border.static } as const

const styles = stylex.create({
  year: {
    padding: { default: space[32], [breakpoints.phone]: space[20] },
    borderColor: color.ruleDefault,
    borderStyle: 'solid',
    borderWidth: border.static,
    backgroundColor: color.surfacePaper,
    marginBlockStart: space[32],
  },
  yearHead: {
    marginBlockEnd: space[20],
    marginBlockStart: `calc(-1 * ${space[8]})`,
    marginInlineEnd: 0,
    marginInlineStart: `calc(-1 * ${space[12]})`,
  },
  stats: { marginBlockEnd: space[32], marginBlockStart: space[28] },
  rhythm: { ...ruledBlock, paddingBlockStart: space[24] },
  emptyYear: ruledBlock,
  /** One section role across the profile: the in-sheet calendar heading matches the sections below it. */
  sectionTitle: { fontFamily: font.display, fontSize: fontSize[22], fontWeight: fontWeight.bold, lineHeight: lineHeight[30] },
  columns: {
    columnGap: space[48],
    display: 'grid',
    gridTemplateColumns: { default: 'repeat(2, minmax(0, 1fr))', [breakpoints.lg]: 'minmax(0, 1fr)' },
    marginBlockStart: space[32],
    rowGap: 0,
  },
  column: { ...ruledBlock, paddingBlock: space[28], minInlineSize: 0, },
  /** The two score breakdowns share a row; the contest list, the longest column, runs full width below them. */
  fullRow: { gridColumnEnd: '-1',
 gridColumnStart: '1' },
  breakdown: { margin: 0, padding: 0, listStyle: 'none', },
  breakdownItem: { marginBlockStart: { default: space[20], ':first-child': 0 } },
  breakdownLabel: {
    gap: space[16],
    display: 'flex',
    fontSize: fontSize[15],
    justifyContent: 'space-between',
    marginBlockEnd: space[8],
    overflowWrap: 'anywhere',
  },
  breakdownScore: { fontVariantNumeric: 'tabular-nums', whiteSpace: 'nowrap' },
  unit: { color: color.textMuted, fontWeight: fontWeight.regular },
  bar: { backgroundColor: color.ruleDefault, blockSize: size[8], display: 'block' },
  /** Bars use the calendar's fourth heat step, so the data colour holds about 3.9:1 on the track in both themes. */
  barFill: { backgroundColor: color.heat3, blockSize: '100%', display: 'block', minInlineSize: size[2] },
  barWidth: (width: string) => ({ inlineSize: width }),
  contestTitle: { alignItems: 'center', columnGap: space[12], display: 'flex', flexWrap: 'wrap', rowGap: space[4] },
  contestPages: { marginBlockStart: space[20] },
})

type Found = Extract<Statistics, { status: 'found' }>

/** A profile's year: its scores, daily activity, activity split and contest history, as Webv2's statistics show them. */
export function ProfileStatistics({ statistics, year, page, owner }: {
  statistics: Found
  year: number
  page: number
  owner: boolean
}) {
  const { profile, activity } = statistics
  const links = useLinks()
  const name = profile.display_name
  const empty = activity.ok && activity.data.total_updates === 0

  return (
    <div>
      <ProfileHeader profile={profile} owner={owner} view="overview" overviewHref={`/profile/${profile.id}/statistics/${year}`} />
      <section {...stylex.props(styles.year)} aria-labelledby="profile-year-heading">
        <div {...stylex.props(styles.yearHead)}>
          <h2 id="profile-year-heading" {...stylex.props(a11y.visuallyHidden)}>{name}’s activity in {year}</h2>
          <YearControl statistics={statistics} year={year} />
        </div>
        {empty ? (
          <EmptyState
            variant="plain"
            xstyle={styles.emptyYear}
            headingLevel={3}
            title={`No activity recorded in ${year}`}
            actions={owner ? <a {...stylex.props(...buttonStyles())} href={links.logs}>Log activity</a> : null}
          />
        ) : (
          <>
            <YearTotals statistics={statistics} year={year} />
            <div {...stylex.props(styles.rhythm)}>
              <SectionHeading level={3} id="profile-rhythm-heading" title="Daily activity" titleXstyle={styles.sectionTitle} />
              {activity.ok ? (<>
                <ActivityCalendar
                  userId={profile.id}
                  year={year}
                  today={statistics.today}
                  activity={activity.data}
                  labelledBy="profile-rhythm-heading"
                />
              </>) : <p {...stylex.props(pageStyles.muted)}>Could not retrieve updates for {year}.</p>}
            </div>
          </>
        )}
      </section>
      <div {...stylex.props(styles.columns)}>
        {!empty ? (
          <>
            <Breakdown id="profile-languages" title="Score by language" section={statistics.scores} items={scores => scores.scores.map(score => ({ label: score.language_name ?? score.language_code, score: score.score }))} empty="No scores recorded this year." />
            <Breakdown id="profile-activities" title="Score by activity" section={statistics.split} items={split => split.activities.map(item => ({ label: item.activity_name, score: item.score }))} empty="Not enough data to show the activity split." />
          </>
        ) : null}
        <ContestHistorySection history={statistics.contests} userId={profile.id} year={year} page={page} fullRow={!empty} />
      </div>
    </div>
  )
}

function YearTotals({ statistics, year }: { statistics: Found; year: number }) {
  const { scores, activity } = statistics
  const total = scores.ok ? scores.data.scores.reduce((sum, score) => sum + score.score, 0) : undefined
  const days = daysCounted(year, statistics.today)
  const average = total && days ? `${scoreFormat.format(total / days)} per day` : undefined

  return (
    <StatGrid
      xstyle={styles.stats}
      stats={[
        { label: 'Overall score', value: total === undefined ? 'Unavailable' : scoreFormat.format(total), detail: average },
        ...(activity.ok ? [{ label: 'Updates', value: activity.data.total_updates, detail: `${activeDays(activity.data.scores)} active days` }] : []),
        ...(scores.ok ? [{ label: 'Languages', value: scores.data.scores.length }] : []),
      ]}
    />
  )
}

/** Steps through the years from the profile's first year to the current one. */
function YearControl({ statistics, year }: { statistics: Found; year: number }) {
  return (
    <YearStepper
      year={year}
      first={Number(statistics.profile.created_at.slice(0, 4))}
      current={Number(statistics.today.slice(0, 4))}
      href={target => `/profile/${statistics.profile.id}/statistics/${target}`}
      label="Profile year"
    />
  )
}

function Breakdown<T>({ id, title, section, items, empty }: {
  id: string
  title: string
  section: Section<T>
  items: (data: T) => { label: string; score: number }[]
  empty: string
}) {
  const rows = section.ok ? items(section.data).sort((left, right) => right.score - left.score) : []
  const total = rows.reduce((sum, row) => sum + row.score, 0)

  return (
    <section {...stylex.props(styles.column)} aria-labelledby={id}>
      <SectionHeading level={2} id={id} title={title} titleXstyle={styles.sectionTitle} />
      {!section.ok ? <p {...stylex.props(pageStyles.muted)}>Could not retrieve this breakdown.</p> : rows.length ? (
        <ol {...stylex.props(styles.breakdown)}>
          {rows.map(row => (
            <li key={row.label} {...stylex.props(styles.breakdownItem)}>
              <div {...stylex.props(styles.breakdownLabel)}>
                <span>{row.label}</span>
                <strong {...stylex.props(styles.breakdownScore)}>
                  {scoreFormat.format(row.score)} <small {...stylex.props(styles.unit)}>{pointsUnit(row.score)}</small>
                </strong>
              </div>
              <span {...stylex.props(styles.bar)} aria-hidden="true">
                <span {...stylex.props(styles.barFill, styles.barWidth(`${total ? row.score / total * 100 : 0}%`))} />
              </span>
            </li>
          ))}
        </ol>
      ) : <p {...stylex.props(pageStyles.muted)}>{empty}</p>}
    </section>
  )
}

function ContestHistorySection({ history, userId, year, page, fullRow }: {
  history: Section<ContestHistory>
  /** Below the two breakdowns the list takes the full width. */
  fullRow: boolean
  userId: string
  year: number
  page: number
}) {
  const links = useLinks()
  const pages = history.ok ? Math.ceil(history.data.total / contestPageSize) : 0

  return (
    <section {...stylex.props(styles.column, fullRow && styles.fullRow)} aria-labelledby="profile-contests">
      <SectionHeading level={2} id="profile-contests" title="Contests" titleXstyle={styles.sectionTitle} />
      {!history.ok ? <p {...stylex.props(pageStyles.muted)}>Could not retrieve contests.</p> : history.data.contests.length ? (
        <ul {...stylex.props(entryStyles.contestList)}>
          {history.data.contests.map(contest => (
            <li key={contest.id} {...stylex.props(entryStyles.contestItem)}>
              <span {...stylex.props(styles.contestTitle)}>
                <a {...stylex.props(...entryTitle)} href={links.contestProfile(contest.id, userId)}>{contest.title}</a>
                {contest.private ? <StatusTag>Unlisted</StatusTag> : null}
              </span>
              <span {...stylex.props(entryStyles.meta)}>{[formatPeriod(contest), ...contest.languages].join(' · ')}</span>
            </li>
          ))}
        </ul>
      ) : <p {...stylex.props(pageStyles.muted)}>Hasn’t participated in any contests in {year}.</p>}
      {pages > 1 && page <= pages ? (
        <Pagination
          label="Contest pages"
          totalPages={pages}
          currentPage={page}
          getHref={target => `/profile/${userId}/statistics/${year}?contests=${target}`}
          xstyle={styles.contestPages}
        />
      ) : null}
    </section>
  )
}

/** The UTC calendar days the daily average spans: all of a past year, the current year up to today, none ahead. */
function daysCounted(year: number, today: string) {
  const current = Number(today.slice(0, 4))
  if (year > current) return 0
  const end = year === current ? Date.parse(`${today}T00:00:00Z`) : Date.UTC(year, 11, 31)
  return (end - Date.UTC(year, 0, 1)) / 86_400_000 + 1
}
