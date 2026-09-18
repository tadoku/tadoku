const calendarDate = /^\d{4}-\d{2}-\d{2}$/
const dateOptions: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short', year: 'numeric' }

// Activity dates are calendar labels. UTC here prevents them shifting a day;
// actual timestamps use the viewer’s time zone and locale.
export function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    ...dateOptions,
    ...(calendarDate.test(value) ? { timeZone: 'UTC' } : {}),
  }).format(new Date(value))
}

export function formatDateRange(start: string, end: string) {
  return new Intl.DateTimeFormat(undefined, {
    ...dateOptions,
    ...(calendarDate.test(start) && calendarDate.test(end) ? { timeZone: 'UTC' } : {}),
  }).formatRange(new Date(start), new Date(end))
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    ...dateOptions,
    hour: 'numeric', minute: '2-digit', timeZoneName: 'short',
  }).format(new Date(value))
}

// Sample contests store inclusive UTC calendar boundaries. Convert them to
// instants before displaying their schedule in a viewer’s local time zone.
export function contestStart(date: string) { return `${date}T00:00:00Z` }
export function contestEnd(date: string) { return `${date}T23:59:59.999Z` }
