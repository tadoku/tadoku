import { createServer } from 'node:http'
import { test, expect } from '@playwright/test'
import { holdScripts, waitForHydration } from './support'

const fixtureEnabled = process.env.AUTHV2_SESSION_FIXTURE === 'true'
let status = 200
let contentType = 'application/json'
let session: unknown = { account_app: 'v2', account: null }
const requests: { cookie?: string; branch?: string; method?: string; path?: string }[] = []
const profileUpdates: unknown[] = []
const fixture = createServer((request, response) => {
  const cors = request.headers.origin
    ? { 'Access-Control-Allow-Origin': request.headers.origin, 'Access-Control-Allow-Credentials': 'true' }
    : {}
  if (request.url === '/api/internal/identity/settings/profile') {
    if (request.method === 'OPTIONS') {
      response.writeHead(204, { ...cors, 'Access-Control-Allow-Methods': 'PATCH', 'Access-Control-Allow-Headers': 'content-type' }).end()
      return
    }
    const chunks: Buffer[] = []
    request.on('data', chunk => chunks.push(chunk))
    request.on('end', () => {
      const update = JSON.parse(Buffer.concat(chunks).toString('utf8'))
      profileUpdates.push(update)
      const account = (session as { account: object }).account
      response.writeHead(200, { ...cors, 'Content-Type': 'application/json' }).end(JSON.stringify({ ...account, ...update }))
    })
    return
  }
  if (request.url !== '/api/internal/identity/session') {
    response.writeHead(404).end()
    return
  }
  requests.push({ cookie: request.headers.cookie, branch: request.headers['x-dev-branch']?.toString(), method: request.method, path: request.url })
  response.writeHead(status, { 'Content-Type': contentType }).end(typeof session === 'string' ? session : JSON.stringify(session))
})

test.beforeAll(async () => {
  if (fixtureEnabled) await new Promise<void>((resolve) => fixture.listen(5195, '127.0.0.1', resolve))
})

test.afterAll(async () => {
  if (fixtureEnabled) await new Promise<void>((resolve, reject) => fixture.close(error => error ? reject(error) : resolve()))
})

test('account screen, session routing and unavailable API', async ({ page, request }, testInfo) => {
  await page.goto('/login')
  await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Log in', exact: true })).toBeEnabled()
  await page.setViewportSize({ width: 390, height: 844 })
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('login-mobile.png'), fullPage: true })
  expect((await request.get('/healthz')).status()).toBe(200)

  if (!fixtureEnabled) return

  await test.step('SSR forwards cookies and branch selection to the session endpoint', async () => {
    const response = await request.get('/login', { headers: { Cookie: 'tadoku_session_fixture=synthetic', 'X-Dev-Branch': 'g-smoke' } })
    expect(response.status()).toBe(200)
    expect.soft(requests).toContainEqual({ cookie: 'tadoku_session_fixture=synthetic', branch: 'g-smoke', method: 'GET', path: '/api/internal/identity/session' })
  })

  await test.step('v1 returns to the matching page of the existing account app', async () => {
    session = { account_app: 'v1', account: null }
    for (const [path, location] of [
      ['/login?return_to=%2Flogs%2Fnew', '/login?return_to=%2Flogs%2Fnew'],
      ['/register', '/register'],
      ['/', '/'],
      ['/settings?return_to=https%3A%2F%2Ftadoku.app%2F', '/?return_to=https%3A%2F%2Ftadoku.app%2F'],
      ['/recovery', '/account-recovery'],
    ]) {
      const response = await request.get(path, { maxRedirects: 0 })
      expect.soft(response.status(), path).toBe(307)
      expect.soft(response.headers().location, path).toBe(`http://account.localhost:5195${location}`)
    }
  })

  await test.step('v1 legacy account details do not prevent the redirect', async () => {
    for (const email of ['reader@example.test', 'reader@localhost']) {
      session = {
        account_app: 'v1',
        account: {
          id: '00000000-0000-4000-8000-000000000001',
          email,
          display_name: '😀'.repeat(25),
          created_at: '2026-01-01T00:00:00Z',
          authenticated_at: '2026-01-01T00:00:00Z',
        },
      }
      const response = await request.get('/login?return_to=%2Flogs%2Fnew', { maxRedirects: 0 })
      expect.soft(response.status()).toBe(307)
      expect.soft(response.headers().location).toBe('http://account.localhost:5195/login?return_to=%2Flogs%2Fnew')
    }
  })

  await test.step('a missing, failing or malformed API keeps a visible unavailable state', async () => {
    for (const [code, body, type] of [
      [404, {}, 'application/json'],
      [503, {}, 'application/json'],
      [200, { account_app: 'unknown' }, 'application/json'],
      [200, '<!doctype html><title>Gateway</title>', 'text/html'],
      [200, '', 'application/json'],
    ] as const) {
      status = code
      session = body
      contentType = type
      await page.goto('/login')
      await expect.soft(page.getByRole('heading', { level: 1, name: 'We can’t reach your account right now' })).toBeVisible()
      const retry = page.getByRole('link', { name: 'Try again' })
      await expect.soft(retry).toHaveAttribute('href', '/login')
      const label = await retry.evaluate(link => [getComputedStyle(link).color, getComputedStyle(link).backgroundColor])
      expect.soft(label[0], 'retry label differs from its background').not.toBe(label[1])
      await expect.soft(page.getByLabel('Email')).toHaveCount(0)
      expect((await request.get('/healthz')).status()).toBe(200)
    }
  })
})

const account = (id: string, email: string) => ({
  account_app: 'v2',
  account: { id, email, display_name: 'Reader', created_at: '2026-01-01T00:00:00Z', authenticated_at: '2026-01-01T00:00:00Z' },
})

test('Settings shows and resubmits stored account values unchanged', async ({ page, request }) => {
  test.skip(!fixtureEnabled, 'needs the local session fixture')
  status = 200
  contentType = 'application/json'

  await test.step('any well-formed account ID reaches Settings; a malformed one stays unavailable', async () => {
    // The import keeps source IDs without checking their UUID version, so a version-0 ID is a real stored value.
    for (const [id, ready] of [
      ['01000000-0000-0000-0000-000000000000', true],
      ['00000000-0000-4000-8000-000000000001', true],
      ['00000000-0000-8000-8000-000000000001', true],
      ['not-a-uuid', false],
    ] as const) {
      session = account(id, 'reader@example.test')
      const html = await (await request.get('/settings')).text()
      expect.soft(html.includes('Change email'), id).toBe(ready)
      expect.soft(html.includes('We can’t reach your account right now'), id).toBe(!ready)
    }
  })

  // Imported addresses keep their raw stored form: a long ideographic-space prefix, and 321 bytes that only fit
  // the 320-byte limit once Go lowercases them.
  for (const [name, email] of [
    ['long ideographic-space prefix', '\u3000'.repeat(350) + 'reader@example.test'],
    ['short ideographic-space prefix', '\u3000reader@example.test'],
    ['Kelvin sign at 321 bytes', 'a'.repeat(316) + '\u212A@x'],
    ['dotted capital I at 321 bytes', 'a'.repeat(317) + '\u0130@x'],
  ]) {
    await test.step(`${name}: server-rendered, hydrated and submitted unchanged`, async () => {
      session = account('00000000-0000-4000-8000-000000000001', email)
      const html = await (await request.get('/settings')).text()
      expect.soft(html.includes('\u0000'), 'server HTML has no NUL').toBe(false)
      expect.soft(html.includes('\uFFFD'), 'server HTML has no replacement character').toBe(false)
      expect.soft(html.includes(`value="${email}"`), 'server-rendered email input').toBe(true)

      await page.goto('/settings')
      await waitForHydration(page)
      await expect.soft(page.getByRole('textbox', { name: 'Email', exact: true })).toHaveValue(email)
      profileUpdates.length = 0
      await page.getByRole('button', { name: 'Change email' }).click()
      await expect.soft(page.getByText('Email changed.')).toBeVisible()
      expect.soft(profileUpdates).toEqual([{ email }])
    })
  }

  await test.step('an edit typed before hydration is kept and submitted', async () => {
    session = account('00000000-0000-4000-8000-000000000001', '\u3000'.repeat(350) + 'reader@example.test')
    const scripts = await holdScripts(page)
    await page.goto('/settings', { waitUntil: 'domcontentloaded' })
    await page.getByRole('textbox', { name: 'Email', exact: true }).fill('edited@example.test')
    await scripts.release()
    await expect.soft(page.getByRole('textbox', { name: 'Email', exact: true })).toHaveValue('edited@example.test')
    profileUpdates.length = 0
    await page.getByRole('button', { name: 'Change email' }).click()
    await expect.soft(page.getByText('Email changed.')).toBeVisible()
    expect.soft(profileUpdates).toEqual([{ email: 'edited@example.test' }])
  })
})
