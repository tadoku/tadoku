import { afterEach, describe, expect, it, vi } from 'vitest'

const { useMutation } = vi.hoisted(() => ({ useMutation: vi.fn() }))

vi.mock('next/config', () => ({
  default: () => ({
    publicRuntimeConfig: {
      homeUrl: 'https://tadoku.test',
      adminUrl: 'https://admin.tadoku.test',
      apiEndpoint: 'https://tadoku.test/api/internal',
    },
  }),
}))

vi.mock('@tanstack/react-query', () => ({
  useMutation,
  useQuery: vi.fn(),
  useQueryClient: vi.fn(),
}))

import { useCreateContest } from './contest-api'

const form = {
  title: 'Contest',
  description: 'Description',
  contest_start: '2026-01-01',
  contest_end: '2026-01-31',
  registration_end: '2026-01-30',
  private: false,
  official: false,
  language_code_allow_list: [],
  activity_type_id_allow_list: [],
}

function answerCreatedContest(contest: Record<string, unknown>) {
  vi.stubGlobal('fetch', async () => new Response(
    JSON.stringify({ ...form, ...contest }),
    { status: 200, headers: { 'Content-Type': 'application/json' } },
  ))
}

async function submit() {
  const options = useMutation.mock.lastCall?.[0]
  options.onSuccess(await options.mutationFn(form))
}

describe('useCreateContest', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('fails like a malformed answer when the created contest has no ID', async () => {
    answerCreatedContest({})
    const navigate = vi.fn()
    useCreateContest(navigate)

    await expect(submit()).rejects.toThrow('unavailable')
    expect(navigate).not.toHaveBeenCalled()
  })

  it('navigates to the created contest', async () => {
    answerCreatedContest({ id: '00000000-0000-4000-8000-000000000001' })
    const navigate = vi.fn()
    useCreateContest(navigate)

    await submit()
    expect(navigate).toHaveBeenCalledWith('00000000-0000-4000-8000-000000000001')
  })
})
