// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { LogConfigurationOptions } from '@app/immersion/api'

Reflect.set(globalThis, 'React', React)

const { updateLog } = vi.hoisted(() => ({ updateLog: vi.fn() }))

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))
vi.mock('next/router', () => ({ useRouter: () => ({ replace: vi.fn() }) }))
vi.mock('@app/common/session', () => ({ useSessionOrRedirect: vi.fn() }))
vi.mock('@app/immersion/api', () => ({
  fetchTagSuggestions: vi.fn(async () => []),
  useUpdateLog: () => ({ mutate: updateLog }),
}))
vi.mock('@app/immersion/scoring-api', () => ({
  useScorePreview: () => ({ data: undefined }),
}))

import { EditLogForm } from './Form'

type EditedLog = React.ComponentProps<typeof EditLogForm>['log']

const reading = { id: 1, name: 'Reading', input_type: 'amount_primary' as const }
const listening = { id: 2, name: 'Listening', input_type: 'time_primary' as const }
const unit = (id: string, unit_key: string, log_activity_id: number, name: string) => ({
  id,
  unit_key,
  log_activity_id,
  name,
  modifier: 1,
})

const options: LogConfigurationOptions = {
  languages: [{ code: 'jpn', name: 'Japanese' }],
  activities: [reading, listening],
  units: [
    unit('a0000000-0000-4000-8000-000000000001', 'reading_page', 1, 'Page'),
    unit('a0000000-0000-4000-8000-000000000002', 'reading_character', 1, 'Character'),
    unit('a0000000-0000-4000-8000-000000000003', 'listening_minute', 2, 'Minute'),
  ],
  user_language_codes: [],
  scoring_engine_enabled: true,
}

const storedLog = (overrides: Partial<EditedLog>): EditedLog => ({
  id: 'b0000000-0000-4000-8000-000000000001',
  language: { code: 'jpn', name: 'Japanese' },
  activity: reading,
  tags: [],
  amount: 0,
  description: 'before',
  ...overrides,
})

const minutesInput = () =>
  document.querySelector<HTMLInputElement>('input[name="durationMinutes"]')
const unitSelect = () =>
  document.querySelector<HTMLSelectElement>('select[name="amountUnit"]')
const save = () => fireEvent.click(screen.getByRole('button', { name: 'Save' }))

afterEach(() => {
  cleanup()
  updateLog.mockReset()
})

describe('editing a log', () => {
  it('saves the unit key of an amount log', async () => {
    const log = storedLog({ amount: 12, unit_key: 'reading_character' })
    render(<EditLogForm options={options} log={log} />)

    expect(unitSelect()!.value).toBe('reading_character')
    save()

    await waitFor(() =>
      expect(updateLog).toHaveBeenCalledWith({
        logId: log.id,
        payload: { tags: [], description: 'before', amount: 12, unit_key: 'reading_character' },
      }),
    )
  })

  it.each([
    { stored: 'a minute amount', duration_seconds: undefined, minutes: '15', seconds: 900 },
    { stored: 'a minute amount and a duration', duration_seconds: 600, minutes: '10', seconds: 600 },
  ])('opens a timed log stored with $stored as its duration', async ({ duration_seconds, minutes, seconds }) => {
    const log = storedLog({
      activity: listening,
      amount: 15,
      unit_key: 'listening_minute',
      duration_seconds,
    })
    render(<EditLogForm options={options} log={log} />)

    expect(unitSelect()).toBeNull()
    expect(minutesInput()!.value).toBe(minutes)
    save()

    await waitFor(() =>
      expect(updateLog).toHaveBeenCalledWith({
        logId: log.id,
        payload: { tags: [], description: 'before', duration_seconds: seconds },
      }),
    )
  })

  it('does not save a timed log without a duration', async () => {
    const log = storedLog({ activity: listening, duration_seconds: 600 })
    render(<EditLogForm options={options} log={log} />)

    fireEvent.change(minutesInput()!, { target: { value: '' } })
    save()

    await screen.findByText('Time spent is required')
    expect(updateLog).not.toHaveBeenCalled()
  })
})
