import { randomUUID } from 'node:crypto'
import { test, expect } from '@playwright/test'
import {
  adminEmail, adminPassword, expectSettings, logIn, newPage, origins, readerEmail, readerPassword, saveReaderPassword,
} from './support'

// Failure modes: recovery answers that reveal whether an account exists; a held email whose link differs from its
// body or leaves the tenant; a recovery link that does not land signed in on Settings, or continues before the new
// password is saved; old passwords, sessions or links still working afterwards.

test('recovery through a held email sets a new password and continues', async ({ browser }, testInfo) => {
  const { authv2, main } = origins(testInfo)
  const initial = readerPassword()
  const destination = `${main}/pages/dev-welcome?w2=recovered#main`

  const earlier = await newPage(browser, testInfo)
  await logIn(earlier, readerEmail, initial)
  await expectSettings(earlier)
  const admin = await newPage(browser, testInfo)
  await logIn(admin, adminEmail, adminPassword())
  await expectSettings(admin)

  const held = async () => {
    const list = await admin.request.get(`${main}/api/internal/emails?recipient=${encodeURIComponent(readerEmail)}&template=account.recovery&status=held`)
    expect(list.status()).toBe(200)
    return ((await list.json()) as { emails: { id: string }[] }).emails.map(message => message.id)
  }
  const earlierEmails = new Set(await held())

  const page = await newPage(browser, testInfo, { viewport: { width: 1280, height: 900 } })
  const confirmations: string[] = []
  for (const email of [`w2-unknown-${randomUUID()}@example.test`, readerEmail]) {
    await page.goto(`/recovery?return_to=${encodeURIComponent(destination)}`)
    await page.getByLabel('Email').fill(email)
    await page.getByRole('button', { name: 'Send recovery link' }).click()
    await expect(page.getByRole('heading', { level: 1, name: 'Check your email' })).toBeVisible()
    confirmations.push(await page.locator('main').innerText())
  }
  expect(confirmations[0]).toBe(confirmations[1])
  await page.screenshot({ path: testInfo.outputPath('recovery-sent.png'), fullPage: true })

  let id: string | undefined
  await expect.poll(async () => {
    id = (await held()).find(candidate => !earlierEmails.has(candidate))
    return id
  }, { timeout: 30_000 }).toBeTruthy()
  const detail = await admin.request.get(`${main}/api/internal/emails/${id}`)
  expect(detail.status()).toBe(200)
  const message = await detail.json() as { status: string; body: string; params: { recovery_url: string } }
  expect(message.status).toBe('held')
  const link = message.params.recovery_url
  expect(message.body).toContain(link)
  expect(new URL(link).origin).toBe(main)

  await page.goto(link)
  await expectSettings(page)
  expect(new URL(page.url()).origin).toBe(authv2)
  expect(new URL(page.url()).searchParams.get('return_to')).toBe(destination)
  await expect(page.getByRole('banner')).toContainText('Log out')
  await page.screenshot({ path: testInfo.outputPath('recovered-settings.png'), fullPage: true })

  const replacement = `Recovered-${randomUUID()}`
  const password = page.getByRole('region', { name: 'Password' })
  await password.getByLabel('New password').fill(replacement)
  await password.getByRole('button', { name: 'Change password and continue' }).click()
  await expect(page).toHaveURL(destination)
  saveReaderPassword(replacement)

  await earlier.goto('/')
  await expect(earlier.getByRole('heading', { level: 1, name: 'Log in' })).toBeVisible()
  await page.goto(authv2)
  await page.getByRole('button', { name: 'Log out', exact: true }).click()
  await logIn(page, readerEmail, initial)
  await expect(page.getByText('That email and password don’t match.')).toBeVisible()
  await logIn(page, readerEmail, replacement)
  await expectSettings(page)

  const reused = await admin.request.get(link, { maxRedirects: 0 })
  expect(reused.status()).toBe(400)
  expect((await reused.json() as { code: string }).code).toBe('invalid_recovery_token')
  for (const context of [earlier.context(), admin.context(), page.context()]) await context.close()
})
