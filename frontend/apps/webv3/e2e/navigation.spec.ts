import type { Page } from '@playwright/test'
import { accountId, api, expect, hydrated, reader, test, webv3 } from './logging-fixtures'

/** Every document request the page makes from now on; client navigation makes none. */
function documentRequests(page: Page) {
  const urls: string[] = []
  page.on('request', request => { if (request.resourceType() === 'document') urls.push(request.url()) })
  return urls
}

/** The router has rendered the page it navigated to. */
const settled = (page: Page) => expect(page.getByRole('main')).not.toHaveAttribute('aria-busy', 'true')

/** After a client navigation the page title is focused and announced, so keyboard and screen reader users follow. */
async function expectArrival(page: Page) {
  await expect(page.getByRole('main').getByRole('heading', { level: 1 })).toBeFocused()
  await expect(page.getByRole('status').filter({ hasText: await page.title() })).toHaveCount(1)
}

test('home → contest → profile → leaderboard navigates without loading a new document', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/')
  await hydrated(page)
  const documents = documentRequests(page)

  await page.getByRole('main').getByRole('link', { name: /^(Contest details|Full leaderboard|See final board)$/ }).first().click()
  await expect(page).toHaveURL(/\/contests\/[^/]+\/leaderboard\/1$/)
  await settled(page)
  await page.getByRole('table').getByRole('link').first().click()
  await expect(page).toHaveURL(/\/contests\/[^/]+\/profile\/[^/]+\/1$/)
  await settled(page)
  await page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Leaderboard', exact: true }).first().click()
  await expect(page).toHaveURL(/\/leaderboard\/latest\/1$/)
  expect(documents).toEqual([])
})

test('a client navigation focuses and announces the new page, and the first load does not', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/')
  await hydrated(page)
  await expect(page.getByRole('main').getByRole('heading', { level: 1 })).not.toBeFocused()

  await page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Contests', exact: true }).first().click()
  await expect(page).toHaveURL(/\/contests\/official\/1$/)
  await expectArrival(page)
  await page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Leaderboard', exact: true }).first().click()
  await expect(page).toHaveURL(/\/leaderboard\/latest\/1$/)
  await expectArrival(page)
})

test('the calendar\'s date field opens that day\'s activity', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const id = await accountId(readerContext)
  const today = new Date().toISOString().slice(0, 10)
  await page.goto(`/profile/${id}/statistics`)
  await hydrated(page)

  await page.getByLabel(/^Open a day/).fill(today)
  await page.getByRole('button', { name: 'Open day', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/profile/${id}/updates/1?date=${today}`)
})

test.describe('without JavaScript', () => {
  test.use({ javaScriptEnabled: false })

  test('the mobile menu opens and lists the navigation', async ({ page }) => {
    test.skip(!webv3, 'Set WEBV3_BASE_URL')
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/')
    await page.getByRole('link', { name: 'Open menu', exact: true }).click()
    await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Contests', exact: true }))
      .toBeVisible()
  })

  test('the account menu opens with the account\'s pages and Log out', async ({ page, readerContext }) => {
    test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
    void readerContext
    await page.goto('/')
    await page.getByLabel(reader.name, { exact: true }).click()
    const menu = page.getByRole('navigation', { name: 'Main navigation' })
    await expect(menu.getByRole('link', { name: 'Profile', exact: true })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'Account settings', exact: true })).toBeVisible()
    await expect(menu.getByRole('button', { name: 'Log out', exact: true })).toBeVisible()
  })
})
