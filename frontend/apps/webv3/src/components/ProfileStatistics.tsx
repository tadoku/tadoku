import { EmptyState, Pagination, SectionHeading, StatGrid, StatusTag, buttonClassName } from 'paper-ui'
import { formatPeriod, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { activeDays, contestPageSize, type ContestHistory, type ProfileStatistics as Statistics, type Section } from '../lib/profile'
import { ActivityCalendar } from './ActivityCalendar'
import { ProfileHeader } from './ProfileHeader'
import { YearStepper } from './YearStepper'

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
    <div className="profile-page">
      <ProfileHeader profile={profile} owner={owner} view="overview" overviewHref={`/profile/${profile.id}/statistics/${year}`} />
      <section className="profile-year" aria-labelledby="profile-year-heading">
        <div className="profile-year-head">
          <h2 id="profile-year-heading" className="paper-visually-hidden">{name}’s activity in {year}</h2>
          <YearControl statistics={statistics} year={year} />
        </div>
        {empty ? (
          <EmptyState
            variant="plain"
            className="profile-empty-year"
            headingLevel={3}
            title={`No activity recorded in ${year}`}
            actions={owner ? <a className={buttonClassName()} href={links.logs}>Log activity</a> : null}
          />
        ) : (
          <>
            <YearTotals statistics={statistics} year={year} />
            <div className="profile-rhythm">
              <SectionHeading level={3} id="profile-rhythm-heading" title="Daily activity" />
              {activity.ok ? (<>
                <ActivityCalendar
                  userId={profile.id}
                  year={year}
                  today={statistics.today}
                  activity={activity.data}
                  labelledBy="profile-rhythm-heading"
                />
              </>) : <p className="profile-muted">Could not retrieve updates for {year}.</p>}
            </div>
          </>
        )}
      </section>
      <div className="profile-columns">
        {!empty ? (
          <>
            <Breakdown id="profile-languages" title="Score by language" section={statistics.scores} items={scores => scores.scores.map(score => ({ label: score.language_name ?? score.language_code, score: score.score }))} empty="No scores recorded this year." />
            <Breakdown id="profile-activities" title="Score by activity" section={statistics.split} items={split => split.activities.map(item => ({ label: item.activity_name, score: item.score }))} empty="Not enough data to show the activity split." />
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
  const average = total && days ? `${scoreFormat.format(total / days)} per day` : undefined

  return (
    <StatGrid
      className="profile-stats"
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
    <section aria-labelledby={id}>
      <SectionHeading level={2} id={id} title={title} />
      {!section.ok ? <p className="profile-muted">Could not retrieve this breakdown.</p> : rows.length ? (
        <ol className="profile-breakdown">
          {rows.map(row => (
            <li key={row.label}>
              <div><span>{row.label}</span><strong>{scoreFormat.format(row.score)} <small>{pointsUnit(row.score)}</small></strong></div>
              <span className="profile-bar" aria-hidden="true">
                <span style={{ width: `${total ? row.score / total * 100 : 0}%` }} />
              </span>
            </li>
          ))}
        </ol>
      ) : <p className="profile-muted">{empty}</p>}
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
      {!history.ok ? <p className="profile-muted">Could not retrieve contests.</p> : history.data.contests.length ? (
        <ul className="profile-contests">
          {history.data.contests.map(contest => (
            <li key={contest.id}>
              <span className="profile-contest-title">
                <a className="profile-entry-title" href={links.contestProfile(contest.id, userId)}>{contest.title}</a>
                {contest.private ? <StatusTag>Unlisted</StatusTag> : null}
              </span>
              <span className="profile-entry-meta">{[formatPeriod(contest), ...contest.languages].join(' · ')}</span>
            </li>
          ))}
        </ul>
      ) : <p className="profile-muted">Hasn’t participated in any contests in {year}.</p>}
      {pages > 1 && page <= pages ? (
        <Pagination label="Contest pages" totalPages={pages} currentPage={page} getHref={target => `/profile/${userId}/statistics/${year}?contests=${target}`} />
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
