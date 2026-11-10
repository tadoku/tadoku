import { randomUUID } from 'node:crypto'
import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { accountId, admin, amountInput, api, contestStepUrl, expect, fillDraft, hydrated, linkTo, reader, saveButton, savedLogId, savedLogUrl, statValue, test, webv3 } from './logging-fixtures'
import { apiJson, score, standingRows } from './standings'

interface Contest { id: string; title: string }
interface Log { id: string; score: number }

const year = new Date().getUTCFullYear()
const day = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
const tag = () => randomUUID().slice(0, 8)
const soon = { timeout: 5_000 }

/** Counts full document loads from now on, so a test can prove it only ever moved through the router. */
function documentLoads(page: Page) {
  let loads = 0
  page.on('request', request => {
    if (request.isNavigationRequest() && request.frame() === page.mainFrame()) loads++
  })
  return () => loads
}

/** An open community contest, created through the API by an account old enough to create one. */
async function createContest(context: BrowserContext): Promise<Contest> {
  const title = `Webv3 freshness check ${tag()}`
  const response = await context.request.post(`${api}/immersion/contests`, {
    data: {
      title,
      description: 'Created by the Webv3 mutation freshness journey.',
      contest_start: day(),
      contest_end: day(30),
      registration_end: day(14),
      private: false,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return { ...(await response.json()), title }
}

/** Registers the context's account for the contest and submits one reading log to it through the API. */
async function submitLog(context: BrowserContext, contestId: string, amount: number) {
  const joined = await context.request.post(`${api}/immersion/contests/${contestId}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.status(), 'registration').toBe(200)
  const registration = await apiJson<{ id: string }>(context.request, `/immersion/contests/${contestId}/registration`)
  const created = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount, unit_key: 'pages', tags: [], description: `freshness-${tag()}`, registration_ids: [registration.id] },
  })
  expect(created.ok(), `log answered ${created.status()}`).toBeTruthy()
  return await created.json() as Log
}

async function yearTotal(request: APIRequestContext, userId: string) {
  const { scores } = await apiJson<{ scores: { score: number }[] }>(request, `/immersion/users/${userId}/scores/${year}`)
  return scores.reduce((sum, item) => sum + item.score, 0)
}

const contestScore = async (request: APIRequestContext, contestId: string, userId: string) =>
  (await apiJson<{ score: number }>(request, `/immersion/contests/${contestId}/leaderboard/${userId}`)).score

const banner = (page: Page, name: string) => page.getByRole('banner').getByRole('link', { name, exact: true })
const contestViews = (page: Page, name: string) => page.getByRole('navigation', { name: 'Contest views' }).getByRole('link', { name, exact: true })
const submissions = (page: Page) => page.getByRole('region', { name: 'Contests' })
const overallScore = (page: Page) => statValue(page, 'Overall score')
const rowScore = (page: Page, name: string) => standingRows(page).filter({ hasText: name }).getByRole('cell').last()
const feedEntry = (page: Page, logId: string) => linkTo(page, `/logs/${logId}`)

test.beforeEach(() => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and both fixture identities')
  test.setTimeout(120_000)
})

test('saving a log and submitting it refreshes standings, the feed, home and profile without a reload', async ({ readerContext, adminContext }) => {
  const contest = await createContest(adminContext)
  const readerId = await accountId(readerContext)
  const seed = await submitLog(readerContext, contest.id, 2)
  const before = await contestScore(readerContext.request, contest.id, readerId)

  const page = await readerContext.newPage()
  await page.goto(`/logs/${seed.id}`)
  await hydrated(page)
  const loads = documentLoads(page)

  // Each page's answer enters the client's cache.
  await submissions(page).getByRole('link', { name: contest.title, exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contest.id}/leaderboard/1$`))
  await expect(rowScore(page, reader.name)).toHaveText(score.format(before))
  await contestViews(page, 'Activity').click()
  await expect(feedEntry(page, seed.id)).toBeVisible()
  await banner(page, 'Home').click()
  await expect(overallScore(page)).toBeVisible()
  await page.getByRole('link', { name: 'Your profile', exact: true }).click()
  await expect(page).toHaveURL(/\/statistics$/)
  await expect(overallScore(page)).toBeVisible()

  await banner(page, 'Log activity').click()
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
  await page.getByLabel('Language', { exact: true }).selectOption('jpn')
  await page.getByLabel('Activity', { exact: true }).selectOption({ label: 'Reading' })
  await fillDraft(page, '40', `freshness-saved-${tag()}`)
  await saveButton(page).click()
  await page.waitForURL(contestStepUrl)
  const logId = savedLogId(page)
  await expect(page.getByRole('checkbox', { name: new RegExp(contest.title) })).toBeChecked()
  await page.getByRole('button', { name: 'Save submissions', exact: true }).click()
  await page.waitForURL(savedLogUrl)

  const after = await contestScore(readerContext.request, contest.id, readerId)
  expect(after, 'the save added to the contest score').toBeGreaterThan(before)
  const total = await yearTotal(readerContext.request, readerId)

  // Back through the router: every page shows the save.
  await submissions(page).getByRole('link', { name: contest.title, exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contest.id}/leaderboard/1$`))
  await expect.soft(rowScore(page, reader.name), 'standings').toHaveText(score.format(after), soon)
  await contestViews(page, 'Activity').click()
  await expect.soft(feedEntry(page, logId), 'contest feed').toBeVisible(soon)
  await banner(page, 'Home').click()
  await expect.soft(overallScore(page), 'home totals').toContainText(score.format(total), soon)
  await page.getByRole('link', { name: 'Your profile', exact: true }).click()
  await expect(page).toHaveURL(/\/statistics$/)
  await expect.soft(overallScore(page), 'profile totals').toContainText(score.format(total), soon)
  expect(loads(), 'full document loads').toBe(0)
})

test('an administrator\'s edit of another learner\'s log refreshes that learner\'s views without a reload', async ({ readerContext, adminContext }) => {
  const contest = await createContest(adminContext)
  const readerId = await accountId(readerContext)
  const seed = await submitLog(readerContext, contest.id, 2)

  const page = await adminContext.newPage()
  await page.goto(`/logs/${seed.id}`)
  await hydrated(page)
  const loads = documentLoads(page)

  await submissions(page).getByRole('link', { name: contest.title, exact: true }).click()
  await expect(rowScore(page, reader.name)).toHaveText(score.format(seed.score))
  await page.goBack()
  await expect(page).toHaveURL(new RegExp(`/logs/${seed.id}$`))
  await page.getByRole('link', { name: 'Activity', exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${readerId}/updates/1$`))
  await expect(page.locator(`li:has(a[href="/logs/${seed.id}"])`)).toContainText(score.format(seed.score))
  await page.getByRole('navigation', { name: 'Profile views' }).getByRole('link', { name: 'Overview', exact: true }).click()
  await expect(overallScore(page)).toBeVisible()
  await page.goBack()
  await expect(page).toHaveURL(new RegExp(`/profile/${readerId}/updates/1$`))

  await feedEntry(page, seed.id).click()
  await page.getByRole('link', { name: 'Edit log', exact: true }).click()
  await expect(amountInput(page)).toBeEnabled({ timeout: 15_000 })
  await amountInput(page).fill('50')
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await page.waitForURL(savedLogUrl)

  const edited = await apiJson<Log>(adminContext.request, `/immersion/logs/${seed.id}`)
  expect(edited.score, 'the edit changed the score').toBeGreaterThan(seed.score)
  const total = await yearTotal(adminContext.request, readerId)

  await page.getByRole('link', { name: 'Activity', exact: true }).click()
  await expect.soft(page.locator(`li:has(a[href="/logs/${seed.id}"])`), 'owner activity').toContainText(score.format(edited.score), soon)
  await page.getByRole('navigation', { name: 'Profile views' }).getByRole('link', { name: 'Overview', exact: true }).click()
  await expect.soft(overallScore(page), 'owner totals').toContainText(score.format(total), soon)
  await page.goBack()
  await feedEntry(page, seed.id).click()
  await submissions(page).getByRole('link', { name: contest.title, exact: true }).click()
  await expect.soft(rowScore(page, reader.name), 'standings').toHaveText(score.format(edited.score), soon)
  expect(loads(), 'full document loads').toBe(0)
})

test('joining a contest shows it in Joined and in the profile\'s contest history without a reload', async ({ returningLearner: learner, adminContext }) => {
  const contest = await createContest(adminContext)
  const page = await learner.context.newPage()
  await page.goto(`/contests/${contest.id}/leaderboard/1`)
  await hydrated(page)
  const loads = documentLoads(page)

  const joined = page.getByRole('navigation', { name: 'Contest collections' }).getByRole('link', { name: 'Joined', exact: true })
  const account = page.getByRole('button', { name: learner.name, exact: true })
  const history = () => page.getByRole('region', { name: 'Contests' }).getByRole('link', { name: contest.title, exact: true })

  await banner(page, 'Contests').click()
  await joined.click()
  await expect(page).toHaveURL(/\/contests\/joined\/1$/)
  await expect(page.getByRole('main')).not.toContainText(contest.title)
  await account.click()
  await page.getByRole('menuitem', { name: 'Profile', exact: true }).click()
  await expect(page).toHaveURL(/\/statistics$/)
  await expect(page.getByRole('heading', { name: 'Contests', exact: true })).toBeVisible()
  await expect(history()).toHaveCount(0)
  await page.goBack()
  await page.goBack()
  await page.goBack()
  await expect(page).toHaveURL(new RegExp(`/contests/${contest.id}/leaderboard/1$`))

  await page.getByRole('link', { name: 'Join contest', exact: true }).first().click()
  await expect(page.getByRole('button', { name: 'Register' })).toBeEnabled({ timeout: 15_000 })
  await page.getByRole('combobox', { name: 'Languages' }).fill('Japa')
  await page.getByRole('option', { name: 'Japanese', exact: true }).click()
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contest.id}/leaderboard/1$`))

  await banner(page, 'Contests').click()
  await joined.click()
  await expect.soft(page.getByRole('link', { name: contest.title, exact: true }), 'Joined list').toBeVisible(soon)
  await account.click()
  await page.getByRole('menuitem', { name: 'Profile', exact: true }).click()
  await expect(page).toHaveURL(/\/statistics$/)
  await expect.soft(history(), 'profile contest history').toBeVisible(soon)
  expect(loads(), 'full document loads').toBe(0)
})

test('dropping a registration language takes its logs out of the contest without a reload', async ({ readerContext, adminContext }) => {
  const contest = await createContest(adminContext)
  const seed = await submitLog(readerContext, contest.id, 2)
  const widened = await readerContext.request.post(`${api}/immersion/contests/${contest.id}/registration`, { data: { language_codes: ['jpn', 'kor'] } })
  expect(widened.status(), 'adding a second language').toBe(200)

  const page = await readerContext.newPage()
  await page.goto(`/logs/${seed.id}`)
  await hydrated(page)
  const loads = documentLoads(page)
  const submitted = submissions(page).getByRole('link', { name: contest.title, exact: true })
  await expect(submitted).toBeVisible()

  await submitted.click()
  await expect(rowScore(page, reader.name)).toHaveText(score.format(seed.score))
  await contestViews(page, 'Activity').click()
  await expect(feedEntry(page, seed.id)).toBeVisible()

  await page.getByRole('button', { name: 'Contest options', exact: true }).click()
  await page.getByRole('menuitem', { name: 'Update registration', exact: true }).click()
  await page.getByRole('button', { name: 'Remove Japanese', exact: true }).click()
  await page.getByRole('button', { name: 'Update registration', exact: true }).click()
  await expect(page).not.toHaveURL(/\/registration$/)

  await contestViews(page, 'Activity').click()
  await expect.soft(feedEntry(page, seed.id), 'contest feed').toHaveCount(0, soon)

  for (let back = 0; back < 8 && !new URL(page.url()).pathname.endsWith(`/logs/${seed.id}`); back++) await page.goBack()
  await expect(page).toHaveURL(new RegExp(`/logs/${seed.id}$`))
  await expect.soft(submitted, 'the log\'s submissions').toHaveCount(0, soon)
  expect(loads(), 'full document loads').toBe(0)
})

test('removing a log from a contest updates the log\'s submissions and the standings without a reload', async ({ readerContext, adminContext }) => {
  const contest = await createContest(adminContext)
  const seed = await submitLog(readerContext, contest.id, 2)

  const page = await adminContext.newPage()
  await page.goto(`/logs/${seed.id}`)
  await hydrated(page)
  const loads = documentLoads(page)
  await expect(submissions(page).getByRole('link', { name: contest.title, exact: true })).toBeVisible()

  await submissions(page).getByRole('link', { name: contest.title, exact: true }).click()
  await expect(rowScore(page, reader.name)).toHaveText(score.format(seed.score))
  await contestViews(page, 'Activity').click()
  await page.getByRole('button', { name: 'Remove from contest' }).click()
  const dialog = page.getByRole('dialog', { name: 'Remove log from contest' })
  await dialog.getByLabel('Reason').fill('Freshness journey moderation')
  await dialog.getByRole('button', { name: 'Yes, remove it' }).click()
  await expect(page.getByText('Log removed from contest.')).toBeVisible()

  await contestViews(page, 'Leaderboard').click()
  const remaining = await apiJson<{ entries: { user_display_name: string; score: number }[] }>(adminContext.request, `/immersion/contests/${contest.id}/leaderboard?page_size=50&page=0`)
  const entry = remaining.entries.find(item => item.user_display_name === reader.name)
  if (entry) await expect.soft(rowScore(page, reader.name), 'standings').toHaveText(score.format(entry.score), soon)
  else await expect.soft(rowScore(page, reader.name), 'standings').toHaveCount(0, soon)

  await page.goBack()
  await page.goBack()
  await page.goBack()
  await expect(page).toHaveURL(new RegExp(`/logs/${seed.id}$`))
  await expect.soft(submissions(page), 'the log\'s submissions').toContainText('Saved to the learner\'s personal record only.', soon)
  expect(loads(), 'full document loads').toBe(0)
})
