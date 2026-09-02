import { describe, expect, it, vi } from 'vitest'

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))

import { filterUnits, NewLogV2APISchema } from './domain'

const amountActivity = {
  id: 101,
  name: 'Amount activity',
  input_type: 'amount_primary' as const,
}
const timeActivity = {
  id: 202,
  name: 'Time activity',
  input_type: 'time_primary' as const,
}
const amountUnit = {
  id: 'amount-unit',
  unit_key: 'test_amount_unit',
  log_activity_id: amountActivity.id,
  name: 'Amount unit',
  modifier: 1,
}
const timeUnit = {
  id: 'time-unit',
  unit_key: 'test_time_unit',
  log_activity_id: timeActivity.id,
  name: 'Time unit',
  modifier: 0.4,
}

const validFormValues = {
  languageCode: 'jpn',
  allUnits: [amountUnit, timeUnit],
  allActivities: [amountActivity, timeActivity],
  tags: [] as string[],
  description: '',
}

describe('NewLogV2APISchema', () => {
  it('ignores a stale zero amount when previewing a duration-only listening log', () => {
    const result = NewLogV2APISchema.parse({
      ...validFormValues,
      activityId: timeActivity.id,
      amountValue: 0,
      amountUnit: '00000000-0000-0000-0000-000000000000',
      durationMinutes: 15,
    })

    expect(result).toEqual({
      language_code: 'jpn',
      activity_id: timeActivity.id,
      duration_seconds: 900,
      tags: [],
      description: '',
    })
  })

  it('still rejects a zero amount for an amount-primary activity', () => {
    const result = NewLogV2APISchema.safeParse({
      ...validFormValues,
      activityId: amountActivity.id,
      amountValue: 0,
      amountUnit: amountUnit.id,
    })

    expect(result.success).toBe(false)
  })

  it('still supports a legacy amount-only listening log', () => {
    const result = NewLogV2APISchema.parse({
      ...validFormValues,
      activityId: timeActivity.id,
      amountValue: 15,
      amountUnit: timeUnit.id,
    })

    expect(result).toEqual({
      language_code: 'jpn',
      activity_id: timeActivity.id,
      amount: 15,
      unit_id: timeUnit.id,
      tags: [],
      description: '',
    })
  })

  it('rejects a zero amount for a legacy amount-only listening log', () => {
    const result = NewLogV2APISchema.safeParse({
      ...validFormValues,
      activityId: timeActivity.id,
      amountValue: 0,
      amountUnit: timeUnit.id,
    })

    expect(result.success).toBe(false)
  })
})

describe('filterUnits', () => {
  it('keeps legacy reading page variants out of the V2 amount selector', () => {
    const readingUnits = [
      {
        id: 'page',
        unit_key: 'reading_page',
        log_activity_id: 1,
        name: 'Page',
        modifier: 1,
      },
      {
        id: 'two-column-page',
        unit_key: 'reading_two_column_page',
        log_activity_id: 1,
        name: '2 Column page',
        modifier: 1.6,
        language_code: 'jpn',
      },
      {
        id: 'comic-page',
        unit_key: 'reading_comic_page',
        log_activity_id: 1,
        name: 'Comic page',
        modifier: 0.2,
      },
      {
        id: 'sentence',
        unit_key: 'reading_sentence',
        log_activity_id: 1,
        name: 'Sentence',
        modifier: 0.05,
      },
    ]

    expect(filterUnits(readingUnits, 1, 'jpn').map(unit => unit.unit_key)).toEqual([
      'reading_page',
      'reading_sentence',
    ])
  })
})
