import { DateTime, Interval } from 'luxon'
import { HeatmapChart } from 'paper-ui'

const heatmapYear = 2023

const heatmapData = Interval.fromDateTimes(
  DateTime.fromObject({ year: heatmapYear, month: 1, day: 1 }),
  DateTime.fromObject({ year: heatmapYear, month: 12, day: 31 }).endOf('day'),
)
  .splitBy({ day: 1 })
  .flatMap((interval, index) => {
    const date = interval.start
    if (date === null) return []
    const value = index % 5 === 0 ? 0 : ((index * 37) % 100) + 1
    return [
      {
        date: date.toISODate() ?? '',
        value,
        tooltip: `${value} points on ${date.toLocaleString(
          DateTime.DATE_FULL,
        )}`,
      },
    ]
  })

export default function Example() {
  return (
    <HeatmapChart id="reading-activity" year={heatmapYear} data={heatmapData} />
  )
}
