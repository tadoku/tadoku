import { z } from 'zod'
import type { LogOptions, Registration } from './logs'

/**
 * What the activity form holds. Amount activities use `amount` and `unitId`, timed ones `minutes`; `toNewLog` sends
 * only the measure that applies.
 */
export type LogDraft = {
  languageCode: string
  activityId: string
  amount: number
  unitId: string
  minutes: number
  description: string
  tags: string
  registrationIds: string[]
}

const amountRule = z.number({ error: 'Enter an amount.' }).positive('Enter an amount greater than zero.')
const unitRule = z.guid('Choose a unit.')
const minutesRule = z.number({ error: 'Enter minutes.' })
  .positive('Enter minutes greater than zero.')
  .max(1440, 'Enter at most 1440 minutes.')
  .refine(minutes => Math.round(minutes * 60) > 0, 'Enter at least one second.')

// The API counts description characters as Unicode code points.
const descriptionRule = z.string()
  .refine(description => [...description].length <= 255, 'Use a description of at most 255 characters.')

// The API counts each normalized tag's UTF-8 bytes, then the distinct tags.
const tagRule = z.string()
  .min(1)
  .refine(
    tag => new TextEncoder().encode(tag).length <= 50,
    'Use shorter tags. A tag allows 50 bytes, and most non-Latin characters take 2 to 4 bytes.',
  )
const tagsRule = z.array(tagRule).max(10, 'Use at most 10 different tags.')

/** The activity a viewer saves, as the API receives it. */
export const newLogSchema = z.object({
  viewerId: z.guid(),
  languageCode: z.string().length(3),
  activityId: z.int().positive(),
  measure: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('amount'), amount: amountRule, unitId: unitRule }),
    z.object({ kind: z.literal('duration'), seconds: z.int().positive().max(86_400) }),
  ]),
  description: descriptionRule,
  tags: tagsRule,
  registrationIds: z.array(z.guid()),
})

export type NewLog = z.infer<typeof newLogSchema>

/**
 * Validates the form's values with the API's rules, reporting each problem on the field that holds it. Only the
 * measure that applies to the chosen activity is checked.
 */
export function logDraftSchema(options: LogOptions) {
  return z.custom<LogDraft>().superRefine((draft, context) => {
    const check = (field: keyof LogDraft, rule: z.ZodType, value: unknown) => {
      const result = rule.safeParse(value)
      if (!result.success) context.addIssue({ code: 'custom', path: [field], message: result.error.issues[0].message })
    }

    if (isTimed(options, draft.activityId)) {
      check('minutes', minutesRule, draft.minutes)
    } else {
      check('amount', amountRule, draft.amount)
      // AmountWithUnit shows one error for the amount and its unit, so a missing unit is reported on the amount.
      check('amount', unitRule, draft.unitId)
    }
    check('description', descriptionRule, draft.description.trim())
    check('tags', tagsRule, parseTags(draft.tags))
  })
}

/** A new draft: the viewer's first own language, the first activity, its first unit and every eligible contest. */
export function firstDraft(options: LogOptions): LogDraft {
  const { languages, activities, user_language_codes: ownLanguages = [] } = options.configuration
  const offered = (code: string) => languages.some(language => language.code === code)
  const languageCode = ownLanguages.find(offered) ?? languages[0].code
  const activityId = String(activities[0].id)

  return {
    ...dependentFields(options, languageCode, activityId),
    languageCode,
    activityId,
    amount: Number.NaN,
    minutes: Number.NaN,
    description: '',
    tags: '',
  }
}

/** The unit and contests to select again after the language or activity changes. */
export function dependentFields(options: LogOptions, languageCode: string, activityId: string) {
  return {
    unitId: unitsFor(options, languageCode, activityId)[0]?.id ?? '',
    registrationIds: eligibleRegistrations(options, languageCode, activityId).map(registration => registration.id),
  }
}

export function toNewLog(draft: LogDraft, options: LogOptions, viewerId: string): NewLog {
  const selected = eligibleRegistrations(options, draft.languageCode, draft.activityId)
    .filter(registration => draft.registrationIds.includes(registration.id))

  return {
    viewerId,
    languageCode: draft.languageCode,
    activityId: Number(draft.activityId),
    measure: isTimed(options, draft.activityId)
      ? { kind: 'duration', seconds: Math.round(draft.minutes * 60) }
      : { kind: 'amount', amount: draft.amount, unitId: draft.unitId },
    description: draft.description.trim(),
    tags: parseTags(draft.tags),
    registrationIds: selected.map(registration => registration.id),
  }
}

export function isTimed(options: LogOptions, activityId: string) {
  const activity = options.configuration.activities.find(activity => String(activity.id) === activityId)
  return activity?.input_type === 'time_primary'
}

// Two-column and comic pages are legacy reading units; Webv2's log form records them as tag modifiers instead.
const legacyUnitKeys = new Set(['reading_two_column_page', 'reading_comic_page'])

/** The activity's units, where a unit specific to the language replaces the general unit of the same name. */
export function unitsFor(options: LogOptions, languageCode: string, activityId: string) {
  const units = options.configuration.units.filter(unit =>
    String(unit.log_activity_id) === activityId && !legacyUnitKeys.has(unit.unit_key))
  const specific = new Set(units.filter(unit => unit.language_code === languageCode).map(unit => unit.name))

  return units.filter(unit => unit.language_code === languageCode || (!unit.language_code && !specific.has(unit.name)))
}

/** The viewer's ongoing registrations whose contest accepts this language and activity. */
export function eligibleRegistrations(options: LogOptions, languageCode: string, activityId: string): Registration[] {
  return options.registrations.filter(registration =>
    registration.languages.some(language => language.code === languageCode) &&
    registration.contest.allowed_activities.some(activity => String(activity.id) === activityId))
}

/**
 * Comma-separated tags as the API stores them. Go's strings.TrimSpace and strings.ToLower trim Unicode white space and
 * lowercase each character on its own, so 'İ' becomes 'i' and 'Σ' always becomes 'σ', unlike String.toLowerCase.
 */
function parseTags(text: string) {
  const tags = text.split(',').map(tag =>
    [...tag.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, '')]
      .map(character => String.fromCodePoint(character.toLowerCase().codePointAt(0)!))
      .join(''))

  return [...new Set(tags.filter(Boolean))]
}
