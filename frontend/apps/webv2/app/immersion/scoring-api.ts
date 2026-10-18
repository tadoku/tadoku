import { z } from 'zod'
import { browserAppUrls } from 'ui/app-urls'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

const root = () => `${browserAppUrls().apiEndpoint}/immersion`

export const Score = z.object({
  language_code: z.string(),
  language_name: z.string().optional(),
  score: z.number(),
})

export type Score = z.infer<typeof Score>

export type ScorePreviewPayload = {
  registration_ids?: string[]
  language_code: string
  activity_id: number
  amount?: number
  unit_id?: string
  unit_key?: string
  duration_seconds?: number
  tags: string[]
}

const AppliedScoringRule = z.object({
  rule_id: z.string(),
  rate: z.number(),
})

const ScoreEstimate = z.object({
  score: z.number(),
  source: z.enum(['amount', 'duration_minutes']),
  rule_set_id: z.string().optional(),
  rules: z.array(AppliedScoringRule),
})

const ScorePreview = z.object({
  platform: ScoreEstimate,
  contests: z.array(
    z.object({
      registration_id: z.string(),
      contest_id: z.string(),
      estimate: ScoreEstimate,
    }),
  ),
})

export type ScorePreview = z.infer<typeof ScorePreview>

export const useScorePreview = (
  payload: ScorePreviewPayload | undefined,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['log', 'score-preview', payload],
    async (): Promise<ScorePreview> => {
      const response = await fetch(`${root()}/logs/score-preview`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return ScorePreview.parse(await response.json())
    },
    {
      ...options,
      enabled: payload !== undefined && options?.enabled !== false,
      keepPreviousData: true,
    },
  )

const ScoringRule = z.object({
  id: z.string().optional(),
  priority: z.number(),
  stackable: z.boolean(),
  activity_id: z.number(),
  unit_key: z.string().optional(),
  language_code: z.string().optional(),
  tag: z.string().optional(),
  score_source: z.enum(['amount', 'duration_minutes']),
  rate: z.number(),
})

export type ScoringRule = z.infer<typeof ScoringRule>

const ScoringRuleSet = z.object({
  id: z.string(),
  scope: z.enum(['platform', 'contest']),
  contest_id: z.string().optional(),
  version: z.number(),
  status: z.enum(['draft', 'published']),
  active: z.boolean(),
  mode: z.enum(['override', 'replace']).optional(),
  fallback_rule_set_id: z.string().optional(),
  rules: z.array(ScoringRule),
  created_at: z.string(),
  published_at: z.string().optional(),
})

export type ScoringRuleSet = z.infer<typeof ScoringRuleSet>

const ScoringRuleSets = z.object({
  rule_sets: z.array(ScoringRuleSet),
})

export type ScoringRuleSetDraftPayload = {
  mode: 'override' | 'replace'
  fallback_rule_set_id?: string
  rules: Omit<ScoringRule, 'id'>[]
}

export const usePlatformScoringRuleSets = (options?: { enabled?: boolean }) =>
  useQuery(
    ['scoring', 'rule-sets', 'platform'],
    async (): Promise<ScoringRuleSet[]> => {
      const response = await fetch(`${root()}/scoring/rule-sets`)
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return ScoringRuleSets.parse(await response.json()).rule_sets
    },
    options,
  )

export const useContestScoringRuleSets = (
  contestId: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['scoring', 'rule-sets', 'contest', contestId],
    async (): Promise<ScoringRuleSet[]> => {
      const response = await fetch(
        `${root()}/contests/${contestId}/scoring/rule-sets`,
      )
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return ScoringRuleSets.parse(await response.json()).rule_sets
    },
    { ...options, enabled: !!contestId && options?.enabled !== false },
  )

export const useCreateContestScoringRuleSet = (contestId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (payload: ScoringRuleSetDraftPayload) => {
      const response = await fetch(
        `${root()}/contests/${contestId}/scoring/rule-sets`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        },
      )
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return ScoringRuleSet.parse(await response.json())
    },
    onSuccess() {
      queryClient.invalidateQueries([
        'scoring',
        'rule-sets',
        'contest',
        contestId,
      ])
    },
  })
}

export const usePublishScoringRuleSet = (contestId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (ruleSetId: string) => {
      const response = await fetch(
        `${root()}/scoring/rule-sets/${ruleSetId}/publish`,
        { method: 'POST' },
      )
      if (response.status !== 200) {
        throw new Error(response.status.toString())
      }
      return ScoringRuleSet.parse(await response.json())
    },
    onSuccess() {
      queryClient.invalidateQueries([
        'scoring',
        'rule-sets',
        'contest',
        contestId,
      ])
    },
  })
}

export const useActivateScoringRuleSet = (contestId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (ruleSetId: string) => {
      const response = await fetch(
        `${root()}/scoring/rule-sets/${ruleSetId}/activate`,
        { method: 'POST' },
      )
      if (response.status !== 204) {
        throw new Error(response.status.toString())
      }
    },
    onSuccess() {
      queryClient.invalidateQueries([
        'scoring',
        'rule-sets',
        'contest',
        contestId,
      ])
    },
  })
}
