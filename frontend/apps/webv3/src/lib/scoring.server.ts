import {
  immersionContestFindById,
  immersionLogGetConfigurations,
  immersionScoringRuleSetActivate,
  immersionScoringRuleSetCreateContest,
  immersionScoringRuleSetListContest,
  immersionScoringRuleSetListPlatform,
  immersionScoringRuleSetPublish,
} from 'api-client'
import type { LogResult } from './logs'
import { answer, draftOwnerCookie } from './logs.server'
import type { ContestScoring, RuleSetChange } from './scoring'
import type { NewRuleSet } from './scoring-draft'
import { sessionCookie, sessionRequest } from './session.server'

export async function readContestScoring(contestId: string): Promise<LogResult<ContestScoring>> {
  const cookie = sessionCookie()
  if (!cookie) return { ok: false, reason: 'signed-out' }

  const request = sessionRequest(cookie)
  const [contest, options, platform, ruleSets] = await Promise.all([
    answer(immersionContestFindById({ ...request, path: { id: contestId } })),
    answer(immersionLogGetConfigurations(request)),
    answer(immersionScoringRuleSetListPlatform(request)),
    answer(immersionScoringRuleSetListContest({ ...request, path: { id: contestId } })),
  ])
  if (!contest.ok) return contest
  if (!options.ok) return options
  if (!platform.ok) return platform
  if (!ruleSets.ok) return ruleSets
  return {
    ok: true,
    data: {
      contest: contest.data,
      options: options.data,
      platformRuleSets: platform.data.rule_sets,
      contestRuleSets: ruleSets.data.rule_sets,
    },
  }
}

/** Creates a contest rule-set draft. Empty matchers match any input, so they are left out. */
export async function saveRuleSet(draft: NewRuleSet): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(draft.viewerId, { admin: true })
  if (!session.ok) return session

  const { configuration } = draft
  const result = await answer(immersionScoringRuleSetCreateContest({
    ...sessionRequest(session.data),
    path: { id: draft.contestId },
    body: {
      mode: configuration.mode,
      ...(configuration.mode === 'override' ? { fallback_rule_set_id: configuration.fallbackRuleSetId } : {}),
      rules: draft.rules.map((rule, index) => ({
        priority: index * 10,
        stackable: rule.stackable,
        activity_id: rule.activityId,
        ...(rule.unitKey ? { unit_key: rule.unitKey } : {}),
        ...(rule.languageCode ? { language_code: rule.languageCode } : {}),
        ...(rule.tag ? { tag: rule.tag } : {}),
        score_source: rule.scoreSource,
        rate: rule.rate,
      })),
    },
  }))
  return result.ok ? { ok: true, data: null } : result
}

export async function updateRuleSet(change: RuleSetChange): Promise<LogResult<null>> {
  const session = await draftOwnerCookie(change.viewerId, { admin: true })
  if (!session.ok) return session

  const request = { ...sessionRequest(session.data), path: { id: change.ruleSetId } }
  const result = change.action === 'publish'
    ? await answer(immersionScoringRuleSetPublish(request))
    : await answer(immersionScoringRuleSetActivate(request))
  return result.ok ? { ok: true, data: null } : result
}
