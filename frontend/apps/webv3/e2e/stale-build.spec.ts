import type { Page } from '@playwright/test'
import { expect, hydrated, test, webv3 } from './logging-fixtures'

/** A server function ID no build has, as a tab loaded from an earlier release would send after a rename. */
const unknownId = 'f'.repeat(64)

/** The headers TanStack Start's client sends with every server function call. */
const clientHeaders = { 'x-tsr-serverFn': 'true', accept: 'application/x-tss-framed, application/x-ndjson, application/json' }

/** Every document the page loads from now on; a client navigation loads none. */
function documentLoads(page: Page) {
  const urls: string[] = []
  page.on('request', request => { if (request.resourceType() === 'document') urls.push(request.url()) })
  return urls
}

/** Sends every server function call the page makes from now on to the unknown ID, as a tab from an earlier release would. */
const staleTab = (page: Page) => page.route('**/_serverFn/**', route => {
  const url = new URL(route.request().url())
  url.pathname = `/_serverFn/${unknownId}`
  return route.continue({ url: url.toString() })
})

test('a call to an unknown server function answers 409 marking the build as outdated, never a 5xx', async ({ request }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  for (const method of ['GET', 'POST'] as const) {
    const response = await request.fetch(`/_serverFn/${unknownId}`, {
      method, headers: { ...clientHeaders, ...method === 'POST' ? { 'content-type': 'application/json' } : {} },
      data: method === 'POST' ? '{}' : undefined,
    })
    expect(response.status(), method).toBe(409)
    expect(response.headers()['x-tadoku-outdated-build'], method).toBe('true')
    expect(response.headers()['cache-control'], method).toContain('no-store')
  }
})

test('a tab from an earlier build reloads once when a server function is unknown, and never again in a loop', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/')
  await hydrated(page)
  await staleTab(page)
  const documents = documentLoads(page)
  const leaderboard = page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Leaderboard', exact: true })

  // The leaderboard's loader calls a server function from the browser; the unknown ID makes the tab load the page anew.
  await leaderboard.first().click()
  await expect.poll(() => documents.length).toBe(1)
  await expect(page).toHaveURL(/\/leaderboard\/latest\/1$/)
  await hydrated(page)

  // The reloaded tab still meets an unknown ID, so it shows the failure instead of reloading again.
  const outdated = page.waitForResponse(response => response.url().includes('/_serverFn/') && response.status() === 409)
  const contests = page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Contests', exact: true })
  await contests.first().click()
  await outdated
  await expect(page.getByRole('heading', { name: 'This page couldn’t load' })).toBeVisible()
  await page.waitForTimeout(1_000)
  expect(documents).toHaveLength(1)
})
