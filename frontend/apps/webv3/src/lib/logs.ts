import { queryOptions } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import { z } from 'zod'

const language = z.object({ code: z.string().length(3), name: z.string() })
const activity = z.object({ id: z.int(), name: z.string(), input_type: z.enum(['amount_primary', 'time_primary']) })
export const configurationSchema = z.object({
  languages: z.array(language).min(1), activities: z.array(activity).min(1),
  units: z.array(z.object({ id: z.uuid(), unit_key: z.string(), log_activity_id: z.int(), name: z.string(), modifier: z.number(), language_code: z.string().optional() })),
  user_language_codes: z.array(z.string()).optional(), scoring_engine_enabled: z.boolean(),
})
export const registrationsSchema = z.object({ registrations: z.array(z.object({
  id: z.uuid(), languages: z.array(language),
  contest: z.object({ title: z.string(), allowed_activities: z.array(activity) }),
})) })
export const logSchema = z.object({
  id: z.uuid(), user_id: z.uuid(), user_display_name: z.string(),
  description: z.string().optional(), language, activity,
  amount: z.number(), unit_id: z.uuid(), unit_name: z.string(), modifier: z.number(),
  duration_seconds: z.number().int().positive().optional(), score: z.number(),
  tags: z.array(z.string()), created_at: z.string().datetime(), deleted: z.boolean(),
  registrations: z.array(z.object({ contest_id: z.uuid(), title: z.string(), score: z.number().optional() })),
})
export type ImmersionLog = z.infer<typeof logSchema>
export type Configuration = z.infer<typeof configurationSchema>
export type Registration = z.infer<typeof registrationsSchema>['registrations'][number]

export const logForm = z.object({
  languageCode: z.string().length(3, 'Choose a language.'), activityId: z.int(),
  measure: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('amount'), amount: z.number({ error: 'Enter an amount.' }).positive('Enter an amount greater than zero.'), unitId: z.uuid('Choose a unit.') }),
    z.object({ kind: z.literal('time'), minutes: z.number({ error: 'Enter minutes.' }).positive('Enter minutes greater than zero.').max(1440, 'Enter at most 1440 minutes.').refine(value => Math.round(value * 60) > 0, 'Enter at least one second.') }),
  ]),
  tags: z.array(z.string().trim().min(1).max(50, 'Use tags of at most 50 characters.')).max(10, 'Use at most 10 tags.'),
  description: z.string().trim().max(255, 'Use a description of at most 255 characters.'),
  registrationIds: z.array(z.uuid()),
})
export type LogFormValue = z.infer<typeof logForm>
export type Failure = 'invalid' | 'signed-out' | 'forbidden' | 'deleting' | 'unavailable' | 'not-found' | 'switched'
export type Result<T> = { ok: true; data: T } | { ok: false; reason: Failure }

export const getLogFormData = createServerFn({ method: 'GET' }).handler(async () => {
  const { readLogFormData } = await import('./logs.server')
  return readLogFormData()
})
export const getLog = createServerFn({ method: 'GET' }).inputValidator(z.uuid()).handler(async ({ data }) => {
  const { readLog } = await import('./logs.server')
  return readLog(data)
})
export const createLog = createServerFn({ method: 'POST' }).inputValidator(logForm.extend({ viewerId: z.uuid() })).handler(async ({ data }) => {
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
