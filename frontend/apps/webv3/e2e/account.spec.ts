import type { BrowserContext } from '@playwright/test'
import { admin, api, authenticate, expect, logOut, test, webv3 as base } from './logging-fixtures'

const branch = base ? new URL(base).hostname.includes('.next.') : false

/** The account host's URL for `path` that returns to `returnTo`. */
function accountHref(loginHref: string, path: string, returnTo: string) {
  return `${new URL(loginHref).origin}${path}?return_to=${encodeURIComponent(returnTo)}`
}

/** The tenant's session cookie: `tadoku_session` canonically, `tadoku_session_<tenant>` on a branch. */
async function sessionCookie(context: BrowserContext) {
  return (await context.cookies())
    .find(cookie => branch ? cookie.name.startsWith('tadoku_session_') : cookie.name === 'tadoku_session')
}

test('anonymous login return is SSR stable and includes the fragment after hydration', async ({ browser, page }) => {
  test.skip(!base, 'Set WEBV3_BASE_URL')
  const destination = `${base}/pages/manual?source=next&note=reading%20%26%20listening#entry`
  const context = await browser.newContext({ javaScriptEnabled: false })
  const ssr = await context.newPage()
  const response = await ssr.goto(destination)
  expect(response?.headers()['cache-control']).toContain('no-store')
  const link = ssr.getByRole('link', { name: 'Log in', exact: true })
  const ssrReturn = destination.replace(/^http:\/\//, 'https://').split('#')[0]
  expect(new URL((await link.getAttribute('href'))!).searchParams.get('return_to')).toBe(ssrReturn)
  const loginHref = (await link.getAttribute('href'))!
  await expect(ssr.getByRole('link', { name: 'Sign up', exact: true })).toHaveAttribute('href', accountHref(loginHref, '/register', ssrReturn))
  await context.close()
  await page.goto(destination)
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(destination)
  await page.getByRole('link', { name: 'Skip to content' }).focus()
  await page.keyboard.press('Enter')
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(`${destination.split('#')[0]}#main-content`)
})

test('login, exact return, SSR reload, foreign-origin refusal and logout invalidate the session', async ({ page, context, playwright, loginLearner }, testInfo) => {
  const { name } = loginLearner
  const destination = `${base}/pages/manual?source=next&note=reading%20%26%20listening#entry`
  await page.goto(destination)
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(destination)
  await page.getByRole('link', { name: 'Log in', exact: true }).click()
  await authenticate(page, loginLearner)
  await expect(page).toHaveURL(destination)
  await expect(page.getByRole('button', { name, exact: true })).toBeVisible()
  await page.reload()
  await expect(page.getByRole('button', { name, exact: true })).toBeVisible()
  const rejected = await context.request.post(`${base}/logout`, { headers: { Origin: 'https://evil.tadoku.dev.lab', 'Sec-Fetch-Site': 'same-site' } })
  expect(rejected.status()).toBe(403)
  await page.reload()
  await expect(page.getByRole('button', { name, exact: true })).toBeVisible()
  if (branch) {
    // The branch API selects its own tenant's cookie, so a canonical session cookie beside it changes nothing.
    const domain = new URL(base!).hostname.split('.next.')[1]
    await context.addCookies([{ name: 'tadoku_session', value: 'canonical', domain: `.${domain}`, path: '/', secure: true, sameSite: 'Lax' }])
    await page.reload()
    await expect(page.getByRole('button', { name, exact: true })).toBeVisible()
  }
  await page.screenshot({ path: testInfo.outputPath('signed-in.png'), fullPage: true })

  // Webv2's account menu: profile, account settings returning here, the theme switch, and no admin entry for a learner.
  await expect(page.getByRole('button', { name, exact: true })).toBeEnabled()
  await page.getByRole('button', { name, exact: true }).click()
  const menu = page.getByRole('menu')
  await expect(menu.getByRole('menuitem')).toHaveText(['Profile', 'Account settings', 'Use dark theme', 'Log out'])
  await page.screenshot({ path: testInfo.outputPath('account-menu.png') })
  await menu.getByRole('menuitem', { name: 'Profile', exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`^${base}/profile/[a-f0-9-]{36}/statistics$`))
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible()
  await page.goto(destination)
  await page.setViewportSize({ width: 390, height: 900 })
  await page.getByRole('button', { name: 'Open menu', exact: true }).click()
  const mobile = page.getByRole('region', { name: `${name} account` })
  const settings = new URL((await mobile.getByRole('link', { name: 'Account settings', exact: true }).getAttribute('href'))!)
  expect([settings.hostname, settings.pathname, settings.searchParams.get('return_to')])
    .toEqual([new URL(base!).hostname.replace(/(^|\.)next\./, '$1account.'), '/settings', destination])
  await expect(mobile.getByRole('link', { name: 'Admin', exact: true })).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('account-menu-390.png'), fullPage: true })
  await page.setViewportSize({ width: 1280, height: 720 })

  const oldCookie = (await sessionCookie(context))!
  await logOut(page, loginLearner)
  await expect(page).toHaveURL(`${base}/`)
  await expect(page.getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
  expect(await sessionCookie(context)).toBeUndefined()
  const anonymous = await playwright.request.newContext({ extraHTTPHeaders: { Cookie: `${oldCookie.name}=${oldCookie.value}` } })
  const who = await anonymous.get(`${api}/identity/session`)
  expect(who.status()).toBe(200)
  expect((await who.json()).account).toBeNull()
  await anonymous.dispose()
})

test('an expired session cookie renders anonymous', async ({ context, page, loginLearner }) => {
  await page.goto('/pages/manual')
  await page.getByRole('link', { name: 'Log in', exact: true }).click()
  await authenticate(page, loginLearner)
  await expect(page.getByRole('button', { name: loginLearner.name, exact: true })).toBeVisible()
  const cookie = (await sessionCookie(context))!
  await context.addCookies([{ ...cookie, value: 'expired' }])
  await page.reload()
  await expect(page.getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
})

test('thirty interleaved requests isolate two identities and anonymous HTML', async ({ browser, loginLearner, adminContext }) => {
  test.setTimeout(90_000)
  test.skip(!admin.email, 'Set E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD')
  const [learner, anonymous] = await Promise.all([browser.newContext(), browser.newContext()])
  const contexts = [learner, adminContext, anonymous]
  try {
    const page = await learner.newPage()
    await page.goto(`${base}/pages/manual`)
    await page.getByRole('link', { name: 'Log in', exact: true }).click()
    await authenticate(page, loginLearner)
    await expect(page).toHaveURL(`${base}/pages/manual`)
    await Promise.all(Array.from({ length: 30 }, async (_, index) => {
      const kind = index % 3
      const response = await contexts[kind].request.get(`${base}/pages/manual`)
      expect(response.status()).toBe(200)
      expect(response.headers()['cache-control']).toContain('no-store')
      const html = await response.text()
      expect(html.includes(loginLearner.name)).toBe(kind === 0)
      expect(html.includes('Dev Admin')).toBe(kind === 1)
      if (kind === 2) expect(html).toContain('>Log in<')
    }))
    await logOut(page, loginLearner)
    await page.getByRole('link', { name: 'Log in', exact: true }).click()
    await authenticate(page, admin)
    await expect(page.getByRole('button', { name: 'Dev Admin', exact: true })).toBeVisible()
    expect(await page.getByRole('navigation', { name: 'Main navigation' }).textContent()).not.toContain(loginLearner.name)

    // An administrator's menu adds Webv2's Admin entry, which opens the matching admin host.
    await expect(page.getByRole('button', { name: 'Dev Admin', exact: true })).toBeEnabled()
    await page.getByRole('button', { name: 'Dev Admin', exact: true }).click()
    await expect(page.getByRole('menu').getByRole('menuitem')).toHaveText(['Profile', 'Admin', 'Account settings', 'Use dark theme', 'Log out'])
    await page.getByRole('menuitem', { name: 'Admin', exact: true }).click()
    await page.waitForURL(url => url.hostname === new URL(base!).hostname.replace(/(^|\.)next\./, '$1admin.'))
  } finally {
    await Promise.all([learner.close(), anonymous.close()])
  }
})
