import { expect, test, type Page } from '@playwright/test'

const base = process.env.WEBV3_BASE_URL
const api = process.env.TADOKU_API_URL
const email = process.env.E2E_READER_EMAIL
const password = process.env.E2E_READER_PASSWORD

test.use({ trace: 'off', video: 'off', ignoreHTTPSErrors: false })

async function signIn(page: Page, identifier = email!, secret = password!) {
  await page.goto('/logs/new')
  await page.getByRole('link', { name: 'Log in', exact: true }).first().click()
  await page.locator('input[name="identifier"]').fill(identifier)
  await page.locator('input[name="password"]').fill(secret)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL(`${base}/logs/new`)
  await expect(page.getByRole('button', { name: 'Save activity', exact: true })).toBeEnabled({ timeout: 15_000 })
}

test('native activity link opens an anonymous sign-in prompt', async ({ page }) => {
  test.skip(!base, 'Set WEBV3_BASE_URL')
  await page.goto('/plan')
  const link = page.getByRole('link', { name: 'Log activity', exact: true }).first()
  await expect(link).toHaveAttribute('href', '/logs/new')
  await link.click()
  await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
})

for (const [activity, field, quantity] of [['Reading', 'Amount', '17'], ['Listening', 'Minutes', '2.5']]) {
  test(`${activity}: validation, one save, persisted API detail and reload`, async ({ page, context }, testInfo) => {
    test.skip(!base || !api || !email || !password, 'Set development URLs and fixture credentials')
    await signIn(page)
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
    const response = await context.request.get(`${api}/immersion/logs/${id}`)
    expect(response.status()).toBe(200)
    expect(response.headers()['x-dev-selected']).toBe(new URL(base!).hostname.split('.next.')[0])
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
    const listed = await context.request.get(`${api}/immersion/users/${record.user_id}/logs?page_size=100&page=0`)
    expect((await listed.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(1)
    await testInfo.attach('persisted-record', { body: JSON.stringify(record, null, 2), contentType: 'application/json' })
    await page.screenshot({ path: testInfo.outputPath(`${activity}-saved.png`), fullPage: true })
  })
}

test('transport failure keeps values and never navigates or claims success', async ({ page }) => {
  test.skip(!base || !email || !password, 'Set development fixture credentials')
  await signIn(page)
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
  test(`keyboard-only activity form at ${width}px`, async ({ page }, testInfo) => {
    test.skip(!base || !email || !password, 'Set development fixture credentials')
    await page.setViewportSize({ width, height: 900 })
    const errors: string[] = []
    page.on('pageerror', error => errors.push(error.message))
    await signIn(page)
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

test('expired session keeps draft, rejects save and recovers through existing account login', async ({ page, context }) => {
  test.skip(!base || !email || !password, 'Set development fixture credentials')
  await signIn(page)
  const description = `expired-session-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('4')
  await page.getByLabel('Description', { exact: true }).fill(description)
  const cookie = (await context.cookies()).find(cookie => cookie.name === 'ory_kratos_session')!
  await context.addCookies([{ ...cookie, value: 'expired' }])
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Your session ended')
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('4')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
  const opened = context.waitForEvent('page')
  await page.getByRole('link', { name: 'Log in in a new tab', exact: true }).click()
  const login = await opened
  await login.locator('input[name="identifier"]').fill(email!)
  await login.locator('input[name="password"]').fill(password!)
  await login.locator('button[type="submit"]').click()
  await expect(login.getByRole('button', { name: 'Save activity', exact: true })).toBeEnabled()
  await login.close()
  await page.bringToFront()
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
})

test('server function rejects foreign and absent Origin without creating activity', async ({ page, context }) => {
  test.skip(!base || !api || !email || !password, 'Set development fixture credentials')
  await signIn(page)
  const description = `csrf-check-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('5')
  await page.getByLabel('Description', { exact: true }).fill(description)
  let refused = 0
  await page.route('**/*', async route => {
    const request = route.request()
    if (request.method() !== 'POST' || !request.url().includes('_serverFn')) return route.continue()
    const scenarios: Record<string, string>[] = [{ Origin: 'https://evil.tadoku.dev.lab', 'Sec-Fetch-Site': 'same-site' }, { 'Sec-Fetch-Site': 'same-origin' }]
    for (const headers of scenarios) {
      const response = await context.request.post(request.url(), { data: request.postData()!, headers: { 'Content-Type': request.headers()['content-type'], ...headers } })
      expect(response.status()).toBe(403)
      refused++
    }
    await route.abort()
  })
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('could not confirm')
  expect(refused).toBe(2)
  const configuration = await context.request.get(`${api}/immersion/logs/configuration-options`)
  expect(configuration.status()).toBe(200)
  const who = await context.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')
  const userId = (await who.json()).identity.id
  const list = await context.request.get(`${api}/immersion/users/${userId}/logs?page_size=100&page=0`)
  expect((await list.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(0)
})

test('API rejection preserves entered activity instead of showing success', async ({ page }) => {
  test.skip(!base || !email || !password, 'Set development fixture credentials')
  await signIn(page)
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

test('public detail hides contest management and the canonical tenant cannot read branch activity', async ({ page, context, browser }) => {
  test.skip(!base || !api || !email || !password, 'Set development fixture credentials')
  await signIn(page)
  await page.getByLabel('Amount', { exact: true }).fill('6')
  await page.getByLabel('Description', { exact: true }).fill(`isolation-${Date.now()}`)
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  const id = new URL(page.url()).pathname.split('/').pop()!
  const owner = await context.request.get(`${api}/immersion/logs/${id}`)
  const log = await owner.json()
  expect(log.registrations.length).toBeGreaterThan(0)
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toBeVisible()
  const publicContext = await browser.newContext()
  try {
    const publicPage = await publicContext.newPage()
    await publicPage.goto(`${base}/logs/${id}`)
    await expect(publicPage.getByText(id, { exact: true })).toBeVisible()
    await expect(publicPage.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
    const publicRecord = await publicContext.request.get(`${api}/immersion/logs/${id}`)
    expect((await publicRecord.json()).registrations).toEqual([])
    const otherTenant = await context.request.get(`https://tadoku.dev.lab/api/internal/immersion/logs/${id}`)
    expect(otherTenant.status()).toBe(404)
  } finally { await publicContext.close() }
})

test('a different signed-in learner cannot change another learner\'s activity', async ({ page, context, browser }) => {
  const adminEmail = process.env.E2E_SECOND_EMAIL
  const adminPassword = process.env.E2E_SECOND_PASSWORD
  test.skip(!base || !api || !email || !password || !adminEmail || !adminPassword, 'Set both development fixture identities')
  await signIn(page)
  const adminContext = await browser.newContext()
  try {
    const admin = await adminContext.newPage()
    await signIn(admin, adminEmail!, adminPassword!)
    await admin.getByLabel('Amount', { exact: true }).fill('9')
    await admin.getByLabel('Description', { exact: true }).fill(`ownership-${Date.now()}`)
    await admin.getByRole('button', { name: 'Save activity', exact: true }).click()
    await expect(admin).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
    const id = new URL(admin.url()).pathname.split('/').pop()!
    await page.goto(`${base}/logs/${id}`)
    await expect(page.getByText(id, { exact: true })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
    const path = `${api}/immersion/logs/${id}`
    const before = await (await context.request.get(path)).json()
    expect(before.registrations).toEqual([])
    const update = await context.request.put(path, { data: { language_code: before.language.code, activity_id: before.activity.id, amount: 99, unit_id: before.unit_id, tags: [], description: 'unauthorized change' } })
    expect(update.status()).toBe(403)
    const removal = await context.request.delete(path)
    expect(removal.status()).toBe(403)
    const after = await (await context.request.get(path)).json()
    expect(after).toEqual(before)
  } finally { await adminContext.close() }
})

test('branch-only suspension rejects an open draft and renders the suspended account state', async ({ page, context, browser }, testInfo) => {
  const adminEmail = process.env.E2E_SECOND_EMAIL
  const adminPassword = process.env.E2E_SECOND_PASSWORD
  test.skip(!base || !api || !email || !password || !adminEmail || !adminPassword, 'Set both development fixture identities')
  await signIn(page)
  await page.getByLabel('Amount', { exact: true }).fill('7')
  await page.getByLabel('Description', { exact: true }).fill('suspended draft stays here')
  const who = await context.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')
  const userId = (await who.json()).identity.id
  const adminContext = await browser.newContext()
  let suspended = false
  try {
    const admin = await adminContext.newPage()
    await signIn(admin, adminEmail!, adminPassword!)
    const response = await adminContext.request.put(`${api}/authz/users/${userId}/role`, { data: { role: 'banned', reason: 'Branch-only Webv3 verification fixture' } })
    expect(response.status()).toBe(200)
    expect(response.headers()['x-dev-selected']).toBe(new URL(base!).hostname.split('.next.')[0])
    suspended = true
    await page.getByRole('button', { name: 'Save activity', exact: true }).click()
    await expect(page.getByRole('alert')).toContainText('cannot save activity')
    await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('7')
    await expect(page.getByLabel('Description', { exact: true })).toHaveValue('suspended draft stays here')
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Your account is suspended' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Log out', exact: true })).toBeVisible()
    await page.screenshot({ path: testInfo.outputPath('suspended.png'), fullPage: true })
  } finally {
    if (suspended) {
      const restored = await adminContext.request.put(`${api}/authz/users/${userId}/role`, { data: { role: 'user', reason: 'Finish branch-only Webv3 verification fixture' } })
      expect(restored.status()).toBe(200)
      const role = await context.request.get(`${api}/authz/current-user/role`)
      expect((await role.json()).role).toBe('user')
    }
    await adminContext.close()
  }
})

test('a draft cannot be saved under an account switched in another tab', async ({ page, context }, testInfo) => {
  const adminEmail = process.env.E2E_SECOND_EMAIL
  const adminPassword = process.env.E2E_SECOND_PASSWORD
  test.skip(!base || !api || !email || !password || !adminEmail || !adminPassword, 'Set both development fixture identities')
  await signIn(page)
  const reader = await (await context.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')).json()
  const description = `account-switch-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('11')
  await page.getByLabel('Description', { exact: true }).fill(description)
  for (const checkbox of await page.getByRole('group', { name: 'Contests', exact: true }).getByRole('checkbox').all()) await checkbox.uncheck()
  const other = await context.newPage()
  await other.goto(`${base}/plan`)
  await other.getByRole('button', { name: 'Log out', exact: true }).click()
  await signIn(other, adminEmail!, adminPassword!)
  const admin = await (await context.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')).json()
  await page.bringToFront()
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('different account. Nothing was saved')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('11')
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Admin', { exact: true })).toBeVisible()
  for (const userId of [reader.identity.id, admin.identity.id]) {
    const list = await context.request.get(`${api}/immersion/users/${userId}/logs?page_size=100&page=0`)
    expect((await list.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(0)
  }
  await page.screenshot({ path: testInfo.outputPath('account-switch-blocked.png'), fullPage: true })
})

test('public activity detail still loads with an expired session cookie and has a title', async ({ page, context }) => {
  test.skip(!base || !api || !email || !password, 'Set development fixture credentials')
  await signIn(page)
  await page.getByLabel('Amount', { exact: true }).fill('2')
  await page.getByLabel('Description', { exact: true }).fill(`public-expired-${Date.now()}`)
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  const id = new URL(page.url()).pathname.split('/').pop()!
  const cookie = (await context.cookies()).find(cookie => cookie.name === 'ory_kratos_session')!
  await context.addCookies([{ ...cookie, value: 'expired' }])
  await page.reload()
  await expect(page.getByText(id, { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
  await expect(page).toHaveTitle('Activity · Tadoku')
})

for (const change of ['expiry', 'account switch']) {
  test(`an anonymously opened form keeps its first signed-in draft after ${change}`, async ({ page, context }, testInfo) => {
    const adminEmail = process.env.E2E_SECOND_EMAIL
    const adminPassword = process.env.E2E_SECOND_PASSWORD
    test.skip(!base || !api || !email || !password || !adminEmail || !adminPassword, 'Set both development fixture identities')
    await page.clock.install()
    await page.goto('/logs/new')
    await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
    const other = await context.newPage()
    await signIn(other)
    await page.clock.fastForward(31_000)
    await page.bringToFront()
    await page.evaluate(() => window.dispatchEvent(new Event('visibilitychange')))
    await expect(page.getByRole('button', { name: 'Save activity', exact: true })).toBeEnabled()
    const description = `anonymous-entry-${change}-${Date.now()}`
    await page.getByLabel('Amount', { exact: true }).fill('13')
    await page.getByLabel('Description', { exact: true }).fill(description)
    await other.bringToFront()
    await other.getByRole('button', { name: 'Log out', exact: true }).click()
    if (change === 'account switch') await signIn(other, adminEmail!, adminPassword!)
    await page.bringToFront()
    if (change === 'expiry') {
      await page.clock.fastForward(31_000)
      await page.evaluate(() => window.dispatchEvent(new Event('visibilitychange')))
      await expect(page.getByRole('link', { name: 'Log in', exact: true }).first()).toBeVisible()
    }
    await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('13')
    await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
    await page.getByRole('button', { name: 'Save activity', exact: true }).click()
    await expect(page.getByRole('alert')).toContainText(change === 'expiry' ? 'Your session ended' : 'different account. Nothing was saved')
    await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('13')
    await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
    if (change === 'account switch') {
      await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Admin', { exact: true })).toBeVisible()
      await expect(page.getByRole('link', { name: 'Start a new activity for this account' })).toBeVisible()
    }
    await page.screenshot({ path: testInfo.outputPath(`anonymous-entry-${change.replace(' ', '-')}.png`), fullPage: true })
  })
}
