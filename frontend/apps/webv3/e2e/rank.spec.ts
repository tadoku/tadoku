import type { BrowserContext } from '@playwright/test'
import { api, expect, reader, test, webv3 } from './logging-fixtures'

const ordinal = new Intl.PluralRules('en-US', { type: 'ordinal' })
const suffix: Record<string, string> = { one: 'st', two: 'nd', few: 'rd', other: 'th' }
const place = (rank: number) => `${rank}${suffix[ordinal.select(rank)]}`

/** Joins the latest official round and logs a little to it, which ranks the learner below the seeded field. */
async function joinWithSmallLog(context: BrowserContext) {
  const latest = await (await context.request.get(`${api}/immersion/contests/latest-official`)).json() as { id: string; title: string }
  const registered = await context.request.post(`${api}/immersion/contests/${latest.id}/registration`, { data: { language_codes: ['jpn'] } })
  expect(registered.status(), 'registration').toBe(200)
  const { id: registrationId } = await (await context.request.get(`${api}/immersion/contests/${latest.id}/registration`)).json() as { id: string }

  const saved = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount: 1, unit_key: 'pages', tags: [], description: 'Below the first page', registration_ids: [registrationId] },
  })
  expect(saved.ok(), `log answered ${saved.status()}`).toBeTruthy()
  return latest
}

test('a participant ranked below the first standings page sees their rank on home and on the participant page', async ({ learner }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  test.setTimeout(60_000)
  const { context, id: userId } = learner
  const contest = await joinWithSmallLog(context)

  // The API's answer, whichever rank the field above leaves the learner, must be beyond the first page of 50.
  await expect.poll(async () => (await context.request.get(`${api}/immersion/contests/${contest.id}/leaderboard/${userId}`)).status(), { timeout: 20_000 }).toBe(200)
  const { rank, is_tie: tie } = await (await context.request.get(`${api}/immersion/contests/${contest.id}/leaderboard/${userId}`)).json() as { rank: number; is_tie: boolean }
  expect(rank, 'the seeded field ranks above the learner').toBeGreaterThan(50)

  const page = await context.newPage()
  await page.goto('/')
  await expect.soft(page.getByRole('heading', { level: 1 })).toHaveText(`${tie ? 'Tied ' : ''}${place(rank)} place`)
  await page.screenshot({ path: testInfo.outputPath('home-rank.png'), fullPage: true })

  await page.goto(`/contests/${contest.id}/profile/${userId}/1`)
  await expect.soft(page.getByText(new RegExp(`^${place(rank)} · `))).toBeVisible()
  await page.screenshot({ path: testInfo.outputPath('participant-rank.png'), fullPage: true })
})
