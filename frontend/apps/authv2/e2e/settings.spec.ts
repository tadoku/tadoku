import { randomUUID } from 'node:crypto'
import { existsSync, writeFileSync } from 'node:fs'
import { test, expect } from '@playwright/test'
import {
  adminEmail, adminPassword, expectContained, expectSettings, logIn, newPage, openSetting, origins, readerEmail,
  readerPassword, settingRow, settingValue, waitForHydration,
} from './support'

// Failure modes: guests stranded on Settings or returned without their continuation; profile and email changes that
// do not persist or are not explained; the old address never told about an email change; crafted continuations
// rendered as links; foreign-origin native posts accepted; logout that leaves a usable session; stale sessions
// without a way to log in again and come back to the editor they were using.

test('guests log in and come back to Settings with its continuation', async ({ browser }, testInfo) => {
  const { authv2, main } = origins(testInfo)
  const page = await newPage(browser, testInfo)
  const destination = `${main}/pages/dev-welcome?w2=settings`
  const settings = `${authv2}/settings?return_to=${encodeURIComponent(destination)}`

  await page.goto(settings)
  await expect(page).toHaveURL(`${authv2}/login?return_to=${encodeURIComponent(settings)}`)
  await logIn(page, readerEmail, readerPassword(), new URL(page.url()).pathname + new URL(page.url()).search)
  await expect(page).toHaveURL(settings)
  await expectSettings(page)
  await expect(page.getByRole('link', { name: 'Back to Tadoku' })).toHaveAttribute('href', destination)

  await page.goto(`/?return_to=${encodeURIComponent('https://evil.example/steal')}`)
  await expectSettings(page)
  await expect(page.locator('a[href*="evil.example"], input[value*="evil.example"]')).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Back to Tadoku' })).toHaveCount(0)
  await page.context().close()
})

test('profile and email changes persist, explain problems and notify the old address', async ({ browser }, testInfo) => {
  const { main } = origins(testInfo)
  const page = await newPage(browser, testInfo, { viewport: { width: 1280, height: 900 } })
  await logIn(page, readerEmail, readerPassword())
  await expectSettings(page)
  await waitForHydration(page)

  const profile = settingRow(page, 'Display name')
  await openSetting(page, 'Display name')
  await profile.getByRole('textbox', { name: 'Display name' }).fill('x')
  await profile.getByRole('button', { name: 'Save display name' }).click()
  await expect(profile.getByRole('textbox', { name: 'Display name' })).toHaveAccessibleDescription(/Use 2 to 32 characters/)
  await profile.getByRole('textbox', { name: 'Display name' }).fill('W2 Reader')
  await profile.getByRole('button', { name: 'Save display name' }).click()
  await expect(profile.getByRole('status')).toHaveText('Display name saved.')
  await expect(settingValue(page, 'Display name')).toHaveText('W2 Reader')
  await expect(page.getByRole('banner')).toContainText('W2 Reader')
  await page.reload()
  await waitForHydration(page)
  await expect(settingValue(page, 'Display name')).toHaveText('W2 Reader')

  const email = settingRow(page, 'Email')
  await openSetting(page, 'Email')
  await email.getByRole('textbox', { name: 'New email' }).fill(adminEmail)
  await email.getByRole('button', { name: 'Save email' }).click()
  await expect(email.getByRole('textbox', { name: 'New email' })).toHaveAccessibleDescription(/already registered/)
  const changed = `w2-reader-${randomUUID()}@example.test`
  await email.getByRole('textbox', { name: 'New email' }).fill(changed)
  await email.getByRole('button', { name: 'Save email' }).click()
  await expect(email.getByRole('status')).toContainText('Email changed.')
  await expect(settingValue(page, 'Email')).toHaveText(changed)
  await page.reload()
  await waitForHydration(page)
  await expect(settingValue(page, 'Email')).toHaveText(changed)
  await page.screenshot({ path: testInfo.outputPath('settings-desktop.png'), fullPage: true })

  const admin = await newPage(browser, testInfo)
  await logIn(admin, adminEmail, adminPassword())
  await expectSettings(admin)
  await expect.poll(async () => {
    const list = await admin.request.get(`${main}/api/internal/emails?recipient=${encodeURIComponent(readerEmail)}&template=account.email_changed&status=held`)
    expect(list.status()).toBe(200)
    const body = await list.json() as { emails: { params: Record<string, string> }[] }
    return body.emails.some(message => message.params.new_email === changed)
  }, { timeout: 30_000 }).toBe(true)
  await admin.context().close()

  await openSetting(page, 'Email')
  await email.getByRole('textbox', { name: 'New email' }).fill(readerEmail)
  await email.getByRole('button', { name: 'Save email' }).click()
  await expect(email.getByRole('status')).toContainText('Email changed.')
  await expect(settingValue(page, 'Email')).toHaveText(readerEmail)
  await page.setViewportSize({ width: 390, height: 844 })
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('settings-mobile.png'), fullPage: true })
  await page.context().close()
})

test('password problems are explained and foreign-origin posts are refused', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo)
  await logIn(page, readerEmail, readerPassword())
  await expectSettings(page)
  await waitForHydration(page)

  const password = settingRow(page, 'Password')
  await openSetting(page, 'Password')
  await password.getByLabel('New password').fill(readerEmail)
  await password.getByRole('button', { name: 'Save password' }).click()
  await expect(password.getByLabel('New password')).toHaveAccessibleDescription(/too similar to your email/)

  const foreign = await page.request.post('/actions/password', {
    headers: { Origin: 'https://evil.example' },
    form: { password: `Foreign-${randomUUID()}` },
    maxRedirects: 0,
  })
  expect(foreign.status()).toBe(403)
  await page.context().close()
})

test('logging out ends the session', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo)
  await logIn(page, readerEmail, readerPassword())
  await expectSettings(page)
  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await page.goto('/')
  await expect(page).toHaveURL(/\/login$/)
  await page.context().close()
})

// Run with AUTHV2_STALE_SESSION_DIR set, then age the Reader's sessions once `ready` appears there and create `aged`.
test('stale sessions keep display names editable and log in again for email changes', async ({ browser }, testInfo) => {
  const directory = process.env.AUTHV2_STALE_SESSION_DIR
  test.skip(!directory, 'needs an operator to age the branch Reader session')
  const page = await newPage(browser, testInfo)
  await logIn(page, readerEmail, readerPassword())
  await expectSettings(page)
  writeFileSync(`${directory}/ready`, '')
  await expect.poll(() => existsSync(`${directory}/aged`), { timeout: 600_000, intervals: [2_000] }).toBe(true)

  await page.reload()
  await waitForHydration(page)
  const profile = settingRow(page, 'Display name')
  await openSetting(page, 'Display name')
  await profile.getByRole('textbox', { name: 'Display name' }).fill('Stale Reader')
  await profile.getByRole('button', { name: 'Save display name' }).click()
  await expect(profile.getByRole('status')).toHaveText('Display name saved.')

  const email = settingRow(page, 'Email')
  await openSetting(page, 'Email')
  await email.getByRole('textbox', { name: 'New email' }).fill(`w2-stale-${randomUUID()}@example.test`)
  await email.getByRole('button', { name: 'Save email' }).click()
  await expect(email.getByRole('alert')).toContainText('log in again')
  await page.screenshot({ path: testInfo.outputPath('settings-stale-session.png'), fullPage: true })

  await email.getByRole('link', { name: 'Log in again' }).click()
  await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await page.getByLabel('Email').fill(readerEmail)
  await page.getByLabel('Password').fill(readerPassword())
  await page.getByRole('button', { name: 'Log in', exact: true }).click()
  await expectSettings(page)
  await waitForHydration(page)
  expect(new URL(page.url()).searchParams.get('edit')).toBe('email')
  await expect(email.getByRole('textbox', { name: 'New email' })).toBeFocused()
  await openSetting(page, 'Display name')
  await profile.getByRole('textbox', { name: 'Display name' }).fill('Dev Reader')
  await profile.getByRole('button', { name: 'Save display name' }).click()
  await expect(profile.getByRole('status')).toHaveText('Display name saved.')
  await page.context().close()
})
