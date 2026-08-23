import { Session } from '@ory/client'
import { describe, expect, it, vi } from 'vitest'
import { bootstrapFeatureFlagDecisions } from './bootstrap'

const session = {
  identity: { id: '123e4567-e89b-12d3-a456-426614174000' },
} as unknown as Session

describe('bootstrapFeatureFlagDecisions', () => {
  it('bootstraps the server decision alongside an authenticated session', async () => {
    const request = vi
      .fn()
      .mockResolvedValue(
        new Response(
          JSON.stringify({ decisions: { 'release.log-entry-v2': true } }),
          { status: 200 },
        ),
      )

    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'ory_kratos_session=secret',
        true,
        request,
      ),
    ).resolves.toEqual({ 'release.log-entry-v2': true })
    expect(request).toHaveBeenCalledWith(
      'http://127.0.0.1:3000/api/feature-flags',
      {
        cache: 'no-store',
        headers: { cookie: 'ory_kratos_session=secret' },
        signal: expect.any(AbortSignal),
      },
    )
  })

  it('uses the registry default without loading Flipt when there is no session', async () => {
    const request = vi.fn()

    await expect(
      bootstrapFeatureFlagDecisions(undefined, undefined, true, request),
    ).resolves.toEqual({ 'release.log-entry-v2': false })
    expect(request).not.toHaveBeenCalled()
  })

  it('uses the registry default when the provider cannot load', async () => {
    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'ory_kratos_session=secret',
        true,
        async () => {
          throw new Error('provider unavailable')
        },
      ),
    ).resolves.toEqual({ 'release.log-entry-v2': false })
  })

  it('bounds a stalled same-origin bootstrap request', async () => {
    vi.useFakeTimers()
    const request = vi.fn((_: RequestInfo | URL, init?: RequestInit) =>
      new Promise<Response>((_, reject) => {
        init?.signal?.addEventListener('abort', () => reject(init.signal?.reason))
      }),
    ) as typeof fetch

    const result = bootstrapFeatureFlagDecisions(
      session,
      'ory_kratos_session=secret',
      true,
      request,
      50,
    )
    await vi.advanceTimersByTimeAsync(50)

    await expect(result).resolves.toEqual({ 'release.log-entry-v2': false })
    expect(request).toHaveBeenCalledWith(
      'http://127.0.0.1:3000/api/feature-flags',
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    )
    vi.useRealTimers()
  })

  it('never loads the server evaluator in the browser', async () => {
    const request = vi.fn()

    await expect(
      bootstrapFeatureFlagDecisions(
        session,
        'ory_kratos_session=secret',
        false,
        request,
      ),
    ).resolves.toEqual({ 'release.log-entry-v2': false })
    expect(request).not.toHaveBeenCalled()
  })

  it('does not request decisions for a malformed Kratos subject', async () => {
    const request = vi.fn()
    const malformedSession = {
      identity: { id: 'not-a-uuid' },
    } as unknown as Session

    await expect(
      bootstrapFeatureFlagDecisions(
        malformedSession,
        'ory_kratos_session=secret',
        true,
        request,
      ),
    ).resolves.toEqual({ 'release.log-entry-v2': false })
    expect(request).not.toHaveBeenCalled()
  })
})
