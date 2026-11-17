import { createServer } from 'node:http'
import { test, expect, type Page } from '@playwright/test'
import { openSettingsTab, settingRow, stubAvatars, waitForHydration } from './support'

// Failure modes: a theme choice that doesn't apply, isn't shared with other Tadoku pages through the cookie or is
// lost on reload; System that ignores the device or stops following it; a picker that shows no choice or can't change
// it until the page hydrates, as on a slow phone, or without JavaScript; tabs that hide the row a URL opened, or that
// need JavaScript; Delete account outside Sign-in and security. Runs against the settings fixture, as `settings-email.spec.ts` describes.
const enabled = process.env.AUTHV2_SETTINGS_FIXTURE === 'true'
const session = '{"account":{"authenticated_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z",'
  + '"display_name":"Reader","email":"reader@example.test","id":"00000000-0000-4000-8000-000000000001"},"recent_login":true}'

const fixture = createServer((request, response) => {
  if (request.url !== '/api/internal/identity/session') return response.writeHead(404).end()
  response.writeHead(200, { 'Content-Type': 'application/json' }).end(session)
})

test.beforeAll(async () => {
  if (enabled) await new Promise<void>(resolve => fixture.listen(5195, '127.0.0.1', resolve))
})

test.afterAll(async () => {
  if (enabled) await new Promise<void>((resolve, reject) => fixture.close(error => error ? reject(error) : resolve()))
})

test.beforeEach(async ({ page }) => {
  test.skip(!enabled, 'needs the local settings fixture')
  await stubAvatars(page)
})

const theme = (page: Page) => page.locator('html')
const cookie = async (page: Page) => (await page.context().cookies()).find(entry => entry.name === 'tadoku-theme')?.value

test('Appearance picks System, Light or Dark and every Tadoku page follows it', async ({ page }, testInfo) => {
  await page.emulateMedia({ colorScheme: 'light' })
  await page.goto('/settings')
  await waitForHydration(page)
  await expect(settingRow(page, 'Display name')).toBeVisible()
  await expect(settingRow(page, 'Email')).toHaveCount(0)

  await openSettingsTab(page, 'Appearance')
  await expect.poll(() => new URL(page.url()).search).toBe('?tab=appearance')
  const choice = (name: string) => page.getByRole('radio', { name, exact: true })
  const pick = (name: string) => page.locator('label').filter({ has: choice(name) }).click()
  await expect(choice('System')).toBeChecked()
  await expect(theme(page)).toHaveAttribute('data-theme', 'light')
  // Once hydrated, a choice applies as it is picked.
  await expect(page.getByRole('button', { name: 'Save theme', exact: true })).toHaveCount(0)

  await pick('Dark')
  await expect(theme(page)).toHaveAttribute('data-theme', 'dark')
  expect(await cookie(page)).toBe('dark')
  await page.screenshot({ path: testInfo.outputPath('appearance-dark.png'), fullPage: true })
  await page.reload()
  await waitForHydration(page)
  await expect(theme(page)).toHaveAttribute('data-theme', 'dark')
  await expect(choice('Dark')).toBeChecked()

  await pick('System')
  await expect(theme(page)).toHaveAttribute('data-theme', 'light')
  expect(await cookie(page)).toBe('system')
  await page.emulateMedia({ colorScheme: 'dark' })
  await expect(theme(page)).toHaveAttribute('data-theme', 'dark')

  await pick('Light')
  await expect(theme(page)).toHaveAttribute('data-theme', 'light')
  expect(await cookie(page)).toBe('light')
  await page.screenshot({ path: testInfo.outputPath('appearance-light.png'), fullPage: true })
})

test('a URL that opens or confirms a row shows its tab, and tabs work without JavaScript', async ({ browser }, testInfo) => {
  const context = await browser.newContext({ javaScriptEnabled: false, baseURL: testInfo.project.use.baseURL })
  const page = await context.newPage()
  await stubAvatars(page)

  await page.goto('/settings?edit=password')
  await expect(page.getByRole('link', { name: 'Sign-in and security', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(settingRow(page, 'Password').getByLabel('New password')).toBeVisible()
  await page.goto('/settings?saved=display-name')
  await expect(settingRow(page, 'Display name').getByRole('status')).toHaveText('Display name saved.')

  await expect(page.getByRole('link', { name: 'Delete account', exact: true })).toHaveCount(0)

  await openSettingsTab(page, 'Sign-in and security')
  await expect(settingRow(page, 'Email')).toBeVisible()
  await expect(settingRow(page, 'Display name')).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Delete account', exact: true })).toHaveAttribute('href', '/delete-account')
  await page.screenshot({ path: testInfo.outputPath('security-no-js.png'), fullPage: true })
  await openSettingsTab(page, 'Appearance')
  await expect(page.getByRole('link', { name: 'Delete account', exact: true })).toHaveCount(0)
  await context.close()
})

test('without JavaScript, Appearance shows the stored choice and saves a new one', async ({ browser }, testInfo) => {
  const baseURL = testInfo.project.use.baseURL!
  const context = await browser.newContext({ javaScriptEnabled: false, baseURL })
  await context.addCookies([{ name: 'tadoku-theme', value: 'dark', url: baseURL }])
  const page = await context.newPage()
  await stubAvatars(page)
  const choice = (name: string) => page.getByRole('radio', { name, exact: true })

  await page.goto('/settings?tab=appearance')
  await expect(choice('Dark')).toBeChecked()
  await expect(choice('Light')).toBeEnabled()
  await page.locator('label').filter({ has: choice('Light') }).click()
  await page.getByRole('button', { name: 'Save theme', exact: true }).click()

  await expect.poll(() => new URL(page.url()).searchParams.get('tab')).toBe('appearance')
  await expect(choice('Light')).toBeChecked()
  expect(await cookie(page)).toBe('light')
  await page.screenshot({ path: testInfo.outputPath('appearance-no-js.png'), fullPage: true })
  await context.close()
})
