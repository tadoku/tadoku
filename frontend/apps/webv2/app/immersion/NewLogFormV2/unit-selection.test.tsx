// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

const { createLog } = vi.hoisted(() => ({ createLog: vi.fn() }))

Reflect.set(globalThis, 'React', React)

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))
vi.mock('next/router', () => ({ useRouter: () => ({ replace: vi.fn() }) }))
vi.mock('@app/common/session', () => ({ useSessionOrRedirect: vi.fn() }))
vi.mock('@app/immersion/api', () => ({
  fetchTagSuggestions: vi.fn(async () => []),
  useCreateLogV2: () => ({ mutate: createLog }),
}))
vi.mock('@app/immersion/contest-api', () => ({
  useOngoingContestRegistrations: () => ({ data: undefined }),
}))
vi.mock('@app/immersion/scoring-api', () => ({
  useScorePreview: () => ({ data: undefined }),
}))

import { LogFormV2 } from './Form'

const reading = { id: 1, name: 'Reading', input_type: 'amount_primary' as const }
const listening = { id: 2, name: 'Listening', input_type: 'time_primary' as const }
const writing = { id: 3, name: 'Writing', input_type: 'amount_primary' as const }
const unit = (
  unit_key: string,
  log_activity_id: number,
  name: string,
  language_code?: string,
) => ({ unit_key, log_activity_id, name, modifier: 1, language_code })
const readingPage = unit('pages', 1, 'Page')
const readingCharacter = unit('characters', 1, 'Character')
const readingCharacterJpn = unit('characters', 1, 'Character', 'jpn')
const readingCharacterKor = unit('characters', 1, 'Character', 'kor')
const writingPage = unit('pages', 3, 'Page')

const options = {
  languages: [
    { code: 'jpn', name: 'Japanese' },
    { code: 'kor', name: 'Korean' },
    { code: 'eng', name: 'English' },
  ],
  activities: [reading, listening, writing],
  units: [
    readingPage,
    readingCharacter,
    readingCharacterJpn,
    readingCharacterKor,
    writingPage,
  ],
  tags: [],
  user_language_codes: [],
  scoring_engine_enabled: true,
}

const field = (name: string) =>
  document.querySelector<HTMLSelectElement>(`select[name="${name}"]`)!

afterEach(() => {
  cleanup()
  createLog.mockReset()
})

describe('new log unit selection', () => {
  it('keeps the selected unit when the language changes', () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)
    const language = field('languageCode')
    const unitSelect = field('amountUnit')

    fireEvent.change(unitSelect, { target: { value: 'characters' } })
    fireEvent.change(language, { target: { value: 'kor' } })
    expect(unitSelect.value).toBe('characters')

    fireEvent.change(language, { target: { value: 'eng' } })
    expect(unitSelect.value).toBe('characters')
  })

  it('selects the first unit when the activity has no matching unit', () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)
    const unitSelect = field('amountUnit')

    fireEvent.change(unitSelect, { target: { value: 'characters' } })
    fireEvent.change(field('activityId'), { target: { value: String(writing.id) } })
    expect(unitSelect.value).toBe('pages')
  })

  it('submits the selected unit key', async () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)

    fireEvent.change(field('amountUnit'), { target: { value: 'characters' } })
    fireEvent.change(screen.getByLabelText('Amount', { selector: 'input' }), {
      target: { value: '12' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Create' }))

    await waitFor(() =>
      expect(createLog).toHaveBeenCalledWith({
        language_code: 'jpn',
        activity_id: reading.id,
        amount: 12,
        unit_key: 'characters',
        tags: [],
        description: '',
      }),
    )
  })

  it('offers no unit for a timed activity and submits only its duration', async () => {
    render(<LogFormV2 options={options} defaultValues={{ languageCode: 'jpn' }} />)

    fireEvent.change(field('activityId'), { target: { value: String(listening.id) } })
    expect(field('amountUnit')).toBeNull()

    fireEvent.change(document.querySelector('input[name="durationMinutes"]')!, {
      target: { value: '15' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'Create' }))

    await waitFor(() =>
      expect(createLog).toHaveBeenCalledWith({
        language_code: 'jpn',
        activity_id: listening.id,
        duration_seconds: 900,
        tags: [],
        description: '',
      }),
    )
  })
})
