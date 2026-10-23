import { createServer } from 'node:http'
import { test, expect, type Page, type Route } from '@playwright/test'
import { openSetting, settingRow, settingValue, waitForHydration } from './support'

// Settings for imported addresses the API sends exactly as stored because net/mail can't read them. Email inputs
// drop line breaks and surrounding ASCII whitespace from such an address, so the input can't hold it unchanged.
// Native submissions only accept an https account origin, so the server runs like the registration fixture's:
// TADOKU_WEB_URL=https://tadoku.dev.lab and TADOKU_API_URL pointing at the fixture below. Browser calls to that web
// host's API are answered in the test and never leave it.
const enabled = process.env.AUTHV2_SETTINGS_FIXTURE === 'true'
const host = 'account.tadoku.dev.lab'
const unchanged = 'This is already your email address. Enter a different one to change it.'

// Byte for byte what the API's generated session writer emits, with the email exactly as encoding/json escapes it.
const emitted = (email: string) => '{"account":{"authenticated_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z",'
  + `"display_name":"Reader","email":${email},"id":"00000000-0000-4000-8000-000000000001"},"account_app":"v2"}\n`

const accounts = [
  { name: 'ordinary address', body: emitted('"reader@example.test"'), email: 'reader@example.test', shown: 'reader@example.test' },
  { name: 'leading vertical tab', body: emitted('"\\u000breader@example.test"'), email: '\vreader@example.test', shown: '\vreader@example.test' },
  { name: 'leading form feed', body: emitted('"\\freader@example.test"'), email: '\freader@example.test', shown: 'reader@example.test' },
  { name: 'leading CRLF', body: emitted('"\\r\\nreader@example.test"'), email: '\r\nreader@example.test', shown: 'reader@example.test' },
  { name: 'trailing LF', body: emitted('"reader@example.test\\n"'), email: 'reader@example.test\n', shown: 'reader@example.test' },
]

let session = accounts[0].body
type Write = { method?: string; path?: string; origin?: string; cookie?: string; branch?: string; type?: string; body: unknown }
const writes: Write[] = []

const fixture = createServer((request, response) => {
  if (request.url === '/api/internal/identity/session') {
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(session)
    return
  }
  if (!['/api/internal/identity/settings/profile', '/api/internal/identity/settings/password'].includes(request.url ?? '')) {
    response.writeHead(404).end()
    return
  }
  const chunks: Buffer[] = []
  request.on('data', chunk => chunks.push(chunk))
  request.on('end', () => {
    const { origin, cookie, 'x-dev-branch': branch, 'content-type': type } = request.headers
    const body = JSON.parse(Buffer.concat(chunks).toString('utf8'))
    writes.push({ method: request.method, path: request.url, origin, cookie, branch: branch?.toString(), type, body })
    if (request.method === 'PUT') return response.writeHead(204).end()
    const account = (JSON.parse(session) as { account: object }).account
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(JSON.stringify({ ...account, ...body }))
  })
})

async function answerBrowser(route: Route) {
  const request = route.request()
  const cors = { 'Access-Control-Allow-Origin': await request.headerValue('origin') ?? '*', 'Access-Control-Allow-Credentials': 'true' }
  if (request.method() === 'OPTIONS') {
    return route.fulfill({ status: 204, headers: { ...cors, 'Access-Control-Allow-Methods': 'PATCH, PUT', 'Access-Control-Allow-Headers': 'content-type' } })
  }
  const path = new URL(request.url()).pathname
  writes.push({ method: request.method(), path, type: await request.headerValue('content-type') ?? undefined, body: request.postDataJSON() })
  if (request.method() === 'PUT') return route.fulfill({ status: 204, headers: cors })
  const account = (JSON.parse(session) as { account: object }).account
  return route.fulfill({ status: 200, contentType: 'application/json', headers: cors, body: JSON.stringify({ ...account, ...request.postDataJSON() }) })
}

test.beforeAll(async () => {
  if (enabled) await new Promise<void>(resolve => fixture.listen(5195, '127.0.0.1', resolve))
})

test.afterAll(async () => {
  if (enabled) await new Promise<void>((resolve, reject) => fixture.close(error => error ? reject(error) : resolve()))
})

test.beforeEach(() => {
  test.skip(!enabled, 'needs the local settings fixture')
})

test.use({ actionTimeout: 10_000 })

const destination = 'https://tadoku.dev.lab/pages/dev-welcome?w2=settings'

/** A hydrated save lands where a native one redirects: the row collapsed, showing its confirmation. */
async function expectSaved(page: Page, name: string, message: string, search: Record<string, string>) {
  await expect.poll(() => {
    const url = new URL(page.url())
    return { pathname: url.pathname, search: Object.fromEntries(url.searchParams) }
  }).toEqual({ pathname: '/settings', search })
  await expect.soft(settingRow(page, name).getByRole('status')).toHaveText(message)
  await expect.soft(page.locator('main input:not([type="hidden"])')).toHaveCount(0)
}

/** The row's text exactly as the DOM holds it; Playwright's text matchers would normalize its whitespace. */
async function storedValue(page: Page, name: string) {
  await expect(settingValue(page, name)).toHaveCount(1)
  return settingValue(page, name).evaluate(value => value.textContent)
}

test('hydrated Settings shows stored addresses and only sends the fields a visitor changed', async ({ page, request }, testInfo) => {
  await page.route('https://tadoku.dev.lab/api/internal/identity/settings/*', answerBrowser)
  const errors: string[] = []
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()) })

  // The router's dehydrated match IDs contain NUL separators, so count NULs against an ASCII address's page.
  const nuls = (html: string) => html.split('\u0000').length - 1
  session = accounts[0].body
  const routerNuls = nuls(await (await request.get('/settings?edit=email')).text())

  for (const account of accounts) {
    await test.step(account.name, async () => {
      session = account.body
      const html = await (await request.get('/settings?edit=email')).text()
      expect.soft(nuls(html), 'server HTML gains no NUL').toBe(routerNuls)
      expect.soft(html.includes('�'), 'server HTML has no replacement character').toBe(false)
      expect.soft(html.includes(`value="${account.email}"`), 'server-rendered email input').toBe(true)

      const load = async (path: string) => {
        await page.goto(path)
        await waitForHydration(page)
        writes.length = 0
      }
      await load('/settings')
      expect.soft(await storedValue(page, 'Email'), 'the hydrated row shows the stored address').toBe(account.email)

      await load('/settings?edit=email')
      const email = settingRow(page, 'Email').getByRole('textbox', { name: 'New email', exact: true })
      await expect.soft(email).toHaveValue(account.shown)
      await expect.soft(email).toBeFocused()
      await page.screenshot({ path: testInfo.outputPath(`hydrated-${account.name.replaceAll(' ', '-')}.png`), fullPage: true })

      await settingRow(page, 'Email').getByRole('button', { name: 'Save email' }).click()
      await expect.soft(page.locator('#email-email-error')).toHaveText(unchanged)
      await expect.soft(email).toHaveValue(account.shown)
      expect.soft(writes, 'an unchanged address sends nothing').toEqual([])

      await load(`/settings?edit=display-name&return_to=${encodeURIComponent(destination)}`)
      await settingRow(page, 'Display name').getByRole('textbox', { name: 'Display name' }).fill('Renamed Reader')
      await settingRow(page, 'Display name').getByRole('button', { name: 'Save display name' }).click()
      await expectSaved(page, 'Display name', 'Display name saved.', { return_to: destination, saved: 'display-name' })
      expect.soft(writes, 'a display name save sends only the name').toEqual([
        { method: 'PATCH', path: '/api/internal/identity/settings/profile', type: 'application/json', body: { display_name: 'Renamed Reader' } },
      ])

      await load('/settings?edit=password')
      await settingRow(page, 'Password').getByLabel('New password').fill('a-long-reading-passphrase')
      await settingRow(page, 'Password').getByRole('button', { name: 'Save password' }).click()
      await expectSaved(page, 'Password', 'Password changed. Your other devices have been signed out.', { saved: 'password' })
      expect.soft(writes, 'a password save sends only the password').toEqual([
        { method: 'PUT', path: '/api/internal/identity/settings/password', type: 'application/json', body: { password: 'a-long-reading-passphrase' } },
      ])

      await load('/settings?edit=email')
      await email.fill('edited@example.test')
      await settingRow(page, 'Email').getByRole('button', { name: 'Save email' }).click()
      await expectSaved(page, 'Email', 'Email changed. We’ve let your previous address know.', { saved: 'email' })
      expect.soft(writes, 'an edited address is sent as typed').toEqual([
        { method: 'PATCH', path: '/api/internal/identity/settings/profile', type: 'application/json', body: { email: 'edited@example.test' } },
      ])
    })
  }
  expect.soft(errors, 'no hydration or runtime errors').toEqual([])
})

test('native Settings forms send only what the visitor changed', async ({ browser, request }, testInfo) => {
  // The browser posts each native form as it would without JavaScript. The test captures that exact body, then
  // replays it as the account host so the server's Origin check applies.
  const context = await browser.newContext({ baseURL: testInfo.project.use.baseURL, javaScriptEnabled: false })
  const page = await context.newPage()
  let posted: { path: string; type?: string; body: string } | undefined
  await page.route('**/actions/*', async route => {
    const submitted = route.request()
    posted = { path: new URL(submitted.url()).pathname, type: await submitted.headerValue('content-type') ?? undefined, body: submitted.postData() ?? '' }
    await route.fulfill({ status: 204 })
  })
  const submit = async (section: string, button: string) => {
    posted = undefined
    await settingRow(page, section).getByRole('button', { name: button }).click()
    await expect.poll(() => posted).toBeDefined()
    writes.length = 0
    const response = await request.post(posted!.path, {
      data: posted!.body,
      headers: {
        Host: host,
        Origin: `https://${host}`,
        'Content-Type': posted!.type ?? '',
        Cookie: 'tadoku_session_fixture=visitor',
        'X-Dev-Branch': 'w2-settings',
      },
      maxRedirects: 0,
    })
    expect.soft(response.status()).toBe(303)
    expect.soft(response.headers()['cache-control']).toBe('private, no-store')
    return new URL(response.headers().location)
  }
  const forwarded = (method: string, path: string, body: unknown) => ({
    method, path, origin: `https://${host}`, cookie: 'tadoku_session_fixture=visitor', branch: 'w2-settings', type: 'application/json', body,
  })

  for (const account of accounts) {
    await test.step(account.name, async () => {
      session = account.body
      await page.goto('/settings')
      // Without hydration the row is the parsed server HTML, which turns CRLF into LF.
      expect.soft(await storedValue(page, 'Email'), 'the server-rendered row shows the stored address')
        .toBe(account.email.replace(/\r\n/g, '\n'))
      await openSetting(page, 'Email')
      const email = settingRow(page, 'Email').getByRole('textbox', { name: 'New email', exact: true })
      await expect.soft(email).toHaveValue(account.shown)

      const location = await submit('Email', 'Save email')
      expect.soft(new URLSearchParams(posted?.body).get('email'), 'the browser posts the address its input shows').toBe(account.shown)
      expect.soft(location.pathname).toBe('/')
      expect.soft(location.searchParams.get('error')).toBe('unchanged_email')
      expect.soft(location.searchParams.get('form')).toBe('email')
      expect.soft(writes, 'an unchanged address sends nothing').toEqual([])
      await page.goto(`/settings${location.search}`)
      await expect.soft(settingRow(page, 'Email').locator('.paper-flash--danger')).toHaveText(unchanged)
      await page.screenshot({ path: testInfo.outputPath(`native-${account.name.replaceAll(' ', '-')}.png`), fullPage: true })

      await page.goto('/settings')
      await openSetting(page, 'Display name')
      await settingRow(page, 'Display name').getByRole('textbox', { name: 'Display name' }).fill('Renamed Reader')
      expect.soft((await submit('Display name', 'Save display name')).search).toBe('?saved=display-name')
      expect.soft(writes).toEqual([forwarded('PATCH', '/api/internal/identity/settings/profile', { display_name: 'Renamed Reader' })])

      await openSetting(page, 'Password')
      await settingRow(page, 'Password').getByLabel('New password').fill('a-long-reading-passphrase')
      expect.soft((await submit('Password', 'Save password')).search).toBe('?saved=password')
      expect.soft(writes).toEqual([forwarded('PUT', '/api/internal/identity/settings/password', { password: 'a-long-reading-passphrase' })])

      await openSetting(page, 'Email')
      await email.fill('edited@example.test')
      expect.soft((await submit('Email', 'Save email')).search).toBe('?saved=email')
      expect.soft(writes).toEqual([forwarded('PATCH', '/api/internal/identity/settings/profile', { email: 'edited@example.test' })])
    })
  }
  await context.close()
})
