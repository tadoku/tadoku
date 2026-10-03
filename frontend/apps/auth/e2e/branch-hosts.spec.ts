import { expect, test } from '@playwright/test'

const app = process.env.BRANCH_APP_URL
const auth = process.env.BRANCH_AUTH_URL
const admin = process.env.BRANCH_ADMIN_URL
const email = process.env.E2E_ADMIN_EMAIL
const password = process.env.E2E_ADMIN_PASSWORD

test('branch hosts retain navigation and authenticated API selection', async ({ page, request }) => {
  test.skip(!app || !auth || !admin || !email || !password, 'Set branch URLs and development administrator credentials')
  const route = new URL(app!).hostname.split('.')[0]
  const html = await request.get(app!)
  expect(html.status()).toBe(200)
  expect(html.headers()['x-dev-selected']).toBe(route)
  expect(await html.text()).toContain(`${auth}/register?return_to=`)
  expect(await html.text()).not.toContain('https://account.tadoku.dev.lab/register')

  const requests: string[] = []
  page.on('request', r => requests.push(r.url()))
  await page.goto(app!)
  const login = page.getByRole('link', { name: 'Log in', exact: true })
  await expect(login).toHaveAttribute('href', `${auth}/login?return_to=${app}/`)
  await login.click()
  await page.locator('input[name="identifier"]').fill(email!)
  await page.locator('input[name="password"]').fill(password!)
  await page.locator('button[type="submit"]').click()
  await expect(page).toHaveURL(`${app}/`)
  await expect(page.getByRole('button', { name: /Open navigation menu/ })).toBeVisible()

  const roleResponse = page.waitForResponse(r => r.url() === `${app}/api/internal/authz/current-user/role`)
  await page.goto(admin!)
  const role = await roleResponse
  expect(role.status()).toBe(200)
  expect(role.headers()['x-dev-selected']).toBe(route)
  expect((await role.json()).role).toBe('admin')
  await expect(page.getByRole('heading', { name: 'Dashboard', exact: true })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Back to Tadoku', exact: false })).toHaveAttribute('href', app!)
  expect(requests.filter(url => url.startsWith('https://tadoku.dev.lab/api/internal') || url.startsWith('https://admin.tadoku.dev.lab'))).toEqual([])

  await page.goto(app!)
  await page.getByRole('button', { name: /Open navigation menu/ }).click()
  await page.getByRole('menuitem', { name: 'Log out' }).click()
  await expect(page).toHaveURL(new RegExp(`^${auth}/login`))
  expect((await page.request.get('https://account.tadoku.dev.lab/kratos/sessions/whoami')).status()).toBe(401)
})
