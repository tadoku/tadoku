import { randomUUID } from 'node:crypto'
import { existsSync, writeFileSync } from 'node:fs'
import { test, expect, type Browser, type BrowserContext, type Page, type TestInfo } from '@playwright/test'
import {
  adminEmail, adminPassword, expectContained, expectSettings, firstWrite, holdScripts, logIn, newPage, origins,
  readerEmail, readerPassword, waitForHydration,
} from './support'

// Failure modes: a delete control on Settings other than the quiet link, or a link that drops return_to; guests or
// old logins shown the form instead of a login that comes back to it; the fresh-login banner missing or the email not
// prefilled; a page that shows the address to type; a hydrated button that enables before the typed address matches, or
// a native form that can't submit; a mistyped address, an administrator or a stale session that deletes anything; a
// deletion that leaves the cookie or password usable, or doesn't land on /deleted; pages that overflow 390px.

const confirmLabel = 'Type your email address to confirm'
const banner = 'Privileged actions require a fresh login. Log in again to continue.'

type Disposable = { readonly email: string; readonly password: string }

function disposable(local = 'w2-delete'): Disposable {
  return { email: `${local}-${randomUUID()}@example.test`, password: `Delete-${randomUUID()}` }
}

async function register(page: Page, account: Disposable) {
  await page.goto('/register')
  await page.getByLabel('Display name').fill('Deleting Reader')
  await page.getByLabel('Email').fill(account.email)
  await page.getByLabel('Password').fill(account.password)
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expectSettings(page)
}

/** The account the context's cookie reads as on the tenant's main-host identity API, or null for a guest. */
async function sessionAccount(page: Page, testInfo: TestInfo) {
  const response = await page.request.get(`${origins(testInfo).main}/api/internal/identity/session`)
  expect(response.status()).toBe(200)
  return (await response.json() as { account: { id: string; email: string } | null }).account
}

async function expectDeletePage(page: Page) {
  await expect(page.getByRole('heading', { level: 1, name: 'Delete your account' })).toBeVisible()
  await expect(page.getByRole('textbox', { name: confirmLabel })).toBeVisible()
}

async function expectDeleted(page: Page, testInfo: TestInfo) {
  await expect(page).toHaveURL(`${origins(testInfo).authv2}/deleted`)
  await expect(page.getByRole('heading', { level: 1, name: 'Your account is deleted' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Go to Tadoku', exact: true })).toHaveAttribute('href', origins(testInfo).main)
  await expect(page.getByRole('button', { name: 'Log out', exact: true })).toHaveCount(0)
}

type Cookies = Awaited<ReturnType<BrowserContext['cookies']>>

/** After a deletion: the cleared context and a replay of its old cookies both read as guests, and login is refused. */
async function expectGone(browser: Browser, page: Page, testInfo: TestInfo, account: Disposable, cookies: Cookies) {
  expect(await sessionAccount(page, testInfo)).toBeNull()
  expect((await page.context().cookies()).filter(cookie => cookie.name.startsWith('tadoku_session'))).toEqual([])
  const replay = await newPage(browser, testInfo)
  await replay.context().addCookies(cookies)
  expect(await sessionAccount(replay, testInfo)).toBeNull()
  await logIn(replay, account.email, account.password)
  await expect(replay.getByText('That email and password don’t match.')).toBeVisible()
  await replay.context().close()
}

async function expectFreshLoginBanner(page: Page, email: string) {
  await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await expect(page.getByRole('status').filter({ hasText: banner })).toBeVisible()
  await expect(page.getByText('You’re logged in as')).toHaveCount(0)
  await expect(page.getByLabel('Email')).toHaveValue(email)
}

test('Settings links quietly to the delete page, and guests log in before it', async ({ browser }, testInfo) => {
  const { authv2, main } = origins(testInfo)
  const destination = `${main}/pages/dev-welcome?w2=delete`
  const deletePage = `${authv2}/delete-account?return_to=${encodeURIComponent(destination)}`
  const page = await newPage(browser, testInfo)

  await page.goto(deletePage)
  await expect(page).toHaveURL(`${authv2}/login?return_to=${encodeURIComponent(deletePage)}`)
  await expect(page.getByText(banner)).toHaveCount(0)
  await logIn(page, readerEmail, readerPassword(), new URL(page.url()).pathname + new URL(page.url()).search)
  await expect(page).toHaveURL(deletePage)
  await expectDeletePage(page)

  await page.getByRole('link', { name: 'Cancel', exact: true }).click()
  await expectSettings(page)
  expect(new URL(page.url()).searchParams.get('return_to')).toBe(destination)
  await waitForHydration(page)
  const settings = page.getByRole('main')
  const link = settings.getByRole('link', { name: 'Delete account', exact: true })
  const href = new URL(await link.getAttribute('href') ?? '', authv2)
  expect({ path: href.pathname, return_to: href.searchParams.get('return_to') })
    .toEqual({ path: '/delete-account', return_to: destination })
  await expect(settings.getByRole('link', { name: /delete/i })).toHaveCount(1)
  await expect(settings.getByRole('button', { name: /delete/i })).toHaveCount(0)
  await expect(page.locator('form[action="/actions/delete-account"]')).toHaveCount(0)

  await link.click()
  await expect(page).toHaveURL(deletePage)
  await expectDeletePage(page)

  await page.goto(`/delete-account?return_to=${encodeURIComponent('https://evil.example/steal')}`)
  await expectDeletePage(page)
  await expect(page.locator('a[href*="evil.example"], input[value*="evil.example"]')).toHaveCount(0)

  // A continuation Settings accepts, but too long to nest in the delete page's own, still comes back to the page.
  const long = `${main}/pages/dev-welcome?w2=${'x'.repeat(1_820)}`
  await page.goto(`/?return_to=${encodeURIComponent(long)}`)
  await expect(page.getByRole('link', { name: 'Back to Tadoku' })).toHaveAttribute('href', long)
  const guest = await newPage(browser, testInfo)
  await guest.goto(`/delete-account?return_to=${encodeURIComponent(long)}`)
  await expect(guest).toHaveURL(`${authv2}/login?return_to=${encodeURIComponent(`${authv2}/delete-account`)}`)
  await guest.context().close()
  await page.context().close()
})

test('a hydrated deletion waits for the address, ends the session and lands on /deleted', async ({ browser }, testInfo) => {
  const { main } = origins(testInfo)
  // The API lowercases one character at a time: Σ to σ even where JavaScript would write a final ς, and İ to i.
  const account = disposable('w2-delete-iοσ')
  const page = await newPage(browser, testInfo, { viewport: { width: 1280, height: 900 } })
  await register(page, account)
  await waitForHydration(page)
  await page.getByRole('link', { name: 'Delete account', exact: true }).click()
  await expectDeletePage(page)

  await expect(page.getByText(account.email)).toHaveCount(0)

  const field = page.getByRole('textbox', { name: confirmLabel })
  const submit = page.getByRole('button', { name: 'Delete my account', exact: true })
  await expect(field).toHaveAttribute('autocomplete', 'off')
  await expect(field).toHaveAttribute('spellcheck', 'false')
  const signedIn = await sessionAccount(page, testInfo)
  expect(signedIn).not.toBeNull()
  await expect(page.getByRole('link', { name: /export your logs/i }))
    .toHaveAttribute('href', `${main}/profile/${signedIn?.id}/updates`)

  // Before hydration the native form must stay usable; once hydrated it waits for a matching address.
  const scripts = await holdScripts(page)
  await page.reload({ waitUntil: 'domcontentloaded' })
  await expect(submit).toBeEnabled()
  await field.fill(`x${account.email}`)
  await scripts.release()
  await expect(field).toHaveValue(`x${account.email}`)
  await expect(submit).toBeDisabled()
  await field.fill(account.email.slice(0, -1))
  await expect(submit).toBeDisabled()
  const typed = `  ${account.email.toUpperCase().replace('I', 'İ')} `
  await field.fill(typed)
  await expect(submit).toBeEnabled()
  await page.screenshot({ path: testInfo.outputPath('delete-account-desktop-light.png'), fullPage: true })
  await page.emulateMedia({ colorScheme: 'dark' })
  await page.reload()
  await waitForHydration(page)
  await field.fill(typed)
  await page.screenshot({ path: testInfo.outputPath('delete-account-desktop-dark.png'), fullPage: true })

  const cookies = await page.context().cookies()
  const request = await firstWrite(page, () => submit.click())
  expect(request.url()).toBe(`${main}/api/internal/identity/settings/account-deletion`)
  expect(request.headers()['content-type']).toBe('application/json')
  expect(request.postDataJSON()).toEqual({ email: typed })
  await expectDeleted(page, testInfo)
  await page.screenshot({ path: testInfo.outputPath('deleted-desktop-dark.png'), fullPage: true })
  await expectGone(browser, page, testInfo, account, cookies)
  await page.context().close()
})

test('a deletion without JavaScript refuses a mistyped address, then deletes and lands on /deleted', async ({ browser }, testInfo) => {
  const { authv2 } = origins(testInfo)
  const account = disposable()
  const page = await newPage(browser, testInfo, { javaScriptEnabled: false, viewport: { width: 390, height: 844 } })
  await register(page, account)
  await page.getByRole('link', { name: 'Delete account', exact: true }).click()
  await expectDeletePage(page)
  await expectContained(page)
  await expect(page.getByText(account.email)).toHaveCount(0)

  const field = page.getByRole('textbox', { name: confirmLabel })
  const submit = page.getByRole('button', { name: 'Delete my account', exact: true })
  await field.fill(`other-${account.email}`)
  await submit.click()
  await expect(page).toHaveURL(`${authv2}/delete-account?error=confirmation_mismatch&form=delete-account`)
  await expect(page.getByRole('alert')).toContainText('doesn’t match')
  expect(await sessionAccount(page, testInfo)).toMatchObject({ email: account.email })
  await page.screenshot({ path: testInfo.outputPath('delete-account-mobile-mismatch.png'), fullPage: true })

  const cookies = await page.context().cookies()
  await field.fill(account.email)
  const request = await firstWrite(page, () => submit.click())
  expect(request.url()).toBe(`${authv2}/actions/delete-account`)
  await expectDeleted(page, testInfo)
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('deleted-mobile-light.png'), fullPage: true })
  await expectGone(browser, page, testInfo, account, cookies)
  await page.context().close()
})

test('administrators are refused and keep their account, with and without JavaScript', async ({ browser }, testInfo) => {
  for (const javaScriptEnabled of [true, false]) {
    const page = await newPage(browser, testInfo, { javaScriptEnabled })
    await logIn(page, adminEmail, adminPassword())
    await expectSettings(page)
    await page.goto('/delete-account')
    await expectDeletePage(page)
    if (javaScriptEnabled) await waitForHydration(page)
    await page.getByRole('textbox', { name: confirmLabel }).fill(adminEmail)
    await page.getByRole('button', { name: 'Delete my account', exact: true }).click()
    await expect(page.getByRole('alert')).toContainText('Administrator')
    if (!javaScriptEnabled) expect(new URL(page.url()).searchParams.get('error')).toBe('admin_account')
    expect(await sessionAccount(page, testInfo)).toMatchObject({ email: adminEmail })
    await page.context().close()
  }
})

// Run with AUTHV2_STALE_SESSION_DIR set. Once `deletion-ready` appears there, move every session of the account whose
// email it holds more than an hour back, then create `deletion-aged`.
test('old logins log in again under the fresh-login banner and come back to the page', async ({ browser }, testInfo) => {
  const directory = process.env.AUTHV2_STALE_SESSION_DIR
  test.skip(!directory, 'needs an operator to age the account’s sessions')
  const { authv2 } = origins(testInfo)
  const account = disposable()
  const reauth = `${authv2}/login?reauth=1&return_to=${encodeURIComponent(`${authv2}/delete-account`)}`

  // Each page is open before the hour runs out: two on the delete form, one on Settings.
  const hydrated = await newPage(browser, testInfo)
  await register(hydrated, account)
  await hydrated.goto('/delete-account')
  await waitForHydration(hydrated)
  const native = await newPage(browser, testInfo, { javaScriptEnabled: false })
  await logIn(native, account.email, account.password)
  await expectSettings(native)
  await native.goto('/delete-account')
  await expectDeletePage(native)
  const arriving = await newPage(browser, testInfo, { javaScriptEnabled: false, viewport: { width: 390, height: 844 } })
  await logIn(arriving, account.email, account.password)
  await expectSettings(arriving)

  writeFileSync(`${directory}/deletion-ready`, account.email)
  await expect.poll(() => existsSync(`${directory}/deletion-aged`), { timeout: 600_000, intervals: [2_000] }).toBe(true)

  // Opening the page with an old login redirects during server rendering.
  await arriving.getByRole('link', { name: 'Delete account', exact: true }).click()
  await expect(arriving).toHaveURL(reauth)
  await expectFreshLoginBanner(arriving, account.email)
  await expectContained(arriving)
  await arriving.screenshot({ path: testInfo.outputPath('login-reauth-mobile.png'), fullPage: true })
  await arriving.getByLabel('Password').fill('not-the-password')
  await arriving.getByRole('button', { name: 'Log in', exact: true }).click()
  await expect(arriving.getByRole('alert')).toContainText('don’t match')
  await expectFreshLoginBanner(arriving, account.email)
  await arriving.getByLabel('Password').fill(account.password)
  await arriving.getByRole('button', { name: 'Log in', exact: true }).click()
  await expect(arriving).toHaveURL(`${authv2}/delete-account`)
  await expectDeletePage(arriving)

  // A form opened while the login was recent sends its submission to the same login, both ways.
  for (const page of [native, hydrated]) {
    await page.getByRole('textbox', { name: confirmLabel }).fill(account.email)
    await page.getByRole('button', { name: 'Delete my account', exact: true }).click()
    await expect(page).toHaveURL(reauth)
    await expectFreshLoginBanner(page, account.email)
  }

  // A hydrated visit through the Settings link redirects in the browser too, and logging in comes back.
  await hydrated.goto('/')
  await waitForHydration(hydrated)
  await hydrated.getByRole('link', { name: 'Delete account', exact: true }).click()
  await expect(hydrated).toHaveURL(reauth)
  await expectFreshLoginBanner(hydrated, account.email)
  await hydrated.getByLabel('Password').fill(account.password)
  await hydrated.getByRole('button', { name: 'Log in', exact: true }).click()
  await expect(hydrated).toHaveURL(`${authv2}/delete-account`)
  await expectDeletePage(hydrated)
  expect(await sessionAccount(hydrated, testInfo)).toMatchObject({ email: account.email })
  for (const page of [hydrated, native, arriving]) await page.context().close()
})
