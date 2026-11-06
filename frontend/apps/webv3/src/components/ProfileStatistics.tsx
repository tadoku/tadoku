import { useEffect, useRef, type ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { Button, EmptyState, HeatmapChart, Pagination, SectionHeading, StatGrid, StatusTag, buttonClassName } from 'paper-ui'
import { ChevronLeftIcon, ChevronRightIcon } from 'paper-ui/icons'
import { pointsUnit } from '../lib/format'
import { useLinks } from '../lib/links'
import { contestPageSize, type ContestHistory, type ProfileStatistics as Statistics, type Section } from '../lib/profile'
import { ProfileHeader } from './ProfileHeader'

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
            description={owner
              ? 'A few pages or minutes are a good place to start. Your personal record grows with every entry.'
              : 'Scores and the activity calendar appear after the first update of the year.'}
            actions={owner ? <a className={buttonClassName()} href={links.logs}>Log activity</a> : null}
          />
        ) : (
          <>
            <YearTotals statistics={statistics} year={year} />
            <div className="profile-rhythm">
              <SectionHeading level={3} id="profile-rhythm-heading" title="Daily activity" />
              {activity.ok ? (<>
                <HeatmapScroll latest={year === Number(statistics.today.slice(0, 4))}>
                  <HeatmapChart
                    today={statistics.today}
                    id={`profile-${profile.id}-${year}`}
                    year={year}
                    getHref={date => links.profileUpdates(profile.id, date)}
                    data={activity.data.scores.map(day => ({
                      date: day.date,
                      value: day.score,
                      tooltip: `${scoreFormat.format(day.score)} ${pointsUnit(day.score)} on ${formatDay(day.date)}`,
                    }))}
                  />
                </HeatmapScroll>
                <form className="profile-day-picker" method="get" action={links.profileUpdates(profile.id)}>
                  <label htmlFor="profile-day">Show a day’s updates</label>
                  <input
                    id="profile-day"
                    className="paper-input"
                    type="date"
                    name="date"
                    required
                    min={`${year}-01-01`}
                    max={year === Number(statistics.today.slice(0, 4)) ? statistics.today : `${year}-12-31`}
                  />
                  <button type="submit" className={buttonClassName({ variant: 'outline' })}>Show</button>
                </form>
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

/** On narrow screens the year scrolls sideways; the current year opens at its latest weeks. */
function HeatmapScroll({ latest, children }: { latest: boolean; children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (latest && ref.current) ref.current.scrollLeft = ref.current.scrollWidth
  }, [latest])
  return (
    <div ref={ref} className="profile-heatmap-scroll paper-focus-ring" role="region" aria-labelledby="profile-rhythm-heading" tabIndex={0}>
      {children}
    </div>
  )
}

/** Steps through the years from the profile's first year to the current one. */
function YearControl({ statistics, year }: { statistics: Found; year: number }) {
  const first = Number(statistics.profile.created_at.slice(0, 4))
  const current = Number(statistics.today.slice(0, 4))
  const previous = year > first ? Math.min(year - 1, current) : undefined
  const next = year < current ? Math.max(year + 1, first) : undefined
  const step = (target: number | undefined, label: string, icon: ReactNode) => target === undefined
    ? <Button variant="ghost" aria-label={label} disabled>{icon}</Button>
    : (
      <Link
        className={buttonClassName({ variant: 'ghost' })}
        aria-label={`${label}, ${target}`}
        to="/profile/$id/statistics/{-$year}"
        params={{ id: statistics.profile.id, year: String(target) }}
      >
        {icon}
      </Link>
    )

  return (
    <div className="profile-year-control" role="group" aria-label="Profile year">
      {step(previous, 'Previous year', <ChevronLeftIcon className="paper-icon-default" aria-hidden="true" />)}
      <strong aria-live="polite">{year}</strong>
      {step(next, 'Next year', <ChevronRightIcon className="paper-icon-default" aria-hidden="true" />)}
    </div>
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
              <a className="profile-entry-title" href={links.contestProfile(contest.id, userId)}>{contest.title}</a>
              {contest.private ? <StatusTag>Private</StatusTag> : null}
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

const scoreFormat = new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 })
const dayFormat = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeZone: 'UTC' })

/** For example `Nov 4, 2026` for `2026-11-04`. */
function formatDay(date: string) {
  return dayFormat.format(new Date(`${date}T00:00:00Z`))
}

function activeDays(days: { score: number }[]) {
  return days.filter(day => day.score > 0).length
}

/** The UTC calendar days the daily average spans: all of a past year, the current year up to today, none ahead. */
function daysCounted(year: number, today: string) {
  const current = Number(today.slice(0, 4))
  if (year > current) return 0
  const end = year === current ? Date.parse(`${today}T00:00:00Z`) : Date.UTC(year, 11, 31)
  return (end - Date.UTC(year, 0, 1)) / 86_400_000 + 1
}
