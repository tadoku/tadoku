import { format, utcToZonedTime } from 'date-fns-tz'
import { addDays } from 'date-fns'

// Will format the date correctly in utc
export function prettyDateInUTC(date: Date): string {
  return format(utcToZonedTime(date, 'utc'), 'uuuu-MM-dd', { timeZone: 'UTC' })
}

export function prettyDateIgnoreTimezone(date: Date): string {
  return format(date, 'uuuu-MM-dd')
}

export function formatUTC(date: Date, pattern: string): string {
  return format(utcToZonedTime(date, 'utc'), pattern, { timeZone: 'UTC' })
}

export function getDates(startDate: Date, endDate: Date) {
  const dates = []

  let currentDate = utcToZonedTime(startDate, 'UTC')
  const deadline = utcToZonedTime(endDate, 'UTC')

  while (currentDate < deadline) {
    dates.push(currentDate)
    currentDate = addDays(currentDate, 1)
  }

  return dates
}
