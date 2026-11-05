import { queryOptions, useQuery, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionLanguage } from 'api-client'
import { z } from 'zod'
import type { Read } from './leaderboard'
import { viewerKey, viewerQueryOptions } from './session'

/** The signed-in viewer's registration for one contest, or `unavailable` when it can't be read. */
export type RegistrationRead =
  | { kind: 'registered'; languages: ImmersionLanguage[] }
  | { kind: 'not-registered' }
  | { kind: 'unavailable' }

/** A registration save for the account that opened the form, with one to three distinct language codes. */
export const registrationSaveSchema = z.object({
  contestId: z.guid(),
  viewerId: z.guid(),
  languageCodes: z.array(z.string().min(1)).min(1).max(3).refine(codes => new Set(codes).size === codes.length),
})

export type RegistrationSave = z.infer<typeof registrationSaveSchema>

export type RegistrationFailure =
  | 'signed-out'
  | 'switched'
  | 'account-unavailable'
  | 'suspended'
  | 'closed'
  | 'invalid'
  | 'not-found'
  | 'deleting'
  | 'unavailable'

export type RegistrationResult = { ok: true } | { ok: false; reason: RegistrationFailure }

const getRegistration = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readRegistration } = await import('./registration.server')
    return readRegistration(data)
  })

const getContestLanguages = createServerFn({ method: 'GET' }).handler(async () => {
  const { readContestLanguages } = await import('./registration.server')
  return readContestLanguages()
})

const putRegistration = createServerFn({ method: 'POST' })
  .inputValidator(registrationSaveSchema)
  .handler(async ({ data }) => {
    const { saveRegistration } = await import('./registration.server')
    return saveRegistration(data)
  })

/** The signed-in viewer `viewerKey` names's registration for the contest. */
export const registrationQuery = (contestId: string, viewerKey: string) => queryOptions({
  queryKey: ['registration', contestId, viewerKey],
  queryFn: (): Promise<RegistrationRead> => getRegistration({ data: contestId }),
  staleTime: 30_000,
  retry: false,
})

/** Every language a contest without a language restriction accepts. */
export const contestLanguagesQuery = () => queryOptions({
  queryKey: ['contest-languages'],
  queryFn: (): Promise<Read<ImmersionLanguage[]>> => getContestLanguages(),
  staleTime: 300_000,
  retry: false,
})

/**
 * Registers the account that opened the form, or changes its languages. A save refreshes what the registration
 * changes: the registration itself, the contest's board and standings and the activity form's contests.
 */
export function useSaveRegistration() {
  const queryClient = useQueryClient()

  return async (save: RegistrationSave): Promise<RegistrationResult> => {
    const result = await putRegistration({ data: save })
      .catch((): RegistrationResult => ({ ok: false, reason: 'unavailable' }))

    if (result.ok) {
      await Promise.all([
        // Read again now, so the contest page the form returns to shows the saved registration.
        queryClient.invalidateQueries({ queryKey: ['registration', save.contestId], refetchType: 'all' }),
        queryClient.invalidateQueries({ queryKey: ['contest-board', save.contestId], refetchType: 'all' }),
        queryClient.invalidateQueries({ queryKey: ['standings', { kind: 'contest', contestId: save.contestId }] }),
        queryClient.invalidateQueries({ queryKey: ['log-options'] }),
      ])
    } else if (result.reason === 'switched' || result.reason === 'signed-out') {
      await queryClient.invalidateQueries({ queryKey: ['viewer'] })
    }
    return result
  }
}

/** The viewer and whether they have joined the contest. A registration that can't be read counts as not joined, as in Webv2. */
export function useParticipation(contestId: string) {
  const { data: viewer } = useQuery(viewerQueryOptions())
  const registration = useQuery({
    ...registrationQuery(contestId, viewerKey(viewer)),
    enabled: viewer?.status === 'signed-in',
  })
  return { viewer, registration: registration.data, registered: registration.data?.kind === 'registered' }
}
