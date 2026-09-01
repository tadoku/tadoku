// @vitest-environment jsdom

import { Provider, useSetAtom } from 'jotai'
import {
  act,
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
  useLatchedFeatureFlag,
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

vi.mock('next/config', () => ({
  default: () => ({
    publicRuntimeConfig: { apiEndpoint: 'https://tadoku.test/api/internal' },
  }),
}))

const FlagValue = () => {
  const enabled = useFeatureFlag('release-log-entry-v2')
  return <output>{enabled ? 'enabled' : 'disabled'}</output>
}

const LatchedFlagValue = ({
  enabledOverride = false,
}: {
  enabledOverride?: boolean
}) => {
  const enabled = useLatchedFeatureFlag(
    'release-log-entry-v2',
    true,
    enabledOverride,
  )
  const setDecisions = useSetAtom(featureFlagDecisionsAtom)

  return (
    <>
      <button
        type="button"
        onClick={() => setDecisions({ 'release-log-entry-v2': false })}
      >
        Disable
      </button>
      {enabled ? (
        <label>
          V2 draft
          <input defaultValue="before decision change" />
        </label>
      ) : (
        <output>latched disabled</output>
      )}
    </>
  )
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
      [featureFlagDecisionsAtom, { 'release-log-entry-v2': enabled }],
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

  it('keeps a mounted feature flow stable until the next mount', () => {
    const mounted = render(
      <Wrapper enabled>
        <LatchedFlagValue />
      </Wrapper>,
    )

    const input = screen.getByLabelText('V2 draft') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'unsaved form value' } })
    fireEvent.click(screen.getByRole('button', { name: 'Disable' }))

    expect(input.value).toBe('unsaved form value')
    expect(screen.getByLabelText('V2 draft')).toBe(input)
    mounted.unmount()

    render(
      <Wrapper>
        <LatchedFlagValue />
      </Wrapper>,
    )

    expect(screen.getByText('latched disabled')).toBeTruthy()
  })

  it('allows an existing behavior to override a disabled flag', () => {
    render(
      <Wrapper>
        <LatchedFlagValue enabledOverride />
      </Wrapper>,
    )

    expect(screen.getByLabelText('V2 draft')).toBeTruthy()
  })

  it('uses the legacy default while refreshing after navigation without remounting a form', async () => {
    let resolveRefresh: (response: Response) => void = () => {}
    const fetchMock = vi.mocked(fetch).mockReturnValue(
      new Promise(resolve => {
        resolveRefresh = resolve
      }),
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

    act(() => routerEvents.emit())

    await waitFor(() => expect(screen.getByText('disabled')).toBeTruthy())
    expect(input.value).toBe('unsaved form value')
    expect(fetchMock).toHaveBeenCalledWith(
      'https://tadoku.test/api/internal/immersion/feature-flags',
      {
        credentials: 'include',
        cache: 'no-store',
        signal: expect.any(AbortSignal),
      },
    )

    resolveRefresh(
      new Response(
        JSON.stringify({ decisions: { 'release-log-entry-v2': true } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    await waitFor(() => expect(screen.getByText('enabled')).toBeTruthy())
    expect(input.value).toBe('unsaved form value')
  })

  it('keeps the legacy default when refresh fails', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('provider unavailable'))

    render(
      <Wrapper enabled>
        <FeatureFlagRefresh>
          <FlagValue />
        </FeatureFlagRefresh>
      </Wrapper>,
    )
    act(() => routerEvents.emit())

    await waitFor(() => expect(fetch).toHaveBeenCalled())
    expect(screen.getByText('disabled')).toBeTruthy()
  })

  it('bounds a stalled browser refresh and keeps the form mounted', async () => {
    vi.useFakeTimers()
    vi.mocked(fetch).mockImplementation((_, init) => {
      const signal = init?.signal
      return new Promise<Response>((_, reject) => {
        signal?.addEventListener('abort', () => reject(signal.reason))
      })
    })

    render(
      <Wrapper enabled>
        <FeatureFlagRefresh timeoutMilliseconds={50}>
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

    act(() => routerEvents.emit())
    await act(() => vi.advanceTimersByTimeAsync(50))

    expect(screen.getByText('disabled')).toBeTruthy()
    expect(input.value).toBe('unsaved form value')
    vi.useRealTimers()
  })
})
