import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('next/config', () => ({
  default: () => ({
    publicRuntimeConfig: { apiEndpoint: 'https://tadoku.test/api/internal' },
  }),
}))

import { getFeatureAccess, updateFeatureAccess } from './api'

const targetUserId = '0198f6c5-c4af-7b1d-9776-884c065d72db'
const result = {
  enabled: true,
  changed: true,
  environment: 'production',
  revision: 'a'.repeat(40),
}

describe('feature access browser API', () => {
  beforeEach(() => vi.restoreAllMocks())

  it('calls the public immersion endpoint directly', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(new Response(JSON.stringify(result), { status: 200 }))

    await expect(
      getFeatureAccess({ targetUserId, flagKey: 'release-log-entry-v2' }),
    ).resolves.toEqual(result)

    expect(fetchMock).toHaveBeenCalledWith(
      `https://tadoku.test/api/internal/immersion/admin/feature-flags/release-log-entry-v2/users/${targetUserId}`,
      { credentials: 'include' },
    )
  })

  it.each([
    [true, 'PUT'],
    [false, 'DELETE'],
  ])('uses the resource method for enabled=%s', async (enabled, method) => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(new Response(JSON.stringify(result), { status: 200 }))

    await updateFeatureAccess({
      targetUserId,
      flagKey: 'release-log-entry-v2',
      enabled,
    })

    expect(fetchMock).toHaveBeenCalledWith(
      `https://tadoku.test/api/internal/immersion/admin/feature-flags/release-log-entry-v2/users/${targetUserId}`,
      { method, credentials: 'include' },
    )
  })
})
