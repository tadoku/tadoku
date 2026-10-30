import { DateTime } from 'luxon'
import { describe, expect, it, vi } from 'vitest'

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))

import {
  ContestRegistrationFormSchema,
  registrationChangesClosed,
} from './ContestRegistration'

describe('ContestRegistrationFormSchema', () => {
  it('rejects duplicate language codes', () => {
    const result = ContestRegistrationFormSchema.safeParse({
      contest_id: 'contest-id',
      new_languages: [
        { code: 'jpn', name: 'Japanese' },
        { code: 'jpn', name: 'Duplicate display name' },
      ],
    })

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            message: 'Cannot select the same language more than once',
          }),
        ]),
      )
    }
  })

  it('accepts one to three unique language codes', () => {
    for (const newLanguages of [
      [{ code: 'jpn', name: 'Japanese' }],
      [
        { code: 'jpn', name: 'Japanese' },
        { code: 'kor', name: 'Korean' },
      ],
      [
        { code: 'jpn', name: 'Japanese' },
        { code: 'kor', name: 'Korean' },
        { code: 'zho', name: 'Chinese' },
      ],
    ]) {
      const result = ContestRegistrationFormSchema.safeParse({
        contest_id: 'contest-id',
        new_languages: newLanguages,
      })

      expect(result.success).toBe(true)
    }
  })
})

describe('registrationChangesClosed', () => {
  it('keeps changes open through the last UTC second of contest_end', () => {
    const now = DateTime.fromISO('2026-09-30T23:59:59.999Z')

    expect(registrationChangesClosed('2026-09-30', now)).toBe(false)
  })

  it('closes changes from the first instant after contest_end in UTC', () => {
    const now = DateTime.fromISO('2026-10-01T00:00:00.000Z')

    expect(registrationChangesClosed('2026-09-30', now)).toBe(true)
  })

  it('uses the UTC day when the browser is ahead of UTC', () => {
    const now = DateTime.fromISO('2026-10-01T05:00:00', { zone: 'Asia/Tokyo' })

    expect(registrationChangesClosed('2026-09-30', now)).toBe(false)
  })
})
