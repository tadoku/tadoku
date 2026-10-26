import { createServer } from 'node:http'
import { test, expect, type Locator } from '@playwright/test'
import { expectContained, settingRow, waitForHydration } from './support'

// Failure modes: an editor that can't be opened or cancelled without JavaScript; two editors open at once, or an
// input at rest; Change or Cancel links that drop return_to or keep a stale error or confirmation; a native error
// that doesn't reopen its editor, or shows in another row; a confirmation missing from its collapsed row; focus that
// doesn't land in the opened input; more than one filled button on screen; a row that overflows 390px; the recovery
// landing with the Password editor closed or without its continuing Save.
// Runs on the Settings fixture's server (see e2e/README.md); only the session read reaches the fixture below.
const enabled = process.env.AUTHV2_SETTINGS_FIXTURE === 'true'
const session = JSON.stringify({
  account: {
    id: '00000000-0000-4000-8000-000000000001',
    email: 'reader@example.test',
    display_name: 'Reader',
    created_at: '2026-01-01T00:00:00Z',
    authenticated_at: '2026-01-01T00:00:00Z',
  },
  recent_login: true,
})
const fixture = createServer((request, response) => {
  if (request.url === '/api/internal/identity/session') {
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(session)
    return
  }
  response.writeHead(404).end()
})

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

const destination = 'https://tadoku.dev.lab/pages/dev-welcome?w2=rows'
const returnTo = encodeURIComponent(destination)
const rows = [['display-name', 'Display name'], ['email', 'Email'], ['password', 'Password']] as const
type Editor = (typeof rows)[number][0]

type State = {
  readonly path: string
  readonly open?: { readonly editor: Editor; readonly save: string; readonly problem?: RegExp }
  readonly saved?: { readonly editor: Editor; readonly message: string }
}

// Every URL Settings can be on: at rest, one editor open, a native error, a confirmation and the recovery landing.
const states: State[] = [
  { path: '/' },
  { path: '/settings' },
  { path: `/settings?return_to=${returnTo}` },
  { path: '/settings?edit=display-name', open: { editor: 'display-name', save: 'Save display name' } },
  { path: `/?edit=email&return_to=${returnTo}`, open: { editor: 'email', save: 'Save email' } },
  { path: '/settings?edit=password', open: { editor: 'password', save: 'Save password' } },
  {
    path: '/settings?form=email&error=email_taken&edit=password',
    open: { editor: 'email', save: 'Save email', problem: /already registered/ },
  },
  {
    path: `/?form=display-name&error=invalid_display_name&return_to=${returnTo}`,
    open: { editor: 'display-name', save: 'Save display name', problem: /Use 2 to 32 characters/ },
  },
  {
    path: `/settings?saved=email&return_to=${returnTo}`,
    saved: { editor: 'email', message: 'Email changed. We’ve let your previous address know.' },
  },
  { path: `/?edit=password&return_to=${returnTo}`, open: { editor: 'password', save: 'Save password and continue' } },
]

/** The link leads to exactly this Settings URL, with nothing else carried over. */
async function expectTarget(link: Locator, pathname: string, search: Record<string, string>) {
  await expect(link).toHaveCount(1)
  const url = new URL((await link.getAttribute('href')) ?? '', 'https://account.tadoku.dev.lab')
  expect({ pathname: url.pathname, search: Object.fromEntries(url.searchParams) }).toEqual({ pathname, search })
}

for (const javaScriptEnabled of [true, false]) {
  test.describe(javaScriptEnabled ? 'hydrated' : 'without JavaScript', () => {
    test.use({ javaScriptEnabled })

    for (const state of states) {
      test(state.path, async ({ page }) => {
        const errors: string[] = []
        page.on('pageerror', error => errors.push(error.message))
        page.on('console', message => { if (message.type() === 'error') errors.push(message.text()) })
        const url = new URL(state.path, 'https://account.tadoku.dev.lab')
        const kept: Record<string, string> = url.searchParams.has('return_to') ? { return_to: destination } : {}

        await page.goto(state.path)
        if (javaScriptEnabled) await waitForHydration(page)
        if (state.open) await expect(page.locator('main input:visible')).toHaveCount(1)
        else await expect(page.locator('main input:not([type="hidden"])')).toHaveCount(0)
        // The skip link sits outside the header and main, off screen until it has focus.
        const filled = page.locator('header, main').locator('.paper-button--default')
        await expect(filled).toHaveCount(state.open ? 1 : 0)
        if (state.open) await expect(filled).toHaveAccessibleName(state.open.save)

        for (const [editor, name] of rows) {
          const row = settingRow(page, name)
          const open = state.open?.editor === editor
          await expect(row).toHaveCount(1)
          await expect(row.locator('input:visible')).toHaveCount(open ? 1 : 0)
          if (open) {
            await expect(row.locator('input:visible')).toBeFocused()
            await expectTarget(row.getByRole('link', { name: 'Cancel', exact: true }), url.pathname, kept)
            await expect(row.getByRole('link', { name: /^Change/ })).toHaveCount(0)
          } else {
            const change = row.getByRole('link', { name: `Change ${name.toLowerCase()}`, exact: true })
            await expectTarget(change, url.pathname, { edit: editor, ...kept })
          }

          const problem = open ? state.open?.problem : undefined
          await expect(row.getByRole('alert')).toHaveCount(problem ? 1 : 0)
          if (problem) await expect(row.getByRole('alert')).toContainText(problem)
          const saved = state.saved?.editor === editor ? state.saved.message : undefined
          await expect(row.getByRole('status')).toHaveCount(saved ? 1 : 0)
          if (saved) await expect(row.getByRole('status')).toHaveText(saved)
        }
        await expect(page.getByRole('link', { name: 'Back to Tadoku' })).toHaveCount(kept.return_to ? 1 : 0)

        await expectContained(page)
        await page.setViewportSize({ width: 390, height: 844 })
        await expectContained(page)
        expect(errors, 'no hydration or runtime errors').toEqual([])
      })
    }

    test('the keyboard opens and cancels each editor and keeps return_to', async ({ page }) => {
      await page.goto(`/settings?return_to=${returnTo}`)
      if (javaScriptEnabled) await waitForHydration(page)
      const change = (name: string) => settingRow(page, name)
        .getByRole('link', { name: `Change ${name.toLowerCase()}`, exact: true })
      await expect(change('Display name')).toBeVisible()
      await change('Display name').focus()
      for (const [, name] of rows.slice(1)) {
        await page.keyboard.press('Tab')
        await expect(change(name)).toBeFocused()
      }

      for (const [editor, name] of rows) {
        const row = settingRow(page, name)
        await change(name).focus()
        await page.keyboard.press('Enter')
        await expect.poll(() => Object.fromEntries(new URL(page.url()).searchParams))
          .toEqual({ edit: editor, return_to: destination })
        await expect(row.locator('input:visible')).toBeFocused()
        await page.keyboard.press('Tab')
        await expect(row.locator('.paper-button--default')).toBeFocused()
        await page.keyboard.press('Tab')
        await expect(row.getByRole('link', { name: 'Cancel', exact: true })).toBeFocused()
        await page.keyboard.press('Enter')
        await expect.poll(() => Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ return_to: destination })
        await expect(page.locator('main input:not([type="hidden"])')).toHaveCount(0)
      }
    })
  })
}
