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
    await expect(page.locator('.paper-navbar__mobile-trigger')).toBeEnabled()
    await page.evaluate(() => document.fonts.ready)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`homepage-${width}.png`), fullPage: true })
  }
  expect(failures).toEqual([])
})

test('the retired local goal page is not found', async ({ page }) => {
  expect((await page.goto('/plan'))?.status()).toBe(404)
})

test('returns a useful not-found page for an unknown route', async ({ page }) => {
  const response = await page.goto('/this-route-does-not-exist')
  expect(response?.status()).toBe(404)
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  await page.getByRole('link', { name: /home/i }).first().click()
  await expect(page).toHaveURL(/\/$/)
  await expect(page.locator('table')).toBeVisible()
})

test('preserves the dark theme and opens the blog through mobile navigation', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 390, height: 900 })
  // The appearance switch belongs to the account menu; a visitor's stored or system choice still applies.
  await page.goto('/')
  await page.evaluate(() => localStorage.setItem('tadoku-theme', 'dark'))
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  await expect(page.getByRole('button', { name: 'Open menu', exact: true })).toBeEnabled()
  await page.evaluate(() => document.fonts.ready)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('homepage-dark-mobile.png'), fullPage: true })
  await page.getByRole('button', { name: 'Open menu', exact: true }).click()
  await page.getByRole('link', { name: 'Blog', exact: true }).first().click()
  await expect(page).toHaveURL(/\/blog\/1$/)
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('blog-dark-mobile.png'), fullPage: true })
})

test('offers the mobile menu as a link before hydration', async ({ browser }, testInfo) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 900 } })
  const page = await context.newPage()
  await page.goto(testInfo.project.use.baseURL || 'http://127.0.0.1:5183')
  await expect(page.getByRole('link', { name: 'Open menu', exact: true })).toBeVisible()
  await page.screenshot({ path: testInfo.outputPath('homepage-before-hydration.png'), fullPage: true })
  await context.close()
})
