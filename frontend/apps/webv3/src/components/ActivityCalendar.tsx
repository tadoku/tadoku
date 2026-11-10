import type { ImmersionUserActivity } from 'api-client'
import { HeatmapChart, focusRing } from 'paper-ui'
import { useEffect, useRef, type ReactNode } from 'react'
import { formatDay, pointsUnit, scoreFormat } from '../lib/format'
import { useLinks } from '../lib/links'


/**
 * A learner's year of daily scores, each day linking to its updates. The region is named by the heading `labelledBy`;
 * on narrow screens the year scrolls sideways and the current year opens at its latest weeks.
 */
export function ActivityCalendar({ userId, year, today, activity, labelledBy, surface }: {
  userId: string
  year: number
  /** The server's UTC date, so the calendar marks the same today on the server and after hydration. */
  today: string
  activity: ImmersionUserActivity
  labelledBy: string
  surface?: 'paper' | 'canvas'
}) {
  const links = useLinks()
  return (
    <>
    <p className="mb-2 mt-0 text-metadata text-muted tablet:hidden">Swipe for earlier months.</p>
    <HeatmapScroll latest={year === Number(today.slice(0, 4))} labelledBy={labelledBy}>
      <HeatmapChart
        today={today}
        surface={surface}
        scrolls
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
    </>
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
      const start = (scroller.querySelector('.paper-heatmap__weekdays')?.getBoundingClientRect().right ?? 0) - 4
      const end = scroller.getBoundingClientRect().right + 1
      for (const label of scroller.querySelectorAll<SVGTextElement>('svg.paper-heatmap text')) {
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
    <div ref={ref} className={focusRing({ className: 'max-w-full overflow-x-auto pb-3 pt-6' })} role="region" aria-labelledby={labelledBy} tabIndex={0}>
      {children}
    </div>
  )
}
