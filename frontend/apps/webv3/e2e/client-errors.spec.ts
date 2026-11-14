import type { Page, Request } from '@playwright/test'
import { expect, hydrated, test, webv3 } from './logging-fixtures'

const endpoint = '/client-error'

/** Every report the page sends from now on. */
function reports(page: Page) {
  const sent: Request[] = []
  page.on('request', request => { if (new URL(request.url()).pathname === endpoint) sent.push(request) })
  return sent
}

// `frontend/scripts/smoke-webv3-image.mjs` finds the `client_error` lines these reports write in the container's log.
test('an uncaught error and an unhandled rejection in the browser are each reported once, without the query', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  const sent = reports(page)
  await page.goto('/?token=e2e-secret')
  await hydrated(page)

  await page.evaluate(() => { setTimeout(() => { throw new Error('e2e client error') }) })
  await expect.poll(() => sent.length).toBe(1)
  await page.evaluate(() => { void Promise.reject(new Error('e2e client rejection')) })
  await expect.poll(() => sent.length).toBe(2)

  expect(sent.map(request => request.postDataJSON())).toMatchObject([
    { kind: 'error', message: 'e2e client error', path: '/', stack: expect.any(String), build: expect.any(String) },
    { kind: 'unhandledrejection', message: 'e2e client rejection', path: '/' },
  ])
  for (const request of sent) {
    expect(request.postData()).not.toContain('e2e-secret')
    expect((await request.response())?.status()).toBe(204)
  }
})

test('the report endpoint answers 400 to a report it cannot accept, never a 5xx', async ({ request }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  const valid = { kind: 'error', message: 'e2e rejected report', path: '/', build: 'e2e' }
  const invalid = [
    'not json',
    '[]',
    JSON.stringify({ ...valid, kind: 'other' }),
    JSON.stringify({ ...valid, path: 'https://tadoku.app/' }),
    JSON.stringify({ ...valid, extra: true }),
    JSON.stringify({ ...valid, message: 'x'.repeat(20_000) }),
    'x'.repeat(1_000_000),
  ]
  for (const data of invalid) {
    const response = await request.post(endpoint, { data, headers: { 'content-type': 'text/plain' } })
    expect(response.status(), data.slice(0, 40)).toBe(400)
  }
})
