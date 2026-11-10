import type { BrowserContext, Page } from '@playwright/test'
import {
  admin,
  amountInput,
  api,
  contestStepUrl,
  expect,
  joinedContest,
  logLinks,
  logOut,
  savedLogId,
  saveButton,
  signIn,
  statValue,
  test,
  webv3,
} from './logging-fixtures'
import { score, standingRows } from './standings'

/** The log's score in the contest, as the API saved it. */
async function contestScore(context: BrowserContext, logId: string, contestId: string) {
  const { registrations = [] } = await (await context.request.get(`${api}/immersion/logs/${logId}`)).json() as {
    registrations?: { contest_id: string; score?: number }[]
  }
  return registrations.find(registration => registration.contest_id === contestId)?.score
}

/** The learner's standings row agrees with the API's entry: its score, or no row once nothing counts. */
async function expectStanding(page: Page, context: BrowserContext, contestId: string, learner: { id: string; name: string }) {
  await page.goto(`/contests/${contestId}/leaderboard/1`)
  const entry = await context.request.get(`${api}/immersion/contests/${contestId}/leaderboard/${learner.id}`)
  const row = standingRows(page).filter({ has: page.getByRole('link', { name: learner.name, exact: true }) })
  if (entry.status() === 404) {
    await expect(row).toHaveCount(0)
    return undefined
  }
  const { score: points } = await entry.json() as { score: number }
  await expect(row.getByRole('cell').last()).toHaveText(score.format(points))
  return points
}

test('a learner logs in, logs, submits, sees it in profile and standings, edits and deletes it, and logs out', async ({
  browser,
  learner,
  adminContext,
}, testInfo) => {
  test.skip(!webv3 || !api || !admin.email, 'Set development URLs and the Dev Admin fixture account')
  test.setTimeout(120_000)
  const { id: contestId, title } = await joinedContest(adminContext, learner.context)
  const context = await browser.newContext({ baseURL: webv3 })
  const page = await context.newPage()
  const shot = (name: string) => page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true })

  // Login through authv2 returns to the activity form.
  await signIn(page, learner)
  await expect(page.getByRole('button', { name: learner.name, exact: true })).toBeVisible()

  const description = `release-journey-${Date.now()}`
  await page.getByLabel('Language', { exact: true }).selectOption('jpn')
  await page.getByLabel('Activity', { exact: true }).selectOption({ label: 'Reading' })
  await amountInput(page).fill('7')
  await page.getByLabel(/^Description/).fill(description)
  await saveButton(page).click()

  // A learner with an ongoing registration reaches the contest step with that contest chosen.
  await page.waitForURL(contestStepUrl)
  const logId = savedLogId(page)
  const contest = page.getByRole('checkbox', { name: title })
  await expect(contest).toBeChecked({ timeout: 15_000 })
  await page.getByRole('button', { name: 'Save submissions', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/logs/${logId}`)
  const submitted = await contestScore(context, logId, contestId)
  expect(submitted, 'the API counts the log in the contest').toBeGreaterThan(0)
  const contests = page.getByRole('region', { name: 'Contests' })
  await expect(contests.getByRole('link', { name: title, exact: true })).toBeVisible()
  await expect(contests).toContainText(`${score.format(submitted!)} points`)
  await shot('1-submitted')

  await page.getByRole('button', { name: learner.name, exact: true }).click()
  await page.getByRole('menuitem', { name: 'Profile', exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${learner.id}/statistics`))
  await expect(statValue(page, 'Overall score')).toContainText(score.format(submitted!))
  await expect(statValue(page, 'Logs')).toContainText('1')
  await page.goto(`/profile/${learner.id}/updates/1`)
  await expect(logLinks(page)).toHaveCount(1)
  await expect(logLinks(page)).toHaveAttribute('href', `/logs/${logId}`)
  await shot('2-profile')

  expect(await expectStanding(page, context, contestId, learner)).toBe(submitted)
  await shot('3-standings')

  await page.goto(`/logs/${logId}`)
  await page.getByRole('link', { name: 'Edit log', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Save', exact: true })).toBeEnabled({ timeout: 15_000 })
  await amountInput(page).fill('12')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/logs/${logId}`)
  const edited = await contestScore(context, logId, contestId)
  expect(edited, 'the edit rescores the ongoing contest').toBeGreaterThan(submitted!)
  expect(await expectStanding(page, context, contestId, learner)).toBe(edited)
  await shot('4-edited-standings')

  await page.goto(`/logs/${logId}`)
  await page.getByRole('button', { name: 'Delete log', exact: true }).click()
  await page.getByRole('dialog', { name: 'Delete this log?' }).getByRole('button', { name: 'Delete log', exact: true }).click()
  await page.waitForURL(new RegExp(`/profile/${learner.id}/statistics`))
  expect((await context.request.get(`${api}/immersion/logs/${logId}`)).status()).toBe(404)
  expect(await expectStanding(page, context, contestId, learner) ?? 0, 'the deleted log no longer counts').toBe(0)
  await shot('5-deleted-standings')

  const cookies = await context.cookies()
  await logOut(page, learner)
  await expect(page.getByRole('banner').getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
  await shot('6-logged-out')
  const stale = await browser.newContext()
  await stale.addCookies(cookies)
  const session = await stale.request.get(`${api}/identity/session`)
  expect(session.ok() && Boolean((await session.json()).account), 'the API no longer accepts the old session').toBe(false)
  await stale.close()
  await context.close()
})
