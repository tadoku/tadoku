import { beforeEach, describe, expect, it, vi } from 'vitest'

const { useQuery } = vi.hoisted(() => ({ useQuery: vi.fn() }))

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { apiEndpoint: '' } }),
}))

vi.mock('react-query', () => ({
  useMutation: vi.fn(),
  useQuery,
  useQueryClient: vi.fn(),
}))

import { useLog } from './api'

describe('useLog', () => {
  beforeEach(() => {
    useQuery.mockClear()
  })

  it('does not retry a not found response', () => {
    useLog('deleted-log')

    const options = useQuery.mock.calls[0]?.[2]
    expect(options).toMatchObject({ retry: expect.any(Function) })
    expect(options.retry(0, new Error('404'))).toBe(false)
  })

  it('keeps the default retry limit for other errors', () => {
    useLog('temporarily-unavailable-log')

    const options = useQuery.mock.calls[0]?.[2]
    expect(options.retry(0, new Error('500'))).toBe(true)
    expect(options.retry(2, new Error('500'))).toBe(true)
    expect(options.retry(3, new Error('500'))).toBe(false)
  })
})
