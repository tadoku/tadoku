import { SectionHeading } from 'paper-ui'
import { countLabel, pointsUnit, scoreFormat } from '../lib/format'
import type { Read } from '../lib/leaderboard'
import { muted } from './standings-styles'

type Row = { label: string; score: number; fill?: string }
type TagRow = { label: string; score: number; logs: number }

const shownTags = 6

/**
 * A score split such as Score by language: one row per item, highest first, with its points, its share of the total and
 * a bar. `fill` colours the row's bar, for example by activity; it defaults to a neutral.
 */
export function Breakdown<T>({ id, title, section, items, empty }: {
  id: string
  title: string
  section: Read<T>
  items: (data: T) => Row[]
  empty: string
}) {
  const rows = section.ok ? items(section.data).sort((left, right) => right.score - left.score) : []
  const total = rows.reduce((sum, row) => sum + row.score, 0)
  const share = (score: number) => (total ? score / total * 100 : 0)

  return (
    <BreakdownSection id={id} title={title} failed={!section.ok} empty={rows.length ? undefined : empty}>
      <ol className={list}>
        {rows.map(row => (
          <BreakdownRow key={row.label} label={row.label} detail={`${pointsUnit(row.score)} · ${Math.round(share(row.score))}%`} score={row.score} width={share(row.score)} fill={row.fill} />
        ))}
      </ol>
    </BreakdownSection>
  )
}

/**
 * Score by tag: the top six tags with their points and logs, then a line naming the rest. A log counts toward each of
 * its tags, so tags overlap: rows show no share, and bars are measured against `total`, which must be at least the largest tag's score.
 */
export function TagBreakdown<T>({ id, title, section, items, total, empty }: {
  id: string
  title: string
  section: Read<T>
  items: (data: T) => TagRow[]
  total: (data: T) => number
  empty: string
}) {
  const rows = section.ok ? items(section.data).sort((left, right) => right.score - left.score) : []
  const base = section.ok ? total(section.data) : 0
  const rest = rows.slice(shownTags)

  return (
    <BreakdownSection id={id} title={title} failed={!section.ok} empty={rows.length ? undefined : empty}>
      <ol className={list}>
        {rows.slice(0, shownTags).map(row => (
          <BreakdownRow key={row.label} label={row.label} detail={`${pointsUnit(row.score)} · ${countLabel(row.logs, 'log')}`} score={row.score} width={base ? row.score / base * 100 : 0} />
        ))}
      </ol>
      {rest.length ? (
        <p className={muted}>Also {rest.map(row => `${row.label} ${scoreFormat.format(row.score)}`).join(' · ')}</p>
      ) : null}
    </BreakdownSection>
  )
}

const list = 'm-0 list-none p-0 [&>li+li]:mt-5'

function BreakdownSection({ id, title, failed, empty, children }: {
  id: string
  title: string
  failed: boolean
  empty: string | undefined
  children: React.ReactNode
}) {
  return (
    <section aria-labelledby={id}>
      <SectionHeading level={2} id={id} title={title} />
      {failed ? <p className={muted}>Could not retrieve this breakdown.</p> : empty ? <p className={muted}>{empty}</p> : children}
    </section>
  )
}

function BreakdownRow({ label, detail, score, width, fill }: { label: string; detail: string; score: number; width: number; fill?: string }) {
  return (
    <li>
      <div className="mb-2 flex justify-between gap-4 text-ui wrap-anywhere">
        <span>{label}</span>
        <strong className="whitespace-nowrap tabular-nums">
          {scoreFormat.format(score)} <small className="font-regular text-muted">{detail}</small>
        </strong>
      </div>
      <span className="block h-2 bg-rule-subtle" aria-hidden="true">
        <span className={`block h-full min-w-0.5 ${fill ?? 'bg-muted'}`} style={{ width: `${width}%` }} />
      </span>
    </li>
  )
}
