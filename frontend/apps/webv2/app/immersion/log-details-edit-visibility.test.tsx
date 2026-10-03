import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Log } from '@app/immersion/api'

Reflect.set(globalThis, 'React', React)

const { useSession, useUserRole } = vi.hoisted(() => ({
  useSession: vi.fn(),
  useUserRole: vi.fn(),
}))

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: {} }),
}))

vi.mock('@app/immersion/api', () => ({
  useDeleteLog: () => ({ mutate: vi.fn() }),
}))

vi.mock('@app/common/session', () => ({ useSession, useUserRole }))

vi.mock('next/router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

import { LogDetailsV2 } from '@app/immersion/LogDetailsV2'

const log = {
  id: 'log-id',
  user_id: 'owner-id',
  user_display_name: 'owner',
  created_at: '2026-01-01T00:00:00Z',
  activity: { id: 1, name: 'Reading' },
  language: { code: 'jpa', name: 'Japanese' },
  tags: [],
  amount: 10,
  modifier: 1,
  score: 10,
  unit_name: 'page',
  registrations: [],
} as unknown as Log

describe('log details edit visibility', () => {
  beforeEach(() => {
    useSession.mockReset()
    useUserRole.mockReset()
  })

  it.each([
    { viewer: 'owner', id: 'owner-id', role: 'user', canEdit: true },
    { viewer: 'admin', id: 'admin-id', role: 'admin', canEdit: true },
    { viewer: 'other user', id: 'other-id', role: 'user', canEdit: false },
    { viewer: 'anonymous', id: undefined, role: undefined, canEdit: false },
  ])('$viewer sees edit links: $canEdit', ({ id, role, canEdit }) => {
    useSession.mockReturnValue([id ? { identity: { id } } : undefined])
    useUserRole.mockReturnValue(role)

    const markup = renderToStaticMarkup(<LogDetailsV2 log={log} />)

    expect(markup).toContain('Submitted to contests')
    expect(markup.includes('href="/logs/log-id/edit"')).toBe(canEdit)
    expect(markup.includes('href="/logs/log-id/contests"')).toBe(canEdit)
  })
})
