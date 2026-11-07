import { randomUUID } from 'node:crypto'
import type { APIRequestContext, Page } from '@playwright/test'
import {
  accountId,
  createLog,
  api,
  branch,
  expect,
  fillDraft,
  openForm,
  reader,
  test,
  webv3,
} from './logging-fixtures'

// The fixtures' own test.use applies only to the first file that imports them; authenticated runs keep traces off.
test.use({ trace: 'off', video: 'off', ignoreHTTPSErrors: false })

const year = new Date().getUTCFullYear()
const scoreFormat = new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 })

type Contest = { contest: { id: string; title: string; private: boolean } }

/** A profile's year as the API reports it to `request`'s account. */
async function apiYear(request: APIRequestContext, userId: string, target: number) {
  const read = async <T>(path: string) => {
    const response = await request.get(`${api}/immersion/users/${userId}/${path}/${target}`)
    expect(response.status()).toBe(200)
    // Development branches select their tenant by host; the container's synthetic API has no gateway.
    if (reader.email) expect(response.headers()['x-dev-selected']).toBe(branch)
    return await response.json() as T
  }
  const [activity, scores, split, contests] = await Promise.all([
    read<{ total_updates: number; scores: { score: number }[] }>('activity'),
    read<{ scores: { language_name?: string; language_code: string; score: number }[] }>('scores'),
    read<{ activities: { activity_name: string; score: number }[] }>('activity-split'),
    read<{ registrations: Contest[] }>('contest-registrations'),
  ])
  return { activity, scores, split, contests: contests.registrations.map(registration => registration.contest) }
}

/** The page shows the year's totals, breakdowns and contests exactly as the API reports them. */
async function expectYear(page: Page, request: APIRequestContext, userId: string, target: number) {
  const expected = await apiYear(request, userId, target)
  await expect(page.getByRole('group', { name: 'Profile year' })).toContainText(String(target))

  if (expected.activity.total_updates === 0) {
    await expect(page.getByRole('heading', { name: `No activity recorded in ${target}` })).toBeVisible()
    await expect(page.locator('.paper-stat-grid')).toHaveCount(0)
  } else {
    const total = expected.scores.scores.reduce((sum, score) => sum + score.score, 0)
    await expect(stat(page, 'Overall score')).toContainText(scoreFormat.format(total))
    await expect(stat(page, 'Updates')).toContainText(String(expected.activity.total_updates))
    await expect(stat(page, 'Languages')).toContainText(String(expected.scores.scores.length))
    const languages = page.getByRole('region', { name: 'Score by language' })
    for (const score of expected.scores.scores) {
      await expect(languages.getByRole('listitem').filter({ hasText: score.language_name ?? score.language_code }))
        .toContainText(`${scoreFormat.format(score.score)} points`)
    }
    const activities = page.getByRole('region', { name: 'Score by activity' })
    for (const activity of expected.split.activities) {
      await expect(activities.getByRole('listitem').filter({ hasText: activity.activity_name }))
        .toContainText(`${scoreFormat.format(activity.score)} points`)
    }
  }

  const contests = page.getByRole('region', { name: 'Contests' })
  await expect(contests.getByRole('link')).toHaveText(expected.contests.map(contest => contest.title))
  return expected
}

const stat = (page: Page, label: string) => page.locator('.paper-stat-grid__item', { hasText: label }).locator('dd')

test('an owner sees API totals, switches years and gets the export and day links', async ({ readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await fillDraft(page, '12', `webv3-profile-${Date.now()}`)
  await createLog(page)
  const userId = await accountId(readerContext)

  await page.goto(`/profile/${userId}/statistics/${year}`)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Dev Reader')
  const current = await expectYear(page, readerContext.request, userId, year)
  expect(current.activity.total_updates).toBeGreaterThan(0)
  for (const contest of current.contests.filter(contest => contest.private)) {
    await expect(page.getByRole('listitem').filter({ hasText: contest.title })).toContainText('Unlisted')
  }

  await expect(page.getByRole('link', { name: 'Download CSV' })).toHaveAttribute('href', /\/api\/internal\/immersion\/logs\/export\?format=csv$/)
  await expect(page.getByRole('link', { name: 'Download JSON' })).toHaveAttribute('href', /\/api\/internal\/immersion\/logs\/export\?format=json$/)
  const csv = await readerContext.request.get((await page.getByRole('link', { name: 'Download CSV' }).getAttribute('href'))!)
  expect(csv.status()).toBe(200)
  expect(csv.headers()['content-disposition']).toMatch(/attachment/)
  await expect(page.getByRole('link', { name: 'Activity', exact: true })).toHaveAttribute('href', new RegExp(`/profile/${userId}/updates/1$`))
  const today = new Date().toISOString().slice(0, 10)
  const day = page.locator(`a.paper-heatmap__link[href$="/profile/${userId}/updates/1?date=${today}"]`)
  await expect(day).toHaveAttribute('aria-label', / points on /)
  await page.screenshot({ path: testInfo.outputPath('owner-current-year.png'), fullPage: true })

  // Year switching: the previous year, then back through the year control without a reload.
  await page.goto(`/profile/${userId}/statistics/${year - 1}`)
  const previous = await expectYear(page, readerContext.request, userId, year - 1)
  expect(previous.activity.total_updates).not.toBe(current.activity.total_updates)
  await page.getByRole('link', { name: `Next year, ${year}` }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${userId}/statistics/${year}$`))
  await expectYear(page, readerContext.request, userId, year)
})

test('a visitor opens a standings profile, server-rendered without owner tools', async ({ browser }) => {
  test.skip(!webv3 || !api, 'Set development URLs')
  const context = await browser.newContext({ baseURL: webv3, javaScriptEnabled: false })
  const page = await context.newPage()
  await page.goto('/')
  const participant = page.locator('a.home-person-link').first()
  const name = (await participant.textContent())!
  await participant.click()

  await expect(page).toHaveURL(/\/profile\/[0-9a-f-]{36}\/statistics$/)
  const userId = new URL(page.url()).pathname.split('/')[2]
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(name)
  const visible = await expectYear(page, context.request, userId, year)
  expect(visible.contests.every(contest => !contest.private)).toBe(true)
  if (visible.activity.total_updates > 0) await expect(page.locator('.paper-heatmap rect[data-date]')).not.toHaveCount(0)
  await expect(page.getByRole('link', { name: /Download/ })).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Activity', exact: true }))
    .toHaveAttribute('href', `/profile/${userId}/updates/1`)
  await context.close()
})

test('a year without activity shows the empty state', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const userId = await accountId(readerContext)
  await page.goto(`/profile/${userId}/statistics/2000`)
  const expected = await expectYear(page, page.request, userId, 2000)
  expect(expected.activity.total_updates).toBe(0)
  await expect(page.getByText('Hasn’t participated in any contests in 2000.')).toBeVisible()
})

test('an unknown or malformed profile is not found', async ({ page }) => {
  test.skip(!webv3 || !api, 'Set development URLs')
  const unknown = randomUUID()
  expect((await page.request.get(`${api}/immersion/users/${unknown}/profile`)).status()).toBe(404)

  for (const path of [`/profile/${unknown}/statistics`, '/profile/not-a-user/statistics', `/profile/${unknown}/statistics/20x6`]) {
    const response = await page.goto(path)
    expect(response?.status()).toBe(404)
  }
  await page.goto(`/profile/${unknown}/statistics/${year}`)
  await expect(page.getByRole('heading', { name: 'Profile not found' })).toBeVisible()
})

test('the profile fits at 390px and its calendar scrolls within its region', async ({ readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const userId = await accountId(readerContext)
  const page = await readerContext.newPage()
  await page.setViewportSize({ width: 390, height: 900 })
  await page.goto(`/profile/${userId}/statistics/${year}`)

  const calendar = page.getByRole('region', { name: 'Daily activity' })
  await expect(calendar).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  expect(await calendar.evaluate(element => element.scrollWidth > element.clientWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('owner-390.png'), fullPage: true })
})
