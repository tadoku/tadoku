import { queryOptions, useQueryClient } from '@tanstack/react-query'
import { createServerFn } from '@tanstack/react-start'
import type { ImmersionContestView, ImmersionLogConfigurationOptions, ImmersionScoringRuleSet } from 'api-client'
import { z } from 'zod'
import type { LogResult } from './logs'
import { newRuleSetSchema, type NewRuleSet } from './scoring-draft'
import { viewerQueryOptions } from './session'

/** What the scoring page shows an administrator: the contest, its rule-set versions and what new rules can match. */
export type ContestScoring = {
  contest: ImmersionContestView
  options: ImmersionLogConfigurationOptions
  platformRuleSets: ImmersionScoringRuleSet[]
  contestRuleSets: ImmersionScoringRuleSet[]
}

/** Publishing makes a draft immutable; activating applies a published version to new contest submissions. */
export type RuleSetChange = { viewerId: string; contestId: string; ruleSetId: string; action: 'publish' | 'activate' }

const getContestScoring = createServerFn({ method: 'GET' })
  .inputValidator(z.guid())
  .handler(async ({ data }) => {
    const { readContestScoring } = await import('./scoring.server')
    return readContestScoring(data)
  })

const createRuleSet = createServerFn({ method: 'POST' })
  .inputValidator(newRuleSetSchema)
  .handler(async ({ data }) => {
    const { saveRuleSet } = await import('./scoring.server')
    return saveRuleSet(data)
  })

const changeRuleSet = createServerFn({ method: 'POST' })
  .inputValidator(z.object({
    viewerId: z.guid(),
    contestId: z.guid(),
    ruleSetId: z.guid(),
    action: z.enum(['publish', 'activate']),
  }))
  .handler(async ({ data }) => {
    const { updateRuleSet } = await import('./scoring.server')
    return updateRuleSet(data)
  })

/** A contest's scoring as one administrator reads it. */
export const contestScoringQuery = (contestId: string, viewerId: string) => queryOptions({
  queryKey: ['contest-scoring', contestId, viewerId],
  queryFn: (): Promise<LogResult<ContestScoring>> => getContestScoring({ data: contestId }),
  retry: false,
  staleTime: 30_000,
})

/**
 * Creates a draft, or publishes or activates a version, for the administrator who opened the page, then reads the
 * contest's versions again. A refusal because another account signed in refreshes the viewer. A failed request is
 * `unavailable`, because the change may or may not have been made.
 */
export function useChangeScoring() {
  const queryClient = useQueryClient()

  return async (change: { create: NewRuleSet } | { update: RuleSetChange }): Promise<LogResult<null>> => {
    const { viewerId, contestId } = 'create' in change ? change.create : change.update
    const result = await ('create' in change ? createRuleSet({ data: change.create }) : changeRuleSet({ data: change.update }))
      .catch((): LogResult<null> => ({ ok: false, reason: 'unavailable' }))

    if (!result.ok && result.reason === 'switched') {
      await queryClient.invalidateQueries({ queryKey: viewerQueryOptions().queryKey })
    } else {
      await queryClient.invalidateQueries({ queryKey: contestScoringQuery(contestId, viewerId).queryKey })
    }
    return result
  }
}
