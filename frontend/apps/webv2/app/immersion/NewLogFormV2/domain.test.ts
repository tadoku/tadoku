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
  unit_key: 'test_amount_unit',
  log_activity_id: amountActivity.id,
  name: 'Amount unit',
  modifier: 1,
}
const timeUnit = {
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
  it('replaces the general unit with the language-specific unit of the same key', () => {
    const characterUnits = [
      {
        unit_key: 'characters',
        log_activity_id: 1,
        name: 'Character',
        modifier: 0.0025,
      },
      {
        unit_key: 'characters',
        log_activity_id: 1,
        name: 'Kanji character',
        modifier: 0.005,
        language_code: 'jpn',
      },
    ]

    expect(filterUnits(characterUnits, 1, 'jpn').map(unit => unit.name)).toEqual([
      'Kanji character',
    ])
    expect(filterUnits(characterUnits, 1, 'kor').map(unit => unit.name)).toEqual([
      'Character',
    ])
  })
})
