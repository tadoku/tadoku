import type { Page } from '@playwright/test'
import {
  accountId,
  admin,
  api,
  authenticate,
  branch,
  expect,
  expectDraftKept,
  fillDraft,
  isSaveRequest,
  logsDescribed,
  openForm,
  reader,
  saveButton,
  savedLogId,
  savedLogUrl,
  sessionCookie,
  signIn,
  test,
  webv3,
} from './logging-fixtures'

const header = (page: Page) => page.getByRole('navigation', { name: 'Main navigation' })

test('expired session keeps draft, rejects save and recovers through authv2 login', async ({ page, readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const description = `expired-session-${Date.now()}`
  await openForm(page)
  await fillDraft(page, '4', description)
  await readerContext.addCookies([{ ...await sessionCookie(readerContext), value: 'expired' }])

  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText('Your session ended')
  await expectDraftKept(page, '4', description)

  const opened = readerContext.waitForEvent('page')
  await page.getByRole('link', { name: 'Log in in a new tab', exact: true }).click()
  const login = await opened
  await authenticate(login)
  await expect(saveButton(login)).toBeEnabled()
  await login.close()

  await page.bringToFront()
  await saveButton(page).click()
  await expect(page).toHaveURL(savedLogUrl)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(description)
})

test('server function rejects foreign and absent Origin without creating activity', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  const description = `csrf-check-${Date.now()}`
  await openForm(page)
  await fillDraft(page, '5', description)

  // Replay the browser's save with a foreign Origin and with none, then drop the original request.
  let refused = 0
  await page.route('**/*', async route => {
    const request = route.request()
    if (!isSaveRequest(request)) return route.continue()
    const attempts: Record<string, string>[] = [
      { Origin: 'https://evil.tadoku.dev.lab', 'Sec-Fetch-Site': 'same-site' },
      { 'Sec-Fetch-Site': 'same-origin' },
    ]
    for (const headers of attempts) {
      const response = await readerContext.request.post(request.url(), {
        data: request.postData()!,
        headers: { 'Content-Type': request.headers()['content-type'], ...headers },
      })
      expect(response.status()).toBe(403)
      refused++
    }
    await route.abort()
  })

  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText('could not confirm')
  expect(refused).toBe(2)
  expect(await logsDescribed(readerContext, await accountId(readerContext), description)).toHaveLength(0)
})

test('branch-only suspension rejects an open draft and renders the suspended account state', async ({ page, readerContext, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  await openForm(page)
  await fillDraft(page, '7', 'suspended draft stays here')
  const role = `${api}/authz/users/${await accountId(readerContext)}/role`

  // Prove the administrator's requests reach this branch before changing any role.
  const current = await adminContext.request.get(`${api}/authz/current-user/role`)
  expect(current.headers()['x-dev-selected']).toBe(branch)
  expect((await current.json()).role).toBe('admin')

  let suspended = false
  try {
    const response = await adminContext.request.put(role, {
      data: { role: 'banned', reason: 'Branch-only Webv3 verification fixture' },
    })
    suspended = response.ok()
    expect(response.status()).toBe(200)

    await saveButton(page).click()
    await expect(page.getByRole('alert')).toContainText('cannot save activity')
    await expectDraftKept(page, '7', 'suspended draft stays here')
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Your account is suspended' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Log out', exact: true })).toBeVisible()
    await page.screenshot({ path: testInfo.outputPath('suspended.png'), fullPage: true })
  } finally {
    if (suspended) {
      const restored = await adminContext.request.put(role, {
        data: { role: 'user', reason: 'Finish branch-only Webv3 verification fixture' },
      })
      expect(restored.status()).toBe(200)
      const own = await readerContext.request.get(`${api}/authz/current-user/role`)
      expect((await own.json()).role).toBe('user')
    }
  }
})

test('a draft cannot be saved under an account switched in another tab', async ({ page, context }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  const description = `account-switch-${Date.now()}`
  await signIn(page)
  const readerId = await accountId(context)
  await fillDraft(page, '11', description)
  for (const contest of await page.getByRole('group', { name: 'Contests', exact: true }).getByRole('checkbox').all()) {
    await contest.uncheck()
  }

  const other = await context.newPage()
  await other.goto(`${webv3}/plan`)
  await other.getByRole('button', { name: 'Log out', exact: true }).click()
  await signIn(other, admin)
  const adminId = await accountId(context)

  await page.bringToFront()
  await saveButton(page).click()
  await expect(page.getByRole('alert')).toContainText('different account. Nothing was saved')
  await expectDraftKept(page, '11', description)
  await expect(header(page).getByText('Dev Admin', { exact: true })).toBeVisible()
  for (const userId of [readerId, adminId]) {
    expect(await logsDescribed(context, userId, description)).toHaveLength(0)
  }
  await page.screenshot({ path: testInfo.outputPath('account-switch-blocked.png'), fullPage: true })
})

test('public activity detail still loads with an expired session cookie and has a title', async ({ page, readerContext }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development fixture credentials')
  await openForm(page)
  await fillDraft(page, '2', `public-expired-${Date.now()}`)
  await saveButton(page).click()
  await expect(page).toHaveURL(savedLogUrl)
  const id = savedLogId(page)

  await readerContext.addCookies([{ ...await sessionCookie(readerContext), value: 'expired' }])
  await page.reload()
  await expect(page.getByText(id, { exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Manage contest submissions' })).toHaveCount(0)
  await expect(page).toHaveTitle('Activity · Tadoku')
})

for (const change of ['expiry', 'account switch']) {
  test(`an anonymously opened form keeps its first signed-in draft after ${change}`, async ({ page, context }, testInfo) => {
    test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
    const description = `anonymous-entry-${change}-${Date.now()}`
    const refocusAfterViewerGoesStale = async () => {
      await page.clock.fastForward(31_000)
      await page.bringToFront()
      await page.evaluate(() => window.dispatchEvent(new Event('visibilitychange')))
    }

    // Open the form signed out, log in in another tab, and let this tab notice the session on refocus.
    await page.clock.install()
    await page.goto('/logs/new')
    await expect(page.getByRole('heading', { name: 'Sign in to log activity' })).toBeVisible()
    const other = await context.newPage()
    await signIn(other)
    await refocusAfterViewerGoesStale()
    await expect(saveButton(page)).toBeEnabled()
    await fillDraft(page, '13', description)

    // End or switch the session in the other tab.
    await other.bringToFront()
    await other.getByRole('button', { name: 'Log out', exact: true }).click()
    if (change === 'account switch') await signIn(other, admin)
    await page.bringToFront()
    if (change === 'expiry') {
      await refocusAfterViewerGoesStale()
      await expect(page.getByRole('link', { name: 'Log in', exact: true }).first()).toBeVisible()
    }

    await expectDraftKept(page, '13', description)
    await saveButton(page).click()
    const refusal = change === 'expiry' ? 'Your session ended' : 'different account. Nothing was saved'
    await expect(page.getByRole('alert')).toContainText(refusal)
    await expectDraftKept(page, '13', description)
    if (change === 'account switch') {
      await expect(header(page).getByText('Dev Admin', { exact: true })).toBeVisible()
      await expect(page.getByRole('link', { name: 'Start a new activity for this account' })).toBeVisible()
    }
    await page.screenshot({ path: testInfo.outputPath(`anonymous-entry-${change.replace(' ', '-')}.png`), fullPage: true })
  })
}
