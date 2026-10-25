import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestRegistration, ImmersionContestView, ImmersionLogConfigurationOptions } from 'api-client'
import { z } from 'zod'

export type { ImmersionLog } from 'api-client'
export type Configuration = ImmersionLogConfigurationOptions
/** A registration the form can offer: the contract leaves its ID and contest optional because request bodies share it. */
export type Registration = ImmersionContestRegistration & { id: string; contest: ImmersionContestView }

/**
 * A tag as the API stores it, after Go's strings.TrimSpace and strings.ToLower: Unicode white space is trimmed and each
 * character is lowercased on its own, so 'İ' becomes 'i' and 'Σ' always becomes 'σ', unlike String.toLowerCase.
 */
function apiTag(tag: string) {
  return [...tag.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, '')]
    .map(character => String.fromCodePoint(character.toLowerCase().codePointAt(0)!))
    .join('')
}

export const logForm = z.object({
  languageCode: z.string().length(3, 'Choose a language.'), activityId: z.int(),
  measure: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('amount'), amount: z.number({ error: 'Enter an amount.' }).positive('Enter an amount greater than zero.'), unitId: z.guid('Choose a unit.') }),
    z.object({ kind: z.literal('time'), minutes: z.number({ error: 'Enter minutes.' }).positive('Enter minutes greater than zero.').max(1440, 'Enter at most 1440 minutes.').refine(value => Math.round(value * 60) > 0, 'Enter at least one second.') }),
  ]),
  // The API normalizes and deduplicates tags, then counts them and each tag's UTF-8 bytes.
  tags: z.array(z.string())
    .transform(tags => [...new Set(tags.map(apiTag).filter(Boolean))])
    .pipe(z.array(z.string().refine(
      tag => new TextEncoder().encode(tag).length <= 50,
      'Use shorter tags. A tag allows 50 bytes, and most non-Latin characters take 2 to 4 bytes.',
    )).max(10, 'Use at most 10 different tags.')),
  // The API counts description characters as Unicode code points.
  description: z.string().trim().refine(value => [...value].length <= 255, 'Use a description of at most 255 characters.'),
  registrationIds: z.array(z.guid()),
})
export type LogFormValue = z.infer<typeof logForm>
export type Failure = 'invalid' | 'signed-out' | 'forbidden' | 'deleting' | 'unavailable' | 'not-found' | 'switched' | 'account-unavailable'
export type Result<T> = { ok: true; data: T } | { ok: false; reason: Failure }

export const getLogFormData = createServerFn({ method: 'GET' }).handler(async () => {
  const { readLogFormData } = await import('./logs.server')
  return readLogFormData()
})
export const getLog = createServerFn({ method: 'GET' }).inputValidator(z.guid()).handler(async ({ data }) => {
  const { readLog } = await import('./logs.server')
  return readLog(data)
})
export const createLog = createServerFn({ method: 'POST' }).inputValidator(logForm.extend({ viewerId: z.guid() })).handler(async ({ data }) => {
  const { saveLog } = await import('./logs.server')
  return saveLog(data)
})
export const formQueryOptions = (viewerId: string) => queryOptions({ queryKey: ['log-form', viewerId], queryFn: () => getLogFormData(), retry: false, staleTime: Infinity })
export const logQueryOptions = (id: string, viewerId: string) => queryOptions({ queryKey: ['log', id, viewerId], queryFn: () => getLog({ data: id }), retry: false, staleTime: 30_000 })

export function unitsFor(configuration: Configuration, activityId: number, languageCode: string) {
  const units = configuration.units.filter(unit => unit.log_activity_id === activityId && !['reading_two_column_page', 'reading_comic_page'].includes(unit.unit_key))
  return units.filter(unit => unit.language_code === languageCode || (!unit.language_code && !units.some(other => other.name === unit.name && other.language_code === languageCode)))
}
export const eligibleRegistrations = (registrations: Registration[], languageCode: string, activityId: number) => registrations.filter(registration =>
  registration.languages.some(language => language.code === languageCode) && registration.contest.allowed_activities.some(activity => activity.id === activityId),
)
