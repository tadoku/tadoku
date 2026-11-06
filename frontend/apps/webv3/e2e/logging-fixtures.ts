import { expect, test as base, type Browser, type BrowserContext, type Cookie, type Page, type Request } from '@playwright/test'

export const webv3 = process.env.WEBV3_BASE_URL
export const api = process.env.TADOKU_API_URL
export const reader = { email: process.env.E2E_READER_EMAIL, password: process.env.E2E_READER_PASSWORD }
export const admin = { email: process.env.E2E_SECOND_EMAIL, password: process.env.E2E_SECOND_PASSWORD }
export const branch = webv3 ? new URL(webv3).hostname.split('.next.')[0] : undefined

export async function authenticate(page: Page, account = reader) {
  await page.getByLabel('Email').fill(account.email!)
  await page.getByLabel('Password').fill(account.password!)
  await page.getByRole('button', { name: 'Log in', exact: true }).click()
}

/** Opens the activity form anonymously, logs in through authv2 and waits for the form on its return. */
export async function signIn(page: Page, account = reader) {
  await page.goto('/logs/new')
  await page.getByRole('link', { name: 'Log in', exact: true }).first().click()
  await authenticate(page, account)
  await page.waitForURL(`${webv3}/logs/new`)
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
}

/** Logs out from the account menu, as Webv2's user dropdown does. */
export async function logOut(page: Page, name = 'Dev Reader') {
  await expect(page.getByRole('button', { name, exact: true })).toBeEnabled()
  await page.getByRole('button', { name, exact: true }).click()
  await page.getByRole('menuitem', { name: 'Log out', exact: true }).click()
}

/** Opens the activity form in a context that already holds a session from `signIn`. */
export async function openForm(page: Page) {
  await page.goto('/logs/new')
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
}

export const savedLogUrl = /\/logs\/[a-f0-9-]{36}$/

/** Where a save of a new log leads an account with ongoing registrations: the contest step, as in Webv2. */
export const contestStepUrl = /\/logs\/[a-f0-9-]{36}\/contests\?preselect=1$/

/** A save: the only POST the activity form sends, to its server function. */
export const isSaveRequest = (request: Request) => request.method() === 'POST' && request.url().includes('_serverFn')

export const saveButton = (page: Page) => page.getByRole('button', { name: 'Create', exact: true })

/**
 * Creates the open draft and returns its ID, leaving the page on the saved log's detail. An account with ongoing
 * registrations first reaches the contest step; the log stays personal because that step is not saved.
 */
export async function createLog(page: Page) {
  await saveButton(page).click()
  await page.waitForURL(url => savedLogUrl.test(url.href) || contestStepUrl.test(url.href))
  const id = savedLogId(page)
  if (!savedLogUrl.test(page.url())) await page.goto(`/logs/${id}`)
  return id
}

/** Adds each tag as a chip by typing it and pressing Enter. */
export async function addTags(page: Page, tags: string[]) {
  const input = page.getByRole('combobox', { name: /^Tags \(up to/ })
  for (const tag of tags) {
    await input.fill(tag)
    await input.press('Enter')
  }
}

/** Paper's amount field is a group labelled "Amount" that holds the amount input and the unit select. */
export const amountInput = (page: Page) => page.getByRole('spinbutton', { name: 'Amount', exact: true })

/** Enters an amount and a description in the open activity form. */
export async function fillDraft(page: Page, amount: string, description: string) {
  await amountInput(page).fill(amount)
  await page.getByLabel('Description', { exact: true }).fill(description)
}

/** The open activity form still holds the draft, unsaved. */
export async function expectDraftKept(page: Page, amount: string, description: string) {
  await expect(amountInput(page)).toHaveValue(amount)
  await expect(page.getByLabel('Description', { exact: true })).toHaveValue(description)
  await expect(page).toHaveURL(/\/logs\/new$/)
}

/** The log ID from its detail page or its contest step. */
export function savedLogId(page: Page) {
  return new URL(page.url()).pathname.split('/')[2]
}

/** The account's logs with this description, read from the API. */
export async function logsDescribed(context: BrowserContext, userId: string, description: string) {
  const response = await context.request.get(`${api}/immersion/users/${userId}/logs?page_size=100&page=0`)
  const { logs } = await response.json() as { logs: { description?: string }[] }
  return logs.filter(log => log.description === description)
}

/** The tenant's session cookie, which the identity API names after the tenant. */
export async function sessionCookie(context: BrowserContext) {
  return (await context.cookies()).find(cookie => cookie.name.startsWith('tadoku_session'))!
}

export async function accountId(context: BrowserContext) {
  return (await (await context.request.get(`${api}/identity/session`)).json()).account.id as string
}

async function session(browser: Browser, account: typeof reader) {
  if (!webv3 || !account.email || !account.password) return []
  const context = await browser.newContext({ baseURL: webv3 })
  try {
    await signIn(await context.newPage(), account)
    return await context.cookies()
  } finally {
    await context.close()
  }
}

/**
 * The identity API admits ten logins per account in 15 minutes, counting successful ones, so journeys where login is
 * incidental share one real session per account per worker, held only in memory. Journeys that log out, switch
 * accounts or test login itself sign in on their own.
 */
export const test = base.extend<{ readerContext: BrowserContext; adminContext: BrowserContext }, {
  readerCookies: Cookie[]
  adminCookies: Cookie[]
}>({
  // Set here rather than with a module-level test.use(), which Playwright attributes to whichever spec loads this file
  // first and which then breaks the fixture digest of the other specs sharing its worker.
  trace: ['off', { scope: 'worker' }],
  video: ['off', { scope: 'worker' }],
  ignoreHTTPSErrors: false,
  readerCookies: [async ({ browser }, provide) => provide(await session(browser, reader)), { scope: 'worker' }],
  adminCookies: [async ({ browser }, provide) => provide(await session(browser, admin)), { scope: 'worker' }],
  readerContext: async ({ context, readerCookies }, provide) => {
    await context.addCookies(readerCookies)
    await provide(context)
  },
  adminContext: async ({ browser, adminCookies }, provide) => {
    const context = await browser.newContext({ baseURL: webv3 })
    await context.addCookies(adminCookies)
    await provide(context)
    await context.close()
  },
})

export { expect }
