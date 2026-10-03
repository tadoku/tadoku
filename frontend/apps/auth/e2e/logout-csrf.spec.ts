import { expect, test } from '@playwright/test'

test('anonymous login creates no logout flow and authenticated logout still works', async ({ page }) => {
  const authUrl = process.env.AUTH_URL
  const email = process.env.E2E_ADMIN_EMAIL
  const password = process.env.E2E_ADMIN_PASSWORD
  test.skip(!authUrl || !email || !password, 'Set development AUTH_URL and administrator credentials')

  const logoutFlows: string[] = []
  page.on('request', request => {
    if (new URL(request.url()).pathname.endsWith('/self-service/logout/browser')) {
      logoutFlows.push(request.url())
    }
  })

  await page.goto(`${authUrl}/login`)
  await expect(page.locator('input[name="identifier"]')).toBeVisible()
  expect(logoutFlows).toEqual([])
  await page.locator('input[name="identifier"]').fill(email!)
  await page.locator('input[name="password"]').fill(password!)
  await page.locator('button[type="submit"]').click()
  await expect(page.getByRole('heading', { name: 'Settings', exact: true })).toBeVisible()
  await expect.poll(() => logoutFlows.length).toBeGreaterThan(0)
  await page.getByRole('button', { name: /Open navigation menu|Dev Admin/ }).click()
  await page.getByRole('menuitem', { name: 'Log out' }).click()
  await expect(page.getByRole('heading', { name: 'Log in', exact: true })).toBeVisible()
  expect((await page.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')).status()).toBe(401)
})
