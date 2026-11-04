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
  id: '00000000-0000-4000-8000-000000000001',
  unit_key: 'test_amount_unit',
  log_activity_id: amountActivity.id,
  name: 'Amount unit',
  modifier: 1,
}
const timeUnit = {
  id: '00000000-0000-4000-8000-000000000002',
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
      amountUnit: amountUnit.unit_key,
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

  it('sends the unit key for an amount-primary activity', () => {
    const result = NewLogV2APISchema.parse({
      ...validFormValues,
      activityId: amountActivity.id,
      amountValue: 12,
      amountUnit: amountUnit.unit_key,
    })

    expect(result).toEqual({
      language_code: 'jpn',
      activity_id: amountActivity.id,
      amount: 12,
      unit_key: amountUnit.unit_key,
      tags: [],
      description: '',
    })
  })

  it('still rejects a zero amount for an amount-primary activity', () => {
    const result = NewLogV2APISchema.safeParse({
      ...validFormValues,
      activityId: amountActivity.id,
      amountValue: 0,
      amountUnit: amountUnit.unit_key,
    })

    expect(result.success).toBe(false)
  })

  it('rejects a unit key of another activity', () => {
    const result = NewLogV2APISchema.safeParse({
      ...validFormValues,
      activityId: amountActivity.id,
      amountValue: 12,
      amountUnit: timeUnit.unit_key,
    })

    expect(result.success).toBe(false)
  })

  it('requires a duration instead of an amount for a time-primary activity', () => {
    const result = NewLogV2APISchema.safeParse({
      ...validFormValues,
      activityId: timeActivity.id,
      amountValue: 15,
      amountUnit: timeUnit.unit_key,
    })

    expect(result.success).toBe(false)
    expect(result.error?.issues.map(issue => issue.path)).toEqual([
      ['durationMinutes'],
    ])
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

  it('replaces the general unit with the language-specific unit of the same key', () => {
    const characterUnits = [
      {
        id: 'character',
        unit_key: 'reading_character',
        log_activity_id: 1,
        name: 'Character',
        modifier: 0.0025,
      },
      {
        id: 'character-jpn',
        unit_key: 'reading_character',
        log_activity_id: 1,
        name: 'Kanji character',
        modifier: 0.005,
        language_code: 'jpn',
      },
    ]

    expect(filterUnits(characterUnits, 1, 'jpn').map(unit => unit.id)).toEqual([
      'character-jpn',
    ])
    expect(filterUnits(characterUnits, 1, 'kor').map(unit => unit.id)).toEqual([
      'character',
    ])
  })
})
