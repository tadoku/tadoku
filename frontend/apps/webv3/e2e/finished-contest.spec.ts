import { randomUUID } from 'node:crypto'
import type { APIRequestContext } from '@playwright/test'
import { accountId, admin, amountInput, api, expect, reader, test, webv3 } from './logging-fixtures'
import { finishedRound } from './seeded'
import { apiJson, score, standingRows } from './standings'

interface Submission { contest_id: string; title: string; score?: number }

const utcDay = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
const submissions = async (request: APIRequestContext, logId: string) =>
  (await apiJson<{ registrations?: Submission[] }>(request, `/immersion/logs/${logId}`)).registrations ?? []

/** Creates a community contest as Dev Admin, whom the API lets date contests in the past. */
async function contest(request: APIRequestContext, start: string, end: string, registrationEnd: string) {
  const response = await request.post(`${api}/immersion/contests`, {
    data: {
      title: `Webv3 final day ${randomUUID().slice(0, 8)}`,
      description: 'Created by the Webv3 finished-contest journey.',
      contest_start: start,
      contest_end: end,
      registration_end: registrationEnd,
      private: true,
      official: false,
      language_code_allow_list: ['jpn'],
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<{ id: string; title: string }>
}

test.beforeEach(() => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and both fixture identities')
})

test('a finished contest keeps its score after an edit, refuses deletion and keeps the submission', async ({ page, readerContext }, testInfo) => {
  const round = finishedRound()
  const { request } = readerContext
  const { contest_end: end } = await apiJson<{ contest_end: string }>(request, `/immersion/contests/${round.contestId}`)
  expect(end < utcDay(), 'the seeded round has finished').toBe(true)
  const standings = async () => {
    await page.goto(`/contests/${round.contestId}/leaderboard/1`)
    await expect(standingRows(page).filter({ hasText: reader.name }).getByRole('cell').last()).toHaveText(score.format(round.score))
  }
  await standings()

  // An edit rescores the personal record only; the finished contest keeps the score it saved.
  const amount = String(11 + Math.floor(Math.random() * 80))
  await page.goto(`/logs/${round.logId}/edit`)
  await expect(page.getByRole('button', { name: 'Save', exact: true })).toBeEnabled({ timeout: 15_000 })
  await amountInput(page).fill(amount)
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/logs/${round.logId}`)
  await expect(page.getByText(`${amount} pages`, { exact: true })).toBeVisible()
  const contests = page.getByRole('region', { name: 'Contests' })
  await expect(contests).toContainText(`${score.format(round.score)} points`)
  expect(await submissions(request, round.logId)).toEqual([expect.objectContaining({ contest_id: round.contestId, score: round.score })])
  await standings()
  await page.screenshot({ path: testInfo.outputPath('frozen-score.png'), fullPage: true })

  await page.goto(`/logs/${round.logId}`)
  await page.getByRole('button', { name: 'Delete log', exact: true }).click()
  await page.getByRole('dialog', { name: 'Delete this log?' }).getByRole('button', { name: 'Delete log', exact: true }).click()
  await expect(page.getByText('This log cannot be deleted. A log submitted to a finished contest stays in its results.')).toBeVisible()
  await expect(page).toHaveURL(`${webv3}/logs/${round.logId}`)
  expect((await request.get(`${api}/immersion/logs/${round.logId}`)).status()).toBe(200)
  await page.screenshot({ path: testInfo.outputPath('deletion-refused.png'), fullPage: true })

  // Saving the contest step without choices keeps the finished submission and its saved score.
  await page.goto(`/logs/${round.logId}/contests`)
  await expect(page.getByText(`Dev Finished Round has finished. Its saved score of ${round.score} points stays.`)).toBeVisible()
  await expect(page.getByRole('checkbox', { name: 'Dev Finished Round' })).toHaveCount(0)
  const save = page.getByRole('button', { name: 'Save submissions', exact: true })
  if (await save.count()) {
    await save.click()
    await expect(page).toHaveURL(`${webv3}/logs/${round.logId}`)
  }
  expect(await submissions(request, round.logId)).toEqual([expect.objectContaining({ contest_id: round.contestId, score: round.score })])
  await standings()
})

test('a contest takes logs through its last UTC day and nothing after it', async ({ page, readerContext, adminContext }, testInfo) => {
  const { request } = readerContext
  const userId = await accountId(readerContext)
  const lastDay = await contest(adminContext.request, utcDay(-3), utcDay(), utcDay())
  const ended = await contest(adminContext.request, utcDay(-3), utcDay(-1), utcDay())

  const joined = await request.post(`${api}/immersion/contests/${lastDay.id}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.status(), 'registration on the last day').toBe(200)
  const refused = await request.post(`${api}/immersion/contests/${ended.id}/registration`, { data: { language_codes: ['jpn'] } })
  expect(refused.status(), 'registration after the last day').toBe(403)

  // On its last day the contest is offered to a new log, and the log counts in its standings.
  await page.goto('/logs/new')
  await expect(page.getByRole('button', { name: 'Log activity', exact: true })).toBeEnabled({ timeout: 15_000 })
  await page.getByLabel('Language', { exact: true }).selectOption('jpn')
  await page.getByLabel('Activity', { exact: true }).selectOption({ label: 'Reading' })
  await amountInput(page).fill('4')
  await page.getByLabel(/^Description/).fill(`final-day-${Date.now()}`)
  await page.getByRole('button', { name: 'Log activity', exact: true }).click()
  await page.waitForURL(/\/logs\/[a-f0-9-]{36}\/contests\?preselect=1$/)
  const logId = new URL(page.url()).pathname.split('/')[2]
  const choice = page.getByRole('checkbox', { name: lastDay.title })
  await expect(choice).toBeEnabled({ timeout: 15_000 })
  // Dev Reader's other ongoing contests stay out of this log.
  const checked = page.getByRole('checkbox', { checked: true })
  while (await checked.count()) await checked.first().uncheck()
  await choice.check()
  await page.getByRole('button', { name: 'Save submissions', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/logs/${logId}`)
  const [submission] = await submissions(request, logId)
  expect(submission).toMatchObject({ contest_id: lastDay.id })

  await page.goto(`/contests/${lastDay.id}/leaderboard/1`)
  await expect(standingRows(page).filter({ hasText: reader.name }).getByRole('cell').last()).toHaveText(score.format(submission.score!))
  await page.screenshot({ path: testInfo.outputPath('last-day.png'), fullPage: true })

  // The day after its end the contest page offers no way in.
  await page.goto(`/contests/${ended.id}/leaderboard/1`)
  await expect(page.getByRole('heading', { level: 1, name: ended.title })).toBeVisible()
  await expect(page.getByRole('main').getByRole('link', { name: 'Join contest', exact: true })).toHaveCount(0)
  expect([204, 404], 'not on the board').toContain((await request.get(`${api}/immersion/contests/${ended.id}/leaderboard/${userId}`)).status())
  await page.screenshot({ path: testInfo.outputPath('ended.png'), fullPage: true })
})
