import type { Page } from '@playwright/test'
import {
  admin,
  amountInput,
  api,
  branch,
  expect,
  isSaveRequest,
  logsDescribed,
  openForm,
  reader,
  saveButton,
  savedLogId,
  savedLogUrl,
  test,
  webv3,
} from './logging-fixtures'

test('native activity link opens an anonymous sign-in prompt', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/plan')
  const link = page.getByRole('link', { name: 'Log activity', exact: true }).first()
  await expect(link).toHaveAttribute('href', '/logs/new')

  await link.click()
  await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
})

for (const activity of ['Reading', 'Listening']) {
  test(`${activity}: validation, one save, persisted API detail and reload`, async ({ page, readerContext }, testInfo) => {
    test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
    const timed = activity === 'Listening'
    const quantity = timed ? page.getByLabel('Minutes', { exact: true }) : amountInput(page)
    const description = page.getByLabel('Description', { exact: true })
    const tags = page.getByLabel('Tags', { exact: true })
    await openForm(page)
    await page.getByLabel('Activity', { exact: true }).selectOption({ label: activity })
    await page.getByLabel('Language', { exact: true }).selectOption('jpn')

    await saveButton(page).click()
    await expect(page.getByRole('alert').first()).toBeVisible()
    await expect(quantity).toBeFocused()

    await quantity.fill(timed ? '2.5' : '17')
    await description.fill('x'.repeat(256))
    await saveButton(page).click()
    await expect(description).toBeFocused()

    const text = `webv3-e2e-${activity}-${Date.now()}`
    await description.fill(text)
    await tags.fill(Array.from({ length: 11 }, (_, index) => `tag${index}`).join(','))
    await saveButton(page).click()
    await expect(tags).toBeFocused()

    // Two submissions in the same tick, each delayed at the browser boundary, must make one save.
    await tags.fill('book, fiction')
    let saves = 0
    await page.route('**/*', async route => {
      if (isSaveRequest(route.request())) {
        saves++
        await new Promise(resolve => setTimeout(resolve, 600))
      }
      await route.continue()
    })
    await page.evaluate(() => {
      const form = document.querySelector<HTMLFormElement>('form[aria-busy]')!
      form.requestSubmit()
      form.requestSubmit()
    })
    await expect(page).toHaveURL(savedLogUrl)
    expect(saves).toBe(1)

    const id = savedLogId(page)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(text)
    await page.reload()
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(text)

    const response = await readerContext.request.get(`${api}/immersion/logs/${id}`)
    expect(response.status()).toBe(200)
    expect(response.headers()['x-dev-selected']).toBe(branch)
    const record = await response.json()
    expect(record).toMatchObject({ id, description: text, tags: ['book', 'fiction'] })
    expect(record.language.code).toBe('jpn')
    expect(record.activity.name).toBe(activity)
    if (timed) expect(record.duration_seconds).toBe(150)
    else expect(record.amount).toBe(17)
    expect(record.registrations.length).toBeGreaterThan(0)
    await expect(page.getByText(record.id, { exact: true })).toBeVisible()
    for (const registration of record.registrations) {
      await expect(page.getByRole('link', { name: registration.title, exact: true })).toBeVisible()
    }
    expect(await logsDescribed(readerContext, record.user_id, text)).toHaveLength(1)

    await testInfo.attach('persisted-record', { body: JSON.stringify(record, null, 2), contentType: 'application/json' })
    await page.screenshot({ path: testInfo.outputPath(`${activity}-saved.png`), fullPage: true })
  })
}

test('text limits follow the API: description characters, tag bytes and normalized tags', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const tags = page.getByLabel('Tags', { exact: true })
  await openForm(page)
  await amountInput(page).fill('1')

  // 128 emoji are 256 UTF-16 units but 128 characters; the API allows 255 characters.
  const description = `${'😀'.repeat(128)}${Date.now()}`
  await page.getByLabel('Description', { exact: true }).fill(description)
  // 18 characters but 54 UTF-8 bytes; the API allows 50 bytes per tag.
  await tags.fill('日本語'.repeat(6))
  await saveButton(page).click()
  await expect(tags).toBeFocused()
  await expect(page.getByLabel('Description', { exact: true })).not.toHaveAttribute('aria-invalid', 'true')

  // Fourteen entries that normalize to five tags, as Go's strings.TrimSpace and strings.ToLower do: they trim U+0085
  // but keep U+FEFF, 'İ' becomes the single byte 'i' and 'Σ' becomes 'σ' wherever it stands.
  await tags.fill(`${'Book, book, BOOK, '.repeat(3)}Fiction, \u0085fiction\u2003, ${'İ'.repeat(17)}, ΟΣ, \ufeffbook`)
  await saveButton(page).click()
  await expect(page).toHaveURL(savedLogUrl)

  const record = await (await readerContext.request.get(`${api}/immersion/logs/${savedLogId(page)}`)).json()
  expect(record.description).toBe(description)
  // The API returns tags in its collation order, which ignores U+FEFF.
  expect([...record.tags].sort()).toEqual(['book', 'fiction', 'i'.repeat(17), 'οσ', '\ufeffbook'].sort())
})

test('changing the activity clears the quantity entered for the previous one', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  const activity = page.getByLabel('Activity', { exact: true })
  const quantity = page.getByRole('spinbutton').first()
  await openForm(page)

  for (const [from, to] of [['Reading', 'Writing'], ['Listening', 'Speaking']]) {
    await activity.selectOption({ label: from })
    await quantity.fill('20')
    await activity.selectOption({ label: to })
    await expect(quantity).toHaveValue('')
  }
})

test('transport failure keeps values and never navigates or claims success', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await amountInput(page).fill('12')
  await page.getByLabel('Description', { exact: true }).fill('preserve my draft')
  await page.route('**/*', route => isSaveRequest(route.request()) ? route.abort() : route.continue())

  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText(/check your activity/i)
  await expect(amountInput(page)).toHaveValue('12')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue('preserve my draft')
  await expect(page).toHaveURL(/\/logs\/new$/)
})

for (const width of [390, 1280]) {
  test(`keyboard-only activity form at ${width}px`, async ({ readerContext }, testInfo) => {
    test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
    const page = await readerContext.newPage()
    await page.setViewportSize({ width, height: 900 })
    const errors: string[] = []
    page.on('pageerror', error => errors.push(error.message))
    await openForm(page)

    await page.getByLabel('Language', { exact: true }).focus()
    await page.keyboard.press('j')
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Activity', { exact: true })).toBeFocused()
    await page.keyboard.press('Home')
    await page.keyboard.press('Tab')
    await expect(amountInput(page)).toBeFocused()
    await page.keyboard.type('3')
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Unit', { exact: true })).toBeFocused()
    await page.keyboard.press('Tab')
    const description = `keyboard-${width}-${Date.now()}`
    await page.keyboard.type(description)
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Tags', { exact: true })).toBeFocused()
    await page.keyboard.type('keyboard')
    await tabTo(page, saveButton(page))
    expect(await fitsViewport(page)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`keyboard-${width}-form.png`), fullPage: true })

    await page.keyboard.press('Enter')
    await expect(page).toHaveURL(savedLogUrl)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
    await expect(page.getByRole('heading', { level: 1 })).toBeFocused()
    expect(await fitsViewport(page)).toBe(true)
    expect(errors).toEqual([])
    await page.screenshot({ path: testInfo.outputPath(`keyboard-${width}-saved.png`), fullPage: true })
  })
}

test('API rejection preserves entered activity instead of showing success', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await amountInput(page).fill('8')
  await page.getByLabel('Description', { exact: true }).fill('rejected activity draft')
  let changed = false
  await page.route('**/*', async route => {
    if (!isSaveRequest(route.request())) return route.continue()
    changed = true
    await route.continue({ postData: withUnknownRegistration(route.request().postData()!) })
  })

  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText('not saved')
  expect(changed).toBe(true)
  await expect(amountInput(page)).toHaveValue('8')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue('rejected activity draft')
  await expect(page).toHaveURL(/\/logs\/new$/)
})

test('public detail hides contest management and the canonical tenant cannot read branch activity', async ({ page, readerContext, browser }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  await amountInput(page).fill('6')
  await page.getByLabel('Description', { exact: true }).fill(`isolation-${Date.now()}`)
  await saveButton(page).click()
  await expect(page).toHaveURL(savedLogUrl)
  const id = savedLogId(page)
  const log = await (await readerContext.request.get(`${api}/immersion/logs/${id}`)).json()
  expect(log.registrations.length).toBeGreaterThan(0)
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toBeVisible()

  const publicContext = await browser.newContext({ baseURL: webv3 })
  try {
    const publicPage = await publicContext.newPage()
    await publicPage.goto(`${webv3}/logs/${id}`)
    await expect(publicPage.getByText(id, { exact: true })).toBeVisible()
    await expect(publicPage.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
    const publicRecord = await publicContext.request.get(`${api}/immersion/logs/${id}`)
    expect((await publicRecord.json()).registrations).toEqual([])
  } finally {
    await publicContext.close()
  }
  const otherTenant = await readerContext.request.get(`https://tadoku.dev.lab/api/internal/immersion/logs/${id}`)
  expect(otherTenant.status()).toBe(404)
})

test('a different signed-in learner cannot change another learner\'s activity', async ({ page, readerContext, adminContext }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  const adminPage = await adminContext.newPage()
  await openForm(adminPage)
  await amountInput(adminPage).fill('9')
  await adminPage.getByLabel('Description', { exact: true }).fill(`ownership-${Date.now()}`)
  await saveButton(adminPage).click()
  await expect(adminPage).toHaveURL(savedLogUrl)
  const id = savedLogId(adminPage)

  await page.goto(`${webv3}/logs/${id}`)
  await expect(page.getByText(id, { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)

  const path = `${api}/immersion/logs/${id}`
  const before = await (await readerContext.request.get(path)).json()
  expect(before.registrations).toEqual([])
  const change = {
    language_code: before.language.code,
    activity_id: before.activity.id,
    amount: 99,
    unit_id: before.unit_id,
    tags: [],
    description: 'unauthorized change',
  }
  expect((await readerContext.request.put(path, { data: change })).status()).toBe(403)
  expect((await readerContext.request.delete(path)).status()).toBe(403)
  expect(await (await readerContext.request.get(path)).json()).toEqual(before)
})

/**
 * Replaces the save's registration IDs with one the API doesn't know, inside TanStack Start's serialized
 * server-function payload, so that the API itself rejects the save.
 */
function withUnknownRegistration(postData: string) {
  const payload = JSON.parse(postData)
  const log = payload.t.p.v[0]
  log.p.v[log.p.k.indexOf('registrationIds')].a = [{ t: 1, s: '00000000-0000-4000-8000-000000000000' }]
  return JSON.stringify(payload)
}

async function tabTo(page: Page, target: ReturnType<Page['locator']>) {
  for (let step = 0; step < 8 && !await target.evaluate(element => element === document.activeElement); step++) {
    await page.keyboard.press('Tab')
  }
  await expect(target).toBeFocused()
}

function fitsViewport(page: Page) {
  return page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)
}
