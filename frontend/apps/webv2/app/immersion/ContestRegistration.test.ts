import { describe, expect, it, vi } from 'vitest'

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))

import { ContestRegistrationFormSchema } from './ContestRegistration'

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
