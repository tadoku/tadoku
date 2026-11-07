import { randomUUID } from 'node:crypto'
import type { BrowserContext, Page } from '@playwright/test'
import {
  addTags,
  amountInput,
  api,
  branch,
  createLog,
  expect,
  openForm,
  reader,
  savedLogUrl,
  test,
  webv3,
} from './logging-fixtures'

const tagsInput = (page: Page) => page.getByRole('combobox', { name: /^Tags \(/ })
const chip = (page: Page, tag: string) => page.getByRole('button', { name: `Remove ${tag}`, exact: true })
const language = (page: Page) => page.getByLabel('Language', { exact: true })
const activity = (page: Page) => page.getByLabel('Activity', { exact: true })
// Required for timed activities, when its label also shows the required mark.
const timeSpent = (page: Page) => page.getByRole('spinbutton', { name: /^Time spent/ })

function fitsViewport(page: Page) {
  return page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)
}

async function record(context: BrowserContext, id: string) {
  const response = await context.request.get(`${api}/immersion/logs/${id}`)
  expect(response.headers()['x-dev-selected']).toBe(branch)
  return response.json()
}

test('tags are chips with the viewer\'s own tags and usage counts suggested by the API', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const tag = `sugg${Date.now()}`
  await openForm(page)
  await expect(page.getByRole('link', { name: 'Cancel', exact: true })).toHaveAttribute('href', '/')
  await amountInput(page).fill('2')
  await addTags(page, [tag])
  await expect(chip(page, tag)).toBeVisible()
  const first = await createLog(page)
  expect((await record(readerContext, first)).tags).toEqual([tag])

  await openForm(page)
  await amountInput(page).fill('3')
  await tagsInput(page).pressSequentially(tag.slice(0, -2))
  const suggestion = page.getByRole('option', { name: `${tag} (1×)`, exact: true })
  // Only the API knows the count.
  await expect(suggestion).toBeVisible()
  await suggestion.click()
  await expect(chip(page, tag)).toBeVisible()
  await expect(tagsInput(page)).toHaveValue('')

  const second = await createLog(page)
  expect((await record(readerContext, second)).tags).toEqual([tag])
})

test('the tag sidebar toggles tags, stops at ten and follows the activity, beside or below the form', async ({ readerContext }, testInfo) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await page.setViewportSize({ width: 1280, height: 900 })
  await openForm(page)
  await activity(page).selectOption({ label: 'Reading' })
  const aside = page.getByRole('complementary', { name: 'Suggested tags' })
  const modifiers = aside.getByRole('region', { name: 'Score modifiers' })
  await expect(modifiers).toBeVisible()
  await expect(aside.getByRole('region', { name: 'Common tags' }).getByRole('button', { name: 'ebook' })).toBeVisible()

  const manga = modifiers.getByRole('button', { name: 'manga', exact: true })
  await manga.click()
  await expect(chip(page, 'manga')).toBeVisible()
  await expect(manga).toHaveAttribute('aria-pressed', 'true')
  await page.screenshot({ path: testInfo.outputPath('new-1280.png'), fullPage: true })
  expect(await fitsViewport(page)).toBe(true)
  await manga.click()
  await expect(chip(page, 'manga')).toHaveCount(0)
  await expect(manga).toHaveAttribute('aria-pressed', 'false')

  await manga.click()
  await addTags(page, Array.from({ length: 9 }, (_, index) => `limit${index}`))
  await expect(page.getByRole('button', { name: /^Remove / })).toHaveCount(10)
  await expect(tagsInput(page)).toHaveAttribute('placeholder', 'Maximum tags reached')
  await expect(modifiers.getByRole('button', { name: 'comic', exact: true })).toBeDisabled()
  // A selected tag can still be removed from the sidebar at the limit.
  await expect(manga).toBeEnabled()
  await manga.click()
  await expect(modifiers.getByRole('button', { name: 'comic', exact: true })).toBeEnabled()

  await activity(page).selectOption({ label: 'Listening' })
  await expect(modifiers.getByRole('button', { name: 'dense', exact: true })).toBeVisible()
  await expect(modifiers.getByRole('button', { name: 'manga', exact: true })).toHaveCount(0)

  // Narrow screens show the same lists inside the form, below the tags.
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(aside).toBeHidden()
  const narrow = page.locator('form').getByRole('region', { name: 'Score modifiers' })
  await expect(narrow.getByRole('button', { name: 'dense', exact: true })).toBeVisible()
  const below = (await narrow.boundingBox())!.y > (await tagsInput(page).boundingBox())!.y
  expect(below).toBe(true)
  expect(await fitsViewport(page)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('new-390.png'), fullPage: true })
})

test('the last logged language and activity are restored for the next log, and `?amount=` still prefills', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const consoleErrors: string[] = []
  page.on('console', message => { if (message.type() === 'error') consoleErrors.push(message.text()) })
  await openForm(page)
  const initial = await language(page).inputValue()
  const codes = await language(page).locator('option').evaluateAll(options => options.map(option => (option as HTMLOptionElement).value))
  const [listened, read] = [...new Set(codes)].filter(code => code && code !== initial)

  await language(page).selectOption(listened)
  await activity(page).selectOption({ label: 'Listening' })
  await timeSpent(page).fill('20')
  const timed = await createLog(page)
  expect(await record(readerContext, timed)).toMatchObject({ language: { code: listened }, activity: { name: 'Listening' } })

  // A remembered timed activity starts without an amount, even when `?amount=` asks for one.
  await page.goto('/logs/new?amount=12')
  await expect(page.getByRole('button', { name: 'Log activity', exact: true })).toBeEnabled({ timeout: 15_000 })
  await expect(language(page)).toHaveValue(listened)
  await expect(activity(page)).toHaveValue(String((await record(readerContext, timed)).activity.id))
  await expect(amountInput(page)).toHaveCount(0)
  await expect(timeSpent(page)).toHaveValue('')

  await language(page).selectOption(read)
  await activity(page).selectOption({ label: 'Reading' })
  await amountInput(page).fill('4')
  await createLog(page)

  await page.goto('/logs/new?amount=12')
  await expect(page.getByRole('button', { name: 'Log activity', exact: true })).toBeEnabled({ timeout: 15_000 })
  await expect(language(page)).toHaveValue(read)
  await expect(activity(page).locator('option:checked')).toHaveText('Reading')
  await expect(amountInput(page)).toHaveValue('12')
  expect(consoleErrors.filter(error => /hydrat/i.test(error))).toEqual([])
})

test('without browser storage the form starts from its defaults and still saves', async ({ readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const page = await readerContext.newPage()
  await page.addInitScript(() => {
    Object.defineProperty(window, 'localStorage', { get() { throw new DOMException('Storage disabled', 'SecurityError') } })
  })
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  await openForm(page)
  await amountInput(page).fill('5')
  const id = await createLog(page)
  expect((await record(readerContext, id)).amount).toBe(5)
  expect(errors).toEqual([])
})

test('edit keeps language and activity fixed, uses the same tags and sidebar and returns to the log', async ({ page, readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  await openForm(page)
  await activity(page).selectOption({ label: 'Reading' })
  await amountInput(page).fill('6')
  await addTags(page, ['book'])
  const id = await createLog(page)

  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto(`/logs/${id}/edit`)
  const save = page.getByRole('button', { name: 'Save', exact: true })
  await expect(save).toBeEnabled({ timeout: 15_000 })
  // An edit can't change language or activity, so the header states them and the form offers no controls for them.
  await expect(page.locator('[data-page-lead]')).toHaveText(/ · /)
  await expect(language(page)).toHaveCount(0)
  await expect(activity(page)).toHaveCount(0)
  await expect(chip(page, 'book')).toBeVisible()
  await expect(page.getByRole('link', { name: 'Cancel', exact: true })).toHaveAttribute('href', `/logs/${id}`)
  const aside = page.getByRole('complementary', { name: 'Suggested tags' })
  await aside.getByRole('button', { name: 'manga', exact: true }).click()
  await expect(chip(page, 'manga')).toBeVisible()
  await expect(page.locator('[data-estimate]')).toContainText('Estimated score:')
  await page.screenshot({ path: testInfo.outputPath('edit-1280.png'), fullPage: true })
  await page.setViewportSize({ width: 390, height: 844 })
  expect(await fitsViewport(page)).toBe(true)
  await page.screenshot({ path: testInfo.outputPath('edit-390.png'), fullPage: true })

  await save.click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}`)
  expect([...(await record(readerContext, id)).tags].sort()).toEqual(['book', 'manga'])
})

test('a learner without ongoing registrations goes from Create straight to the saved log', async ({ page }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  // Both fixture accounts are registered in the seeded round, so this journey registers a disposable branch account.
  await page.goto(webv3!.replace('.next.', '.account.') + '/register')
  await page.getByLabel('Display name').fill('Webv3 Log Form')
  await page.getByLabel('Email').fill(`webv3-log-form-${randomUUID()}@example.test`)
  await page.getByLabel('Password').fill(`Log-form-${randomUUID()}`)
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expect(page.getByRole('button', { name: 'Create account', exact: true })).toHaveCount(0, { timeout: 15_000 })
  const ongoing = await page.request.get(`${api}/immersion/contests/ongoing-registrations`)
  expect((await ongoing.json()).registrations).toEqual([])

  await openForm(page)
  await amountInput(page).fill('3')
  await page.getByRole('button', { name: 'Log activity', exact: true }).click()
  await expect(page).toHaveURL(savedLogUrl)
  const saved = await record(page.context(), new URL(page.url()).pathname.split('/')[2])
  expect(saved).toMatchObject({ amount: 3, registrations: [] })
})
