import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { accountId, admin, api, expect, reader, test, webv3 } from './logging-fixtures'

interface Contest {
  id: string
  title: string
  contest_start: string
  contest_end: string
  registration_end: string
  private: boolean
  language_code_allow_list: string[] | null
  activity_type_id_allow_list: number[]
}
interface Contests { contests: Contest[]; total_size: number }

const count = new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 })
const pageSize = 25
const collections = {
  official: { path: '/contests/official', official: true },
  community: { path: '/contests/user-contests', official: false },
}

/** One page of a contest list as the request's session, if any, sees it. */
async function contests(request: APIRequestContext, official: boolean, page = 1, userId?: string) {
  const query = `page_size=${pageSize}&page=${page - 1}&official=${official}&include_deleted=false${userId ? `&user_id=${userId}` : ''}`
  const response = await request.get(`${api}/immersion/contests?${query}`)
  expect(response.ok(), `contest list answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<Contests>
}

/** The round tag a contest shows on the UTC day `today`. */
function statuses(contest: Contest, today = new Date().toISOString().slice(0, 10)) {
  if (today > contest.contest_end) return ['Ended']
  return [today >= contest.contest_start ? 'Live now' : 'Upcoming']
}

/** Every API contest is a row in the API's order, linking to its contest page, with its privacy and status tags. */
async function expectRows(page: Page, data: Contests, pageNumber = 1) {
  const rows = page.locator('.contest-row')
  await expect(rows).toHaveCount(data.contests.length)
  for (const [index, contest] of data.contests.entries()) {
    const row = rows.nth(index)
    await expect(row.getByRole('heading', { level: 2 })).toHaveText(contest.title)
    await expect(row.getByRole('link', { name: contest.title, exact: true }))
      .toHaveAttribute('href', `/contests/${contest.id}/leaderboard/1`)
    await expect(row.locator('.contest-row__title .paper-status-tag')).toHaveText(statuses(contest))
    const meta = row.locator('.contest-row__allows')
    if (contest.private) await expect(meta).toContainText('Unlisted')
    else await expect(meta).not.toContainText('Unlisted')
  }
  if (data.total_size === 0) {
    await expect(page.getByRole('heading', { level: 2 })).toHaveText(/No contests found|You don’t organize any contests yet/)
    return
  }
  const first = (pageNumber - 1) * pageSize + 1
  // The range summary appears only beside page links.
  await expect(page.locator('.contests-pagination > p')).toHaveText(data.total_size > pageSize
    ? [`${count.format(first)}–${count.format(first + data.contests.length - 1)} of ${count.format(data.total_size)} contests`]
    : [])
  await expect(page.getByRole('navigation', { name: 'Contest pages', exact: true })).toHaveCount(data.total_size > pageSize ? 1 : 0)
}

async function hydrated(page: Page) {
  await expect(page.getByRole('button', { name: 'Use dark theme', exact: true })).toBeEnabled()
}

/** Gives Dev Admin more than one page of community contests in the selected tenant, one of them unlisted. */
async function ensureOrganizerPages(context: BrowserContext) {
  const userId = await accountId(context)
  const owned = await contests(context.request, false, 1, userId)
  const year = new Date().getUTCFullYear()
  const today = new Date().toISOString().slice(0, 10)
  for (let index = owned.total_size; index < pageSize + 2; index++) {
    const kind = index % 3
    const response = await context.request.post(`${api}/immersion/contests`, {
      data: {
        title: `Webv3 list check ${index + 1}`,
        description: 'Created by the Webv3 contest list journey.',
        contest_start: kind === 0 ? `${year - 1}-01-01` : kind === 1 ? today : `${year + 1}-01-01`,
        contest_end: kind === 0 ? `${year - 1}-01-31` : kind === 1 ? `${year + 1}-01-31` : `${year + 1}-01-31`,
        registration_end: kind === 0 ? `${year - 1}-01-15` : `${year + 1}-01-15`,
        private: index === 0,
        official: false,
        language_code_allow_list: ['jpn'],
        activity_type_id_allow_list: [1, 2],
      },
    })
    expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  }
  return userId
}

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('`/contests` redirects permanently to the official contests, as in Webv2', async ({ page }) => {
  const response = await page.request.get('/contests', { maxRedirects: 0 })
  expect(response.status()).toBe(308)
  expect(new URL(response.headers().location, 'http://relative.invalid').pathname).toBe('/contests/official')
  await page.goto('/contests')
  await expect(page).toHaveURL(/\/contests\/official$/)
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
})

test('server-rendered official and community lists show the API contests and link to their contest pages', async ({ browser, request }, testInfo) => {
  const context = await browser.newContext({ javaScriptEnabled: false })
  // A stale session cookie leaves the visitor a guest, whom the configuration endpoint also answers.
  await context.addCookies([{ name: 'tadoku_session', value: 'expired', url: webv3! }])
  const page = await context.newPage()

  for (const [name, collection] of Object.entries(collections)) {
    const data = await contests(request, collection.official)
    const response = await page.goto(`${collection.path}/1`)
    expect(response?.status()).toBe(200)
    await expect(page.getByRole('heading', { level: 1, name: 'Contests' })).toBeVisible()
    await expect(page.getByRole('navigation', { name: 'Contest collections' }).locator('[aria-current="page"]'))
      .toHaveText(name === 'official' ? 'Official' : 'Community')
    await expect(page.getByRole('link', { name: 'Contests', exact: true }).first()).toHaveAttribute('href', '/contests/official/1')
    await expect(page.getByText('Contest dates are calendar days in UTC.')).toBeVisible()
    // Guests see no create action and no restrictions, as in Webv2; a row may still name its deadline or visibility.
    await expect(page.getByRole('link', { name: 'Create contest' })).toHaveCount(0)
    for (const meta of await page.locator('.contest-row__allows').allTextContents()) {
      expect(meta).toMatch(/^(Registration closes \d+ \w+ \d{4})?( · )?(Unlisted)?$/)
    }
    await expectRows(page, data)
    await page.screenshot({ path: testInfo.outputPath(`ssr-${name}.png`), fullPage: true })
  }
  await context.close()
})

test('collection tabs, pages and the signed-out My contests prompt work in the browser', async ({ page, request }, testInfo) => {
  const failures: string[] = []
  page.on('pageerror', error => failures.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') failures.push(message.text())
  })
  await page.goto('/contests/official')
  await hydrated(page)
  await expectRows(page, await contests(request, true))

  const tabs = page.getByRole('navigation', { name: 'Contest collections' })
  await tabs.getByRole('link', { name: 'Community' }).click()
  await expect(page).toHaveURL(/\/contests\/user-contests\/1$/)
  await expect(page).toHaveTitle('Community contests · Tadoku')
  const community = await contests(request, false)
  await expectRows(page, community)

  if (community.total_size > pageSize) {
    await page.getByRole('navigation', { name: 'Contest pages', exact: true }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(/\/contests\/user-contests\/2$/)
    await expectRows(page, await contests(request, false, 2), 2)
  }

  const beyond = Math.ceil(community.total_size / pageSize) + 1
  await page.goto(`/contests/user-contests/${beyond}`)
  if (community.total_size > 0) {
    await expect(page.getByRole('heading', { name: 'This page is past the end of the list' })).toBeVisible()
    await page.getByRole('link', { name: 'Go to the first page' }).click()
    await expect(page).toHaveURL(/\/contests\/user-contests\/1$/)
    await expectRows(page, community)
  } else {
    await expect(page.getByRole('heading', { name: 'No contests found' })).toBeVisible()
  }

  await hydrated(page)
  await page.getByRole('navigation', { name: 'Contest collections' }).getByRole('link', { name: 'My contests' }).click()
  await expect(page).toHaveURL(/\/contests\/my-contests\/1$/)
  await expect(page.getByRole('heading', { level: 2, name: 'Sign in to see the contests you organize' })).toBeVisible()
  const login = new URL((await page.locator('main').getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!)
  expect(login.pathname).toBe('/login')
  expect(login.searchParams.get('return_to')).toBe(page.url())
  await expect(page.locator('.contest-row')).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('my-contests-anonymous.png'), fullPage: true })
  expect(failures).toEqual([])
})

test('an organizer pages through their own contests, unlisted ones included, and can open contest creation', async ({ adminContext }, testInfo) => {
  test.skip(!admin.email || !admin.password, 'Set E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD')
  const userId = await ensureOrganizerPages(adminContext)
  const page = await adminContext.newPage()
  const mine = await contests(adminContext.request, false, 1, userId)
  expect(mine.total_size).toBeGreaterThan(pageSize)

  await page.goto('/contests/my-contests/1')
  await expect(page.getByRole('heading', { level: 1, name: 'Contests' })).toBeVisible()
  await expectRows(page, mine)
  await expect(page.getByRole('link', { name: 'Create contest' })).toHaveAttribute('href', '/contests/new')
  // Signed-in community and organizer lists name the restrictions, as Webv2 does.
  await expect(page.locator('.contest-row').filter({ hasText: 'Webv3 list check' }).first().locator('.contest-row__allows'))
    .toContainText('Japanese · Reading, Listening')
  await page.screenshot({ path: testInfo.outputPath('my-contests-organizer.png'), fullPage: true })

  await hydrated(page)
  await page.getByRole('navigation', { name: 'Contest pages', exact: true }).getByRole('link', { name: 'Next page' }).click()
  await expect(page).toHaveURL(/\/contests\/my-contests\/2$/)
  await expectRows(page, await contests(adminContext.request, false, 2, userId), 2)

  await page.setViewportSize({ width: 390, height: 900 })
  await page.goto('/contests/my-contests/1')
  await hydrated(page)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('my-contests-organizer-390.png'), fullPage: true })
})

test('a participant sees their unlisted contest in My contests and Community, which anonymous visitors do not', async ({ page, readerContext, request }) => {
  test.skip(!reader.email || !reader.password, 'Set E2E_READER_EMAIL and E2E_READER_PASSWORD')
  const userId = await accountId(readerContext)
  const mine = await contests(readerContext.request, false, 1, userId)
  const unlisted = mine.contests.filter(contest => contest.private)
  expect(unlisted.length, 'Dev Reader owns the seeded private contest').toBeGreaterThan(0)
  const permission = await readerContext.request.get(`${api}/immersion/contests/create-permissions`)

  await page.goto('/contests/my-contests/1')
  await expectRows(page, mine)
  await expect(page.getByRole('link', { name: 'Create contest' })).toHaveCount(permission.status() === 200 ? 1 : 0)

  await page.goto('/contests/user-contests/1')
  await expectRows(page, await contests(readerContext.request, false))
  const anonymous = await contests(request, false)
  expect(anonymous.contests.map(contest => contest.id)).not.toContain(unlisted[0].id)
})

test('contest lists stay within mobile and desktop viewports', async ({ page }, testInfo) => {
  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 900 })
    for (const path of ['/contests/official/1', '/contests/user-contests/1', '/contests/my-contests/1']) {
      await page.goto(path)
      await expect(page.getByRole('heading', { level: 1, name: 'Contests' })).toBeVisible()
      await hydrated(page)
      await page.evaluate(() => document.fonts.ready)
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${path} at ${width}px`).toBe(true)
      await page.screenshot({ path: testInfo.outputPath(`${path.split('/')[2]}-${width}.png`), fullPage: true })
    }
  }
})
