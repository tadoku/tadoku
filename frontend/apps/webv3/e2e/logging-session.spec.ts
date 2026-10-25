import { accountId, admin, api, authenticate, branch, expect, openForm, reader, sessionCookie, signIn, test, webv3 } from './logging-fixtures'

test('expired session keeps draft, rejects save and recovers through authv2 login', async ({ page, readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  const description = `expired-session-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('4')
  await page.getByLabel('Description', { exact: true }).fill(description)
  const cookie = await sessionCookie(readerContext)
  await readerContext.addCookies([{ ...cookie, value: 'expired' }])
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Your session ended')
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('4')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
  const opened = readerContext.waitForEvent('page')
  await page.getByRole('link', { name: 'Log in in a new tab', exact: true }).click()
  const login = await opened
  await authenticate(login)
  await expect(login.getByRole('button', { name: 'Save activity', exact: true })).toBeEnabled()
  await login.close()
  await page.bringToFront()
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
})

test('server function rejects foreign and absent Origin without creating activity', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  const description = `csrf-check-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('5')
  await page.getByLabel('Description', { exact: true }).fill(description)
  let refused = 0
  await page.route('**/*', async route => {
    const request = route.request()
    if (request.method() !== 'POST' || !request.url().includes('_serverFn')) return route.continue()
    const scenarios: Record<string, string>[] = [{ Origin: 'https://evil.tadoku.dev.lab', 'Sec-Fetch-Site': 'same-site' }, { 'Sec-Fetch-Site': 'same-origin' }]
    for (const headers of scenarios) {
      const response = await readerContext.request.post(request.url(), { data: request.postData()!, headers: { 'Content-Type': request.headers()['content-type'], ...headers } })
      expect(response.status()).toBe(403)
      refused++
    }
    await route.abort()
  })
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('could not confirm')
  expect(refused).toBe(2)
  const userId = await accountId(readerContext)
  const list = await readerContext.request.get(`${api}/immersion/users/${userId}/logs?page_size=100&page=0`)
  expect((await list.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(0)
})

test('branch-only suspension rejects an open draft and renders the suspended account state', async ({ page, readerContext, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  await openForm(page)
  await page.getByLabel('Amount', { exact: true }).fill('7')
  await page.getByLabel('Description', { exact: true }).fill('suspended draft stays here')
  const userId = await accountId(readerContext)
  // Prove the administrator's requests reach this branch before changing any role.
  const selected = await adminContext.request.get(`${api}/authz/current-user/role`)
  expect(selected.headers()['x-dev-selected']).toBe(branch)
  expect((await selected.json()).role).toBe('admin')
  const role = `${api}/authz/users/${userId}/role`
  let suspended = false
  try {
    const response = await adminContext.request.put(role, { data: { role: 'banned', reason: 'Branch-only Webv3 verification fixture' } })
    suspended = response.ok()
    expect(response.status()).toBe(200)
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
      const restored = await adminContext.request.put(role, { data: { role: 'user', reason: 'Finish branch-only Webv3 verification fixture' } })
      expect(restored.status()).toBe(200)
      const current = await readerContext.request.get(`${api}/authz/current-user/role`)
      expect((await current.json()).role).toBe('user')
    }
  }
})

test('a draft cannot be saved under an account switched in another tab', async ({ page, context }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  await signIn(page)
  const readerId = await accountId(context)
  const description = `account-switch-${Date.now()}`
  await page.getByLabel('Amount', { exact: true }).fill('11')
  await page.getByLabel('Description', { exact: true }).fill(description)
  for (const checkbox of await page.getByRole('group', { name: 'Contests', exact: true }).getByRole('checkbox').all()) await checkbox.uncheck()
  const other = await context.newPage()
  await other.goto(`${webv3}/plan`)
  await other.getByRole('button', { name: 'Log out', exact: true }).click()
  await signIn(other, admin)
  const adminId = await accountId(context)
  await page.bringToFront()
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('different account. Nothing was saved')
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
  await expect(page.getByLabel('Amount', { exact: true })).toHaveValue('11')
  await expect(page.getByRole('navigation', { name: 'Main navigation' }).getByText('Dev Admin', { exact: true })).toBeVisible()
  for (const userId of [readerId, adminId]) {
    const list = await context.request.get(`${api}/immersion/users/${userId}/logs?page_size=100&page=0`)
    expect((await list.json()).logs.filter((log: { description?: string }) => log.description === description)).toHaveLength(0)
  }
  await page.screenshot({ path: testInfo.outputPath('account-switch-blocked.png'), fullPage: true })
})

test('public activity detail still loads with an expired session cookie and has a title', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  await page.getByLabel('Amount', { exact: true }).fill('2')
  await page.getByLabel('Description', { exact: true }).fill(`public-expired-${Date.now()}`)
  await page.getByRole('button', { name: 'Save activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/[a-f0-9-]{36}$/)
  const id = new URL(page.url()).pathname.split('/').pop()!
  const cookie = await sessionCookie(readerContext)
  await readerContext.addCookies([{ ...cookie, value: 'expired' }])
  await page.reload()
  await expect(page.getByText(id, { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
  await expect(page).toHaveTitle('Activity · Tadoku')
})

for (const change of ['expiry', 'account switch']) {
  test(`an anonymously opened form keeps its first signed-in draft after ${change}`, async ({ page, context }, testInfo) => {
    test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
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
    if (change === 'account switch') await signIn(other, admin)
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
