import type { ImmersionContest, ImmersionLog } from 'api-client'

// Contest dates are UTC calendar days. Formatting them by hand keeps server (Node ICU) and browser markup identical.
export const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sept', 'Oct', 'Nov', 'Dec']

export const numberFormat = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 })


export const dayAndMonth = (date: string) => `${Number(date.slice(8, 10))} ${months[Number(date.slice(5, 7)) - 1]}`
export const formatDate = (date: string) => `${dayAndMonth(date)} ${date.slice(0, 4)}`
export const formatPeriod = ({ contest_start: start, contest_end: end }: Pick<ImmersionContest, 'contest_start' | 'contest_end'>) =>
  `${start.slice(0, 4) === end.slice(0, 4) ? dayAndMonth(start) : formatDate(start)} – ${formatDate(end)}`

/** The unit word for a score: "1 point", "2 points", "0.4 points". */
export const pointsUnit = (score: number) => (score === 1 ? 'point' : 'points')

/** A log's amount in its unit and its duration, as Webv2 lists them, for example `12 pages · 30 min`. */
export function formatTracking(log: Pick<ImmersionLog, 'amount' | 'unit_name' | 'duration_seconds'>) {
  const values = []
  if (log.unit_name) values.push(`${numberFormat.format(log.amount)} ${log.unit_name.toLowerCase()}${log.amount === 1 ? '' : 's'}`)
  if (log.duration_seconds !== undefined) values.push(`${numberFormat.format(log.duration_seconds / 60)} min`)
  return values.join(' · ')
}

/** Every score and point total, always to one decimal, for example `12.5` or `12.0`. */
export const scoreFormat = new Intl.NumberFormat('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 1 })
