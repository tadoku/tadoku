import { buttonClassName, EmptyState, entryLink, listRows, metaLine, Pagination, SectionHeading, StatGrid, StatusTag, Tabbar } from 'paper-ui'
import { activityTone } from '../lib/activity'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { activeDays, contestPageSize, type ContestHistory, type ProfileStatistics as Statistics, type Section } from '../lib/profile'
import { ActivityCalendar } from './ActivityCalendar'
import { ProfileHeader } from './ProfileHeader'
import { renderLink } from './router-link'
import { sheetTabs } from './standings-styles'

type Found = Extract<Statistics, { status: 'found' }>

/** A profile's year: its scores, daily activity, activity split and contest history, as Webv2's statistics show them. */
const muted = 'mb-0 mt-2 text-metadata leading-160 text-muted'
const sectionTitle = 'font-display text-heading leading-section'
const columns = [
  'mt-8 grid grid-cols-halves gap-x-12 gap-y-0 max-lg:grid-cols-single',
  '*:min-w-0 *:border-t *:border-rule *:py-7 [&>:nth-child(3)]:col-span-full [&_h2]:text-heading [&_h2]:leading-section',
].join(' ')
const rows = listRows()

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
    <div className="profile-page">
      <ProfileHeader profile={profile} owner={owner} view="overview" overviewHref={`/profile/${profile.id}/statistics/${year}`} exportable={!empty} />
      <section className="mt-8 border border-rule bg-paper" aria-labelledby="profile-year-heading">
        <h2 id="profile-year-heading" className="sr-only">{name}’s activity in {year}</h2>
        <YearTabs statistics={statistics} year={year} />
        <div className="p-8 max-sm:p-5">
        {empty ? (
          <EmptyState
            variant="plain"
            headingLevel={3}
            title={`No activity recorded in ${year}`}
            actions={owner ? <a className={buttonClassName()} href={links.logs}>Log activity</a> : null}
          />
        ) : (
          <>
            <YearTotals statistics={statistics} year={year} />
            <div className="border-t border-rule pt-6">
              <SectionHeading level={3} id="profile-rhythm-heading" title={`Daily activity in ${year}`} classNames={{ title: sectionTitle }} />
              {activity.ok ? (<>
                <ActivityCalendar
                  userId={profile.id}
                  year={year}
                  today={statistics.today}
                  activity={activity.data}
                  labelledBy="profile-rhythm-heading"
                />
              </>) : <p className={muted}>Could not retrieve activity for {year}.</p>}
            </div>
          </>
        )}
        </div>
      </section>
      <div className={columns}>
        {!empty ? (
          <>
            <Breakdown id="profile-languages" title="Score by language" section={statistics.scores} items={scores => scores.scores.map(score => ({ label: score.language_name ?? score.language_code, score: score.score }))} empty="No scores recorded this year." />
            <Breakdown id="profile-activities" title="Score by activity" section={statistics.split} items={split => split.activities.map(item => ({ label: item.activity_name, score: item.score, fill: activityTone(item.activity_id).fill }))} empty="Not enough data to show the activity split." />
          </>
        ) : null}
        <ContestHistorySection history={statistics.contests} userId={profile.id} year={year} page={page} />
      </div>
    </div>
  )
}

function YearTotals({ statistics, year }: { statistics: Found; year: number }) {
  const { scores, activity } = statistics
  const total = scores.ok ? scores.data.scores.reduce((sum, score) => sum + score.score, 0) : undefined
  const days = daysCounted(year, statistics.today)
  const average = total && days ? `${scoreFormat.format(total / days)} per day in ${year}` : undefined

  return (
    <StatGrid
      className="profile-stats mb-8"
      stats={[
        { label: 'Overall score', value: total === undefined ? 'Unavailable' : scoreFormat.format(total), detail: average },
        ...(activity.ok ? [{ label: 'Logs', value: activity.data.total_updates, detail: daysActive(activeDays(activity.data.scores)) }] : []),
        ...(scores.ok ? [{ label: 'Languages', value: scores.data.scores.length }] : []),
      ]}
    />
  )
}

const daysActive = (days: number) => `${days} active ${days === 1 ? 'day' : 'days'}`

/**
 * One tab per year from the profile's first year to the current one, newest first, as the yearly leaderboard shows
 * them. A year opened by address outside that range gets its own tab, so the page always names the year it shows.
 */
function YearTabs({ statistics, year }: { statistics: Found; year: number }) {
  const first = Number(statistics.profile.created_at.slice(0, 4))
  const current = Number(statistics.today.slice(0, 4))
  const span = Array.from({ length: current - first + 1 }, (_, index) => current - index)
  const years = span.includes(year) ? span : [...span, year].sort((a, b) => b - a)
  if (years.length === 1) return null
  return (
    <Tabbar
      classNames={sheetTabs}
      label="Profile year"
      renderLink={renderLink}
      links={years.map(target => ({
        id: String(target),
        label: String(target),
        href: `/profile/${statistics.profile.id}/statistics/${target}`,
        current: target === year,
      }))}
    />
  )
}

function Breakdown<T>({ id, title, section, items, empty }: {
  id: string
  title: string
  section: Section<T>
  /** `fill` colours the row's bar, for example by activity; it defaults to a neutral. */
  items: (data: T) => { label: string; score: number; fill?: string }[]
  empty: string
}) {
  const rows = section.ok ? items(section.data).sort((left, right) => right.score - left.score) : []
  const total = rows.reduce((sum, row) => sum + row.score, 0)

  return (
    <section aria-labelledby={id}>
      <SectionHeading level={2} id={id} title={title} />
      {!section.ok ? <p className={muted}>Could not retrieve this breakdown.</p> : rows.length ? (
        <ol className="m-0 list-none p-0 [&>li+li]:mt-5">
          {rows.map(row => (
            <li key={row.label}>
              <div className="mb-2 flex justify-between gap-4 text-ui wrap-anywhere">
                <span>{row.label}</span>
                <strong className="whitespace-nowrap tabular-nums">
                  {scoreFormat.format(row.score)} <small className="font-regular text-muted">{pointsUnit(row.score)}</small>
                </strong>
              </div>
              <span className="block h-2 bg-rule-subtle" aria-hidden="true">
                <span className={`block h-full min-w-0.5 ${row.fill ?? 'bg-muted'}`} style={{ width: `${total ? row.score / total * 100 : 0}%` }} />
              </span>
            </li>
          ))}
        </ol>
      ) : <p className={muted}>{empty}</p>}
    </section>
  )
}

function ContestHistorySection({ history, userId, year, page }: {
  history: Section<ContestHistory>
  userId: string
  year: number
  page: number
}) {
  const links = useLinks()
  const pages = history.ok ? Math.ceil(history.data.total / contestPageSize) : 0

  return (
    <section aria-labelledby="profile-contests">
      <SectionHeading level={2} id="profile-contests" title="Contests" />
      {!history.ok ? <p className={muted}>Could not retrieve contests.</p> : history.data.contests.length ? (
        <ul className={rows.root()}>
          {history.data.contests.map(contest => (
            <li key={contest.id} className={rows.item()}>
              <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
                <a className={entryLink({ className: 'profile-entry-title' })} href={links.contestProfile(contest.id, userId)}>{contest.title}</a>
                {contest.private ? <StatusTag>Unlisted</StatusTag> : null}
              </span>
              <span className={metaLine({ className: 'profile-entry-meta' })}>{[formatPeriod(contest), ...contest.languages].join(' · ')}</span>
            </li>
          ))}
        </ul>
      ) : <p className={muted}>Hasn’t participated in any contests in {year}.</p>}
      {pages > 1 && page <= pages ? (
        <Pagination className="mt-5" label="Contest pages" totalPages={pages} currentPage={page} getHref={target => `/profile/${userId}/statistics/${year}?contests=${target}`} />
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
