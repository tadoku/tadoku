import { createServer } from 'node:http'
import { test, expect } from '@playwright/test'

const fixtureEnabled = process.env.AUTHV2_SESSION_FIXTURE === 'true'
let status = 200
let session: unknown = { account_app: 'v2', account: null }
const requests: { cookie?: string; branch?: string; method?: string; path?: string }[] = []
const fixture = createServer((request, response) => {
  if (request.url !== '/api/internal/identity/session') {
    response.writeHead(404).end()
    return
  }
  requests.push({ cookie: request.headers.cookie, branch: request.headers['x-dev-branch']?.toString(), method: request.method, path: request.url })
  response.writeHead(status, { 'Content-Type': 'application/json' }).end(JSON.stringify(session))
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
    for (const [code, body] of [[404, {}], [503, {}], [200, { account_app: 'unknown' }]] as const) {
      status = code
      session = body
      await page.goto('/login')
      await expect.soft(page.getByRole('heading', { level: 1, name: 'We can’t reach your account right now' })).toBeVisible()
      await expect.soft(page.getByRole('link', { name: 'Try again' })).toHaveAttribute('href', '/login')
      await expect.soft(page.getByLabel('Email')).toHaveCount(0)
      expect((await request.get('/healthz')).status()).toBe(200)
    }
  })
})
