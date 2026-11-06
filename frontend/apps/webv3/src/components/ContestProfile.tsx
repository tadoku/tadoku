import { Link } from '@tanstack/react-router'
import type { ImmersionContestProfileActivityRow } from 'api-client'
import { PageHeader, SectionHeading, StatGrid, Table, buttonClassName, type TableColumn } from 'paper-ui'
import type { ContestProfile as Profile } from '../lib/contest-activity'
import { formatDate, numberFormat } from '../lib/format'
import { useLinks } from '../lib/links'
import { ContestLogs } from './ContestUpdates'
import { renderLink } from './router-link'

type Found = Extract<Profile, { status: 'found' }>

/** One participant in one contest: their score per registered language, their daily scores and their logs. */
export function ContestProfile({ profile, page }: { profile: Found; page: number }) {
  const links = useLinks()
  const { registration, scores, daily } = profile
  const { contest, user_id: userId, user_display_name: name } = registration
  const byLanguage = new Map(scores.map(score => [score.language_code, score.score]))
  const total = scores.reduce((sum, score) => sum + score.score, 0)

  return (
    <div className="profile-page contest-profile">
      <PageHeader
        backLink={{ href: links.contest(contest.id), label: contest.title }}
        renderLink={renderLink}
        title={name}
        actions={<Link className={buttonClassName({ variant: 'outline' })} to={links.profile(userId)}>View profile</Link>}
      />
      <StatGrid
        className="profile-stats"
        stats={[
          { label: 'Overall score', value: numberFormat.format(total) },
          ...registration.languages.map(language => ({ label: language.name, value: numberFormat.format(byLanguage.get(language.code) ?? 0) })),
        ]}
      />
      <section className="profile-activity" aria-labelledby="contest-daily-heading">
        <SectionHeading level={2} id="contest-daily-heading" title="Daily scores" />
        {daily.ok
          ? <div className="standings-sheet contest-daily"><DailyScores rows={daily.data} languages={registration.languages} name={name} /></div>
          : <p className="profile-muted">Could not load the daily scores. Try reloading this page.</p>}
      </section>
      <section className="profile-activity" aria-labelledby="contest-profile-updates">
        <SectionHeading level={2} id="contest-profile-updates" title="Updates" />
        <ContestLogs contest={contest} userId={userId} page={page} pageHref={next => links.contestProfile(contest.id, userId, next)} />
      </section>
    </div>
  )
}

type Day = { date: string; scores: Map<string, number>; total: number; cumulative: number }

/** Webv2's activity chart as a table: each active UTC day's score per language, the day's total and the running total. */
function DailyScores({ rows, languages, name }: {
  rows: ImmersionContestProfileActivityRow[]
  languages: { code: string; name: string }[]
  name: string
}) {
  if (rows.length === 0) return <p className="profile-muted">No scores recorded in this contest yet.</p>

  const days = new Map<string, Day>()
  for (const row of rows) {
    const day = days.get(row.date) ?? { date: row.date, scores: new Map(), total: 0, cumulative: 0 }
    day.scores.set(row.language_code, (day.scores.get(row.language_code) ?? 0) + row.score)
    day.total += row.score
    days.set(row.date, day)
  }
  let running = 0
  const ordered = [...days.values()].sort((left, right) => left.date.localeCompare(right.date))
  for (const day of ordered) day.cumulative = running += day.total

  const columns: TableColumn<Day>[] = [
    { id: 'day', header: 'Day', rowHeader: true, cell: day => formatDate(day.date) },
    ...(languages.length > 1
      ? languages.map((language): TableColumn<Day> => ({
        id: language.code,
        header: language.name,
        align: 'end',
        cell: day => numberFormat.format(day.scores.get(language.code) ?? 0),
      }))
      : []),
    { id: 'score', header: 'Score', align: 'end', cell: day => numberFormat.format(day.total) },
    { id: 'cumulative', header: 'Running total', align: 'end', cell: day => numberFormat.format(day.cumulative) },
  ]
  return (
    <Table
      caption={`${name}’s daily scores`}
      captionVisibility="screen-reader"
      columns={columns}
      rows={ordered}
      getRowKey={day => day.date}
      minWidth="18rem"
      tableClassName="standings-table"
    />
  )
}
