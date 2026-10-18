import { z } from 'zod'
import { configurationSchema, logSchema, registrationsSchema } from './logs'
import type { Configuration, ImmersionLog, LogFormValue, Registration, Result } from './logs'
import { readViewer, sessionCookie } from './session.server'
import { requestUrls } from './urls.server'

async function api<T>(path: string, schema: z.ZodType<T>, body?: unknown, cookie = sessionCookie()): Promise<Result<T>> {
  if (body && !cookie) return { ok: false, reason: 'signed-out' }
  try {
    const response = await fetch(`${requestUrls().apiUrl}/immersion/${path}`, {
      method: body ? 'POST' : 'GET',
      headers: { Accept: 'application/json', ...(cookie ? { Cookie: cookie } : {}), ...(body ? { 'Content-Type': 'application/json' } : {}) },
      ...(body ? { body: JSON.stringify(body) } : {}),
      redirect: 'error', signal: AbortSignal.timeout(10_000),
    })
    if (response.ok) return { ok: true, data: schema.parse(await response.json()) }
    void response.body?.cancel()
    switch (response.status) {
      case 400: return { ok: false, reason: 'invalid' }
      case 401: return { ok: false, reason: 'signed-out' }
      case 403: return { ok: false, reason: 'forbidden' }
      case 404: return { ok: false, reason: 'not-found' }
      case 409: return { ok: false, reason: 'deleting' }
      default: return { ok: false, reason: 'unavailable' }
    }
  } catch {
    return { ok: false, reason: 'unavailable' }
  }
}

export async function readLogFormData(): Promise<Result<{ configuration: Configuration; registrations: Registration[] }>> {
  if (!sessionCookie()) return { ok: false, reason: 'signed-out' }
  const [configuration, registrations] = await Promise.all([
    api('logs/configuration-options', configurationSchema),
    api('contests/ongoing-registrations', registrationsSchema),
  ])
  if (!configuration.ok) return configuration
  if (!registrations.ok) return registrations
  return { ok: true, data: { configuration: configuration.data, registrations: registrations.data.registrations } }
}

export async function readLog(id: string) {
  const result = await api(`logs/${id}`, logSchema)
  return !result.ok && result.reason === 'signed-out' && sessionCookie()
    ? api(`logs/${id}`, logSchema, undefined, '')
    : result
}

export async function saveLog(value: LogFormValue & { viewerId: string }): Promise<Result<ImmersionLog>> {
  const viewer = await readViewer()
  if (viewer.status === 'anonymous') return { ok: false, reason: 'signed-out' }
  if (viewer.status === 'unavailable') return { ok: false, reason: 'account-unavailable' }
  if (viewer.status === 'banned') return { ok: false, reason: 'forbidden' }
  if (viewer.userId !== value.viewerId) return { ok: false, reason: 'switched' }
  return api('logs', logSchema, {
    language_code: value.languageCode, activity_id: value.activityId,
    tags: value.tags, description: value.description, registration_ids: value.registrationIds,
    ...(value.measure.kind === 'amount'
      ? { amount: value.measure.amount, unit_id: value.measure.unitId }
      : { duration_seconds: Math.round(value.measure.minutes * 60) }),
  })
}
