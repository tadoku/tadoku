import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'

Reflect.set(globalThis, 'React', React)

const { useLog } = vi.hoisted(() => ({ useLog: vi.fn() }))

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: {} }),
}))

vi.mock('@app/immersion/api', () => ({
  useDeleteLog: vi.fn(),
  useLog,
}))

vi.mock('@app/common/session', () => ({
  useSession: () => [undefined],
  useUserRole: () => 'user',
}))

vi.mock('next/router', () => ({
  useRouter: () => ({ query: { id: 'deleted-log' } }),
}))

vi.mock('next/error', () => ({
  default: ({ statusCode }: { statusCode: number }) => (
    <main>Not found ({statusCode})</main>
  ),
}))

import LogPage from '../../pages/logs/[id]'

describe('log details errors', () => {
  beforeEach(() => {
    useLog.mockReset()
  })

  it('renders the not found page when the log API returns 404', () => {
    useLog.mockReturnValue({
      error: new Error('404'),
      isError: true,
      isIdle: false,
      isLoading: false,
    })

    expect(renderToStaticMarkup(<LogPage />)).toContain('Not found (404)')
  })

  it('keeps the transient error message for other failures', () => {
    useLog.mockReturnValue({
      error: new Error('500'),
      isError: true,
      isIdle: false,
      isLoading: false,
    })

    expect(renderToStaticMarkup(<LogPage />)).toContain(
      'Could not load page, please try again later.',
    )
  })
})
