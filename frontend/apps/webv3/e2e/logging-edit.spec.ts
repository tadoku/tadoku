import type { BrowserContext, Page, Request } from '@playwright/test'
import {
  addTags,
  createLog,
  admin,
  amountInput,
  api,
  branch,
  expect,
  fillDraft,
  isSaveRequest,
  openForm,
  reader,
  sessionCookie,
  test,
  webv3,
} from './logging-fixtures'

const estimate = (page: Page) => page.locator('.log-form__estimate strong')
const timeSpent = (page: Page) => page.getByLabel(/^Time spent/)
const description = (page: Page) => page.getByLabel('Description', { exact: true })
const saveChanges = (page: Page) => page.getByRole('button', { name: 'Save', exact: true })
const isPreviewRequest = (request: Request) => request.method() === 'GET' && request.url().includes('_serverFn')

/** Saves a reading log through the activity form and returns its ID, leaving the page on its detail. */
async function createReadingLog(page: Page, text: string, amount = '5') {
  await openForm(page)
  await fillDraft(page, amount, text)
  return createLog(page)
}

async function readRecord(context: BrowserContext, id: string) {
  const response = await context.request.get(`${api}/immersion/logs/${id}`)
  expect(response.headers()['x-dev-selected']).toBe(branch)
  return response
}

/** The API's own estimate for the draft the form shows, formatted as the form formats it. */
async function apiEstimate(page: Page, context: BrowserContext, body: Record<string, unknown>) {
  const response = await context.request.post(`${api}/immersion/logs/score-preview`, { data: body })
  expect(response.status()).toBe(200)
  const { platform } = await response.json() as { platform: { score: number } }
  return new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 }).format(platform.score)
}

test('owner edits a saved log, sees the estimate and the persisted change after reload', async ({ page, readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const id = await createReadingLog(page, `edit-before-${Date.now()}`)
  const before = await (await readRecord(readerContext, id)).json()

  await page.getByRole('link', { name: 'Edit log', exact: true }).click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}/edit`)
  await expect(saveChanges(page)).toBeEnabled({ timeout: 15_000 })
  // Language and activity can't change on an edit, so the form states them instead of offering controls.
  const fixed = page.locator('.log-form__fixed')
  await expect(fixed).toContainText(before.language.name)
  await expect(fixed).toContainText(before.activity.name)
  await expect(page.getByLabel('Language', { exact: true })).toHaveCount(0)
  await expect(amountInput(page)).toHaveValue('5')
  await expect(page.getByLabel('Unit', { exact: true })).toHaveValue(before.unit_key)
  await expect(page.getByRole('group', { name: 'Contests' })).toHaveCount(0)

  const text = `edit-after-${Date.now()}`
  await amountInput(page).fill('12')
  await timeSpent(page).fill('30')
  await description(page).fill(text)
  await addTags(page, ['Edited', 'e2e'])
  const expected = await apiEstimate(page, readerContext, {
    language_code: before.language.code,
    activity_id: before.activity.id,
    amount: 12,
    unit_key: before.unit_key,
    duration_seconds: 1800,
    tags: ['edited', 'e2e'],
  })
  await expect(estimate(page)).toHaveText(expected)
  await page.screenshot({ path: testInfo.outputPath('edit-form.png'), fullPage: true })

  let saves = 0
  page.on('request', request => { if (isSaveRequest(request)) saves++ })
  await saveChanges(page).click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}`)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(text)
  await page.reload()
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(text)
  await expect(page.getByText(`12 ${before.unit_name.toLowerCase()}s`, { exact: true })).toBeVisible()
  await expect(page.getByText('30 min', { exact: true })).toBeVisible()
  expect(saves).toBe(1)

  const after = await (await readRecord(readerContext, id)).json()
  expect(after).toMatchObject({ id, description: text, amount: 12, duration_seconds: 1800, unit_key: before.unit_key })
  expect([...after.tags].sort()).toEqual(['e2e', 'edited'])
  expect(after.language.code).toBe(before.language.code)
  expect(after.activity.id).toBe(before.activity.id)
  expect(after.registrations.map((registration: { contest_id: string }) => registration.contest_id))
    .toEqual(before.registrations.map((registration: { contest_id: string }) => registration.contest_id))
  await testInfo.attach('edited-record', { body: JSON.stringify({ before, after }, null, 2), contentType: 'application/json' })
})

test('editing only the description keeps a saved tag that holds a comma', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const id = await createReadingLog(page, `edit-comma-${Date.now()}`)
  const before = await (await readRecord(readerContext, id)).json()
  const changed = await readerContext.request.put(`${api}/immersion/logs/${id}`, {
    data: { amount: before.amount, unit_key: before.unit_key, tags: ['reading, fiction'], description: before.description },
  })
  expect(changed.status()).toBe(200)

  await page.goto(`/logs/${id}/edit`)
  await expect(saveChanges(page)).toBeEnabled({ timeout: 15_000 })
  const text = `edit-comma-after-${Date.now()}`
  await description(page).fill(text)
  await saveChanges(page).click()
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(text)

  expect((await (await readRecord(readerContext, id)).json()).tags).toEqual(['reading, fiction'])
})

test('an expired session keeps the edit and changes nothing', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const id = await createReadingLog(page, `edit-expired-${Date.now()}`)
  const before = await (await readRecord(readerContext, id)).json()
  await page.goto(`/logs/${id}/edit`)
  await expect(saveChanges(page)).toBeEnabled({ timeout: 15_000 })
  await amountInput(page).fill('21')
  await readerContext.addCookies([{ ...await sessionCookie(readerContext), value: 'expired' }])

  await saveChanges(page).click()
  await expect(page.getByRole('alert')).toContainText('Your session ended')
  await expect(amountInput(page)).toHaveValue('21')
  await expect(page).toHaveURL(`${webv3}/logs/${id}/edit`)
  expect((await (await page.request.get(`${api}/immersion/logs/${id}`)).json()).amount).toBe(before.amount)
})

test('owner deletes a log after confirming, and its detail is gone', async ({ page, readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const id = await createReadingLog(page, `delete-${Date.now()}`)
  const { user_id: userId } = await (await readRecord(readerContext, id)).json()
  const dialog = page.getByRole('dialog', { name: 'Delete this log?' })

  await page.getByRole('button', { name: 'Delete log', exact: true }).click()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('button', { name: 'Keep log' })).toBeFocused()
  await page.screenshot({ path: testInfo.outputPath('delete-confirmation.png'), fullPage: true })
  await dialog.getByRole('button', { name: 'Keep log' }).click()
  await expect(dialog).toBeHidden()
  expect((await readRecord(readerContext, id)).status()).toBe(200)

  await page.getByRole('button', { name: 'Delete log', exact: true }).click()
  await dialog.getByRole('button', { name: 'Delete log', exact: true }).click()
  await page.waitForURL(new RegExp(`/profile/${userId}/statistics`), { waitUntil: 'commit' })

  expect((await readRecord(readerContext, id)).status()).toBe(404)
  await page.goto(`${webv3}/logs/${id}`)
  await expect(page.getByRole('heading', { name: 'Page not found' })).toBeVisible()
})

test('visitors see no owner actions and cannot open the edit page for another learner\'s log', async ({ readerContext, adminContext, browser }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  const adminPage = await adminContext.newPage()
  const description = `visitor-${Date.now()}`
  const id = await createReadingLog(adminPage, description)
  await expect(adminPage.getByRole('link', { name: 'Edit log', exact: true })).toBeVisible()

  const readerPage = await readerContext.newPage()
  await readerPage.goto(`/logs/${id}`)
  await expect(readerPage.getByRole('heading', { level: 1, name: description })).toBeVisible()
  await expect(readerPage.getByRole('link', { name: 'Edit log' })).toHaveCount(0)
  await expect(readerPage.getByRole('button', { name: 'Delete log' })).toHaveCount(0)
  await readerPage.goto(`/logs/${id}/edit`)
  await expect(readerPage.getByRole('heading', { name: 'This is someone else\'s record' })).toBeVisible()
  await expect(saveChanges(readerPage)).toHaveCount(0)

  const anonymous = await browser.newContext({ baseURL: webv3 })
  try {
    const anonymousPage = await anonymous.newPage()
    await anonymousPage.goto(`/logs/${id}`)
    await expect(anonymousPage.getByRole('heading', { level: 1, name: description })).toBeVisible()
    await expect(anonymousPage.getByRole('link', { name: 'Edit log' })).toHaveCount(0)
    await expect(anonymousPage.getByRole('button', { name: 'Delete log' })).toHaveCount(0)
    await anonymousPage.goto(`/logs/${id}/edit`)
    await expect(anonymousPage.getByRole('heading', { name: 'Sign in to edit activity' })).toBeVisible()
  } finally {
    await anonymous.close()
  }
})

test('score preview shows the API estimate, and a dash while incomplete or failed', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  await openForm(page)
  await expect(estimate(page)).toHaveText('—')
  const body = async (amount: number) => ({
    language_code: await page.getByLabel('Language', { exact: true }).inputValue(),
    activity_id: Number(await page.getByLabel('Activity', { exact: true }).inputValue()),
    amount,
    unit_key: await page.getByLabel('Unit', { exact: true }).inputValue(),
    tags: [],
  })

  await amountInput(page).fill('7')
  await expect(estimate(page)).toHaveText(await apiEstimate(page, readerContext, await body(7)))

  await page.route('**/*', route => isPreviewRequest(route.request()) ? route.abort() : route.continue())
  await amountInput(page).fill('8')
  await expect(estimate(page)).toHaveText('—')
  await page.unroute('**/*')

  await amountInput(page).fill('9')
  await expect(estimate(page)).toHaveText(await apiEstimate(page, readerContext, await body(9)))
  await amountInput(page).fill('')
  await expect(estimate(page)).toHaveText('—')
})
