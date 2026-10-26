// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

Reflect.set(globalThis, 'React', React)

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))
vi.mock('next/router', () => ({ useRouter: () => ({ replace: vi.fn() }) }))
vi.mock('@app/common/session', () => ({ useSessionOrRedirect: vi.fn() }))
vi.mock('@app/immersion/api', () => ({
  fetchTagSuggestions: vi.fn(async () => []),
  useCreateLogV2: () => ({ mutate: vi.fn() }),
}))
vi.mock('@app/immersion/contest-api', () => ({
  useOngoingContestRegistrations: () => ({ data: undefined }),
}))
vi.mock('@app/immersion/scoring-api', () => ({
  useScorePreview: () => ({ data: undefined }),
}))

import { LogFormV2 } from './Form'

const reading = { id: 1, name: 'Reading', input_type: 'amount_primary' as const }
const writing = { id: 3, name: 'Writing', input_type: 'amount_primary' as const }
const unit = (
  id: string,
  unit_key: string,
  log_activity_id: number,
  name: string,
  language_code?: string,
) => ({ id, unit_key, log_activity_id, name, modifier: 1, language_code })
const readingPage = unit('a0000000-0000-4000-8000-000000000001', 'reading_page', 1, 'Page')
const readingCharacter = unit('a0000000-0000-4000-8000-000000000002', 'reading_character', 1, 'Character')
const readingCharacterJpn = unit('a0000000-0000-4000-8000-000000000003', 'reading_character', 1, 'Character', 'jpn')
const readingCharacterKor = unit('a0000000-0000-4000-8000-000000000004', 'reading_character', 1, 'Character', 'kor')
const writingPage = unit('a0000000-0000-4000-8000-000000000005', 'writing_page', 3, 'Page')

const options = {
  languages: [
    { code: 'jpn', name: 'Japanese' },
    { code: 'kor', name: 'Korean' },
    { code: 'eng', name: 'English' },
  ],
  activities: [reading, writing],
  units: [readingPage, readingCharacter, readingCharacterJpn, readingCharacterKor, writingPage],
  tags: [],
  user_language_codes: [],
  scoring_engine_enabled: true,
}

const field = (name: string) =>
  document.querySelector<HTMLSelectElement>(`select[name="${name}"]`)!

afterEach(cleanup)

describe('new log unit selection', () => {
  it('keeps the selected unit when the language changes', () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)
    const language = field('languageCode')
    const unitSelect = field('amountUnit')

    fireEvent.change(unitSelect, { target: { value: readingCharacterJpn.id } })
    fireEvent.change(language, { target: { value: 'kor' } })
    expect(unitSelect.value).toBe(readingCharacterKor.id)

    fireEvent.change(language, { target: { value: 'eng' } })
    expect(unitSelect.value).toBe(readingCharacter.id)
  })

  it('selects the first unit when the activity has no matching unit', () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)
    const unitSelect = field('amountUnit')

    fireEvent.change(unitSelect, { target: { value: readingCharacterJpn.id } })
    fireEvent.change(field('activityId'), { target: { value: String(writing.id) } })
    expect(unitSelect.value).toBe(writingPage.id)
  })
})
