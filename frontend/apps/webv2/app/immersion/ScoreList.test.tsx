// @vitest-environment jsdom

import React from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import { Settings } from 'luxon'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Language, Score } from '@app/immersion/api'
import { ScoreList } from './ScoreList'

Reflect.set(globalThis, 'React', React)

const defaultZone = Settings.defaultZone
const originalTZ = process.env.TZ
const languages: Language[] = [
  { code: 'jpn', name: 'Japanese' },
  { code: 'fra', name: 'French' },
]

const renderScores = (total: number | undefined, year?: number) => {
  const list: Score[] =
    total === undefined
      ? []
      : [
          { language_code: 'jpn', score: total / 2 },
          { language_code: 'fra', score: total / 2 },
        ]

  render(<ScoreList {...{ languages, list, year }} />)
  return screen.getByRole('heading', { name: 'Overall score' }).parentElement!
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-10-18T12:00:00Z'))
})

afterEach(() => {
  cleanup()
  vi.useRealTimers()
  Settings.defaultZone = defaultZone
  if (originalTZ === undefined) delete process.env.TZ
  else process.env.TZ = originalTZ
})

describe('ScoreList daily average', () => {
  it.each([
    {
      name: 'includes idle days and today in the current year',
      year: 2026,
      total: 2910,
      average: '10',
      period: 'the 291 days of 2026 so far',
    },
    {
      name: 'uses the whole past year',
      year: 2025,
      total: 730,
      average: '2',
      period: 'all 365 days of 2025',
    },
    {
      name: 'uses the whole past leap year',
      year: 2024,
      total: 366,
      average: '1',
      period: 'all 366 days of 2024',
    },
    {
      name: 'counts February 29 in the current leap year',
      clock: '2024-03-01T00:00:00Z',
      year: 2024,
      total: 61,
      average: '1',
      period: 'the 61 days of 2024 so far',
    },
    {
      name: 'uses singular day on January 1',
      clock: '2026-01-01T00:00:00Z',
      year: 2026,
      total: 5,
      average: '5',
      period: 'the 1 day of 2026 so far',
    },
    {
      name: 'uses the UTC year when the local year is already next year',
      clock: '2025-12-31T12:00:00Z',
      zone: 'Pacific/Kiritimati',
      year: 2025,
      total: 365,
      average: '1',
      period: 'the 365 days of 2025 so far',
    },
    {
      name: 'rounds the average to one decimal place',
      year: 2026,
      total: 1000,
      average: '3.4',
      period: 'the 291 days of 2026 so far',
    },
    {
      name: 'formats thousands in the average',
      year: 2025,
      total: 1000000,
      average: '2,739.7',
      period: 'all 365 days of 2025',
    },
  ])('$name', ({ year, total, average, period, clock, zone }) => {
    if (clock) vi.setSystemTime(new Date(clock))
    if (zone) {
      Settings.defaultZone = zone
      process.env.TZ = zone
    }

    renderScores(total, year)

    expect(screen.getByText(/per day/).textContent).toBe(
      `${average} per day · average over ${period}`,
    )
  })

  it.each([
    { name: 'future years', year: 2027, total: 100, expectedTotal: '100' },
    { name: 'an empty list', year: 2025, expectedTotal: 'N/A' },
    { name: 'zero scores', year: 2025, total: 0, expectedTotal: 'N/A' },
    {
      name: 'contest profiles without a year',
      total: 2910,
      expectedTotal: '2,910',
    },
    {
      name: 'a future UTC year that is already current locally',
      clock: '2025-12-31T12:00:00Z',
      zone: 'Pacific/Kiritimati',
      year: 2026,
      total: 100,
      expectedTotal: '100',
    },
  ])(
    'hides the average for $name',
    ({ year, total, expectedTotal, clock, zone }) => {
      if (clock) vi.setSystemTime(new Date(clock))
      if (zone) {
        Settings.defaultZone = zone
        process.env.TZ = zone
      }

      const card = renderScores(total, year)

      expect(card.textContent).toContain(expectedTotal)
      expect(card.textContent).not.toContain('per day')
    },
  )
})
