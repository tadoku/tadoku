import type { ImmersionContest } from 'api-client'

// Contest dates are UTC calendar days. Formatting them by hand keeps server (Node ICU) and browser markup identical.
export const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sept', 'Oct', 'Nov', 'Dec']

export const numberFormat = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 })
export const dayAndMonth = (date: string) => `${Number(date.slice(8, 10))} ${months[Number(date.slice(5, 7)) - 1]}`
export const formatDate = (date: string) => `${dayAndMonth(date)} ${date.slice(0, 4)}`
export const formatPeriod = ({ contest_start: start, contest_end: end }: Pick<ImmersionContest, 'contest_start' | 'contest_end'>) =>
  `${start.slice(0, 4) === end.slice(0, 4) ? dayAndMonth(start) : formatDate(start)} – ${formatDate(end)}`

/** The unit word for a score: "1 point", "2 points", "0.4 points". */
export const pointsUnit = (score: number) => (score === 1 ? 'point' : 'points')
