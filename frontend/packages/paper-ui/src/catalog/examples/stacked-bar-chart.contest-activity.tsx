import { StackedBarChart } from 'paper-ui'

const series = [
  { id: 'reading', label: 'Reading', color: 'var(--color-activity-reading)' },
  { id: 'listening', label: 'Listening', color: 'var(--color-activity-listening)' },
  { id: 'writing', label: 'Writing', color: 'var(--color-activity-writing)' },
]

// Deterministic points for a three-month round: most days have reading, some add listening or writing.
const days = Array.from({ length: 92 }, (_, index) => {
  const date = new Date(Date.UTC(2026, 6, 1 + index)).toISOString().slice(0, 10)
  const values = {
    reading: index % 6 === 5 ? 0 : 40 + ((index * 53) % 260),
    listening: index % 3 === 0 ? 30 + ((index * 29) % 140) : 0,
    writing: index % 10 === 4 ? 25 : 0,
  }
  return { date, values }
})

export default function ContestActivityExample() {
  return (
    <StackedBarChart
      start="2026-07-01"
      end="2026-09-30"
      days={days}
      series={series}
      runningTotal
      label="Points per day by activity, with the running total"
      description="Reading, listening and writing from 1 July to 30 September 2026, logged on most days."
    />
  )
}
