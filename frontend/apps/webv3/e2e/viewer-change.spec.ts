import type { Page } from '@playwright/test'
import { admin, api, expect, hydrated, reader, signIn, test, webv3 } from './logging-fixtures'

const errorStates = ['This contest could not be loaded', 'Activity unavailable', 'Activity not available to your account']

declare global {
  interface Window { errorStatesSeen: string[] }
}

/** Opens a page under a clock the test can advance, so the viewer read can go stale without waiting for it. */
async function open(page: Page, path: string) {
  await page.clock.install()
  await page.goto(path)
  await hydrated(page)
}

/** Records every error state that reaches the page from now on, however briefly it stays. */
const watchErrors = (page: Page) => page.evaluate(states => {
  window.errorStatesSeen = []
  new MutationObserver(() => {
    const text = document.querySelector('main')?.textContent ?? ''
    window.errorStatesSeen.push(...states.filter(state => text.includes(state)))
  }).observe(document.body, { subtree: true, childList: true, characterData: true })
}, errorStates)

/** Coming back to the tab after the viewer read went stale re-reads the viewer, which changes every viewer-keyed query. */
async function returnToTab(page: Page) {
  await page.clock.fastForward(31_000)
  await page.evaluate(() => window.dispatchEvent(new Event('visibilitychange')))
}

const accountMenu = (page: Page, name: string) => page.getByRole('button', { name, exact: true })
const errorsSeen = (page: Page) => page.evaluate(() => window.errorStatesSeen)

test('signing in from another tab swaps the contest page without an error state', async ({ page, context, loginLearner }) => {
  test.setTimeout(90_000)
  const contest = await (await context.request.get(`${api}/immersion/contests/latest-official`)).json() as { id: string; title: string }

  await open(page, `/contests/${contest.id}/leaderboard`)
  await watchErrors(page)
  await signIn(await context.newPage(), loginLearner)
  await returnToTab(page)

  await expect(accountMenu(page, loginLearner.name)).toBeVisible({ timeout: 15_000 })
  await expect(page.getByRole('heading', { level: 1, name: contest.title })).toBeVisible({ timeout: 15_000 })
  expect(await errorsSeen(page)).toEqual([])
})

test('switching accounts swaps the registration page without an error state', async ({ page, readerContext, adminCookies }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and fixture credentials')
  test.setTimeout(60_000)
  const contest = await (await readerContext.request.get(`${api}/immersion/contests/latest-official`)).json() as { id: string }

  await open(page, `/contests/${contest.id}/registration`)
  await expect(accountMenu(page, reader.name)).toBeVisible()
  await watchErrors(page)
  await readerContext.addCookies(adminCookies)
  await returnToTab(page)

  await expect(accountMenu(page, admin.name)).toBeVisible({ timeout: 15_000 })
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15_000 })
  expect(await errorsSeen(page)).toEqual([])
})

test('switching accounts swaps a log page without an error state', async ({ page, readerContext, adminCookies }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and fixture credentials')
  test.setTimeout(60_000)
  const saved = await readerContext.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount: 1, unit_key: 'pages', tags: [], description: 'Viewer change', registration_ids: [] },
  })
  expect(saved.ok(), `log answered ${saved.status()}`).toBeTruthy()
  const { id } = await saved.json() as { id: string }

  await open(page, `/logs/${id}`)
  await expect(accountMenu(page, reader.name)).toBeVisible()
  await watchErrors(page)
  await readerContext.addCookies(adminCookies)
  await returnToTab(page)

  await expect(accountMenu(page, admin.name)).toBeVisible({ timeout: 15_000 })
  await expect(page.getByRole('heading', { level: 1, name: 'Viewer change' })).toBeVisible({ timeout: 15_000 })
  expect(await errorsSeen(page)).toEqual([])
})
