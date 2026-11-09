import type { Locator, Page } from '@playwright/test'
import { accountId, addTags, api, createLog, expect, fillDraft, hydrated, openForm, reader, test, webv3 } from './logging-fixtures'

/** Saves a log tagged NSFW and returns its ID and description, leaving the page on its detail. */
async function createNsfwLog(page: Page) {
  const description = `nsfw-${Date.now()}`
  await openForm(page)
  await fillDraft(page, '5', description)
  await addTags(page, ['nsfw'])
  return { id: await createLog(page), description }
}

/** The description is blurred and hidden from assistive technology until the reader reveals it, as in Webv2. */
async function expectMasked(scope: Locator, description: string) {
  await expect(scope.getByRole('button', { name: 'Reveal NSFW description', exact: true })).toBeVisible()
  const text = scope.getByText(description, { exact: true })
  await expect(text).toHaveAttribute('aria-hidden', 'true')
  await expect(text).toHaveCSS('filter', /blur/)
}

async function expectRevealed(scope: Locator, description: string) {
  await scope.getByRole('button', { name: 'Reveal NSFW description', exact: true }).click()
  await expect(scope.getByRole('button', { name: 'Hide NSFW description', exact: true })).toBeVisible()
  const text = scope.getByText(description, { exact: true })
  await expect(text).not.toHaveAttribute('aria-hidden', 'true')
  await expect(text).toHaveCSS('filter', 'none')
}

test.beforeEach(() => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
})

test('an NSFW log\'s detail hides its description until the reader reveals it', async ({ page, readerContext }, testInfo) => {
  void readerContext
  const { description } = await createNsfwLog(page)
  await hydrated(page)
  const main = page.getByRole('main')

  await expect(main.getByRole('heading', { level: 1 })).not.toHaveText(description)
  await expectMasked(main, description)
  await page.screenshot({ path: testInfo.outputPath('detail-masked.png'), fullPage: true })
  await expectRevealed(main, description)
})

test('the edit and contest pages link back to an NSFW log without naming its description', async ({ page, readerContext }) => {
  void readerContext
  const { id, description } = await createNsfwLog(page)

  for (const path of [`/logs/${id}/edit`, `/logs/${id}/contests`]) {
    await page.goto(path)
    await hydrated(page)
    await expect(page.locator(`a[href="/logs/${id}"]`).first()).toBeVisible()
    await expect(page.getByRole('link', { name: description })).toHaveCount(0)
  }
  // The contest step's summary of the log masks the description like the detail page.
  await expectMasked(page.getByRole('region', { name: 'Your log' }), description)
})

test('the activity feed hides an NSFW description until the reader reveals it', async ({ page, readerContext }) => {
  const { description } = await createNsfwLog(page)
  await page.goto(`/profile/${await accountId(readerContext)}/updates`)
  await hydrated(page)
  const entry = page.getByRole('listitem').filter({ hasText: description })

  await expectMasked(entry, description)
  await expectRevealed(entry, description)
})
