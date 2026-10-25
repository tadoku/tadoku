import { expect, test, type BrowserContext, type Page } from '@playwright/test'

const base = process.env.WEBV3_BASE_URL
const email = process.env.E2E_READER_EMAIL
const password = process.env.E2E_READER_PASSWORD
const secondEmail = process.env.E2E_SECOND_EMAIL
const secondPassword = process.env.E2E_SECOND_PASSWORD
const api = process.env.TADOKU_API_URL

test.use({ trace: 'off', video: 'off', ignoreHTTPSErrors: false })

async function login(page: Page, identity = 'reader') {
  await page.getByLabel('Email').fill(identity === 'reader' ? email! : secondEmail!)
  await page.getByLabel('Password').fill(identity === 'reader' ? password! : secondPassword!)
  await page.getByRole('button', { name: 'Log in', exact: true }).click()
}

const branch = base ? new URL(base).hostname.includes('.next.') : false

/** The tenant's session cookie: `tadoku_session` canonically, `tadoku_session_<tenant>` on a branch. */
async function sessionCookie(context: BrowserContext) {
  return (await context.cookies())
    .find(cookie => branch ? cookie.name.startsWith('tadoku_session_') : cookie.name === 'tadoku_session')
}

test('anonymous login return is SSR stable and includes the fragment after hydration', async ({ browser, page }) => {
  test.skip(!base, 'Set WEBV3_BASE_URL')
  const destination = `${base}/plan?source=next&note=reading%20%26%20listening#entry`
  const context = await browser.newContext({ javaScriptEnabled: false })
  const ssr = await context.newPage()
  const response = await ssr.goto(destination)
  expect(response?.headers()['cache-control']).toContain('no-store')
  const link = ssr.getByRole('link', { name: 'Log in', exact: true })
  expect(new URL((await link.getAttribute('href'))!).searchParams.get('return_to')).toBe(destination.replace(/^http:\/\//, 'https://').split('#')[0])
  await context.close()
  await page.goto(destination)
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(destination)
  await page.getByRole('link', { name: 'Skip to content' }).focus()
  await page.keyboard.press('Enter')
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(`${destination.split('#')[0]}#main-content`)
})

test('login, exact return, SSR reload, foreign-origin refusal and logout invalidate the session', async ({ page, context, playwright }, testInfo) => {
  test.skip(!base || !email || !password || !api, 'Set development fixture credentials and TADOKU_API_URL')
  const destination = `${base}/plan?source=next&note=reading%20%26%20listening#entry`
  await page.goto(destination)
  await expect.poll(async () => new URL((await page.getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!).searchParams.get('return_to')).toBe(destination)
  await page.getByRole('link', { name: 'Log in', exact: true }).click()
  await login(page)
  await expect(page).toHaveURL(destination)
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Reader', { exact: true })).toBeVisible()
  await page.reload()
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Reader', { exact: true })).toBeVisible()
  const rejected = await context.request.post(`${base}/logout`, { headers: { Origin: 'https://evil.tadoku.dev.lab', 'Sec-Fetch-Site': 'same-site' } })
  expect(rejected.status()).toBe(403)
  await page.reload()
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Reader', { exact: true })).toBeVisible()
  if (branch) {
    // The branch API selects its own tenant's cookie, so a canonical session cookie beside it changes nothing.
    const domain = new URL(base!).hostname.split('.next.')[1]
    await context.addCookies([{ name: 'tadoku_session', value: 'canonical', domain: `.${domain}`, path: '/', secure: true, sameSite: 'Lax' }])
    await page.reload()
    await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Reader', { exact: true })).toBeVisible()
  }
  await page.screenshot({ path: testInfo.outputPath('signed-in.png'), fullPage: true })
  const oldCookie = (await sessionCookie(context))!
  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await expect(page).toHaveURL(`${base}/`)
  await expect(page.getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
  expect(await sessionCookie(context)).toBeUndefined()
  const anonymous = await playwright.request.newContext({ extraHTTPHeaders: { Cookie: `${oldCookie.name}=${oldCookie.value}` } })
  const who = await anonymous.get(`${api}/identity/session`)
  expect(who.status()).toBe(200)
  expect((await who.json()).account).toBeNull()
  await anonymous.dispose()
})

test('an expired session cookie renders anonymous', async ({ context, page }) => {
  test.skip(!base || !email || !password, 'Set development fixture credentials')
  await page.goto('/plan')
  await page.getByRole('link', { name: 'Log in', exact: true }).click()
  await login(page)
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Reader', { exact: true })).toBeVisible()
  const cookie = (await sessionCookie(context))!
  await context.addCookies([{ ...cookie, value: 'expired' }])
  await page.reload()
  await expect(page.getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
})

test('thirty interleaved requests isolate two identities and anonymous HTML', async ({ browser }) => {
  test.setTimeout(90_000)
  test.skip(!base || !email || !password || !secondEmail || !secondPassword, 'Set both development fixture identities')
  const contexts = await Promise.all([browser.newContext(), browser.newContext(), browser.newContext()])
  try {
    for (const [index, identity] of ['reader', 'second'].entries()) {
      const page = await contexts[index].newPage()
      await page.goto(`${base}/plan`)
      await page.getByRole('link', { name: 'Log in', exact: true }).click()
      await login(page, identity)
      await expect(page).toHaveURL(`${base}/plan`)
    }
    await Promise.all(Array.from({ length: 30 }, async (_, index) => {
      const kind = index % 3
      const response = await contexts[kind].request.get(`${base}/plan`)
      expect(response.status()).toBe(200)
      expect(response.headers()['cache-control']).toContain('no-store')
      const html = await response.text()
      expect(html.includes('Dev Reader')).toBe(kind === 0)
      expect(html.includes('Dev Admin')).toBe(kind === 1)
      if (kind === 2) expect(html).toContain('>Log in<')
    }))
    const page = contexts[0].pages()[0]
    await page.getByRole('button', { name: 'Log out', exact: true }).click()
    await page.getByRole('link', { name: 'Log in', exact: true }).click()
    await login(page, 'second')
    await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Admin', { exact: true })).toBeVisible()
    expect(await page.getByRole('navigation', { name: 'Main navigation' }).textContent()).not.toContain('Dev Reader')
  } finally {
    await Promise.all(contexts.map(context => context.close()))
  }
})
