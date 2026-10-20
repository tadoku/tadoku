import { test, expect } from '@playwright/test'
import { writeFileSync } from 'node:fs'

const stateFile = process.env.IDENTITY_UI_STATE!
const origins = {
  webv2: process.env.IDENTITY_WEB_ORIGIN!,
  admin: process.env.IDENTITY_ADMIN_ORIGIN!,
  auth: process.env.IDENTITY_AUTH_ORIGIN!,
}
const setState = (state: object) => writeFileSync(stateFile, JSON.stringify(state))

for (const status of [403, 501, 503, 'timeout']) {
  test(`session ${status} renders bounded recovery in each application`, async ({ page }) => {
    setState({ status })
    for (const [app, origin] of Object.entries(origins)) {
      const response = await page.goto(`${origin}/${app === 'auth' ? 'login' : ''}`)
      expect(response?.status()).toBe(200)
      await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible()
      await expect(page.getByText('Session temporarily unavailable')).toBeVisible()
      await expect(page.getByText('Unhandled Runtime Error')).toHaveCount(0)
      if (app === 'auth') await expect(page.getByRole('textbox')).toHaveCount(0)
      if (app === 'admin') await expect(page.getByRole('link', { name: 'Users', exact: true })).toHaveCount(0)
      await page.screenshot({ animations: 'disabled', path: `${process.env.IDENTITY_UI_EVIDENCE}/${app}-${status}.png` })
    }
  })
}

test('persistent unauthorized responses stop and explicit retry recovers', async ({ page }) => {
  for (const [app, origin] of Object.entries(origins).filter(([app]) => app !== 'auth')) {
    setState({ status: 200, signedIn: true, role: 'admin', queryStatus: 401 })
    let navigations = 0
    page.on('framenavigated', frame => { if (frame === page.mainFrame()) navigations++ })
    await page.goto(`${origin}/${app === 'admin' ? 'users' : 'logs/new'}`)
    await expect(page.getByText('Session temporarily unavailable')).toBeVisible()
    const stoppedAt = navigations
    await page.waitForTimeout(1500)
    expect(navigations).toBe(stoppedAt)
    setState({ status: 200, signedIn: false })
    await page.getByRole('button', { name: 'Try again' }).click()
    await expect(page.getByText('Session temporarily unavailable')).toHaveCount(0)
  }
})

test('reader cannot render admin controls', async ({ page }) => {
  setState({ status: 200, signedIn: true, role: 'user' })
  await page.goto(`${origins.admin}/users`)
  await expect(page.getByRole('heading', { name: 'Access Denied' })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Users', exact: true })).toHaveCount(0)
  await page.screenshot({ animations: 'disabled', path: `${process.env.IDENTITY_UI_EVIDENCE}/admin-reader-denied.png` })
})

test('banned screen keeps logout failure visible without losing session', async ({ page }) => {
  setState({ status: 200, signedIn: true, role: 'banned', logoutStatus: 503 })
  await page.goto(origins.webv2)
  await expect(page.getByRole('heading', { name: 'Account Suspended' })).toBeVisible()
  await page.getByRole('button', { name: 'Log out' }).click()
  await expect(page.getByText('Log out failed. Please try again.')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Account Suspended' })).toBeVisible()
  await page.screenshot({ animations: 'disabled', path: `${process.env.IDENTITY_UI_EVIDENCE}/banned-logout-failure.png` })
})

test('account links follow tenant mode and auth preserves deep-link redirect', async ({ page, request }) => {
  for (const app of ['v1', 'v2']) {
    setState({ status: 200, signedIn: true, role: 'admin', app })
    await page.goto(origins.webv2)
    await page.getByRole('button', { name: /Open navigation menu/ }).click()
    const settings = page.getByRole('menuitem', { name: 'Account settings' })
    await expect(settings).toHaveAttribute('href', new RegExp(app === 'v1' ? 'account\\.tadoku\\.app' : 'authv2\\.tadoku\\.app'))
  }
  const response = await request.get(`${origins.auth}/login?return_to=${encodeURIComponent('https://proof-0123abcd.tadoku.dev.lab/logs/new?amount=3#entry')}`, { maxRedirects: 0 })
  expect(response.status()).toBe(307)
  const target = new URL(response.headers().location)
  expect(target.hostname).toBe('authv2.tadoku.app')
  expect(target.pathname).toBe('/login')
  expect(target.searchParams.get('return_to')).toBe('https://proof-0123abcd.tadoku.dev.lab/logs/new?amount=3#entry')
})
