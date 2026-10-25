import { Account } from 'ui/identity'
import { describe, expect, it, vi } from 'vitest'
import { bootstrapFeatureFlagDecisions } from './bootstrap'

vi.mock('next/config', () => ({
  default: () => ({
    publicRuntimeConfig: { apiEndpoint: 'https://tadoku.test/api/internal' },
    serverRuntimeConfig: {
      apiEndpoint: 'http://oathkeeper-proxy:4455/api/internal',
    },
  }),
}))

const sent = (request: { mock: { calls: unknown[][] } }) =>
  request.mock.calls[0][0] as Request

const session = {
  id: '123e4567-e89b-12d3-a456-426614174000',
} as Account

describe('bootstrapFeatureFlagDecisions', () => {
  it('bootstraps the server decision alongside an authenticated session', async () => {
    const request = vi
      .fn()
      .mockResolvedValue(
        new Response(
          JSON.stringify({ decisions: { 'release-webv3': true } }),
          { status: 200 },
        ),
      )

    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'tadoku_session=secret',
        true,
        request,
      ),
    ).resolves.toEqual({ 'release-webv3': true })
    expect(sent(request).url).toBe(
      'http://oathkeeper-proxy:4455/api/internal/immersion/feature-flags',
    )
    expect(sent(request).cache).toBe('no-store')
    expect(sent(request).headers.get('cookie')).toBe('tadoku_session=secret')
    expect(sent(request).headers.has('x-dev-branch')).toBe(false)
  })

  it('forwards the incoming branch so a branch host gets its own tenant', async () => {
    const request = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify({ decisions: {} }), { status: 200 }))

    await bootstrapFeatureFlagDecisions(
      session,
      'tadoku_session=secret',
      true,
      request,
      3_000,
      'http://oathkeeper-proxy:4455/api/internal',
      'alice-0123abcd',
    )
    expect(sent(request).url).toBe(
      'http://oathkeeper-proxy:4455/api/internal/immersion/feature-flags',
    )
    expect(sent(request).headers.get('cookie')).toBe('tadoku_session=secret')
    expect(sent(request).headers.get('x-dev-branch')).toBe('alice-0123abcd')
  })

  it('uses the registry default without loading Flipt when there is no session', async () => {
    const request = vi.fn()

    await expect(
      bootstrapFeatureFlagDecisions(undefined, undefined, true, request),
    ).resolves.toEqual({ 'release-webv3': false })
    expect(request).not.toHaveBeenCalled()
  })

  it('uses the registry default when the provider cannot load', async () => {
    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'tadoku_session=secret',
        true,
        async () => {
          throw new Error('provider unavailable')
        },
      ),
    ).resolves.toEqual({ 'release-webv3': false })
  })

  it('bounds a stalled same-origin bootstrap request', async () => {
    vi.useFakeTimers()
    const request = vi.fn(
      (input: RequestInfo | URL) =>
        new Promise<Response>((_, reject) => {
          const { signal } = input as Request
          signal.addEventListener('abort', () => reject(signal.reason))
        }),
    )

    const result = bootstrapFeatureFlagDecisions(
      session,
      'tadoku_session=secret',
      true,
      request as typeof fetch,
      50,
    )
    await vi.advanceTimersByTimeAsync(50)

    await expect(result).resolves.toEqual({ 'release-webv3': false })
    expect(sent(request).url).toBe(
      'http://oathkeeper-proxy:4455/api/internal/immersion/feature-flags',
    )
    expect(sent(request).signal.aborted).toBe(true)
    vi.useRealTimers()
  })

  it('never loads the server evaluator in the browser', async () => {
    const request = vi.fn()

    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'tadoku_session=secret',
        false,
        request,
      ),
    ).resolves.toEqual({ 'release-webv3': false })
    expect(request).not.toHaveBeenCalled()
  })

  it('does not request decisions for a malformed account subject', async () => {
    const request = vi.fn()
    const malformedSession = {
      id: 'not-a-uuid',
    } as Account

    await expect(
      bootstrapFeatureFlagDecisions(
        malformedSession,
        'tadoku_session=secret',
        true,
        request,
      ),
    ).resolves.toEqual({ 'release-webv3': false })
    expect(request).not.toHaveBeenCalled()
  })

  it('uses the registry defaults when the answer holds an unregistered flag', async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          decisions: {
            'release-webv3': true,
            'unregistered-flag': true,
          },
        }),
        { status: 200 },
      ),
    )

    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'tadoku_session=secret',
        true,
        request,
      ),
    ).resolves.toEqual({ 'release-webv3': false })
  })
})
