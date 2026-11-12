import { useQuery } from '@tanstack/react-query'
import { Link, useNavigate } from '@tanstack/react-router'
import { EmptyState, PageHeader, SectionHeading, StackedBarChart, StatGrid, buttonClassName, type FilterMenuHandle } from 'paper-ui'
import { useRef } from 'react'
import { activityColor, activityTone } from '../lib/activity'
import { contestSummaryQuery, type ContestProfile as Profile, type ContestSummary } from '../lib/contest-activity'
import { filterCount, filterSections, noFilters, toLogFilters, type ProfileFilters } from '../lib/contest-filters'
import { countLabel, dayAndMonth, formatDate, numberFormat, place, pointsUnit, scoreFormat } from '../lib/format'
import { contestEntryQuery, type Read } from '../lib/leaderboard'
import { useLinks } from '../lib/links'
import { stringifySearch } from '../lib/search'
import { Breakdown, TagBreakdown } from './Breakdown'
import { ContestLogs } from './ContestUpdates'
import { ProfileAppliedFilters, ProfileFilterMenu } from './ContestProfileFilters'
import { renderLink } from './router-link'
import { muted, workspace } from './standings-styles'

type Found = Extract<Profile, { status: 'found' }>
type Summary = ContestSummary

const failed = { ok: false } as const

/**
 * One participant in one contest: a summary of their score, a chart of their points per day by activity, their logs
 * and their score by language, activity and tag. Filters by language, activity and tag narrow every figure.
 */
export function ContestProfile({ profile, page, filters }: { profile: Found; page: number; filters: ProfileFilters }) {
  const links = useLinks()
  const navigate = useNavigate()
  const menu = useRef<FilterMenuHandle>(null)
  const { registration, scores } = profile
  const { contest, user_id: userId, user_display_name: name, languages } = registration
  const total = scores.reduce((sum, score) => sum + score.score, 0)
  const entry = useQuery(contestEntryQuery(contest.id, userId)).data
  const rank = entry?.ok ? entry.data?.rank : undefined
  const filtered = filterCount(filters) > 0
  const input = { contestId: contest.id, userId }
  const summary = useQuery(contestSummaryQuery({ ...input, filters: toLogFilters(filters) }))
  const whole = useQuery({ ...contestSummaryQuery({ ...input, filters: toLogFilters(noFilters) }), enabled: filtered })
  const data = summary.data?.ok ? summary.data.data : undefined
  const unfiltered = filtered ? whole.data : summary.data
  const href = (next: ProfileFilters, target = 1) => `${links.contestProfile(contest.id, userId, target)}${stringifySearch(next)}`
  const change = (next: Partial<ProfileFilters>) => void navigate({
    to: '/contests/$id/profile/$userId/$',
    params: { id: contest.id, userId, _splat: '1' },
    search: { ...filters, ...next },
    resetScroll: false,
  })
  const resultLabel = countLabel(data?.log_count ?? 0, 'log')
  const empty = filtered && data?.log_count === 0
  const languageName = (code: string) => languages.find(language => language.code === code)?.name ?? code
  const activityName = (slug: string) =>
    data?.activities.find(activity => activityTone(activity.id).slug === slug)?.name ?? slug.charAt(0).toUpperCase() + slug.slice(1)

  return (
    <div>
      <PageHeader
        backLink={{ href: links.contest(contest.id), label: contest.title }}
        renderLink={renderLink}
        title={name}
        lead={`${rank ? `${place(rank)} · ` : ''}${scoreFormat.format(total)} ${pointsUnit(total)} in ${contest.title}`}
        actions={<Link className={buttonClassName({ variant: 'outline' })} to={links.profile(userId)}>View profile</Link>}
      />
      <section className="mt-8" aria-labelledby="contest-profile-activity">
        <SectionHeading
          level={2}
          id="contest-profile-activity"
          title="Activity"
          actions={(
            <ProfileFilterMenu
              menu={menu}
              sections={filterSections(data, filters, languages)}
              action={links.contestProfile(contest.id, userId)}
              resultLabel={resultLabel}
              onChange={change}
            />
          )}
        />
        <ProfileAppliedFilters
          filters={filters}
          labels={{ language: languageName, activity: activityName }}
          resultLabel={`${resultLabel} of ${unfiltered?.ok ? numberFormat.format(unfiltered.data.log_count) : '…'}`}
          onChange={change}
          onEmptied={() => menu.current?.focus()}
        />
        {!summary.data ? <p className={muted} role="status">Loading activity…</p> : !data ? (
          <p className={muted}>Could not load this summary. Try reloading this page.</p>
        ) : empty ? (
          <EmptyState
            variant="plain"
            headingLevel={3}
            title="No logs match these filters"
            description="Remove a filter to see more of this participant’s logs."
            actions={<Link className={buttonClassName({ variant: 'outline' })} to={href(noFilters)} resetScroll={false}>Clear filters</Link>}
          />
        ) : (
          <>
            <Figures summary={data} filtered={filtered} total={total} rank={rank} />
            <ActivityChart summary={data} start={contest.contest_start} end={contest.contest_end} />
          </>
        )}
      </section>
      {empty ? null : (
        <div className={`${workspace} mt-8`}>
          <details className="tablet:hidden">
            <summary className="cursor-pointer font-semibold text-link">
              {languages.length > 1 ? 'Score by language, activity and tag' : 'Score by activity and tag'}
              {filtered ? ` · ${filterCount(filters)} applied` : ''}
            </summary>
            <ScoreLists summary={summary.data ?? failed} languages={languages.length > 1} idPrefix="phone" />
          </details>
          <section className="tablet:col-start-1 tablet:row-start-1" aria-labelledby="contest-profile-logs">
            <SectionHeading level={2} id="contest-profile-logs" title="Logs" />
            <ContestLogs
              contest={contest}
              userId={userId}
              page={page}
              filters={toLogFilters(filters)}
              pageHref={next => href(filters, next)}
            />
          </section>
          <aside className="max-tablet:hidden tablet:col-start-2 tablet:row-start-1">
            <ScoreLists summary={summary.data ?? failed} languages={languages.length > 1} idPrefix="aside" />
          </aside>
        </div>
      )}
    </div>
  )
}

/** Score, active days, logs and the best day. Filtered, the score reads as a share of the whole contest score. */
function Figures({ summary, filtered, total, rank }: { summary: Summary; filtered: boolean; total: number; rank?: number }) {
  const { score, active_days: days, best_day: best } = summary
  const share = total ? score / total * 100 : 0
  return (
    <StatGrid
      variant="sheet"
      className="mb-6"
      stats={[
        {
          label: 'Score',
          value: scoreFormat.format(score),
          detail: filtered
            ? `${share > 0 && share < 1 ? 'under 1' : Math.round(share)}% of ${scoreFormat.format(total)}`
            : rank ? `${place(rank)} place` : undefined,
        },
        { label: 'Active days', value: numberFormat.format(days), detail: days ? `${scoreFormat.format(score / days)} per active day` : undefined },
        { label: 'Logs', value: numberFormat.format(summary.log_count) },
        { label: 'Best day', value: best ? scoreFormat.format(best.score) : '–', detail: best ? dayAndMonth(best.date) : undefined },
      ]}
    />
  )
}

/** Points per day from the contest's start to the earlier of today and its end, stacked by activity. */
function ActivityChart({ summary, start, end }: { summary: Summary; start: string; end: string }) {
  const days = new Map<string, Record<string, number>>()
  for (const day of summary.days) days.set(day.date, { ...days.get(day.date), [day.activity_id]: day.score })
  const series = [...summary.activities]
    .sort((left, right) => left.id - right.id)
    .map(activity => ({ id: String(activity.id), label: activity.name, color: activityColor(activity.id) }))
  const best = summary.best_day

  return (
    <StackedBarChart
      start={start}
      end={summary.today < end ? summary.today : end}
      days={[...days].map(([date, values]) => ({ date, values }))}
      series={series}
      runningTotal
      label="Points per day by activity, with the running total"
      description={[
        `${scoreFormat.format(summary.score)} ${pointsUnit(summary.score)} over ${countLabel(summary.active_days, 'active day')}.`,
        best ? `Best day ${formatDate(best.date)}, ${scoreFormat.format(best.score)} ${pointsUnit(best.score)}.` : '',
      ].join(' ').trim()}
    />
  )
}

/** Score by language (with more than one), activity and tag, each counted under every other filter. */
function ScoreLists({ summary: section, languages, idPrefix }: { summary: Read<Summary>; languages: boolean; idPrefix: string }) {
  return (
    <div className="pt-1 [&>section+section]:mt-7 [&>section+section]:border-t [&>section+section]:border-rule [&>section+section]:pt-6">
      {languages ? (
        <Breakdown
          id={`${idPrefix}-languages`}
          title="Score by language"
          section={section}
          items={data => data.languages.map(item => ({ label: item.name, score: item.score }))}
          empty="No scores recorded yet."
        />
      ) : null}
      <Breakdown
        id={`${idPrefix}-activities`}
        title="Score by activity"
        section={section}
        items={data => data.activities.map(item => ({ label: item.name, score: item.score, fill: activityTone(item.id).fill }))}
        empty="No scores recorded yet."
      />
      <TagBreakdown
        id={`${idPrefix}-tags`}
        title="Score by tag"
        section={section}
        items={data => data.tags.map(item => ({ label: item.tag, score: item.score, logs: item.log_count }))}
        // Tags ignore their own filter, so a tag can outscore the filtered total; bars never pass the largest.
        total={data => Math.max(data.score, ...data.tags.map(item => item.score))}
        empty="No tagged logs."
      />
    </div>
  )
}
