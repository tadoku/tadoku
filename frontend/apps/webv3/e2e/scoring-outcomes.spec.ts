import type { APIRequestContext } from '@playwright/test'
import { activityEntries, admin, api, expect, reader, test, webv3 } from './logging-fixtures'
import { ruleSets } from './scoring-rule-sets'
import { scoringRound } from './seeded'
import { apiJson, score, standingRows } from './standings'

interface Log { id: string; score: number; registrations?: { contest_id: string; score?: number }[] }

async function saveLog(request: APIRequestContext, data: Record<string, unknown>) {
  const response = await request.post(`${api}/immersion/logs`, { data: { language_code: 'jpn', tags: [], ...data } })
  expect(response.ok(), `log answered ${response.status()}: ${await response.text()}`).toBeTruthy()
  const { id } = await response.json() as { id: string }
  return apiJson<Log>(request, `/immersion/logs/${id}`)
}

async function standing(request: APIRequestContext, contestId: string, userId: string) {
  const entry = await request.get(`${api}/immersion/contests/${contestId}/leaderboard/${userId}`)
  return entry.status() === 404 ? 0 : (await entry.json() as { score: number }).score
}

test('a contest’s own rules score real logs: the base rule, a stacked modifier and the platform fallback', async ({ returningLearner: learner, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and both fixture identities')
  test.setTimeout(60_000)
  const round = scoringRound()
  const { request } = learner.context

  // The tenant hook seeds the round with an active override pinned to the platform's active version.
  const [contestRules] = await ruleSets(adminContext.request, `/immersion/contests/${round.contestId}/scoring/rule-sets`)
  const platform = (await ruleSets(adminContext.request, '/immersion/scoring/rule-sets')).find(ruleSet => ruleSet.active)
  expect(contestRules).toMatchObject({ id: round.ruleSetId, mode: 'override', active: true, fallback_rule_set_id: platform?.id })

  const joined = await request.post(`${api}/immersion/contests/${round.contestId}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.status(), 'registration').toBe(200)
  const { id: registrationId } = await apiJson<{ id: string }>(request, `/immersion/contests/${round.contestId}/registration`)
  const before = await standing(request, round.contestId, learner.id)

  // The engine scores each input by the round's rules: the base rule, the stacked modifier and, for listening, which no
  // rule covers, the pinned platform version.
  const inputs = [
    { activity_id: 1, amount: 10, unit_key: 'pages', tags: [], description: 'Scoring base' },
    { activity_id: 1, amount: 10, unit_key: 'pages', tags: [round.bonusTag], description: 'Scoring modifier' },
    { activity_id: 2, duration_seconds: 1_200, tags: [], description: 'Scoring fallback' },
  ]
  const previews = await Promise.all(inputs.map(async input => {
    const response = await request.post(`${api}/immersion/logs/score-preview`, {
      data: { language_code: 'jpn', ...input, registration_ids: [registrationId] },
    })
    expect(response.status(), 'score preview').toBe(200)
    return response.json() as Promise<{ platform: { score: number }; contests: { contest_id: string; estimate: { score: number } }[] }>
  }))
  const ruled = previews.map(preview => preview.contests.find(contest => contest.contest_id === round.contestId)!.estimate.score)
  expect(ruled[0]).toBeCloseTo(10 * round.readingRate, 4)
  expect(ruled[1]).toBeCloseTo(10 * round.readingRate * round.bonusRate, 4)
  expect(previews[2].platform.score, 'the platform scores listening').toBeGreaterThan(0)
  expect(ruled[2]).toBeCloseTo(previews[2].platform.score, 4)

  // Saved scores follow the rules only where the engine is authoritative; in shadow mode, as on development clusters,
  // logs and their submissions keep the legacy score and the engine's answer is only previewed.
  const { scoring_engine_enabled: authoritative } = await apiJson<{ scoring_engine_enabled: boolean }>(request, '/immersion/logs/configuration-options')
  testInfo.annotations.push({ type: 'scoring engine', description: authoritative ? 'authoritative' : 'shadow' })
  const logs = []
  for (const input of inputs) logs.push(await saveLog(request, { ...input, registration_ids: [registrationId] }))
  const expected = new Map(logs.map((log, index) => [log.id, authoritative ? ruled[index] : log.score]))
  for (const log of logs) {
    const saved = log.registrations?.find(registration => registration.contest_id === round.contestId)?.score
    expect(saved, `log ${log.id}`).toBeCloseTo(expected.get(log.id)!, 4)
  }

  const total = before + [...expected.values()].reduce((sum, value) => sum + value, 0)
  await expect.poll(() => standing(request, round.contestId, learner.id)).toBeCloseTo(total, 1)
  const page = await learner.context.newPage()
  await page.goto(`/contests/${round.contestId}/leaderboard/1`)
  const row = standingRows(page).filter({ has: page.getByRole('link', { name: learner.name, exact: true }) })
  await expect(row.getByRole('cell').last()).toHaveText(score.format(total))
  await page.screenshot({ path: testInfo.outputPath('standings.png'), fullPage: true })

  // The participant's contest profile lists each log with the points it earned in this contest.
  await page.goto(`/contests/${round.contestId}/profile/${learner.id}/1`)
  for (const [id, points] of expected) {
    const entry = activityEntries(page).filter({ has: page.locator(`[href="/logs/${id}"]`) })
    await expect(entry, `log ${id}`).toContainText(score.format(points))
  }
  await page.screenshot({ path: testInfo.outputPath('contest-profile.png'), fullPage: true })
})
