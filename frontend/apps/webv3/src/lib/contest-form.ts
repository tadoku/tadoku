import type { ImmersionActivity, ImmersionContestConfigurationOptions, ImmersionLanguage } from 'api-client'
import { z } from 'zod'

/** What the contest form holds. `official` applies only when the viewer may create official rounds. */
export type ContestDraft = {
  title: string
  description: string
  contestStart: string
  contestEnd: string
  registrationEnd: string
  languages: ImmersionLanguage[]
  activities: ImmersionActivity[]
  official: boolean
  unlisted: boolean
}

const codePoints = (text: string) => [...text].length
const date = z.iso.date({ error: 'Choose a date.' })

/**
 * The contest a viewer creates, as the API receives it. Official rounds are public and allow every language; only
 * community contests can be unlisted or limit languages.
 */
export const newContestSchema = z.object({
  viewerId: z.guid(),
  // The API requires more than three characters, and the database stores at most 255.
  title: z.string()
    .refine(title => codePoints(title) > 3, 'Use a title of at least 4 characters.')
    .refine(title => codePoints(title) <= 255, 'Use a title of at most 255 characters.'),
  description: z.string()
    .refine(description => codePoints(description) <= 500, 'Use a description of at most 500 characters.')
    .optional(),
  contestStart: date,
  contestEnd: date,
  registrationEnd: date,
  activityIds: z.array(z.int().positive()).min(1, 'Choose at least one activity.'),
  kind: z.discriminatedUnion('kind', [
    z.object({ kind: z.literal('official') }),
    z.object({ kind: z.literal('community'), unlisted: z.boolean(), languageCodes: z.array(z.string().min(1)) }),
  ]),
})

export type NewContest = z.infer<typeof newContestSchema>

/** A new draft: no dates or languages, and the activities the API marks as defaults. */
export function firstContestDraft(options: ImmersionContestConfigurationOptions): ContestDraft {
  return {
    title: '',
    description: '',
    contestStart: '',
    contestEnd: '',
    registrationEnd: '',
    languages: [],
    activities: options.activities.filter(activity => activity.default),
    official: false,
    unlisted: false,
  }
}

/** Whether the draft is an official round, which the viewer may create only when the API offers it. */
export const isOfficial = (draft: ContestDraft, options: ImmersionContestConfigurationOptions) =>
  options.can_create_official_round && draft.official

/**
 * Validates the form's values with the API's rules, reporting each problem on the field that holds it. Contests by
 * non-administrators must not start or end before the current UTC day.
 */
export function contestDraftSchema(options: ImmersionContestConfigurationOptions, admin: boolean, today: string) {
  const shape = newContestSchema.shape
  return z.custom<ContestDraft>().superRefine((draft, context) => {
    const check = (field: keyof ContestDraft, rule: z.ZodType, value: unknown) => {
      const result = rule.safeParse(value)
      if (!result.success) context.addIssue({ code: 'custom', path: [field], message: result.error.issues[0].message })
      return result.success
    }
    const issue = (field: keyof ContestDraft, message: string) => context.addIssue({ code: 'custom', path: [field], message })

    check('title', shape.title, draft.title)
    check('description', shape.description, draft.description.trim())
    const start = check('contestStart', date, draft.contestStart)
    const end = check('contestEnd', date, draft.contestEnd)
    check('registrationEnd', date, draft.registrationEnd)
    if (start && !admin && draft.contestStart < today) issue('contestStart', 'Choose today or a later date.')
    if (end && !admin && draft.contestEnd < today) issue('contestEnd', 'Choose today or a later date.')
    if (start && end && draft.contestEnd < draft.contestStart) {
      issue('contestEnd', 'Choose an end date on or after the start date.')
    }
    check('activities', shape.activityIds, draft.activities.map(activity => activity.id))
    if (isOfficial(draft, options) && draft.languages.length) {
      issue('languages', 'Official rounds allow every language. Remove the selected languages.')
    }
  })
}

/** The contest to create from a valid draft. */
export function toNewContest(draft: ContestDraft, options: ImmersionContestConfigurationOptions, viewerId: string): NewContest {
  const description = draft.description.trim()
  return {
    viewerId,
    title: draft.title,
    ...(description ? { description } : {}),
    contestStart: draft.contestStart,
    contestEnd: draft.contestEnd,
    registrationEnd: draft.registrationEnd,
    activityIds: draft.activities.map(activity => activity.id),
    kind: isOfficial(draft, options)
      ? { kind: 'official' }
      : { kind: 'community', unlisted: draft.unlisted, languageCodes: draft.languages.map(language => language.code) },
  }
}
