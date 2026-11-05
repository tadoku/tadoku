import { api, expect, reader, test, webv3 } from './logging-fixtures'
import { apiJson, board, expectRows, expectTotal } from './standings'

const year = new Date().getUTCFullYear()

/** Yearly and all-time rows open the learner's native profile statistics. */
const statistics = () => (userId: string) => `/profile/${userId}/statistics`

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('server-rendered boards show the API standings, their scopes and profile links', async ({ browser, request }, testInfo) => {
  const latest = await apiJson<{ id: string; title: string }>(request, '/immersion/contests/latest-official')
  const cases = [
    {
      path: '/leaderboard/latest/1',
      api: `/immersion/contests/${latest.id}/leaderboard`,
      scope: latest.title,
      profile: () => (userId: string) => `/contests/${latest.id}/profile/${userId}/1`,
    },
    {
      path: `/leaderboard/yearly/${year}/1`,
      api: `/immersion/leaderboard/yearly/${year}`,
      scope: `${year} official standings`,
      profile: statistics,
    },
    {
      path: '/leaderboard/all-time/1',
      api: '/immersion/leaderboard/global',
      scope: 'All-time official standings',
      profile: statistics,
    },
  ]
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()

  for (const item of cases) {
    const data = await board(request, item.api)
    const response = await page.goto(item.path)
    expect(response?.status()).toBe(200)
    await expect(page.getByRole('heading', { level: 1, name: 'Leaderboard' })).toBeVisible()
    await expect(page.getByRole('heading', { level: 2, name: item.scope, exact: true })).toBeVisible()
    await expect(page.getByRole('navigation', { name: 'Leaderboard period' }).locator('[aria-current="page"]')).toHaveCount(1)
    await expect(page.getByRole('link', { name: 'Leaderboard', exact: true }).first()).toHaveAttribute('href', '/leaderboard/latest/1')
    await expectRows(page, data.entries, item.profile())
    await expectTotal(page, data)
    await page.screenshot({ path: testInfo.outputPath(`ssr-${item.path.split('/')[2]}.png`), fullPage: true })
  }
  await context.close()
})

test('optional page segments, pages past the end and page links follow the API', async ({ page, request }) => {
  const data = await board(request, '/immersion/leaderboard/global')

  await page.goto('/leaderboard/all-time')
  await expectRows(page, data.entries, statistics())
  await expectTotal(page, data)

  if (data.total_size > 50) {
    await page.getByRole('navigation', { name: 'Standings pages' }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(/\/leaderboard\/all-time\/2$/)
    const second = await board(request, '/immersion/leaderboard/global', 2)
    await expectRows(page, second.entries, statistics())
    await expectTotal(page, second, 2)
  }

  const beyond = Math.ceil(data.total_size / 50) + 1
  await page.goto(`/leaderboard/all-time/${beyond}`)
  if (data.total_size > 0) {
    await expect(page.getByRole('heading', { name: 'This page is past the end of the standings' })).toBeVisible()
    await page.getByRole('link', { name: 'Go to the first page' }).click()
    await expect(page).toHaveURL(/\/leaderboard\/all-time\/1$/)
    await expectRows(page, data.entries, statistics())
  } else {
    await expect(page.getByRole('heading', { name: 'No official scores for this period' })).toBeVisible()
  }
})

test('year and period tabs switch boards in the browser', async ({ page, request }) => {
  const failures: string[] = []
  page.on('pageerror', error => failures.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') failures.push(message.text())
  })
  await page.goto(`/leaderboard/yearly/${year}/1`)
  await expect(page.getByRole('button', { name: 'Use dark theme', exact: true })).toBeEnabled()
  const years = page.getByRole('navigation', { name: 'Year' })
  await expect(years.getByRole('link')).toHaveText(Array.from({ length: year - 2019 }, (_, index) => String(year - index)))
  await expect(years.getByRole('link', { name: String(year) })).toHaveAttribute('aria-current', 'page')

  const previous = await board(request, `/immersion/leaderboard/yearly/${year - 1}`)
  await years.getByRole('link', { name: String(year - 1) }).click()
  await expect(page).toHaveURL(new RegExp(`/leaderboard/yearly/${year - 1}/1$`))
  await expect(page).toHaveTitle(`${year - 1} leaderboard · Tadoku`)
  await expect(page.getByRole('heading', { level: 2, name: `${year - 1} official standings` })).toBeVisible()
  await expectRows(page, previous.entries, statistics())
  await expectTotal(page, previous)

  const periods = page.getByRole('navigation', { name: 'Leaderboard period' })
  await periods.getByRole('link', { name: 'All time' }).click()
  await expect(page).toHaveURL(/\/leaderboard\/all-time\/1$/)
  await expect(page.getByRole('heading', { level: 2, name: 'All-time official standings' })).toBeVisible()
  await periods.getByRole('link', { name: 'Yearly' }).click()
  await expect(page).toHaveURL(new RegExp(`/leaderboard/yearly/${year}/1$`))
  await periods.getByRole('link', { name: 'Latest official' }).click()
  await expect(page).toHaveURL(/\/leaderboard\/latest\/1$/)
  await expect(page.getByRole('heading', { level: 2 }).first()).toBeVisible()
  expect(failures).toEqual([])
})

test('a signed-in participant sees their own row and the contest action their registration allows', async ({ page, readerContext, request }, testInfo) => {
  test.skip(!reader.email || !reader.password, 'Set E2E_READER_EMAIL and E2E_READER_PASSWORD')
  const latest = await apiJson<{ id: string; contest_start: string; contest_end: string; registration_end: string }>(
    request, '/immersion/contests/latest-official',
  )
  const session = await (await readerContext.request.get(`${api}/identity/session`)).json() as { account: { id: string } }
  const registration = await readerContext.request.get(`${api}/immersion/contests/${latest.id}/registration`)
  const registered = registration.status() === 200
  const data = await board(request, `/immersion/contests/${latest.id}/leaderboard`)
  const today = new Date().toISOString().slice(0, 10)
  const ongoing = today >= latest.contest_start && today <= latest.contest_end

  await page.goto('/leaderboard/latest/1')
  await expect(page.getByText('You need to log in to participate in this contest.')).toHaveCount(0)
  const own = page.locator('.standings-table tbody tr[aria-current="true"]')
  if (data.entries.some(entry => entry.user_id === session.account.id)) {
    await expect(own).toHaveCount(1)
    await expect(own).toContainText('You')
  } else {
    await expect(own).toHaveCount(0)
  }
  const header = page.locator('.standings-header')
  if (ongoing && registered) await expect(header.getByRole('link', { name: 'Log activity' })).toHaveAttribute('href', '/logs/new')
  else if (!registered && today <= latest.registration_end) await expect(header.getByRole('link', { name: 'Join contest' })).toBeVisible()
  else await expect(header.getByRole('link')).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('latest-signed-in.png'), fullPage: true })
})

test('leaderboards stay within mobile and desktop viewports', async ({ page }, testInfo) => {
  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 900 })
    for (const path of ['/leaderboard/latest/1', `/leaderboard/yearly/${year}/1`, '/leaderboard/all-time/1']) {
      await page.goto(path)
      await expect(page.getByRole('heading', { level: 1, name: 'Leaderboard' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Use dark theme', exact: true })).toBeEnabled()
      await page.evaluate(() => document.fonts.ready)
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${path} at ${width}px`).toBe(true)
      await page.screenshot({ path: testInfo.outputPath(`${path.split('/')[2]}-${width}.png`), fullPage: true })
    }
  }
})
