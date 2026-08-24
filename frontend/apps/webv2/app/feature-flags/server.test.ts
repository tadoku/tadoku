import { Session } from '@ory/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  init: vi.fn(),
  readFile: vi.fn(),
}))

vi.mock('node:fs/promises', () => ({ readFile: mocks.readFile }))
vi.mock('@flipt-io/flipt-client-js', () => ({
  ErrorStrategy: { Fallback: 'fallback' },
  FliptClient: { init: mocks.init },
}))

const session = {
  identity: { id: '123e4567-e89b-12d3-a456-426614174000' },
} as unknown as Session

let server: typeof import('./server')

beforeEach(async () => {
  vi.resetModules()
  vi.clearAllMocks()
  server = await import('./server')
})

afterEach(() => {
  server.closeFeatureFlagClient()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('server feature flag evaluator', () => {
  it('does not initialize Flipt for an unauthenticated request', async () => {
    await expect(server.decisionsForSession(undefined)).resolves.toEqual({
      'release-log-entry-v2': false,
    })
    expect(mocks.init).not.toHaveBeenCalled()
  })

  it('uses safe defaults when Flipt cannot initialize', async () => {
    vi.useFakeTimers()
    mocks.init.mockRejectedValue(new Error('snapshot returned 404'))

    await expect(server.decisionsForSession(session)).resolves.toEqual({
      'release-log-entry-v2': false,
    })
  })

  it('recovers in the background after a cold empty-state failure', async () => {
    vi.useFakeTimers()
    const evaluateBoolean = vi.fn().mockReturnValue({ enabled: true })
    mocks.init
      .mockRejectedValueOnce(new Error('snapshot returned 404'))
      .mockResolvedValueOnce({ evaluateBoolean, close: vi.fn() })

    await expect(server.decisionsForSession(session)).resolves.toEqual({
      'release-log-entry-v2': false,
    })

    await vi.advanceTimersByTimeAsync(30_000)

    expect(mocks.init).toHaveBeenCalledTimes(2)
    await expect(server.decisionsForSession(session)).resolves.toEqual({
      'release-log-entry-v2': true,
    })
  })

  it('evaluates the allowlisted key with UUID and non-PII context only', async () => {
    const evaluateBoolean = vi.fn().mockReturnValue({ enabled: true })
    mocks.init.mockResolvedValue({ evaluateBoolean, close: vi.fn() })

    await expect(server.decisionsForSession(session)).resolves.toEqual({
      'release-log-entry-v2': true,
    })
    expect(evaluateBoolean).toHaveBeenCalledWith({
      flagKey: 'release-log-entry-v2',
      entityId: '123e4567-e89b-12d3-a456-426614174000',
      context: { authenticated: 'true' },
    })
  })

  it('closes the polling SDK client on teardown', async () => {
    const close = vi.fn()
    mocks.init.mockResolvedValue({
      evaluateBoolean: vi.fn().mockReturnValue({ enabled: false }),
      close,
    })

    await server.decisionsForSession(session)
    server.closeFeatureFlagClient()

    expect(close).toHaveBeenCalledTimes(1)
  })

  it('never logs raw provider errors', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    mocks.init.mockResolvedValue({
      evaluateBoolean: vi.fn(() => {
        throw new Error('sensitive@example.com 123e4567-e89b-12d3-a456-426614174000')
      }),
      close: vi.fn(),
    })

    await expect(server.decisionsForSession(session)).resolves.toEqual({
      'release-log-entry-v2': false,
    })
    expect(JSON.stringify(warn.mock.calls)).not.toContain('sensitive@example.com')
    expect(JSON.stringify(warn.mock.calls)).not.toContain(session.identity.id)
    warn.mockRestore()
  })

  it('uses safe defaults for a malformed identity without initializing Flipt', async () => {
    const malformedSession = {
      identity: { id: 'not-a-uuid' },
    } as unknown as Session

    await expect(server.decisionsForSession(malformedSession)).resolves.toEqual(
      {
        'release-log-entry-v2': false,
      },
    )
    expect(mocks.init).not.toHaveBeenCalled()
  })

  it('exchanges a fresh service token and sends the required snapshot headers', async () => {
    mocks.readFile.mockResolvedValue('projected-token\n')
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ access_token: 'exchange-token' }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(new Response('{}', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    await server.createFliptFetcher()({ etag: 'snapshot-etag' })

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      'http://oathkeeper-proxy.default:4455/token-exchange/flipt-evaluation/frontend-webv2',
      {
        headers: { Authorization: 'Bearer projected-token' },
        signal: expect.any(AbortSignal),
      },
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      'http://oathkeeper-proxy.default:4455/flipt/internal/v1/evaluation/snapshot/namespace/default',
      {
        headers: {
          Accept: 'application/json',
          Authorization: 'Bearer exchange-token',
          'If-None-Match': 'snapshot-etag',
          'x-flipt-accept-server-version': '1.47.0',
          'x-flipt-environment': 'local',
        },
        signal: expect.any(AbortSignal),
      },
    )
    expect(fetchMock.mock.calls[0][1]?.signal).toBe(
      fetchMock.mock.calls[1][1]?.signal,
    )
  })
})
