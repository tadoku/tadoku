import { DateTime, Interval } from 'luxon'

function isoDayOfWeek(dt: Date) {
  let wd = dt.getDay() // 0..6, from sunday
  wd = ((wd + 6) % 7) + 1 // 1..7 from monday
  return '' + wd // string so it gets parsed
}

function generateData() {
  const data = []
  const end = new Date(2023, 1, 1, 0, 0, 0, 0)
  let dt = new Date(2022, 1, 1, 0, 0, 0, 0)
  while (dt <= end) {
    const iso = dt.toISOString().substr(0, 10)
    data.push({
      x: iso,
      y: isoDayOfWeek(dt),
      d: iso,
      v: Math.random() * 50,
    })
    dt = new Date(dt.setDate(dt.getDate() + 1))
  }
  return data
}

const data = generateData()

interface Cell {
  x: DateTime
  y: number
  value: number
}

function HeatmapChart() {
  const year = 2022
  const start = DateTime.fromObject({ year, month: 1, day: 1 })
  const end = DateTime.fromObject({ year, month: 12, day: 31 })

  const dates = Interval.fromDateTimes(start, end.endOf('day'))
    .splitBy({ day: 1 })
    .map(it => it.start)

  const weeks = start.weeksInWeekYear + 1
  const cols: (Cell | undefined)[][] = Array.from(Array(weeks)).map(_ =>
    Array.from(Array(7)),
  )

  let col = 0
  for (const date of dates) {
    cols[col][date.weekday - 1] = {
      x: date,
      y: date.weekday,
      value: Math.random() < 0.3 ? 0 : Math.random() * 100,
    }

    if (date.weekday === 7) {
      col += 1
    }
  }

  const maxValue = Math.max(...cols.flatMap(it => it.map(it => it?.value ?? 0)))

  const colWidth = 10
  const rowHeight = 10
  const padding = 3
  const offset = {
    x: 30,
    y: 15,
  }

  const weekdays = ['Mon', undefined, 'Wed', undefined, 'Fri', undefined, 'Sun']

  let lastMonth = 0
  const months = cols.map(rows => {
    const shouldRender = rows
      .filter(it => it !== undefined)
      .some(cell => cell!.x.month > lastMonth)
    if (!shouldRender) {
      return undefined
    }

    lastMonth += 1
    return DateTime.fromObject({ month: lastMonth }).toFormat('LLL')
  })

  return (
    <svg
      width={weeks * colWidth + (weeks - 1) * padding + offset.x}
      height={7 * rowHeight + 6 * padding + offset.y}
      className=""
    >
      {weekdays.map((label, row) => {
        if (!label) {
          return null
        }

        return (
          <text
            textAnchor="end"
            x={offset.x - padding * 2}
            y={offset.y + rowHeight * row + padding * row}
            className=""
            alignmentBaseline="hanging"
            style={{ fontSize: 10 }}
          >
            {label}
          </text>
        )
      })}
      {months.map((label, col) => {
        if (!label) {
          return null
        }

        return (
          <text
            textAnchor="start"
            x={offset.x + colWidth * col + padding * col}
            y={0}
            className=""
            alignmentBaseline="hanging"
            style={{ fontSize: 10 }}
          >
            {label}
          </text>
        )
      })}
      {cols.map((rows, col) =>
        rows.map((cell, row) => {
          if (!cell) {
            return null
          }
          return (
            <rect
              width={colWidth}
              height={rowHeight}
              x={offset.x + colWidth * col + padding * col}
              y={offset.y + rowHeight * row + padding * row}
              fill={'transparent'}
              className={`${getCellDepthClass(maxValue, cell.value)}`}
              strokeWidth={0}
            >
              <title>{cell.x.toLocaleString(DateTime.DATE_FULL)}</title>
            </rect>
          )
        }),
      )}
    </svg>
  )
}

function getCellDepthClass(max: number, value: number) {
  if (value === 0) {
    return 'fill-stone-200'
  }

  const ratio = value / max

  if (ratio < 0.25) {
    return 'fill-teal-100'
  }
  if (ratio < 0.5) {
    return 'fill-teal-300'
  }
  if (ratio < 0.75) {
    return 'fill-teal-500'
  }

  return 'fill-teal-700'
}

export default HeatmapChart
