// @vitest-environment jsdom

import { Provider } from 'jotai'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import React, { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  FeatureFlagRefresh,
  featureFlagDecisionsAtom,
  useFeatureFlag,
} from './client'

const routerEvents = vi.hoisted(() => {
  const listeners = new Set<() => void>()
  return {
    emit: () => listeners.forEach(listener => listener()),
    on: vi.fn((_: string, listener: () => void) => listeners.add(listener)),
    off: vi.fn((_: string, listener: () => void) => listeners.delete(listener)),
    reset: () => listeners.clear(),
  }
})

vi.mock('next/router', () => ({
  useRouter: () => ({ events: routerEvents }),
}))

const FlagValue = () => {
  const enabled = useFeatureFlag('release.log-entry-v2')
  return <output>{enabled ? 'enabled' : 'disabled'}</output>
}

const Wrapper = ({
  children,
  enabled = false,
}: {
  children: ReactNode
  enabled?: boolean
}) => (
  <Provider
    initialValues={[
      [featureFlagDecisionsAtom, { 'release.log-entry-v2': enabled }],
    ]}
  >
    {children}
  </Provider>
)

beforeEach(() => {
  routerEvents.reset()
  vi.stubGlobal('fetch', vi.fn())
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('feature flag browser state', () => {
  it('renders the hydrated decision on the first render without flicker', () => {
    render(
      <Wrapper enabled>
        <FlagValue />
      </Wrapper>,
    )

    expect(screen.getByText('enabled')).toBeTruthy()
    expect(screen.queryByText('disabled')).toBeNull()
  })

  it('defaults to the legacy behavior', () => {
    render(
      <Wrapper>
        <FlagValue />
      </Wrapper>,
    )

    expect(screen.getByText('disabled')).toBeTruthy()
  })

  it('refreshes a decision after navigation without remounting a form', async () => {
    const fetchMock = vi
      .mocked(fetch)
      .mockResolvedValue(
        new Response(
          JSON.stringify({ decisions: { 'release.log-entry-v2': false } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ),
      )

    render(
      <Wrapper enabled>
        <FeatureFlagRefresh>
          <label>
            Draft
            <input defaultValue="before navigation" />
          </label>
          <FlagValue />
        </FeatureFlagRefresh>
      </Wrapper>,
    )
    const input = screen.getByLabelText('Draft') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'unsaved form value' } })

    routerEvents.emit()

    await waitFor(() => expect(screen.getByText('disabled')).toBeTruthy())
    expect(input.value).toBe('unsaved form value')
    expect(fetchMock).toHaveBeenCalledWith('/api/feature-flags', {
      credentials: 'same-origin',
      cache: 'no-store',
      signal: expect.any(AbortSignal),
    })
  })

  it('keeps the hydrated decision when refresh fails', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('provider unavailable'))

    render(
      <Wrapper enabled>
        <FeatureFlagRefresh>
          <FlagValue />
        </FeatureFlagRefresh>
      </Wrapper>,
    )
    routerEvents.emit()

    await waitFor(() => expect(fetch).toHaveBeenCalled())
    expect(screen.getByText('enabled')).toBeTruthy()
  })
})
