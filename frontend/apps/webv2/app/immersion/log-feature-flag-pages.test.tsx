import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { beforeEach, describe, expect, it, vi } from 'vitest'

Reflect.set(globalThis, 'React', React)

const {
  featureDecision,
  useLatchedFeatureFlag,
  useLog,
  useLogConfigurationOptions,
  useOngoingContestRegistrations,
  useUserRole,
} = vi.hoisted(() => ({
  featureDecision: { enabled: false },
  useLatchedFeatureFlag: vi.fn(),
  useLog: vi.fn(),
  useLogConfigurationOptions: vi.fn(),
  useOngoingContestRegistrations: vi.fn(),
  useUserRole: vi.fn(),
}))

vi.mock('@app/feature-flags/client', () => ({ useLatchedFeatureFlag }))

vi.mock('@app/common/session', () => ({
  useSession: () => [{ identity: { id: 'user-id' } }],
  useSessionOrRedirect: vi.fn(),
  useUserRole,
}))

vi.mock('@app/common/hooks', () => ({
  useCurrentDateTime: () => ({ diff: () => ({ as: () => 1 }) }),
}))

vi.mock('@app/common/format', () => ({
  colorForActivity: () => 'lime',
  formatArray: () => null,
  formatDuration: () => '1 minute',
  formatScore: (value: number) => value.toString(),
  formatUnit: () => 'pages',
  hasTrackedAmount: () => true,
}))

vi.mock('@app/immersion/api', () => ({
  useDeleteLog: () => ({ mutate: vi.fn() }),
  useLog,
  useLogConfigurationOptions,
  useOngoingContestRegistrations,
}))

vi.mock('@app/immersion/NewLogForm/Form', () => ({
  LogForm: () => <div>legacy create form</div>,
}))

vi.mock('@app/immersion/NewLogFormV2/Form', () => ({
  LogFormV2: () => <div>v2 create form</div>,
}))

vi.mock('@app/immersion/LogDetailsV2', () => ({
  LogDetailsV2: () => <div>v2 log details</div>,
}))

vi.mock('@app/immersion/EditLogForm/Form', () => ({
  EditLogForm: () => <div>v2 edit form</div>,
}))

vi.mock('next/router', () => ({
  useRouter: () => ({
    push: vi.fn(),
    query: { amount: '12', id: 'log-id' },
    replace: vi.fn(),
  }),
}))

vi.mock('next/head', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: {} }),
}))

vi.mock('@heroicons/react/20/solid', () => ({
  HomeIcon: () => null,
  TrashIcon: () => null,
}))

vi.mock('@heroicons/react/24/outline', () => ({
  XMarkIcon: () => null,
}))

vi.mock('ui', () => ({
  Breadcrumb: () => null,
  ButtonGroup: () => null,
  Loading: () => <div>loading</div>,
  Modal: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

import LogDetailsPage from '../../pages/logs/[id]'
import EditLogPage from '../../pages/logs/[id]/edit'
import NewLogPage from '../../pages/logs/new'

const optionsResult = {
  data: {},
  isError: false,
  isLoading: false,
  isSuccess: true,
}

const registrationsResult = {
  data: { registrations: [] },
  isError: false,
  isLoading: false,
  isSuccess: true,
}

const log = {
  activity: { id: 1, name: 'Reading' },
  amount: 12,
  created_at: '2026-08-31T00:00:00Z',
  deleted: false,
  description: undefined,
  id: 'log-id',
  language: { name: 'Japanese' },
  modifier: 1,
  registrations: [],
  score: 12,
  tags: [],
  unit_name: 'page',
  user_display_name: 'Reader',
  user_id: 'user-id',
}

describe('release-log-entry-v2 pages', () => {
  beforeEach(() => {
    useLatchedFeatureFlag.mockReset()
    useLog.mockReset()
    useLogConfigurationOptions.mockReset()
    useOngoingContestRegistrations.mockReset()
    useUserRole.mockReset()

    featureDecision.enabled = false
    useLatchedFeatureFlag.mockImplementation(
      (_flag: string, _ready: boolean, enabledOverride = false) =>
        enabledOverride || featureDecision.enabled,
    )
    useUserRole.mockReturnValue('user')
    useLogConfigurationOptions.mockReturnValue(optionsResult)
    useOngoingContestRegistrations.mockReturnValue(registrationsResult)
    useLog.mockReturnValue({
      data: log,
      isError: false,
      isIdle: false,
      isLoading: false,
    })
  })

  it('renders legacy create and details pages for an ordinary user when disabled', () => {
    expect(renderToStaticMarkup(<NewLogPage />)).toContain('legacy create form')
    expect(useOngoingContestRegistrations).toHaveBeenCalledWith({
      enabled: true,
    })

    const details = renderToStaticMarkup(<LogDetailsPage />)
    expect(details).toContain('Log details')
    expect(details).not.toContain('v2 log details')
  })

  it('preserves the existing V2 journey for admins when the flag is disabled', () => {
    useUserRole.mockReturnValue('admin')

    expect(renderToStaticMarkup(<NewLogPage />)).toContain('v2 create form')
    expect(renderToStaticMarkup(<LogDetailsPage />)).toContain('v2 log details')
    expect(renderToStaticMarkup(<EditLogPage />)).toContain('v2 edit form')

    expect(useLatchedFeatureFlag).toHaveBeenCalledWith(
      'release-log-entry-v2',
      true,
      true,
    )
  })

  it('renders V2 create and details pages for a targeted non-admin', () => {
    featureDecision.enabled = true

    expect(renderToStaticMarkup(<NewLogPage />)).toContain('v2 create form')
    expect(useOngoingContestRegistrations).toHaveBeenCalledWith({
      enabled: false,
    })
    expect(renderToStaticMarkup(<LogDetailsPage />)).toContain('v2 log details')
  })

  it('gates editing for ordinary users on the feature decision', () => {
    expect(renderToStaticMarkup(<EditLogPage />)).toContain(
      'This feature is not yet available.',
    )

    featureDecision.enabled = true
    expect(renderToStaticMarkup(<EditLogPage />)).toContain('v2 edit form')
  })
})
