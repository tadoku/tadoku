import { expect, test } from '@playwright/test'

const app = process.env.BRANCH_APP_URL
const auth = process.env.BRANCH_AUTH_URL
const email = process.env.E2E_READER_EMAIL
const password = process.env.E2E_READER_PASSWORD

test.use({ ignoreHTTPSErrors: false, trace: 'off', video: 'off', screenshot: 'off' })

test('anonymous navigation preserves the full return URL in login and signup links', async ({ page }) => {
  test.skip(!app || !auth, 'Set development branch URLs')
  expect(new URL(app!).hostname.endsWith('.tadoku.dev.lab')).toBe(true)
  const returnUrl = `${app}/pages/manual?source=next&note=reading%20%26%20listening#entry`
  await page.goto(returnUrl)
  for (const name of ['Log in', 'Sign up']) {
    const link = page.getByRole('link', { name, exact: true }).first()
    await expect.poll(() => link.getAttribute('href')).not.toBe(`${auth}/${name === 'Log in' ? 'login' : 'register'}?return_to=`)
    expect(new URL((await link.getAttribute('href'))!).searchParams.get('return_to')).toBe(returnUrl)
  }
})

for (const destination of ['logs', 'registration']) {
  test(`login returns to the protected ${destination} page with its complete URL`, async ({ page }, testInfo) => {
    test.skip(!app || !auth || !email || !password, 'Set branch URLs and development reader credentials')
    expect(new URL(app!).hostname.endsWith('.tadoku.dev.lab')).toBe(true)
    expect(new URL(auth!).hostname.endsWith('.account.tadoku.dev.lab')).toBe(true)

    let path = '/logs/new'
    if (destination === 'registration') {
      const response = await page.request.get(`${app}/api/internal/immersion/contests/latest-official`)
      expect(response.status()).toBe(200)
      expect(response.headers()['x-dev-selected']).toBe(new URL(app!).hostname.split('.')[0])
      path = `/contests/${(await response.json()).id}/registration`
    }
    const returnUrl = `${app}${path}?source=next&note=reading%20%26%20listening#entry`
    await page.goto(returnUrl)
    await page.waitForURL(url => url.origin === auth && url.pathname === '/login')
    expect(new URL(page.url()).searchParams.get('return_to')).toBe(returnUrl)

    await page.locator('input[name="identifier"]').fill(email!)
    await page.locator('input[name="password"]').fill(password!)
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(returnUrl)
    await expect(page.getByRole('heading', { name: 'Settings', exact: true })).toHaveCount(0)
    await page.reload()
    await expect(page).toHaveURL(returnUrl)
    await expect(page).toHaveTitle(destination === 'logs' ? 'New log - Tadoku' : 'Contest registration - Tadoku')
    await page.screenshot({ path: testInfo.outputPath(`${destination}-after-login.png`), fullPage: true })
  })
}
