import type { ImmersionContest, ImmersionLog } from 'api-client'

// Contest dates are UTC calendar days. Formatting them by hand keeps server (Node ICU) and browser markup identical.
export const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sept', 'Oct', 'Nov', 'Dec']

/** The one locale for numbers and plurals, so the server and every browser render the same text. */
export const numberLocale = 'en-US'

/** Amounts, durations and counts, to at most one decimal, for example `2.3`. */
export const numberFormat = new Intl.NumberFormat(numberLocale, { maximumFractionDigits: 1 })
const amountPlural = new Intl.PluralRules(numberLocale, { maximumFractionDigits: 1 })

/** Today's UTC date, `YYYY-MM-DD`. Read it in a loader, never while rendering, so server and browser agree. */
export const utcToday = () => new Date().toISOString().slice(0, 10)

export const dayAndMonth = (date: string) => `${Number(date.slice(8, 10))} ${months[Number(date.slice(5, 7)) - 1]}`
const dayFormat = new Intl.DateTimeFormat(numberLocale, { dateStyle: 'medium', timeZone: 'UTC' })

/** For example `Nov 4, 2026` for `2026-11-04`, as heatmap tooltips name a day. */
export const formatDay = (date: string) => dayFormat.format(new Date(`${date}T00:00:00Z`))
export const formatDate = (date: string) => `${dayAndMonth(date)} ${date.slice(0, 4)}`
export const formatPeriod = ({ contest_start: start, contest_end: end }: Pick<ImmersionContest, 'contest_start' | 'contest_end'>) =>
  `${start.slice(0, 4) === end.slice(0, 4) ? dayAndMonth(start) : formatDate(start)} – ${formatDate(end)}`

/** An amount in its unit, plural as the amount is shown: `1 page`, `2.3 pages`, `0 pages`. */
export const formatAmount = (amount: number, unitName: string) =>
  `${numberFormat.format(amount)} ${unitName.toLowerCase()}${amountPlural.select(amount) === 'one' ? '' : 's'}`

/** A duration in minutes to at most one decimal, for example `1.7 min`. */
export const formatMinutes = (seconds: number) => `${numberFormat.format(seconds / 60)} min`

/** A log's amount in its unit and its duration, as Webv2 lists them, for example `12 pages · 30 min`. */
export function formatTracking(log: Pick<ImmersionLog, 'amount' | 'unit_name' | 'duration_seconds'>) {
  const values = []
  if (log.unit_name) values.push(formatAmount(log.amount, log.unit_name))
  if (log.duration_seconds !== undefined) values.push(formatMinutes(log.duration_seconds))
  return values.join(' · ')
}

/** Webv2 hides a description tagged NSFW until the reader chooses to see it. */
export const isNsfw = (log: Pick<ImmersionLog, 'tags'>) => log.tags.some(tag => tag.toLowerCase() === 'nsfw')

/** What links and titles call a log: its description, or its language and activity when it has none or is NSFW. */
export const logName = (log: Pick<ImmersionLog, 'description' | 'tags' | 'language' | 'activity'>) =>
  log.description && !isNsfw(log) ? log.description : `${log.language.name} ${log.activity.name.toLowerCase()}`

/** Every score and point total, always to one decimal, for example `12.5` or `12.0`. */
export const scoreFormat = new Intl.NumberFormat(numberLocale, { minimumFractionDigits: 1, maximumFractionDigits: 1 })
const scorePlural = new Intl.PluralRules(numberLocale, { minimumFractionDigits: 1, maximumFractionDigits: 1 })

/** The unit word for a score as `scoreFormat` shows it: `1.0 points` takes the plural, as English does. */
export const pointsUnit = (score: number) => (scorePlural.select(score) === 'one' ? 'point' : 'points')

/** A whole count with its noun, for example `1 log` or `1,204 logs`. */
export const countLabel = (count: number, noun: string) => `${numberFormat.format(count)} ${noun}${count === 1 ? '' : 's'}`

const ordinal = new Intl.PluralRules('en', { type: 'ordinal' })
const ordinalSuffix: Record<string, string> = { one: 'st', two: 'nd', few: 'rd', other: 'th' }

/** A rank in words, for example `1st`. */
export const place = (rank: number) => `${rank}${ordinalSuffix[ordinal.select(rank)]}`
