import { admin, api, branch, expect, openForm, reader, test, webv3 } from './logging-fixtures'

test('native activity link opens an anonymous sign-in prompt', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/plan')
  const link = page.getByRole('link', { name: 'Log activity', exact: true }).first()
  await expect(link).toHaveAttribute('href', '/logs/new')
  await link.click()
  await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
})

for (const [activity, field, quantity] of [['Reading', 'Amount', '17'], ['Listening', 'Minutes', '2.5']]) {
  test(`${activity}: validation, one save, persisted API detail and reload`, async ({ page, readerContext }, testInfo) => {
    test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
    await openForm(page)
    await page.getByLabel('Activity', { exact: true }).selectOption({ label: activity })
    await page.getByLabel('Language', { exact: true }).selectOption('jpn')
    const save = page.getByRole('button', { name: 'Save activity', exact: true })
    await save.click()
    await expect(page.getByRole('alert').first()).toBeVisible()
    await expect(page.getByLabel(field, { exact: true })).toBeFocused()
    const description = `webv3-e2e-${activity}-${Date.now()}`
    await page.getByLabel(field, { exact: true }).fill(quantity)
    await page.getByLabel('Description', { exact: true }).fill('x'.repeat(256))
    await save.click()
    await expect(page.getByLabel('Description', { exact: true })).toBeFocused()
    await page.getByLabel('Description', { exact: true }).fill(description)
    await page.getByLabel('Tags', { exact: true }).fill(Array.from({ length: 11 }, (_, index) => `tag${index}`).join(','))
    await save.click()
    await expect(page.getByLabel('Tags', { exact: true })).toBeFocused()
    await page.getByLabel('Tags', { exact: true }).fill('book, fiction')
    let posts = 0
    await page.route('**/*', async route => {
      if (route.request().method() === 'POST' && route.request().url().includes('_serverFn')) {
        posts++
        await new Promise(resolve => setTimeout(resolve, 600))
      }
      await route.continue()
    })
    await page.evaluate(() => { const form = document.querySelector<HTMLFormElement>('form[aria-busy]')!; form.requestSubmit(); form.requestSubmit() })
    await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
    expect(posts).toBe(1)
    const id = new URL(page.url()).pathname.split('/').pop()!
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
    await page.reload()
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
    const response = await readerContext.request.get(`${api}/immersion/logs/${id}`)
    expect(response.status()).toBe(200)
    expect(response.headers()['x-dev-selected']).toBe(branch)
    const record = await response.json()
    expect(record.id).toBe(id)
    expect(record.description).toBe(description)
    expect(record.tags).toEqual(['book', 'fiction'])
    expect(record.language.code).toBe('jpn')
    expect(record.activity.name).toBe(activity)
    expect(record.registrations.length).toBeGreaterThan(0)
    await expect(page.getByText(record.id, { exact: true })).toBeVisible()
    for (const registration of record.registrations) await expect(page.getByRole('link', { name: registration.title, exact: true })).toBeVisible()
    if (activity === 'Listening') expect(record.duration_seconds).toBe(150)
    else expect(record.amount).toBe(17)
    const listed = await readerContext.request.get(`${api}/immersion/users/${record.user_id}/logs?page_size=100&page=0`)
    expect((await listed.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(1)
    await testInfo.attach('persisted-record', { body: JSON.stringify(record, null, 2), contentType: 'application/json' })
    await page.screenshot({ path: testInfo.outputPath(`${activity}-saved.png`), fullPage: true })
  })
}

test('transport failure keeps values and never navigates or claims success', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await page.getByLabel('Amount', { exact: true }).fill('12')
  await page.getByLabel('Description', { exact: true }).fill('preserve my draft')
  await page.route('**/*', route => route.request().method() === 'POST' && route.request().url().includes('_serverFn') ? route.abort() : route.continue())
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText(/check your activity/i)
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('12')
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
    await expect(page.getByLabel('Amount', { exact: true })).toBeFocused()
    await page.keyboard.type('3')
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Unit', { exact: true })).toBeFocused()
    await page.keyboard.press('Tab')
    const description = `keyboard-${width}-${Date.now()}`
    await page.keyboard.type(description)
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Tags', { exact: true })).toBeFocused()
    await page.keyboard.type('keyboard')
    for (let step = 0; step < 8 && !await page.getByRole('button', { name: 'Save activity', exact: true }).evaluate(element => element === document.activeElement); step++) await page.keyboard.press('Tab')
    await expect(page.getByRole('button', { name: 'Save activity', exact: true })).toBeFocused()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`keyboard-${width}-form.png`), fullPage: true })
    await page.keyboard.press('Enter')
    await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
    await expect(page.getByRole('heading', { level: 1 })).toBeFocused()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    expect(errors).toEqual([])
    await page.screenshot({ path: testInfo.outputPath(`keyboard-${width}-saved.png`), fullPage: true })
  })
}

test('API rejection preserves entered activity instead of showing success', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await page.getByLabel('Amount', { exact: true }).fill('8')
  await page.getByLabel('Description', { exact: true }).fill('rejected activity draft')
  let changed = false
  await page.route('**/*', async route => {
    const request = route.request()
    if (request.method() !== 'POST' || !request.url().includes('_serverFn')) return route.continue()
    const original = request.postData()!
    const payload = JSON.parse(original)
    const data = payload.t.p.v[0]
    data.p.v[data.p.k.indexOf('registrationIds')].a = [{ t: 1, s: '00000000-0000-4000-8000-000000000000' }]
    const body = JSON.stringify(payload)
    expect(body).not.toBe(original)
    changed = true
    await route.continue({ postData: body })
  })
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('not saved')
  expect(changed).toBe(true)
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('8')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue('rejected activity draft')
  await expect(page).toHaveURL(/\/logs\/new$/)
})

test('public detail hides contest management and the canonical tenant cannot read branch activity', async ({ page, readerContext, browser }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  await page.getByLabel('Amount', { exact: true }).fill('6')
  await page.getByLabel('Description', { exact: true }).fill(`isolation-${Date.now()}`)
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  const id = new URL(page.url()).pathname.split('/').pop()!
  const owner = await readerContext.request.get(`${api}/immersion/logs/${id}`)
  const log = await owner.json()
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
    const otherTenant = await readerContext.request.get(`https://tadoku.dev.lab/api/internal/immersion/logs/${id}`)
    expect(otherTenant.status()).toBe(404)
  } finally { await publicContext.close() }
})

test('a different signed-in learner cannot change another learner\'s activity', async ({ page, readerContext, adminContext }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  const adminPage = await adminContext.newPage()
  await openForm(adminPage)
  await adminPage.getByLabel('Amount', { exact: true }).fill('9')
  await adminPage.getByLabel('Description', { exact: true }).fill(`ownership-${Date.now()}`)
  await adminPage.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(adminPage).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  const id = new URL(adminPage.url()).pathname.split('/').pop()!
  await page.goto(`${webv3}/logs/${id}`)
  await expect(page.getByText(id, { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
  const path = `${api}/immersion/logs/${id}`
  const before = await (await readerContext.request.get(path)).json()
  expect(before.registrations).toEqual([])
  const update = await readerContext.request.put(path, { data: { language_code: before.language.code, activity_id: before.activity.id, amount: 99, unit_id: before.unit_id, tags: [], description: 'unauthorized change' } })
  expect(update.status()).toBe(403)
  const removal = await readerContext.request.delete(path)
  expect(removal.status()).toBe(403)
  const after = await (await readerContext.request.get(path)).json()
  expect(after).toEqual(before)
})
