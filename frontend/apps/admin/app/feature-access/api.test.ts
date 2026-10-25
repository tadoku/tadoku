import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('next/config', () => ({
  default: () => ({
    publicRuntimeConfig: { homeUrl: 'https://tadoku.test', adminUrl: 'https://admin.tadoku.test', apiEndpoint: 'https://tadoku.test/api/internal' },
  }),
}))

import { getFeatureAccess, updateFeatureAccess } from './api'

const targetUserId = '0198f6c5-c4af-7b1d-9776-884c065d72db'
const endpoint = `https://tadoku.test/api/internal/immersion/admin/feature-flags/release-webv3/users/${targetUserId}`
const result = {
  enabled: true,
  changed: true,
  environment: 'production',
  revision: 'a'.repeat(40),
}
const answer = (body: unknown) =>
  vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(body), { status: 200 }))
const sent = (fetchMock: { mock: { calls: unknown[][] } }) => fetchMock.mock.calls[0][0] as Request

describe('feature access browser API', () => {
  beforeEach(() => vi.restoreAllMocks())

  it('calls the public immersion endpoint directly', async () => {
    const fetchMock = answer(result)

    await expect(
      getFeatureAccess({ targetUserId, flagKey: 'release-webv3' }),
    ).resolves.toEqual(result)

    expect(sent(fetchMock).url).toBe(endpoint)
    expect(sent(fetchMock).method).toBe('GET')
    expect(sent(fetchMock).credentials).toBe('include')
  })

  it.each([
    [true, 'PUT'],
    [false, 'DELETE'],
  ])('uses the resource method for enabled=%s', async (enabled, method) => {
    const fetchMock = answer(result)

    await updateFeatureAccess({
      targetUserId,
      flagKey: 'release-webv3',
      enabled,
    })

    expect(sent(fetchMock).url).toBe(endpoint)
    expect(sent(fetchMock).method).toBe(method)
    expect(sent(fetchMock).credentials).toBe('include')
  })

  it.each([
    ['personal data', { ...result, email: 'private@example.test' }],
    ['a missing change flag', { enabled: true, environment: 'production', revision: result.revision }],
    ['an unknown environment', { ...result, environment: 'attacker-controlled' }],
  ])('rejects an answer with %s', async (_, body) => {
    answer(body)

    await expect(
      getFeatureAccess({ targetUserId, flagKey: 'release-webv3' }),
    ).rejects.toThrow('feature access unavailable')
  })
})
