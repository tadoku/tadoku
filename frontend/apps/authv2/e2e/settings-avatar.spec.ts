import { createServer } from 'node:http'
import { test, expect, type Page, type Route } from '@playwright/test'
import { avatarImage, expectContained, settingRow, waitForHydration } from './support'

// Failure modes: the header showing Webv3's site navigation instead of the account header (brand, Account, name, Log
// out), or overflowing a phone; the Avatar row missing, pointing at another account's image or missing its Gravatar
// instructions; Fetch latest not refreshing through the API, without JavaScript or once hydrated, or not keeping
// return_to; the page still showing the old image after a fetch; an API failure or ended session not shown in the
// Avatar row. Runs on the Settings fixture's server (see e2e/README.md). The browser's calls to the web host's API are
// answered in the test; the native route's call reaches the fixture below.
const enabled = process.env.AUTHV2_SETTINGS_FIXTURE === 'true'
const host = 'account.tadoku.dev.lab'
const id = '00000000-0000-4000-8000-000000000001'
const avatar = `https://tadoku.dev.lab/api/internal/immersion/users/${id}/avatar`
const refresh = '/api/internal/immersion/profile/avatar/refresh'
const destination = 'https://tadoku.dev.lab/pages/dev-welcome?w2=avatar'
const session = JSON.stringify({
  account: { id, email: 'reader@example.test', display_name: 'Reader One', created_at: '2026-01-01T00:00:00Z', authenticated_at: '2026-01-01T00:00:00Z' },
  recent_login: true,
})

let refreshStatus = 204
const forwarded: { method?: string; origin?: string; cookie?: string }[] = []
const fixture = createServer((request, response) => {
  if (request.url === '/api/internal/identity/session') {
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(session)
    return
  }
  if (request.url === refresh) {
    forwarded.push({ method: request.method, origin: request.headers.origin, cookie: request.headers.cookie })
    response.writeHead(refreshStatus).end()
    return
  }
  response.writeHead(404).end()
})

test.beforeAll(async () => {
  if (enabled) await new Promise<void>(resolve => fixture.listen(5195, '127.0.0.1', resolve))
})

test.afterAll(async () => {
  if (enabled) await new Promise<void>((resolve, reject) => fixture.close(error => error ? reject(error) : resolve()))
})

test.beforeEach(() => {
  test.skip(!enabled, 'needs the local settings fixture')
})

test.use({ actionTimeout: 10_000 })

/** Answers the browser's avatar and refresh calls, recording each. */
async function answerBrowser(page: Page, status = 204) {
  const calls: { method: string; url: string }[] = []
  const record = (route: Route) => calls.push({ method: route.request().method(), url: route.request().url() })
  await page.route(`${avatar}*`, async route => {
    record(route)
    await route.fulfill({ status: 200, contentType: 'image/png', headers: { 'Cache-Control': 'public, max-age=86400' }, body: avatarImage })
  })
  await page.route(`https://tadoku.dev.lab${refresh}`, async route => {
    const origin = await route.request().headerValue('origin') ?? '*'
    const cors = { 'Access-Control-Allow-Origin': origin, 'Access-Control-Allow-Credentials': 'true' }
    if (route.request().method() === 'OPTIONS') return route.fulfill({ status: 204, headers: cors })
    record(route)
    return route.fulfill({ status, headers: cors })
  })
  return calls
}

const backgroundUrl = (page: Page) => page.locator('#avatar-label + div [style*="background-image"]')
  .evaluate(element => (element as HTMLElement).style.backgroundImage)

test('the header is the account header, at desktop and phone widths', async ({ page }) => {
  await answerBrowser(page)
  await page.goto('/settings')
  const header = page.getByRole('banner')
  await expect(header.getByRole('link', { name: /Tadoku/ })).toHaveAttribute('href', 'https://tadoku.dev.lab')
  await expect(header).toContainText('Account')
  await expect(header).toContainText('Reader One')
  await expect(header.getByRole('button', { name: 'Log out', exact: true })).toBeVisible()
  for (const name of ['Home', 'Leaderboard', 'Contests', 'Blog', 'Guide']) {
    await expect(header.getByRole('link', { name, exact: true })).toHaveCount(0)
  }
  await expect(header.getByRole('button', { name: /menu/i })).toHaveCount(0)
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(header.getByRole('button', { name: 'Log out', exact: true })).toBeVisible()
  await expectContained(page)
  await page.screenshot({ path: test.info().outputPath('header-390.png') })
})

test('the Avatar row shows the account\'s Gravatar with instructions', async ({ page }) => {
  const calls = await answerBrowser(page)
  await page.goto('/settings')
  const row = settingRow(page, 'Avatar')
  await expect(row).toContainText('RO')
  await expect(row.getByRole('link', { name: 'Gravatar' })).toHaveAttribute('href', 'https://gravatar.com')
  await expect(row).toContainText('within a day')
  await expect(row.getByRole('button', { name: 'Fetch latest avatar' })).toBeVisible()
  expect(await backgroundUrl(page)).toBe(`url("${avatar}")`)
  await expect.poll(() => calls.filter(call => call.method === 'GET').length).toBeGreaterThan(0)
  await expectContained(page)
  await page.setViewportSize({ width: 390, height: 844 })
  await expectContained(page)
  await page.screenshot({ path: test.info().outputPath('settings-390.png'), fullPage: true })
})

test('a hydrated Fetch latest refreshes through the API and reloads the image', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()) })
  const calls = await answerBrowser(page)
  await page.goto(`/settings?return_to=${encodeURIComponent(destination)}`)
  await waitForHydration(page)
  const before = calls.length

  await settingRow(page, 'Avatar').getByRole('button', { name: 'Fetch latest avatar' }).click()
  await expect(settingRow(page, 'Avatar').getByRole('status')).toHaveText('Fetched your latest avatar.')
  expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ saved: 'avatar', return_to: destination })
  const after = calls.slice(before)
  expect(after[0]).toMatchObject({ method: 'POST', url: `https://tadoku.dev.lab${refresh}` })
  expect(after.some(call => call.method === 'GET' && call.url === avatar), 'the plain URL is reloaded').toBe(true)
  expect(await backgroundUrl(page)).toMatch(new RegExp(`^url\\("${avatar}\\?v=\\d+"\\)$`))
  expect(errors).toEqual([])
  await page.screenshot({ path: test.info().outputPath('fetched.png') })
})

test('a hydrated Fetch latest shows a failure in the Avatar row', async ({ page }) => {
  await answerBrowser(page, 503)
  await page.goto('/settings')
  await waitForHydration(page)
  await settingRow(page, 'Avatar').getByRole('button', { name: 'Fetch latest avatar' }).click()
  await expect(settingRow(page, 'Avatar').getByRole('alert')).toContainText('Something went wrong on our side')
  await expect(page.getByRole('status')).toHaveCount(0)
})

test('without JavaScript, Fetch latest refreshes through the native route', async ({ browser, request }, testInfo) => {
  const context = await browser.newContext({ baseURL: testInfo.project.use.baseURL, javaScriptEnabled: false })
  const page = await context.newPage()
  await answerBrowser(page)
  let posted: { path: string; type?: string; body: string } | undefined
  await page.route('**/actions/*', async route => {
    const submitted = route.request()
    posted = { path: new URL(submitted.url()).pathname, type: await submitted.headerValue('content-type') ?? undefined, body: submitted.postData() ?? '' }
    await route.fulfill({ status: 204 })
  })
  const replay = async () => {
    const response = await request.post(posted!.path, {
      data: posted!.body,
      headers: { Host: host, Origin: `https://${host}`, 'Content-Type': posted!.type ?? '', Cookie: 'tadoku_session_fixture=visitor' },
      maxRedirects: 0,
    })
    expect(response.status()).toBe(303)
    return new URL(response.headers().location)
  }

  await page.goto(`/settings?return_to=${encodeURIComponent(destination)}`)
  await settingRow(page, 'Avatar').getByRole('button', { name: 'Fetch latest avatar' }).click()
  await expect.poll(() => posted?.path).toBe('/actions/avatar')
  forwarded.length = 0
  refreshStatus = 204
  const shownAfterFetch = async () => {
    const saved = await replay()
    expect(saved.pathname).toBe('/')
    const { fetched, ...search } = Object.fromEntries(saved.searchParams)
    expect(search).toEqual({ saved: 'avatar', return_to: destination })
    await page.goto(`/settings${saved.search}`)
    await expect(settingRow(page, 'Avatar').getByRole('status')).toHaveText('Fetched your latest avatar.')
    expect(await backgroundUrl(page)).toBe(`url("${avatar}?v=${fetched}")`)
    return fetched
  }
  const first = await shownAfterFetch()
  expect(forwarded).toEqual([{ method: 'POST', origin: `https://${host}`, cookie: 'tadoku_session_fixture=visitor' }])
  // A second fetch must not reuse the first one's image URL, which the browser has cached for a day.
  await new Promise(resolve => setTimeout(resolve, 5))
  expect(await shownAfterFetch()).not.toBe(first)

  for (const [status, message] of [[503, 'Something went wrong on our side'], [401, 'Your session has ended']] as const) {
    refreshStatus = status
    const failed = await replay()
    expect(Object.fromEntries(failed.searchParams)).toMatchObject({ form: 'avatar', return_to: destination })
    await page.goto(`/settings${failed.search}`)
    await expect(settingRow(page, 'Avatar').getByRole('alert')).toContainText(message)
    await expect(settingRow(page, 'Display name').getByRole('alert')).toHaveCount(0)
  }
  await context.close()
})
