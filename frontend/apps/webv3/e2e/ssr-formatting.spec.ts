import type { Page } from '@playwright/test'
import { accountId, api, expect, hydrated, reader, test, webv3 } from './logging-fixtures'

/** React's hydration warnings and errors, which a server and browser rendering different markup produce. */
function hydrationErrors(page: Page) {
  const errors: string[] = []
  page.on('console', message => {
    if (message.type() === 'error' && /hydrat|did not match|didn't match/i.test(message.text())) errors.push(message.text())
  })
  page.on('pageerror', error => { if (/hydrat|did not match|didn't match/i.test(error.message)) errors.push(error.message) })
  return errors
}

for (const locale of ['ja-JP', 'en-GB']) {
  test.describe(locale, () => {
    test.use({ locale })

    // Hydration smoke coverage: luxon 3.7.2 already formats the heatmap's months in English under every locale, so this
    // guards the en-US pin rather than reproducing a known mismatch.
    test(`hydration smoke: a signed-in profile's statistics and the home calendar hydrate without mismatches in ${locale}`, async ({ page, readerContext }) => {
      test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
      const errors = hydrationErrors(page)

      await page.goto(`/profile/${await accountId(readerContext)}/statistics`)
      await hydrated(page)
      // The browser really runs in this locale: both the language it reports and Intl's default follow it.
      expect(await page.evaluate(() => [navigator.language, Intl.DateTimeFormat().resolvedOptions().locale])).toEqual([locale, locale])
      await expect(page.locator('svg.paper-heatmap')).toBeVisible()
      await page.goto('/')
      await hydrated(page)
      await expect(page.locator('svg.paper-heatmap')).toBeVisible()
      expect(errors).toEqual([])
    })
  })
}

test('a log\'s detail rounds its measure as the lists do and dates it from one UTC instant', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const description = `formatting-${Date.now()}`
  const created = await readerContext.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, unit_key: 'pages', amount: 2.25, duration_seconds: 100, description, tags: [] },
  })
  expect(created.ok()).toBe(true)
  const log = await created.json() as { id: string; user_id: string; created_at: string }
  const instant = new Date(log.created_at).toISOString()

  await page.goto(`/profile/${log.user_id}/updates/1`)
  await expect(page.getByRole('listitem').filter({ hasText: description })).toContainText('2.3 pages · 1.7 min')

  await page.goto(`/logs/${log.id}`)
  const stats = page.getByRole('main')
  await expect(stats).toContainText('2.3 pages')
  await expect(stats).toContainText('1.7 min')
  await expect(stats).toContainText(`${instant.slice(11, 16)} UTC`)
  await expect(page.locator('time, dd').filter({ hasText: /\d{4}/ }).first()).toBeVisible()
})
