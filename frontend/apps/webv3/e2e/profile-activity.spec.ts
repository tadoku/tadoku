import { randomUUID } from 'node:crypto'
import type { APIRequestContext, Page } from '@playwright/test'
import { accountId, activityEntries, api, branch, expect, linkTo, logDescription, logLinks, reader, test, webv3 } from './logging-fixtures'

const pageSize = 25

type Log = { id: string; created_at: string; description?: string; tags: string[]; language: { name: string }; activity: { name: string } }

/** One page of a profile's logs, counted from 1, as the API lists them. */
async function apiLogs(request: APIRequestContext, userId: string, page: number, date?: string) {
  const query = new URLSearchParams({ page_size: String(pageSize), page: String(page - 1), include_deleted: 'false', ...(date ? { date } : {}) })
  const response = await request.get(`${api}/immersion/users/${userId}/logs?${query}`)
  expect(response.status()).toBe(200)
  // Development branches select their tenant by host; the container's synthetic API has no gateway.
  if (reader.email) expect(response.headers()['x-dev-selected']).toBe(branch)
  return await response.json() as { logs: Log[]; total_size: number }
}

/** The page lists exactly the API's logs, in its newest-first order, each linking to its native detail page. */
async function expectEntries(page: Page, expected: { logs: Log[] }) {
  const titles = logLinks(page)
  await expect(titles).toHaveCount(expected.logs.length)
  // Retried: after a client navigation the URL changes before the new page's rows replace the old ones.
  await expect.poll(() => titles.evaluateAll(links => links.map(link => link.getAttribute('href'))))
    .toEqual(expected.logs.map(log => `/logs/${log.id}`))
  await expect(titles).toHaveText(expected.logs.map(log => {
    const nsfw = log.tags.some(tag => tag.toLowerCase() === 'nsfw')
    return nsfw || !log.description ? `${log.language.name} ${log.activity.name.toLowerCase()}` : log.description
  }))
  const created = expected.logs.map(log => Date.parse(log.created_at))
  expect(created).toEqual([...created].sort((left, right) => right - left))
}

const homeParticipant = (page: Page) => page.getByRole('table').getByRole('rowheader').getByRole('link').first()
const fitsWidth = (page: Page) => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)

test('a visitor follows a standings profile to its activity, pages through it and opens an entry', async ({ page }, testInfo) => {
  test.skip(!webv3 || !api, 'Set development URLs')
  await page.goto('/')
  const participant = homeParticipant(page)
  const name = (await participant.textContent())!
  await participant.click()
  await expect(page).toHaveURL(/\/profile\/[0-9a-f-]{36}\/statistics$/)
  await page.getByRole('link', { name: 'Activity', exact: true }).click()

  await expect(page).toHaveURL(/\/profile\/[0-9a-f-]{36}\/updates\/1$/)
  const userId = new URL(page.url()).pathname.split('/')[2]
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(name)
  await expect(page.getByRole('link', { name: 'Activity', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(page.getByRole('link', { name: /Download/ })).toHaveCount(0)
  const first = await apiLogs(page.request, userId, 1)
  expect(first.logs.length).toBeGreaterThan(0)
  await expectEntries(page, first)
  await page.screenshot({ path: testInfo.outputPath('visitor-activity.png') })

  if (first.total_size > pageSize) {
    await page.getByRole('navigation', { name: 'Update pages', exact: true }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(new RegExp(`/profile/${userId}/updates/2$`))
    await expectEntries(page, await apiLogs(page.request, userId, 2))
  }

  const entry = logLinks(page).first()
  const href = (await entry.getAttribute('href'))!
  await entry.click()
  await expect(page).toHaveURL(new RegExp(`${href}$`))
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  await expect(page.getByText('Logged', { exact: true })).toBeVisible()
})

test('an owner pages through their newest-first activity, opens a day from the calendar and an entry', async ({ readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const userId = await accountId(readerContext)
  const request = readerContext.request

  // Pagination needs more than one page and the calendar a log today: top the account up through the API's own create
  // operation, with at least one log.
  const existing = await apiLogs(request, userId, 1)
  for (let index = Math.min(existing.total_size, pageSize); index <= pageSize; index++) {
    const response = await request.post(`${api}/immersion/logs`, {
      data: { language_code: 'jpn', activity_id: 2, duration_seconds: 60, tags: [], description: `webv3-activity-${index}` },
    })
    expect(response.status()).toBe(200)
  }

  const page = await readerContext.newPage()
  await page.goto(`/profile/${userId}/updates`)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Dev Reader')
  await expect(page.getByRole('link', { name: 'Download CSV' })).toBeVisible()
  const first = await apiLogs(request, userId, 1)
  expect(first.total_size).toBeGreaterThan(pageSize)
  await expectEntries(page, first)
  const pages = page.getByRole('navigation', { name: 'Update pages', exact: true })
  await expect(page.getByRole('navigation', { name: /pages/ })).toHaveCount(1)
  await expect(pages.getByRole('link', { name: 'Page 2' })).toHaveAttribute('href', `/profile/${userId}/updates/2`)
  await page.screenshot({ path: testInfo.outputPath('owner-activity.png') })

  await pages.getByRole('link', { name: 'Page 2' }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${userId}/updates/2$`))
  const second = await apiLogs(request, userId, 2)
  await expectEntries(page, second)
  expect(Date.parse(second.logs[0].created_at)).toBeLessThanOrEqual(Date.parse(first.logs.at(-1)!.created_at))
  await pages.getByRole('link', { name: 'Previous page' }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${userId}/updates/1$`))
  await expectEntries(page, first)

  await page.goto(`/profile/${userId}/updates/99`)
  await expect(page.getByRole('heading', { name: 'This page is past the end of the activity' })).toBeVisible()

  // A calendar day opens that UTC day's updates natively, with a way back to all of them.
  const today = new Date().toISOString().slice(0, 10)
  await page.goto(`/profile/${userId}/statistics`)
  await linkTo(page, `/profile/${userId}/updates/1?date=${today}`).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${userId}/updates/1\\?date=${today}$`))
  await expect(page.getByText('Showing activity on')).toBeVisible()
  const day = await apiLogs(request, userId, 1, today)
  expect(day.total_size).toBeGreaterThan(0)
  await expectEntries(page, day)
  if (day.total_size > pageSize) {
    await expect(pages.getByRole('link', { name: 'Page 2' })).toHaveAttribute('href', `/profile/${userId}/updates/2?date=${today}`)
  }
  await page.getByRole('link', { name: 'Show all activity' }).click()
  await expect(page).toHaveURL(new RegExp(`/profile/${userId}/updates/1$`))
  await expectEntries(page, first)

  const entry = first.logs[0]
  await linkTo(page, `/logs/${entry.id}`).click()
  await expect(page).toHaveURL(new RegExp(`/logs/${entry.id}$`))
  if (entry.description) await expect(logDescription(page)).toHaveText(entry.description)
})

test('a new account’s activity shows the empty state to its owner and to visitors', async ({ browser, learner }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const owner = learner.context
  const userId = learner.id
  expect((await apiLogs(owner.request, userId, 1)).total_size).toBe(0)

  const page = await owner.newPage()
  await page.goto(`/profile/${userId}/updates/1`)
  await expect(page.getByRole('heading', { name: 'No activity yet' })).toBeVisible()
  await expect(page.getByRole('main').getByRole('link', { name: 'Log activity' })).toHaveAttribute('href', '/logs/new')
  await expect(activityEntries(page)).toHaveCount(0)
  await expect(page.getByRole('navigation', { name: 'Update pages', exact: true })).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('owner-empty.png'), fullPage: true })

  const visitor = await browser.newContext({ baseURL: webv3 })
  const anonymous = await visitor.newPage()
  await anonymous.goto(`/profile/${userId}/updates/1`)
  await expect(anonymous.getByRole('heading', { name: 'No activity yet' })).toBeVisible()
  await expect(anonymous.getByText('Logs appear here once they have been recorded.')).toBeVisible()
  await expect(anonymous.getByRole('main').getByRole('link', { name: 'Log activity' })).toHaveCount(0)
  await visitor.close()
})

test('an unknown or malformed profile’s activity is not found', async ({ page }) => {
  test.skip(!webv3 || !api, 'Set development URLs')
  const unknown = randomUUID()
  for (const path of [`/profile/${unknown}/updates/1`, `/profile/${unknown}/updates`, '/profile/not-a-user/updates/1']) {
    const response = await page.goto(path)
    expect(response?.status()).toBe(404)
  }
  await expect(page.getByRole('heading', { name: 'Profile not found' })).toBeVisible()
})

test('the activity page fits at 390px', async ({ page }, testInfo) => {
  test.skip(!webv3 || !api, 'Set development URLs')
  await page.setViewportSize({ width: 390, height: 900 })
  await page.goto('/')
  await homeParticipant(page).click()
  await page.getByRole('link', { name: 'Activity', exact: true }).click()
  await expect(activityEntries(page).first()).toBeVisible()
  expect(await fitsWidth(page)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('activity-390.png') })
})
