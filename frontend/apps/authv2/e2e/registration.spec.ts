import { createServer, type ServerResponse } from 'node:http'
import { test, expect, type Route } from '@playwright/test'
import { waitForHydration } from './support'

// Registration against controlled identity API answers, through both submission paths. The native form route only
// accepts an https account origin, so the server runs with TADOKU_WEB_URL=https://tadoku.dev.lab and TADOKU_API_URL
// pointing at the fixture below. Browser calls to that web host's API are answered in the test and never leave it.
const enabled = process.env.AUTHV2_REGISTRATION_FIXTURE === 'true'
const host = 'authv2.tadoku.dev.lab'
const api = 'https://tadoku.dev.lab/api/internal/identity/registration'

type Answer = { readonly status: number; readonly body?: string; readonly type?: string; readonly cookie?: string }
const json = (value: unknown) => JSON.stringify(value)
const created = json({
  account_app: 'v2',
  account: {
    id: '00000000-0000-4000-8000-000000000001',
    email: 'reader@example.test',
    display_name: 'Reader',
    created_at: '2026-01-01T00:00:00Z',
    authenticated_at: '2026-01-01T00:00:00Z',
  },
})
const cookie = 'tadoku_session_fixture=created; Path=/; HttpOnly; Secure; SameSite=Lax'
const unavailable = 'Something went wrong on our side. Try again in a moment.'

// Each answer the API can give, or that something between it and authv2 could, and what the visitor must see.
const failures: readonly {
  name: string
  answer: Answer | 'reset'
  message: string
  field?: 'email' | 'password'
  /** The message after a native submission, whose redirect carries the code but not the password. */
  native?: string
}[] = [
  {
    name: 'registration budget spent (429)',
    answer: { status: 429, body: json({ code: 'throttled', message: 'too many registration attempts; try again later' }) },
    message: 'Too many sign-up attempts from your network. Wait 15 minutes, then try again.',
  },
  {
    name: 'email taken (409)',
    answer: { status: 409, body: json({ code: 'email_taken', message: 'email is already registered' }) },
    message: 'This email is already registered. Log in or reset your password instead.',
    field: 'email',
  },
  {
    name: 'password rejected (400)',
    answer: { status: 400, body: json({ code: 'invalid_password', message: 'password is too similar to email' }) },
    message: 'This password is too similar to your email. Choose something else.',
    field: 'password',
    native: 'Use 8 or more characters, at most 72 bytes, that aren’t too similar to your email.',
  },
  { name: 'admission unavailable (503)', answer: { status: 503, body: json({ message: 'identity service unavailable' }) }, message: unavailable },
  { name: 'admission timeout (504, undocumented)', answer: { status: 504, body: '' }, message: unavailable },
  { name: 'cancelled (499, undocumented)', answer: { status: 499, body: '' }, message: unavailable },
  { name: 'proxy page with 429', answer: { status: 429, body: '<!doctype html><title>Busy</title>', type: 'text/html' }, message: unavailable },
  { name: 'malformed 201', answer: { status: 201, body: json({ account_app: 'v2', account: { id: 'x' } }), cookie }, message: unavailable },
  { name: 'undocumented 200', answer: { status: 200, body: created, cookie }, message: unavailable },
  { name: 'connection reset', answer: 'reset', message: unavailable },
]

let answer: Answer | 'reset' = { status: 201, body: created, cookie }
const sent: { origin?: string; cookie?: string; branch?: string; type?: string; body: unknown }[] = []

function reply(response: ServerResponse) {
  if (answer === 'reset') return response.socket?.destroy()
  const headers: Record<string, string> = { 'Content-Type': answer.type ?? 'application/json' }
  if (answer.cookie) headers['Set-Cookie'] = answer.cookie
  response.writeHead(answer.status, headers).end(answer.body)
}

const fixture = createServer((request, response) => {
  if (request.url === '/api/internal/identity/session') {
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(json({ account_app: 'v2', account: null }))
    return
  }
  if (request.url !== '/api/internal/identity/registration' || request.method !== 'POST') {
    response.writeHead(404).end()
    return
  }
  const chunks: Buffer[] = []
  request.on('data', chunk => chunks.push(chunk))
  request.on('end', () => {
    const { origin, cookie, 'x-dev-branch': branch, 'content-type': type } = request.headers
    sent.push({ origin, cookie, branch: branch?.toString(), type, body: JSON.parse(Buffer.concat(chunks).toString('utf8')) })
    reply(response)
  })
})

async function answerBrowser(route: Route) {
  const request = route.request()
  const cors = { 'Access-Control-Allow-Origin': await request.headerValue('origin') ?? '*', 'Access-Control-Allow-Credentials': 'true' }
  if (request.method() === 'OPTIONS') {
    return route.fulfill({ status: 204, headers: { ...cors, 'Access-Control-Allow-Methods': 'POST', 'Access-Control-Allow-Headers': 'content-type' } })
  }
  sent.push({ type: await request.headerValue('content-type') ?? undefined, body: request.postDataJSON() })
  if (answer === 'reset') return route.abort('connectionreset')
  const headers: Record<string, string> = { ...cors }
  if (answer.cookie) headers['Set-Cookie'] = answer.cookie
  return route.fulfill({ status: answer.status, contentType: answer.type ?? 'application/json', headers, body: answer.body })
}

const values = { display_name: 'Reader', email: 'reader@example.test', password: 'a-long-reading-passphrase' }

test.beforeAll(async () => {
  if (enabled) await new Promise<void>(resolve => fixture.listen(5195, '127.0.0.1', resolve))
})

test.afterAll(async () => {
  if (enabled) await new Promise<void>((resolve, reject) => fixture.close(error => error ? reject(error) : resolve()))
})

test.beforeEach(() => {
  test.skip(!enabled, 'needs the local registration fixture')
})

test('hydrated registration shows each answer and only leaves the form on success', async ({ page }, testInfo) => {
  await page.route(api, answerBrowser)
  const form = page.locator('form#register')
  // A fresh page per answer, so a message left by the previous answer can't satisfy the next one.
  const submit = async (answered: Answer | 'reset') => {
    await page.goto('/register')
    await waitForHydration(page)
    answer = answered
    sent.length = 0
    await form.getByLabel('Display name').fill(values.display_name)
    await form.getByLabel('Email').fill(values.email)
    await form.getByLabel('Password').fill(values.password)
    await form.getByRole('button', { name: 'Create account' }).click()
    await expect.poll(() => sent.length).toBe(1)
  }

  for (const failure of failures) {
    await test.step(failure.name, async () => {
      await submit(failure.answer)
      const shown = failure.field ? page.locator(`#register-${failure.field}-error`) : form.locator('.paper-flash--danger')
      await expect.soft(shown).toHaveText(failure.message)
      await expect.soft(form.getByRole('button', { name: 'Create account' })).toBeEnabled()
      expect.soft(new URL(page.url()).pathname, 'stays on the form').toBe('/register')
      await expect.soft(form.getByLabel('Email')).toHaveValue(values.email)
      expect.soft(sent, 'one request, no automatic retry').toEqual([{ type: 'application/json', body: values }])
      if (failure.name.startsWith('registration budget')) {
        await page.screenshot({ path: testInfo.outputPath('hydrated-429.png'), fullPage: true })
        await page.setViewportSize({ width: 390, height: 844 })
        expect.soft(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'fits 390px').toBe(true)
        await page.screenshot({ path: testInfo.outputPath('hydrated-429-390.png'), fullPage: true })
        await page.setViewportSize({ width: 1280, height: 844 })
      }
    })
  }

  await test.step('201 with a session continues to the account', async () => {
    await submit({ status: 201, body: created, cookie })
    await page.waitForURL(url => url.pathname !== '/register')
    expect.soft(sent).toEqual([{ type: 'application/json', body: values }])
  })
})

test('native registration forwards the request and only passes on a successful session cookie', async ({ page, request }, testInfo) => {
  const post = () => request.post('/actions/register', {
    form: values,
    headers: { Host: host, Origin: `https://${host}`, Cookie: 'tadoku_session_fixture=visitor', 'X-Dev-Branch': 'w2-register' },
    maxRedirects: 0,
  })
  const forwarded = { origin: `https://${host}`, cookie: 'tadoku_session_fixture=visitor', branch: 'w2-register', type: 'application/json', body: values }

  for (const failure of failures) {
    await test.step(failure.name, async () => {
      answer = failure.answer
      sent.length = 0
      const response = await post()
      expect.soft(response.status()).toBe(303)
      expect.soft(response.headers()['cache-control']).toBe('private, no-store')
      expect.soft(response.headers()['set-cookie'], 'no session cookie').toBeUndefined()
      const location = new URL(response.headers().location)
      expect.soft(location.pathname).toBe('/register')
      expect.soft(location.searchParams.get('form')).toBe('register')
      expect.soft(sent, 'one forwarded request, no automatic retry').toEqual([forwarded])

      // A native rejection is shown on the form, not on a field: the redirect carries only the error code.
      await page.goto(`/register${location.search}`)
      await expect.soft(page.locator('form#register .paper-flash--danger')).toHaveText(failure.native ?? failure.message)
      if (failure.name.startsWith('registration budget')) {
        await page.screenshot({ path: testInfo.outputPath('native-429.png'), fullPage: true })
      }
    })
  }

  await test.step('201 with a session passes its cookie on and continues', async () => {
    answer = { status: 201, body: created, cookie }
    sent.length = 0
    const response = await post()
    expect.soft(response.status()).toBe(303)
    expect.soft(response.headers().location).toBe(`https://${host}/`)
    expect.soft(response.headers()['set-cookie']).toBe(cookie)
    expect.soft(response.headers()['cache-control']).toBe('private, no-store')
    expect.soft(sent).toEqual([forwarded])
  })
})
