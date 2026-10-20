import { randomUUID } from 'node:crypto'
import { test, expect } from '@playwright/test'
import {
  expectContained, expectSettings, firstWrite, holdScripts, logIn, newPage, origins, readerEmail, readerPassword,
  waitForHydration,
} from './support'

// Failure modes: server HTML without usable controls; native submissions that cannot create a session; values typed
// before hydration lost or sent elsewhere; enhanced requests proxied through the authv2 server (which collapses every
// visitor into one login-throttle address); return_to ignored after login or followed off the allowlist; wrong
// passwords, throttling, taken emails and invalid fields without a usable message; overflowing mobile layouts.

test('login and registration work without JavaScript', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo, { javaScriptEnabled: false })
  const email = `w2-native-${randomUUID()}@example.test`
  const password = `Native-${randomUUID()}`

  await page.goto('/register')
  await expect(page.getByRole('heading', { level: 1, name: 'Create an account' })).toBeVisible()
  await page.getByLabel('Display name').fill('Native Reader')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expectSettings(page)
  await expect(page.getByRole('textbox', { name: 'Email', exact: true })).toHaveValue(email)
  await page.reload()
  await expect(page.getByLabel('Display name')).toHaveValue('Native Reader')

  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await logIn(page, email, 'not-the-password')
  await expect(page.getByText('That email and password don’t match.')).toBeVisible()
  await logIn(page, email, password)
  await expectSettings(page)
  const emailSection = page.getByRole('region', { name: 'Email address' })
  await emailSection.getByRole('textbox', { name: 'Email' }).fill(readerEmail)
  await emailSection.getByRole('button', { name: 'Change email' }).click()
  await expect(emailSection.getByRole('alert')).toContainText('already registered')
  await expect(page.getByRole('region', { name: 'Profile' }).getByRole('alert')).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('settings-without-javascript.png'), fullPage: true })
  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await page.context().close()
})

test('credentials typed before hydration are kept and sent straight to the identity API', async ({ browser }, testInfo) => {
  const { main } = origins(testInfo)
  const page = await newPage(browser, testInfo)
  const email = `w2-early-${randomUUID()}@example.test`
  const password = `Early-${randomUUID()}`

  for (const [path, button, endpoint] of [
    ['/register', 'Create account', '/api/internal/identity/registration'],
    ['/login', 'Log in', '/api/internal/identity/login'],
  ] as const) {
    const scripts = await holdScripts(page)
    await page.goto(path, { waitUntil: 'domcontentloaded' })
    if (path === '/register') await page.getByLabel('Display name').fill('Early Reader')
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Password').fill(password)
    expect(scripts.held()).toBeGreaterThan(0)
    await scripts.release()
    await expect(page.getByLabel('Email')).toHaveValue(email)
    expect(await page.getByLabel('Password').inputValue() === password).toBe(true)

    const request = await firstWrite(page, () => page.getByRole('button', { name: button, exact: true }).click())
    expect(request.url()).toBe(main + endpoint)
    expect(request.headers()['content-type']).toBe('application/json')
    await expectSettings(page)
    await expect(page.getByRole('textbox', { name: 'Email', exact: true })).toHaveValue(email)
    await page.getByRole('button', { name: 'Log out', exact: true }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  }
  await page.context().close()
})

test('login continues to an allowed return_to and never to a foreign one', async ({ browser }, testInfo) => {
  const { main } = origins(testInfo)
  const page = await newPage(browser, testInfo)
  const destination = `${main}/pages/dev-welcome?w2=login#main`

  await logIn(page, readerEmail, readerPassword(), `/login?return_to=${encodeURIComponent(destination)}`)
  await expect(page).toHaveURL(destination)

  await page.goto(`/login?return_to=${encodeURIComponent('https://evil.example/steal')}`)
  await expect(page.locator('a[href*="evil.example"], input[value*="evil.example"]')).toHaveCount(0)
  await logIn(page, readerEmail, readerPassword(), `/login?return_to=${encodeURIComponent('https://evil.example/steal')}`)
  await expectSettings(page)
  expect(new URL(page.url()).host).not.toContain('evil.example')
  await page.context().close()
})

test('login explains wrong passwords and throttling', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo)
  const unknown = `w2-throttle-${randomUUID()}@example.test`
  await page.goto('/login')
  await waitForHydration(page)
  for (let attempt = 1; attempt <= 11; attempt++) {
    await page.getByLabel('Email').fill(unknown)
    await page.getByLabel('Password').fill('a-wrong-password')
    await page.getByRole('button', { name: 'Log in', exact: true }).click()
    const message = attempt <= 10 ? 'That email and password don’t match.' : 'Too many login attempts.'
    await expect(page.getByRole('alert').getByText(message)).toBeVisible()
    if (attempt === 1) {
      await expect(page.getByLabel('Password')).toHaveValue('')
      await expect(page.getByLabel('Password')).toBeFocused()
    }
  }
  await page.screenshot({ path: testInfo.outputPath('login-throttled.png'), fullPage: true })
  await page.context().close()
})

test('registration maps field problems to the field that caused them', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo)
  await page.goto('/register')
  await waitForHydration(page)

  await page.getByLabel('Display name').fill('x')
  await page.getByLabel('Email').fill(`w2-fields-${randomUUID()}@example.test`)
  await page.getByLabel('Password').fill('short')
  const writes: string[] = []
  page.on('request', request => { if (request.method() !== 'GET') writes.push(request.url()) })
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expect(page.getByLabel('Display name')).toHaveAccessibleDescription(/Use 2 to 32 characters/)
  await expect(page.getByLabel('Password')).toHaveAccessibleDescription(/Use at least 8 characters/)
  await expect(page.getByLabel('Display name')).toBeFocused()
  expect(writes).toEqual([])

  await page.getByLabel('Display name').fill('Field Reader')
  await page.getByLabel('Email').fill(readerEmail)
  await page.getByLabel('Password').fill(`Unrelated-${randomUUID()}`)
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expect(page.getByLabel('Email')).toHaveAccessibleDescription(/already registered/)
  await expect(page.getByLabel('Email')).toBeFocused()

  const email = `w2-similar-${randomUUID()}@example.test`
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(email)
  await page.getByRole('button', { name: 'Create account', exact: true }).click()
  await expect(page.getByLabel('Password')).toHaveAccessibleDescription(/too similar to your email/)
  await page.screenshot({ path: testInfo.outputPath('registration-field-errors.png'), fullPage: true })
  await page.context().close()
})

test('account screens fit a 390px viewport', async ({ browser }, testInfo) => {
  const page = await newPage(browser, testInfo, { viewport: { width: 390, height: 844 } })
  for (const [path, name] of [['/login', 'login'], ['/register', 'register'], ['/recovery', 'recovery']] as const) {
    await page.goto(path)
    await expectContained(page)
    await page.screenshot({ path: testInfo.outputPath(`${name}-mobile.png`), fullPage: true })
  }
  await page.context().close()
})
