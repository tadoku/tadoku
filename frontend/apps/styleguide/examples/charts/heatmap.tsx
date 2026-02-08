import { HeatmapChart } from 'ui'
import { DateTime, Interval } from 'luxon'

export default function HeatmapExample() {
  const year = 2023
  const start = DateTime.fromObject({ year, month: 1, day: 1 })
  const end = DateTime.fromObject({ year, month: 12, day: 31 })

  const data = Interval.fromDateTimes(start, end.endOf('day'))
    .splitBy({ day: 1 })
    .map(it => it.start)
    .map(date => {
      const value = Math.random() < 0.3 ? 0 : Math.random() * 100
      return {
        date: date?.toISODate() ?? 'unknown',
        value,
        tooltip: `${Math.ceil(value)} points on ${
          date?.toLocaleString(DateTime.DATE_FULL) ?? 'unknown'
        }`,
      }
    })

  return <HeatmapChart year={year} data={data} id="demo" />
}
