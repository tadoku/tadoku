import { expect, test, type Page } from '@playwright/test'

const base = process.env.WEBV3_BASE_URL
const email = process.env.E2E_READER_EMAIL
const password = process.env.E2E_READER_PASSWORD
const secondEmail = process.env.E2E_SECOND_EMAIL
const secondPassword = process.env.E2E_SECOND_PASSWORD

test.use({ trace: 'off', video: 'off', ignoreHTTPSErrors: false })

async function login(page: Page, identity = 'reader') {
  await page.locator('input[name="identifier"]').fill(identity === 'reader' ? email! : secondEmail!)
  await page.locator('input[name="password"]').fill(identity === 'reader' ? password! : secondPassword!)
  await page.locator('button[type="submit"]').click()
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

test('login, exact return, SSR reload, foreign-origin refusal and logout invalidate the session', async ({ page, context }, testInfo) => {
  test.skip(!base || !email || !password, 'Set development fixture credentials')
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
  await page.screenshot({ path: testInfo.outputPath('signed-in.png'), fullPage: true })
  const oldCookie = (await context.cookies()).find(cookie => cookie.name === 'ory_kratos_session')!
  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await expect(page).toHaveURL(`${base}/`)
  await expect(page.getByRole('link', { name: 'Log in', exact: true })).toBeVisible()
  const who = await context.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami', { headers: { Cookie: `${oldCookie.name}=${oldCookie.value}` } })
  expect(who.status()).toBe(401)
})

test('expired cookie renders anonymous', async ({ context, page }) => {
  test.skip(!base, 'Set WEBV3_BASE_URL')
  await context.addCookies([{ name: 'ory_kratos_session', value: 'expired', url: base! }])
  await page.goto('/plan')
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
