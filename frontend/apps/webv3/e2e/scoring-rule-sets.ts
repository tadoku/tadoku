import type { APIRequestContext } from '@playwright/test'
import { api, expect } from './logging-fixtures'

export interface RuleSet {
  id: string
  version: number
  status: 'draft' | 'published'
  active: boolean
  mode?: 'override' | 'replace'
  fallback_rule_set_id?: string
  rules: { priority: number; stackable: boolean; activity_id: number; tag?: string; score_source: string; rate: number }[]
}

export async function ruleSets(request: APIRequestContext, path: string) {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), `${path} answered ${response.status()}`).toBeTruthy()
  return (await response.json() as { rule_sets: RuleSet[] }).rule_sets
}

/** The platform rule set a contest's override falls back to, publishing one first on a tenant without any. */
export async function platformFallback(request: APIRequestContext) {
  const platform = (await ruleSets(request, '/immersion/scoring/rule-sets')).filter(ruleSet => ruleSet.status === 'published')
  const existing = platform.find(ruleSet => ruleSet.active) ?? platform[0]
  if (existing) return existing

  const draft = await request.post(`${api}/immersion/scoring/rule-sets`, {
    data: { rules: [{ priority: 0, stackable: false, activity_id: 1, unit_key: 'pages', score_source: 'amount', rate: 1 }] },
  })
  expect(draft.ok(), `rule-set creation answered ${draft.status()}`).toBeTruthy()
  const ruleSet = await draft.json() as RuleSet
  const published = await request.post(`${api}/immersion/scoring/rule-sets/${ruleSet.id}/publish`)
  expect(published.ok(), `rule-set publication answered ${published.status()}`).toBeTruthy()
  return ruleSet
}
