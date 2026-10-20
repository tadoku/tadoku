import {
  immersionScorePreview,
  immersionScoringRuleSetActivate,
  immersionScoringRuleSetCreateContest,
  immersionScoringRuleSetListContest,
  immersionScoringRuleSetListPlatform,
  immersionScoringRuleSetPublish,
  type ImmersionScore,
  type ImmersionScorePreview,
  type ImmersionScorePreviewData,
  type ImmersionScoringRule,
  type ImmersionScoringRuleSet,
  type ImmersionScoringRuleSetDraft,
} from 'api-client'
import { apiData, browserApi } from 'ui/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

export type Score = ImmersionScore
export type ScorePreviewPayload = ImmersionScorePreviewData['body']
export type ScorePreview = ImmersionScorePreview

export const useScorePreview = (
  payload: ScorePreviewPayload | undefined,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['log', 'score-preview', payload],
    async (): Promise<ScorePreview> =>
      apiData(immersionScorePreview({ ...browserApi(), body: payload! })),
    {
      ...options,
      enabled: payload !== undefined && options?.enabled !== false,
      keepPreviousData: true,
    },
  )

export type ScoringRule = ImmersionScoringRule
export type ScoringRuleSet = ImmersionScoringRuleSet
export type ScoringRuleSetDraftPayload = ImmersionScoringRuleSetDraft

export const usePlatformScoringRuleSets = (options?: { enabled?: boolean }) =>
  useQuery(
    ['scoring', 'rule-sets', 'platform'],
    async (): Promise<ScoringRuleSet[]> =>
      (await apiData(immersionScoringRuleSetListPlatform(browserApi())))
        .rule_sets,
    options,
  )

export const useContestScoringRuleSets = (
  contestId: string,
  options?: { enabled?: boolean },
) =>
  useQuery(
    ['scoring', 'rule-sets', 'contest', contestId],
    async (): Promise<ScoringRuleSet[]> =>
      (
        await apiData(
          immersionScoringRuleSetListContest({
            ...browserApi(),
            path: { id: contestId },
          }),
        )
      ).rule_sets,
    { ...options, enabled: !!contestId && options?.enabled !== false },
  )

const useInvalidateContestRuleSets = (contestId: string) => {
  const queryClient = useQueryClient()
  return () =>
    queryClient.invalidateQueries(['scoring', 'rule-sets', 'contest', contestId])
}

export const useCreateContestScoringRuleSet = (contestId: string) => {
  const invalidate = useInvalidateContestRuleSets(contestId)
  return useMutation({
    mutationFn: async (payload: ScoringRuleSetDraftPayload) =>
      apiData(
        immersionScoringRuleSetCreateContest({
          ...browserApi(),
          path: { id: contestId },
          body: payload,
        }),
      ),
    onSuccess: invalidate,
  })
}

export const usePublishScoringRuleSet = (contestId: string) => {
  const invalidate = useInvalidateContestRuleSets(contestId)
  return useMutation({
    mutationFn: async (ruleSetId: string) =>
      apiData(
        immersionScoringRuleSetPublish({
          ...browserApi(),
          path: { id: ruleSetId },
        }),
      ),
    onSuccess: invalidate,
  })
}

export const useActivateScoringRuleSet = (contestId: string) => {
  const invalidate = useInvalidateContestRuleSets(contestId)
  return useMutation({
    mutationFn: async (ruleSetId: string) => {
      await apiData(
        immersionScoringRuleSetActivate({
          ...browserApi(),
          path: { id: ruleSetId },
        }),
      )
    },
    onSuccess: invalidate,
  })
}
