import { z } from 'zod'
import type { ImmersionLog, LogOptions, Registration } from './logs'

/**
 * What the activity form holds. Amount activities use `amount`, `unitKey` and optionally `minutes`, timed ones
 * `minutes`; `toNewLog` and `toLogUpdate` send only the measure that applies.
 */
export type LogDraft = {
  languageCode: string
  activityId: string
  amount: number
  unitKey: string
  minutes: number
  description: string
  tags: string[]
}

const amountRule = z.number({ error: 'Enter an amount.' }).positive('Enter an amount greater than zero.')
const unitRule = z.string().min(1, 'Choose a unit.')
const minutesRule = z.number({ error: 'Enter minutes.' })
  .positive('Enter minutes greater than zero.')
  .max(1440, 'Enter at most 1440 minutes.')
  .refine(minutes => Math.round(minutes * 60) > 0, 'Enter at least one second.')

// The API counts description characters as Unicode code points.
const descriptionRule = z.string()
  .refine(description => [...description].length <= 255, 'Use a description of at most 255 characters.')

// The API counts each normalized tag's UTF-8 bytes, then the distinct tags.
export const tagRule = z.string()
  .min(1)
  .refine(
    tag => new TextEncoder().encode(tag).length <= 50,
    'Use shorter tags. A tag allows 50 bytes, and most non-Latin characters take 2 to 4 bytes.',
  )
const tagsRule = z.array(tagRule).max(10, 'Use at most 10 different tags.')
const secondsRule = z.int().positive().max(86_400)

/** The activity a viewer saves, as the API receives it. An amount may also record the time spent. */
export const newLogSchema = z.object({
  viewerId: z.guid(),
  languageCode: z.string().length(3),
  activityId: z.int().positive(),
  measure: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('amount'), amount: amountRule, unitKey: unitRule, seconds: secondsRule.optional() }),
    z.object({ kind: z.literal('duration'), seconds: secondsRule }),
  ]),
  description: descriptionRule,
  tags: tagsRule,
})

export type NewLog = z.infer<typeof newLogSchema>

/** A change to a saved log. Its language, activity and contest submissions stay as they are. */
export const logUpdateSchema = newLogSchema
  .pick({ viewerId: true, measure: true, description: true, tags: true })
  .extend({ logId: z.guid() })

export type LogUpdate = z.infer<typeof logUpdateSchema>

/** The registrations a saved log is submitted to among the viewer's ongoing ones. */
export const logContestsUpdateSchema = z.object({ viewerId: z.guid(), logId: z.guid(), registrationIds: z.array(z.guid()) })

export type LogContestsUpdate = z.infer<typeof logContestsUpdateSchema>

/** What the API needs to estimate a draft's score. */
export const scorePreviewSchema = newLogSchema.pick({ languageCode: true, activityId: true, measure: true, tags: true })

export type ScorePreviewInput = z.infer<typeof scorePreviewSchema>

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
      check('amount', unitRule, draft.unitKey)
      if (!Number.isNaN(draft.minutes)) check('minutes', minutesRule, draft.minutes)
    }
    check('description', descriptionRule, draft.description.trim())
    check('tags', tagsRule, parseTags(draft.tags))
  })
}

/** The language and activity of the viewer's last saved log, as this browser remembers them. */
export type LastLogged = { languageCode: string | null; activityId: string | null }

/**
 * A new draft: the last logged language and activity when they are still offered, otherwise the viewer's first own
 * language and the first activity; the activity's first unit and `amount` if given. As in Webv2, a remembered timed
 * activity starts without an amount.
 */
export function firstDraft(options: LogOptions, amount = Number.NaN, last?: LastLogged): LogDraft {
  const { languages, activities, user_language_codes: ownLanguages = [] } = options.configuration
  const offered = (code: string | null | undefined) => languages.some(language => language.code === code)
  const languageCode = (offered(last?.languageCode) ? last?.languageCode : ownLanguages.find(offered)) ?? ''
  const remembered = activities.find(activity => String(activity.id) === last?.activityId)
  const activityId = String((remembered ?? activities[0]).id)
  const timed = isTimed(options, activityId)

  return {
    languageCode,
    activityId,
    amount: remembered && timed ? Number.NaN : amount,
    unitKey: timed ? '' : unitFor(options, languageCode, activityId),
    minutes: Number.NaN,
    description: '',
    tags: [],
  }
}

/** The unit to select again after the language or activity changes. */
export function unitFor(options: LogOptions, languageCode: string, activityId: string) {
  return unitsFor(options, languageCode, activityId)[0]?.unit_key ?? ''
}

/**
 * The draft of a saved log. Timed logs saved before durations were recorded hold their minutes as the amount; they are
 * edited and saved as a duration.
 */
export function draftFromLog(options: LogOptions, log: ImmersionLog): LogDraft {
  const languageCode = log.language.code
  const activityId = String(log.activity.id)
  const timed = isTimed(options, activityId)
  const seconds = log.duration_seconds ?? (timed ? log.amount * 60 : undefined)

  return {
    languageCode,
    activityId,
    amount: timed ? Number.NaN : log.amount,
    unitKey: (timed ? undefined : log.unit_key) ?? unitsFor(options, languageCode, activityId)[0]?.unit_key ?? '',
    minutes: seconds === undefined ? Number.NaN : seconds / 60,
    description: log.description ?? '',
    tags: log.tags,
  }
}

export function toNewLog(draft: LogDraft, options: LogOptions, viewerId: string): NewLog {
  return { ...toScorePreview(draft, options), viewerId, description: draft.description.trim() }
}

export function toLogUpdate(draft: LogDraft, options: LogOptions, viewerId: string, log: ImmersionLog): LogUpdate {
  const { measure, tags } = toScorePreview(draft, options)
  return { viewerId, logId: log.id, measure, description: draft.description.trim(), tags }
}

/** The estimate request for a draft that passes `logDraftSchema`. */
export function toScorePreview(draft: LogDraft, options: LogOptions): ScorePreviewInput {
  const seconds = Math.round(draft.minutes * 60)

  return {
    languageCode: draft.languageCode,
    activityId: Number(draft.activityId),
    measure: isTimed(options, draft.activityId)
      ? { kind: 'duration', seconds }
      : { kind: 'amount', amount: draft.amount, unitKey: draft.unitKey, ...(Number.isNaN(seconds) ? {} : { seconds }) },
    tags: parseTags(draft.tags),
  }
}

export function isTimed(options: LogOptions, activityId: string) {
  const activity = options.configuration.activities.find(activity => String(activity.id) === activityId)
  return activity?.input_type === 'time_primary'
}

/** The activity's units, where a unit specific to the language replaces the general unit with the same key. */
export function unitsFor(options: LogOptions, languageCode: string, activityId: string) {
  const units = options.configuration.units.filter(unit => String(unit.log_activity_id) === activityId)
  const specific = new Set(units.filter(unit => unit.language_code === languageCode).map(unit => unit.unit_key))

  return units.filter(unit => unit.language_code === languageCode || (!unit.language_code && !specific.has(unit.unit_key)))
}

/** An ongoing registration a saved log can be submitted to, or why it cannot. */
export type ContestChoice =
  | { kind: 'eligible'; registration: Registration }
  | { kind: 'ineligible'; registration: Registration; reason: string }

/**
 * Webv2's rules for submitting a saved log, in its order. `today` is the UTC date: a contest accepts submissions through
 * the end of its last UTC day.
 */
export function contestChoices(log: ImmersionLog, registrations: Registration[], today: string): ContestChoice[] {
  return registrations.map(registration => {
    const reason = !registration.contest.allowed_activities.some(activity => activity.id === log.activity.id)
      ? 'Activity not allowed'
      : !registration.languages.some(language => language.code === log.language.code)
        ? 'Language not registered'
        : registration.contest.contest_end < today ? 'Ended' : undefined

    return reason ? { kind: 'ineligible', registration, reason } : { kind: 'eligible', registration }
  })
}

/**
 * A tag as the API stores it. Go's strings.TrimSpace and strings.ToLower trim Unicode white space and lowercase each
 * character on its own, so 'İ' becomes 'i' and 'Σ' always becomes 'σ', unlike String.toLowerCase.
 */
export function normalizeTag(tag: string) {
  return [...tag.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, '')]
    .map(character => String.fromCodePoint(character.toLowerCase().codePointAt(0)!))
    .join('')
}

/** Tags as the API stores them. */
export function parseTags(tags: string[]) {
  return [...new Set(tags.map(normalizeTag).filter(Boolean))]
}
