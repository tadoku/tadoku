import { randomUUID } from 'node:crypto'
import { expect, test, Page } from '@playwright/test'

// Use only a disposable development environment: this journey creates an account.
const authUrl = process.env.AUTH_URL
const appUrl = process.env.APP_URL
const mailUrl = process.env.MAILHOG_URL

test('registration, shared session, settings, reauthentication, logout and recovery', async ({
  page,
  request,
}) => {
  test.skip(
    !authUrl || !appUrl || !mailUrl,
    'Set AUTH_URL, APP_URL and MAILHOG_URL for the development GitOps base',
  )
  const email = `kratos-e2e-${randomUUID()}@example.com`
  const password = `Start-${randomUUID()}!`
  const newPassword = `Updated-${randomUUID()}!`
  const recoveredPassword = `Recovered-${randomUUID()}!`
  const pageErrors: string[] = []
  page.on('pageerror', error => pageErrors.push(error.message))

  const submit = (target: Page) =>
    target.locator('button[type="submit"]').click()
  const logout = async () => {
    await page.getByRole('button', { name: /Kratos E2E Updated/ }).click()
    await page.getByRole('menuitem', { name: 'Log out' }).click()
    await expect(
      page.getByRole('heading', { name: 'Log in', exact: true }),
    ).toBeVisible()
    expect(
      (await page.request.get(`${authUrl}/kratos/sessions/whoami`)).status(),
    ).toBe(401)
  }
  const login = async (secret: string) => {
    await page.goto(`${authUrl}/login`)
    await page.locator('input[name="identifier"]').fill(email)
    await page.locator('input[name="password"]').fill(secret)
    await submit(page)
    await expect(
      page.getByRole('heading', { name: 'Settings', exact: true }),
    ).toBeVisible()
  }

  await test.step('register and retain the session after a reload', async () => {
    await page.goto('/register')
    await page.locator('input[name="traits.email"]').fill(email)
    await page.locator('input[name="traits.display_name"]').fill('Kratos E2E')
    await page.locator('input[name="password"]').fill(password)
    await submit(page)
    await expect(
      page.getByRole('heading', { name: 'Settings', exact: true }),
    ).toBeVisible()
    await page.reload()
    await expect(page.locator('input[name="traits.email"]')).toHaveValue(email)
  })

  await test.step('save profile and password using the refreshed settings flow', async () => {
    await page
      .locator('input[name="traits.display_name"]')
      .fill('Kratos E2E Updated')
    await page
      .locator('form')
      .filter({ has: page.locator('input[name="traits.display_name"]') })
      .getByRole('button', { name: 'Save', exact: true })
      .click()
    await expect(page.getByText('Your changes have been saved!')).toBeVisible()
    await page.locator('input[name="password"]').fill(newPassword)
    const updated = page.waitForResponse(
      r =>
        r.url().includes('/self-service/settings?') &&
        r.request().method() === 'POST',
    )
    await page
      .locator('form')
      .filter({ has: page.locator('input[name="password"]') })
      .getByRole('button', { name: 'Save', exact: true })
      .click()
    expect((await updated).status()).toBe(200)
  })

  await test.step('share the session with the app, including server rendering', async () => {
    await page.goto(appUrl!)
    await expect(
      page.getByRole('button', { name: /Kratos E2E Updated/ }),
    ).toBeVisible()
    await page.reload()
    await expect(
      page.getByRole('button', { name: /Kratos E2E Updated/ }),
    ).toBeVisible()
    await logout()
  })

  await test.step('reject the old password and accept the new password on the same flow', async () => {
    await page.locator('input[name="identifier"]').fill(email)
    await page.locator('input[name="password"]').fill(password)
    await submit(page)
    await expect(
      page.getByText('The provided credentials are invalid'),
    ).toBeVisible()
    await page.locator('input[name="password"]').fill(newPassword)
    await submit(page)
    await expect(
      page.getByRole('heading', { name: 'Settings', exact: true }),
    ).toBeVisible()
  })

  await test.step('reauthenticate an existing session', async () => {
    const response = await page.request.get(
      `${authUrl}/kratos/self-service/login/browser`,
      {
        params: { refresh: 'true' },
        headers: { Accept: 'application/json' },
      },
    )
    expect(response.status()).toBe(200)
    const flow = await response.json()
    await page.goto(`${authUrl}/login?flow=${flow.id}`)
    await expect(
      page.getByRole('heading', { name: 'Confirm action' }),
    ).toBeVisible()
    await page.locator('input[name="password"]').fill(newPassword)
    await submit(page)
    await expect(
      page.getByRole('heading', { name: 'Settings', exact: true }),
    ).toBeVisible()
    await logout()
  })

  await test.step('recover by email and log in with the recovered password', async () => {
    await page.goto('/account-recovery')
    await page.locator('input[name="email"]').fill(email)
    await submit(page)
    await expect(page.getByText(/recovery link has been sent/i)).toBeVisible()
    let recoveryLink = ''
    await expect
      .poll(async () => {
        const response = await request.get(`${mailUrl}/api/v2/search`, {
          params: { kind: 'to', query: email },
        })
        const messages = await response.json()
        for (const message of messages.items ?? []) {
          const body =
            message.MIME?.Parts?.map((part: any) => part.Body).join('\n') ??
            message.Content.Body
          // MailHog exposes quoted-printable HTML/text for Kratos messages.
          const decoded = body
            .replace(/=\r?\n/g, '')
            .replace(/=([0-9A-F]{2})/gi, (_: string, hex: string) =>
              String.fromCharCode(parseInt(hex, 16)),
            )
          recoveryLink =
            decoded
              .match(
                /https?:\/\/[^\s<>"']+\/self-service\/recovery\?[^\s<>"']+/,
              )?.[0]
              ?.replace(/&amp;/g, '&') ?? ''
          if (recoveryLink) break
        }
        return recoveryLink
      })
      .not.toBe('')
    await page.goto(recoveryLink)
    await expect(
      page.getByRole('heading', { name: 'Settings', exact: true }),
    ).toBeVisible()
    await page.locator('input[name="password"]').fill(recoveredPassword)
    const updated = page.waitForResponse(
      r =>
        r.url().includes('/self-service/settings?') &&
        r.request().method() === 'POST',
    )
    await page
      .locator('form')
      .filter({ has: page.locator('input[name="password"]') })
      .getByRole('button', { name: 'Save', exact: true })
      .click()
    expect((await updated).status()).toBe(200)
    await logout()
    await login(recoveredPassword)
  })
  expect(pageErrors).toEqual([])
})

test('existing administrator session works across auth and admin', async ({
  page,
}) => {
  const adminUrl = process.env.ADMIN_URL
  const email = process.env.E2E_ADMIN_EMAIL
  const password = process.env.E2E_ADMIN_PASSWORD
  test.skip(
    !authUrl || !appUrl || !adminUrl || !email || !password,
    'Set ADMIN_URL and E2E_ADMIN_EMAIL/E2E_ADMIN_PASSWORD for an existing dev admin',
  )

  await page.goto(`${authUrl}/login?return_to=${encodeURIComponent(adminUrl!)}`)
  await page.locator('input[name="identifier"]').fill(email!)
  await page.locator('input[name="password"]').fill(password!)
  await page.locator('button[type="submit"]').click()
  await expect(
    page.getByRole('heading', { name: 'Dashboard', exact: true }),
  ).toBeVisible()
  await page.reload()
  await expect(
    page.getByRole('heading', { name: 'Dashboard', exact: true }),
  ).toBeVisible()
  const role = await page.request.get(
    `${appUrl}/api/internal/authz/current-user/role`,
  )
  expect(role.status()).toBe(200)
  expect((await role.json()).role).toBe('admin')
  await page.goto(authUrl!)
  await page.getByRole('button', { name: /Open navigation menu/ }).click()
  await page.getByRole('menuitem', { name: 'Log out' }).click()
  await expect(
    page.getByRole('heading', { name: 'Log in', exact: true }),
  ).toBeVisible()
  expect(
    (await page.request.get(`${authUrl}/kratos/sessions/whoami`)).status(),
  ).toBe(401)
})
