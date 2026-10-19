import { expect, test } from '@playwright/test'

test('renders the configured read-only contest and standings before JavaScript runs', async ({ browser, request }, testInfo) => {
  const api = process.env.TADOKU_API_URL || 'https://tadoku.dev.lab/api/internal'
  const latestResponse = await request.get(`${api}/immersion/contests/latest-official`)
  expect(latestResponse.ok()).toBeTruthy()
  const latest = await latestResponse.json()
  const standingsResponse = await request.get(`${api}/immersion/contests/${latest.id}/leaderboard?page_size=5&page=0`)
  expect(standingsResponse.ok()).toBeTruthy()
  const standings = await standingsResponse.json()
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()
  const response = await page.goto(testInfo.project.use.baseURL || 'http://127.0.0.1:5183')
  expect(response?.status()).toBe(200)
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  await expect(page.getByRole('heading', { name: latest.title, exact: true })).toBeVisible()
  for (const entry of standings.entries) {
    await expect(page.getByRole('row').filter({ hasText: entry.user_display_name })).toContainText(new Intl.NumberFormat('en-GB', { maximumFractionDigits: 1 }).format(entry.score))
  }
  await page.screenshot({ path: testInfo.outputPath('homepage-ssr.png'), fullPage: true })
  await context.close()
})

test('hydrates without errors and keeps the homepage within mobile and desktop viewports', async ({ page }, testInfo) => {
  const failures: string[] = []
  page.on('pageerror', error => failures.push(error.message))
  page.on('console', message => {
    if (message.type() === 'error') failures.push(message.text())
  })
  for (const width of [320, 390, 768, 1280]) {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/')
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
    await expect(page.locator('table')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Use dark theme', exact: true })).toBeEnabled()
    await page.evaluate(() => document.fonts.ready)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`homepage-${width}.png`), fullPage: true })
  }
  expect(failures).toEqual([])
})

test('validates and saves an editable local immersion goal without sending writes', async ({ page }, testInfo) => {
  const writes: string[] = []
  page.on('request', request => {
    if (!['GET', 'HEAD', 'OPTIONS'].includes(request.method())) writes.push(`${request.method()} ${new URL(request.url()).pathname}`)
  })
  await page.goto('/')
  await page.getByRole('link', { name: /goal/i }).first().click()
  await expect(page).toHaveURL(/\/plan$/)
  const minutes = page.getByRole('spinbutton', { name: /minutes/i })
  const days = page.getByRole('spinbutton', { name: /days/i })
  const save = page.getByRole('button', { name: /save.*goal/i })
  await minutes.fill('-1')
  await days.fill('8')
  await days.blur()
  await expect(page.getByRole('alert')).toHaveCount(0)
  await save.click()
  await expect(page.getByRole('alert').first()).toBeVisible()
  await expect(minutes).toHaveAttribute('aria-invalid', 'true')
  await minutes.fill('30')
  await days.fill('5')
  await save.click()
  await expect(page.getByRole('status').filter({ hasText: /saved/i })).toBeVisible()
  await page.screenshot({ path: testInfo.outputPath('goal-saved.png'), fullPage: true })
  await page.reload()
  await expect(minutes).toHaveValue('30')
  await expect(days).toHaveValue('5')
  await minutes.fill('45')
  await save.click()
  await page.reload()
  await expect(minutes).toHaveValue('45')
  expect(writes).toEqual([])
})

test('reports unavailable browser storage without claiming the goal was saved', async ({ page }) => {
  await page.addInitScript(() => {
    Storage.prototype.setItem = () => { throw new DOMException('Storage is unavailable', 'QuotaExceededError') }
  })
  await page.goto('/plan')
  await page.getByRole('spinbutton', { name: /minutes/i }).fill('30')
  await page.getByRole('spinbutton', { name: /days/i }).fill('5')
  await page.getByRole('button', { name: /save.*goal/i }).click()
  await expect(page.getByRole('status')).toContainText(/allow saving/i)
  await expect(page.getByRole('status')).not.toContainText('Goal saved')
})

test('returns a useful not-found page for an unknown route', async ({ page }) => {
  const response = await page.goto('/this-route-does-not-exist')
  expect(response?.status()).toBe(404)
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  await page.getByRole('link', { name: /home/i }).first().click()
  await expect(page).toHaveURL(/\/$/)
  await expect(page.locator('table')).toBeVisible()
})

test('preserves the dark theme and opens the local goal through mobile navigation', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 390, height: 900 })
  await page.goto('/')
  await page.getByRole('button', { name: 'Use dark theme', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  await expect(page.getByRole('button', { name: 'Use light theme', exact: true })).toBeEnabled()
  await page.evaluate(() => document.fonts.ready)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('homepage-dark-mobile.png'), fullPage: true })
  await page.getByRole('button', { name: 'Open menu', exact: true }).click()
  await page.getByRole('link', { name: 'Daily goal', exact: true }).click()
  await expect(page).toHaveURL(/\/plan$/)
  await expect(page.getByRole('spinbutton', { name: /minutes/i })).toBeEnabled()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('goal-dark-mobile.png'), fullPage: true })
})

test('keeps client-only mobile controls disabled before hydration', async ({ browser }, testInfo) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 900 } })
  const page = await context.newPage()
  await page.goto(testInfo.project.use.baseURL || 'http://127.0.0.1:5183')
  await expect(page.getByRole('button', { name: 'Use dark theme', exact: true })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'Open menu', exact: true })).toBeDisabled()
  await page.screenshot({ path: testInfo.outputPath('homepage-before-hydration.png'), fullPage: true })
  await context.close()
})
