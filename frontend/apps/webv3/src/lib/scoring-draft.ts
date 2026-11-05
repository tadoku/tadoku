import type { ImmersionLogConfigurationOptions, ImmersionScoringRuleSet } from 'api-client'
import { z } from 'zod'
import { normalizeTag, tagRule } from './log-draft'

/** One ordered rule in the form. Empty unit, language and tag matchers match any input. */
export type RuleDraft = {
  activityId: string
  scoreSource: string
  unitKey: string
  languageCode: string
  tag: string
  rate: number
  stackable: boolean
}

/** What the new rule-set form holds. The pinned platform version applies to override mode only. */
export type ScoringDraft = { mode: string; fallbackRuleSetId: string; rules: RuleDraft[] }

export const scoreSources = [
  { value: 'amount', label: 'Amount' },
  { value: 'duration_minutes', label: 'Duration (minutes)' },
] as const

const rateRule = z.number({ error: 'Enter a rate.' }).min(0, 'Enter a rate of zero or more.')

const ruleSchema = z.object({
  stackable: z.boolean(),
  activityId: z.int().positive(),
  unitKey: z.string(),
  languageCode: z.string(),
  tag: z.string(),
  scoreSource: z.enum(['amount', 'duration_minutes']),
  rate: rateRule,
})

/**
 * A contest rule-set draft as the API receives it, rules in priority order. Override mode falls back to a pinned
 * published platform version; replace mode has no fallback.
 */
export const newRuleSetSchema = z.object({
  viewerId: z.guid(),
  contestId: z.guid(),
  configuration: z.discriminatedUnion('mode', [
    z.object({ mode: z.literal('override'), fallbackRuleSetId: z.guid() }),
    z.object({ mode: z.literal('replace') }),
  ]),
  rules: z.array(ruleSchema),
})

export type NewRuleSet = z.infer<typeof newRuleSetSchema>

/** Published platform versions, which a contest's override can pin. */
export const publishedPlatformRuleSets = (ruleSets: ImmersionScoringRuleSet[]) =>
  ruleSets.filter(ruleSet => ruleSet.status === 'published')

export function emptyRule(options: ImmersionLogConfigurationOptions): RuleDraft {
  return {
    activityId: String(options.activities[0]?.id ?? 1),
    scoreSource: 'amount',
    unitKey: '',
    languageCode: '',
    tag: '',
    rate: 1,
    stackable: false,
  }
}

/** A new draft overrides platform rules, pinning the active published version, with one rule. */
export function firstScoringDraft(
  options: ImmersionLogConfigurationOptions,
  platform: ImmersionScoringRuleSet[],
): ScoringDraft {
  const published = publishedPlatformRuleSets(platform)
  const pinned = published.find(ruleSet => ruleSet.active) ?? published[0]
  return { mode: 'override', fallbackRuleSetId: pinned?.id ?? '', rules: [emptyRule(options)] }
}

/** The activity's distinct unit keys, which a rule's unit matcher can name. */
export function ruleUnits(options: ImmersionLogConfigurationOptions, activityId: string) {
  const units = options.units.filter(unit => String(unit.log_activity_id) === activityId)
  return units.filter((unit, index) => units.findIndex(other => other.unit_key === unit.unit_key) === index)
}

/** Validates the form's values with the API's rules, reporting each problem on the field that holds it. */
export const scoringDraftSchema = z.custom<ScoringDraft>().superRefine((draft, context) => {
  if (draft.mode === 'override' && !draft.fallbackRuleSetId) {
    context.addIssue({ code: 'custom', path: ['fallbackRuleSetId'], message: 'Choose the pinned platform version.' })
  }
  draft.rules.forEach((rule, index) => {
    const rate = rateRule.safeParse(rule.rate)
    if (!rate.success) {
      context.addIssue({ code: 'custom', path: ['rules', index, 'rate'], message: rate.error.issues[0].message })
    }
    const tag = normalizeTag(rule.tag)
    const tagged = tagRule.safeParse(tag)
    if (tag && !tagged.success) {
      context.addIssue({ code: 'custom', path: ['rules', index, 'tag'], message: tagged.error.issues[0].message })
    }
  })
})

/** The draft to create from a valid form, with priorities spaced by 10 in the rules' order. */
export function toNewRuleSet(draft: ScoringDraft, viewerId: string, contestId: string): NewRuleSet {
  return {
    viewerId,
    contestId,
    configuration: draft.mode === 'override'
      ? { mode: 'override', fallbackRuleSetId: draft.fallbackRuleSetId }
      : { mode: 'replace' },
    rules: draft.rules.map(rule => ({
      stackable: rule.stackable,
      activityId: Number(rule.activityId),
      unitKey: rule.unitKey,
      languageCode: rule.languageCode,
      tag: normalizeTag(rule.tag),
      scoreSource: rule.scoreSource === 'duration_minutes' ? 'duration_minutes' : 'amount',
      rate: rule.rate,
    })),
  }
}
