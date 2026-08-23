import { NextApiRequest, NextApiResponse } from 'next'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import handler from '../../pages/api/feature-flags'

const mocks = vi.hoisted(() => ({
  toSession: vi.fn(),
  decisionsForSession: vi.fn(),
}))

vi.mock('@app/common/ory', () => ({
  sdkServer: { toSession: mocks.toSession },
}))

vi.mock('@app/feature-flags/server', () => ({
  decisionsForSession: mocks.decisionsForSession,
}))

const createResponse = () => {
  const response = {
    setHeader: vi.fn(),
    status: vi.fn(),
    json: vi.fn(),
  }
  response.status.mockReturnValue(response)
  return response as unknown as NextApiResponse
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('GET /api/feature-flags', () => {
  it('returns only safe typed decisions when there is no session cookie', async () => {
    const response = createResponse()

    await handler({ method: 'GET', headers: {} } as NextApiRequest, response)

    expect(response.status).toHaveBeenCalledWith(200)
    expect(response.json).toHaveBeenCalledWith({
      decisions: { 'release.log-entry-v2': false },
    })
    expect(mocks.toSession).not.toHaveBeenCalled()
    expect(mocks.decisionsForSession).not.toHaveBeenCalled()
  })

  it('resolves Kratos server-side and returns the allowlisted decision', async () => {
    const response = createResponse()
    const session = {
      identity: { id: '123e4567-e89b-12d3-a456-426614174000' },
    }
    mocks.toSession.mockResolvedValue({ data: session })
    mocks.decisionsForSession.mockResolvedValue({
      'release.log-entry-v2': true,
    })

    await handler(
      {
        method: 'GET',
        headers: { cookie: 'ory_kratos_session=secret' },
      } as NextApiRequest,
      response,
    )

    expect(mocks.toSession).toHaveBeenCalledWith(
      undefined,
      'ory_kratos_session=secret',
    )
    expect(mocks.decisionsForSession).toHaveBeenCalledWith(session)
    expect(response.json).toHaveBeenCalledWith({
      decisions: { 'release.log-entry-v2': true },
    })
  })

  it('fails closed when Kratos or the provider is unavailable', async () => {
    const response = createResponse()
    mocks.toSession.mockRejectedValue(new Error('unavailable'))

    await handler(
      {
        method: 'GET',
        headers: { cookie: 'ory_kratos_session=secret' },
      } as NextApiRequest,
      response,
    )

    expect(response.status).toHaveBeenCalledWith(200)
    expect(response.json).toHaveBeenCalledWith({
      decisions: { 'release.log-entry-v2': false },
    })
  })

  it('rejects non-GET methods without evaluating', async () => {
    const response = createResponse()

    await handler({ method: 'POST', headers: {} } as NextApiRequest, response)

    expect(response.setHeader).toHaveBeenCalledWith('Allow', 'GET')
    expect(response.status).toHaveBeenCalledWith(405)
    expect(mocks.decisionsForSession).not.toHaveBeenCalled()
  })
})
