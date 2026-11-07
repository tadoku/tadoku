import * as stylex from '@stylexjs/stylex'
import type { ImmersionUserActivity } from 'api-client'
import { HeatmapChart, a11y } from 'paper-ui'
import { size, space } from 'paper-ui/tokens.stylex'
import { useEffect, useRef, type ReactNode } from 'react'
import { pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'

const styles = stylex.create({
  scroll: { maxInlineSize: '100%', paddingBlockEnd: space[12], paddingBlockStart: space[24], overflowX: 'auto', },
  frame: { minInlineSize: size[704] },
  /** The legend lives inside the sideways scroller; cap it to the visible width so it wraps on screen. */
  legend: { maxInlineSize: `min(100%, calc(100vw - ${space[80]}))` },
})

const dayFormat = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeZone: 'UTC' })

/** For example `Nov 4, 2026` for `2026-11-04`. */
const formatDay = (date: string) => dayFormat.format(new Date(`${date}T00:00:00Z`))

/**
 * A learner's year of daily scores, each day linking to its updates. The region is named by the heading `labelledBy`;
 * on narrow screens the year scrolls sideways and the current year opens at its latest weeks.
 */
export function ActivityCalendar({ userId, year, today, activity, labelledBy, surface = 'paper' }: {
  userId: string
  year: number
  /** The server's UTC date, so the calendar marks the same today on the server and after hydration. */
  today: string
  activity: ImmersionUserActivity
  labelledBy: string
  /** The surface the calendar sits on, which its pinned weekday labels match. */
  surface?: 'paper' | 'canvas'
}) {
  const links = useLinks()
  return (
    <HeatmapScroll latest={year === Number(today.slice(0, 4))} labelledBy={labelledBy}>
      <HeatmapChart
        today={today}
        surface={surface}
        frameXstyle={styles.frame}
        legendXstyle={styles.legend}
        id={`activity-${userId}-${year}`}
        year={year}
        getHref={date => links.profileUpdates(userId, date)}
        data={activity.scores.map(day => ({
          date: day.date,
          value: day.score,
          tooltip: `${scoreFormat.format(day.score)} ${pointsUnit(day.score)} on ${formatDay(day.date)}`,
        }))}
      />
    </HeatmapScroll>
  )
}

function HeatmapScroll({ latest, labelledBy, children }: { latest: boolean; labelledBy: string; children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const scroller = ref.current
    if (!scroller) return
    // While the year scrolls, month labels under the pinned weekday gutter or past the visible edge would show as
    // clipped fragments, so they hide there. Glyph side bearings sit a few pixels outside a label's anchor.
    const hideCoveredMonths = () => {
      const scrolls = scroller.scrollWidth - scroller.clientWidth > 2
      const start = (scroller.querySelector('[data-heatmap="weekdays"]')?.getBoundingClientRect().right ?? 0) - 4
      const end = scroller.getBoundingClientRect().right + 1
      for (const label of scroller.querySelectorAll<SVGTextElement>('svg[data-heatmap="year"] text')) {
        const { left, right } = label.getBoundingClientRect()
        label.style.visibility = scrolls && (left < start || right > end) ? 'hidden' : ''
      }
    }
    // The current year opens with today near the right edge and about a week after it in view.
    const todayCell = latest && scroller.scrollWidth - scroller.clientWidth > 2 ? scroller.querySelector('[data-today]') : null
    if (todayCell) scroller.scrollLeft += todayCell.getBoundingClientRect().right - (scroller.getBoundingClientRect().right - 48)
    hideCoveredMonths()
    scroller.addEventListener('scroll', hideCoveredMonths, { passive: true })
    return () => scroller.removeEventListener('scroll', hideCoveredMonths)
  }, [latest])
  return (
    <div ref={ref} {...stylex.props(styles.scroll, a11y.focusRing)} role="region" aria-labelledby={labelledBy} tabIndex={0}>
      {children}
    </div>
  )
}
