// @vitest-environment jsdom
import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { FeatureAccessModal } from './FeatureAccessModal'
import { useFeatureAccess, useUpdateFeatureAccess } from './api'

vi.mock('./api', () => ({
  useFeatureAccess: vi.fn(),
  useUpdateFeatureAccess: vi.fn(),
}))

vi.mock('react-toastify', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const mutateAsync = vi.fn()

describe('FeatureAccessModal', () => {
  beforeEach(() => {
    vi.mocked(useFeatureAccess).mockReturnValue({
      data: {
        enabled: false,
        environment: 'production',
        revision: 'a'.repeat(40),
      },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useFeatureAccess>)
    vi.mocked(useUpdateFeatureAccess).mockReturnValue({
      mutateAsync,
      isLoading: false,
    } as unknown as ReturnType<typeof useUpdateFeatureAccess>)
    mutateAsync.mockResolvedValue({
      enabled: true,
      changed: true,
      environment: 'production',
      revision: 'b'.repeat(40),
    })
  })

  it('shows current Tadoku identity but submits only the UUID and allowlisted flag', async () => {
    render(
      <FeatureAccessModal
        isOpen
        setIsOpen={vi.fn()}
        user={{
          id: '0198f6c5-c4af-7b1d-9776-884c065d72db',
          display_name: 'Named User',
          email: 'named@example.test',
          created_at: '2026-09-01T00:00:00Z',
          role: 'admin',
        }}
      />,
    )

    expect(screen.getByText('Named User')).toBeInTheDocument()
    expect(screen.getByText('named@example.test')).toBeInTheDocument()
    expect(
      screen.getByRole('option', { name: 'Release log entry v2' }),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Grant access' }))

    expect(mutateAsync).toHaveBeenCalledWith({
      targetUserId: '0198f6c5-c4af-7b1d-9776-884c065d72db',
      flagKey: 'release-log-entry-v2',
      enabled: true,
    })
    expect(JSON.stringify(mutateAsync.mock.calls)).not.toContain(
      'named@example.test',
    )
    expect(JSON.stringify(mutateAsync.mock.calls)).not.toContain('Named User')
  })
})
