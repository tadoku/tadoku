import { afterEach, describe, expect, it, vi } from 'vitest'
import { formatDate } from '../src/data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../src/dates'

const DateTimeFormat = Intl.DateTimeFormat

function viewer(locale: string, timeZone: string) {
  vi.spyOn(Intl, 'DateTimeFormat').mockImplementation(function (locales, options) {
    return new DateTimeFormat(locales ?? locale, { timeZone, ...options })
  })
}

afterEach(() => vi.restoreAllMocks())

describe('dates in the viewer’s locale and time zone', () => {
  it('keeps a logged calendar date intact west of UTC and uses the viewer’s date order', () => {
    viewer('en-US', 'America/Los_Angeles')
    expect(formatDate('2026-09-01')).toBe('Sep 1, 2026')
  })

  it('shows an event instant on the viewer’s local date when it crosses midnight', () => {
    viewer('en-US', 'America/Los_Angeles')
    expect(formatDate('2026-09-01T00:00:00Z')).toBe('Aug 31, 2026')
  })

  it('collapses repeated month and year in a calendar range', () => {
    viewer('en-US', 'America/Los_Angeles')
    expect(formatDateRange('2026-09-01', '2026-09-30')).toBe('Sep 1\u2009–\u200930, 2026')
  })

  it('localizes both contest boundaries, including a date rollover east of UTC', () => {
    viewer('en-US', 'Asia/Tokyo')
    expect(formatDateTime(contestStart('2026-09-01'))).toBe('Sep 1, 2026, 9:00 AM GMT+9')
    expect(formatDateTime(contestEnd('2026-09-23'))).toBe('Sep 24, 2026, 8:59 AM GMT+9')
    expect(formatDateRange(contestStart('2026-09-01'), contestEnd('2026-09-30'))).toBe('Sep 1\u2009–\u2009Oct 1, 2026')
  })

  it('uses the viewer’s daylight saving offset for each event', () => {
    viewer('en-US', 'America/Los_Angeles')
    expect(formatDateTime(contestEnd('2026-09-23'))).toBe('Sep 23, 2026, 4:59 PM PDT')
    expect(formatDateTime(contestEnd('2026-12-23'))).toBe('Dec 23, 2026, 3:59 PM PST')
  })

  it('follows the viewer’s language and clock convention', () => {
    viewer('de-DE', 'Europe/Berlin')
    expect(formatDate('2026-09-01')).toBe('1. Sept. 2026')
    expect(formatDateTime(contestStart('2026-09-01'))).toBe('1. Sept. 2026, 02:00 MESZ')
  })
})
