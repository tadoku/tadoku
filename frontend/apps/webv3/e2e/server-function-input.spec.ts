import type { APIResponse, Response } from '@playwright/test'
import { expect, expectDraftKept, fillDraft, hydrated, isSaveRequest, openForm, reader, saveButton, test, webv3 } from './logging-fixtures'

/** The headers TanStack Start's client sends with every server function call. */
const clientHeaders = { 'x-tsr-serverFn': 'true', accept: 'application/x-tss-framed, application/x-ndjson, application/json' }

/** A plain 400 that names no error and shows no stack, so a malformed call never counts as a server error. */
async function expectBadRequest(response: APIResponse | Response, label: string) {
  expect(response.status(), label).toBe(400)
  expect(response.headers()['content-type'], label).toContain('text/plain')
  expect(await response.text(), label).not.toMatch(/error|\bat /i)
}

test('a server function call whose payload does not decode answers 400', async ({ page, request }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/')
  await hydrated(page)

  // The leaderboard's loader calls a server function from the browser, which gives this build's ID for it.
  const call = page.waitForRequest(each => each.url().includes('/_serverFn/') && each.method() === 'GET')
  await page.getByRole('navigation', { name: 'Main navigation' }).getByRole('link', { name: 'Leaderboard', exact: true }).first().click()
  const url = new URL((await call).url())

  for (const payload of ['{bad', '{}']) {
    url.searchParams.set('payload', payload)
    await expectBadRequest(await request.get(url.toString(), { headers: clientHeaders }), payload)
  }
})

test('a saved activity whose request body does not decode answers 400 and keeps the draft', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set development fixture credentials')
  const page = await readerContext.newPage()
  await openForm(page)
  await fillDraft(page, '12', 'malformed save')
  await page.route('**/_serverFn/**', route => isSaveRequest(route.request()) ? route.continue({ postData: '{bad' }) : route.continue())

  const save = page.waitForResponse(response => isSaveRequest(response.request()))
  await saveButton(page).click()
  await expectBadRequest(await save, 'POST {bad')
  await expect(page).toHaveURL(/\/logs\/new$/)
  await expectDraftKept(page, '12', 'malformed save')
})
