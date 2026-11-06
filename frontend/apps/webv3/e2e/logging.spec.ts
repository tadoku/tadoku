import type { Page } from '@playwright/test'
import {
  addTags,
  admin,
  amountInput,
  api,
  branch,
  contestStepUrl,
  createLog,
  expect,
  isSaveRequest,
  logsDescribed,
  openForm,
  reader,
  saveButton,
  savedLogId,
  test,
  webv3,
} from './logging-fixtures'

test('the activity form asks an anonymous visitor to sign in, and guests get no header Log activity', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  // As in Webv2, guests get Log in and Sign up in the header instead.
  await page.goto('/plan')
  await expect(page.getByRole('banner').getByRole('link', { name: 'Sign up', exact: true })).toBeVisible()
  await expect(page.getByRole('banner').getByRole('link', { name: 'Log activity', exact: true })).toHaveCount(0)

  await page.goto('/logs/new')
  await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
})

for (const activity of ['Reading', 'Listening']) {
  test(`${activity}: validation, one save, the contest step, persisted API detail and reload`, async ({ page, readerContext }, testInfo) => {
    test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
    const timed = activity === 'Listening'
    const quantity = timed ? page.getByRole('spinbutton', { name: /^Time spent/ }) : amountInput(page)
    const description = page.getByLabel('Description', { exact: true })
    await openForm(page)
    await expect(description).toHaveAttribute('placeholder', 'e.g. One Piece volume 45')
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
    await addTags(page, ['book', 'fiction'])

    // Two submissions in the same tick, each delayed at the browser boundary, must make one save.
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
    // The Dev Reader has ongoing registrations, so the contest step follows with every eligible one selected.
    await expect(page).toHaveURL(contestStepUrl)
    expect(saves).toBe(1)
    await page.unrouteAll()
    const id = savedLogId(page)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Submit to contests')
    const save = page.getByRole('button', { name: 'Save submissions', exact: true })
    await expect(save).toBeEnabled({ timeout: 15_000 })
    const eligible = page.getByRole('checkbox', { disabled: false })
    expect(await eligible.count()).toBeGreaterThan(0)
    for (const contest of await eligible.all()) await expect(contest).toBeChecked()
    await save.click()

    await expect(page).toHaveURL(`${webv3}/logs/${id}`)
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
    await expect(page.getByRole('heading', { level: 1, name: text })).toBeVisible()
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
  const tags = page.getByLabel(/^Tags \(up to/)
  await openForm(page)
  await amountInput(page).fill('1')

  // 128 emoji are 256 UTF-16 units but 128 characters; the API allows 255 characters.
  const description = `${'😀'.repeat(128)}${Date.now()}`
  await page.getByLabel('Description', { exact: true }).fill(description)
  // 18 characters but 54 UTF-8 bytes; the API allows 50 bytes per tag.
  const long = '日本語'.repeat(6)
  await addTags(page, [long])
  await saveButton(page).click()
  await expect(tags).toBeFocused()
  await expect(page.getByLabel('Description', { exact: true })).not.toHaveAttribute('aria-invalid', 'true')
  await page.getByRole('button', { name: `Remove ${long}`, exact: true }).click()

  // The chips ignore case when they add a tag. Seven chips normalize to four tags, as Go's strings.TrimSpace and
  // strings.ToLower do: they trim U+0085, 'İ' becomes the single byte 'i' and 'Σ' becomes 'σ' wherever it stands.
  await addTags(page, ['Book', 'book', 'BOOK', 'Fiction', '\u0085fiction\u2003', 'İ'.repeat(17), 'ΟΣ', 'ος'])
  // The last entry is a duplicate, so its suggestions stay open until dismissed.
  await tags.press('Escape')
  await expect(page.getByRole('button', { name: /^Remove / })).toHaveCount(5)
  const id = await createLog(page)

  const record = await (await readerContext.request.get(`${api}/immersion/logs/${id}`)).json()
  expect(record.description).toBe(description)
  expect([...record.tags].sort()).toEqual(['book', 'fiction', 'i'.repeat(17), 'οσ'].sort())
})

test('`?amount=` prefills a positive whole amount and ignores anything else, as in Webv2', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  for (const [query, expected] of [['12', '12'], ['0', ''], ['-3', ''], ['1.5', ''], ['many', '']]) {
    await page.goto(`/logs/new?amount=${query}`)
    await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
    await expect(amountInput(page)).toHaveValue(expected)
  }
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
    await expect(page.getByLabel(/^Time spent/)).toBeFocused()
    await page.keyboard.press('Tab')
    const description = `keyboard-${width}-${Date.now()}`
    await page.keyboard.type(description)
    await page.keyboard.press('Tab')
    await expect(page.getByLabel(/^Tags \(up to/)).toBeFocused()
    await page.keyboard.type('keyboard')
    await page.keyboard.press('Enter')
    await expect(page.getByRole('button', { name: 'Remove keyboard', exact: true })).toBeVisible()
    await tabTo(page, saveButton(page))
    expect(await fitsViewport(page)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`keyboard-${width}-form.png`), fullPage: true })

    await page.keyboard.press('Enter')
    await expect(page).toHaveURL(contestStepUrl)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Submit to contests')
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
    await route.continue({ postData: withUnknownLanguage(route.request().postData()!) })
  })

  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText('could not be saved')
  expect(changed).toBe(true)
  await expect(amountInput(page)).toHaveValue('8')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue('rejected activity draft')
  await expect(page).toHaveURL(/\/logs\/new$/)
})

test('public detail hides contest management and the canonical tenant cannot read branch activity', async ({ page, readerContext, browser }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  await amountInput(page).fill('6')
  const description = `isolation-${Date.now()}`
  await page.getByLabel('Description', { exact: true }).fill(description)
  const id = await createLog(page)
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toBeVisible()

  const publicContext = await browser.newContext({ baseURL: webv3 })
  try {
    const publicPage = await publicContext.newPage()
    await publicPage.goto(`${webv3}/logs/${id}`)
    await expect(publicPage.getByRole('heading', { level: 1, name: description })).toBeVisible()
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
  const description = `ownership-${Date.now()}`
  await adminPage.getByLabel('Description', { exact: true }).fill(description)
  const id = await createLog(adminPage)

  await page.goto(`${webv3}/logs/${id}`)
  await expect(page.getByRole('heading', { level: 1, name: description })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)

  const path = `${api}/immersion/logs/${id}`
  const before = await (await readerContext.request.get(path)).json()
  expect(before.registrations).toEqual([])
  const change = {
    language_code: before.language.code,
    activity_id: before.activity.id,
    amount: 99,
    unit_key: before.unit_key,
    tags: [],
    description: 'unauthorized change',
  }
  expect((await readerContext.request.put(path, { data: change })).status()).toBe(403)
  expect((await readerContext.request.delete(path)).status()).toBe(403)
  expect(await (await readerContext.request.get(path)).json()).toEqual(before)
})

/**
 * Replaces the save's language with a code the API doesn't offer, inside TanStack Start's serialized server-function
 * payload, so that the API itself rejects the save.
 */
function withUnknownLanguage(postData: string) {
  const payload = JSON.parse(postData)
  const log = payload.t.p.v[0]
  log.p.v[log.p.k.indexOf('languageCode')].s = 'zzz'
  return JSON.stringify(payload)
}

async function tabTo(page: Page, target: ReturnType<Page['locator']>) {
  for (let step = 0; step < 30 && !await target.evaluate(element => element === document.activeElement); step++) {
    await page.keyboard.press('Tab')
  }
  await expect(target).toBeFocused()
}

function fitsViewport(page: Page) {
  return page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)
}
