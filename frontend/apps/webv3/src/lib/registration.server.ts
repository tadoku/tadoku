import {
  immersionContestFindRegistration,
  immersionContestGetConfigurations,
  immersionContestRegistrationUpsert,
  type ImmersionLanguage,
} from 'api-client'
import type { Read } from './leaderboard'
import type { RegistrationFailure, RegistrationRead, RegistrationResult, RegistrationSave } from './registration'
import { readAccount, readViewer, sessionCookie, sessionRequest } from './session.server'

const failed = { data: undefined, error: undefined, response: undefined }

/** The visitor's registration: 204 means they haven't joined. Visitors without a session have none to read. */
export async function readRegistration(contestId: string): Promise<RegistrationRead> {
  const cookie = sessionCookie()
  if (!cookie) return { kind: 'unavailable' }

  const { data, response } = await immersionContestFindRegistration({ ...sessionRequest('immersionContestFindRegistration', cookie), path: { id: contestId } })
    .catch(() => failed)
  if (response?.status === 204) return { kind: 'not-registered' }
  return data ? { kind: 'registered', languages: data.languages } : { kind: 'unavailable' }
}

export async function readContestLanguages(): Promise<Read<ImmersionLanguage[]>> {
  const { data } = await immersionContestGetConfigurations(sessionRequest('immersionContestGetConfigurations', sessionCookie())).catch(() => failed)
  return data?.languages.length ? { ok: true, data: data.languages } : { ok: false }
}

const failures: Partial<Record<number, RegistrationFailure>> = {
  400: 'invalid',
  401: 'signed-out',
  403: 'closed',
  404: 'not-found',
  409: 'deleting',
}

/** Saves the registration for the account that opened the form, refusing it when another account or none is signed in. */
export async function saveRegistration(save: RegistrationSave): Promise<RegistrationResult> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const account = await readAccount(cookie).catch(() => ({ status: 'unavailable' as const }))
  switch (account.status) {
    case 'anonymous': return { ok: false, reason: 'signed-out' }
    case 'unavailable': return { ok: false, reason: 'account-unavailable' }
  }
  if (account.userId !== save.viewerId) return { ok: false, reason: 'switched' }

  const { error, response } = await immersionContestRegistrationUpsert({
    ...sessionRequest('immersionContestRegistrationUpsert', cookie),
    path: { id: save.contestId },
    body: { language_codes: save.languageCodes },
  }).catch(() => failed)
  if (response?.status === 200 && error === undefined) return { ok: true }

  // The API answers a banned account and a closed contest with the same 403, so only a refusal costs a role read.
  if (response?.status === 403 && (await readViewer()).status === 'banned') return { ok: false, reason: 'suspended' }
  return { ok: false, reason: (response && failures[response.status]) || 'unavailable' }
}
