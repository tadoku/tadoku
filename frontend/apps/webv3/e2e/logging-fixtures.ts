import { createHash, randomUUID } from 'node:crypto'
import { mkdir, readFile, rename, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { expect, test as base, type Browser, type BrowserContext, type Cookie, type Page, type Request } from '@playwright/test'

export const webv3 = process.env.WEBV3_BASE_URL
export const api = process.env.TADOKU_API_URL
export const branch = webv3 ? new URL(webv3).hostname.split('.next.')[0] : undefined

/** An account a test can sign in as by typing its credentials into authv2. */
export interface Account { email?: string; password?: string; name: string }

/** A freshly registered account and a context signed in as it. */
export interface Learner extends Required<Account> { id: string; context: BrowserContext }

export const reader: Account = { email: process.env.E2E_READER_EMAIL, password: process.env.E2E_READER_PASSWORD, name: 'Dev Reader' }
export const admin: Account = { email: process.env.E2E_SECOND_EMAIL, password: process.env.E2E_SECOND_PASSWORD, name: 'Dev Admin' }

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
export async function logOut(page: Page, { name } = reader) {
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

/** The saved log's description on its detail page, read from its labelled row. */
/** The log's note: its page heading when short, or the Description row when long. */
export const logDescription = (page: Page) => page.locator('main h1, dt:has-text("Description") + dd').last()

/** Where a save of a new log leads an account with ongoing registrations: the contest step, as in Webv2. */
export const contestStepUrl = /\/logs\/[a-f0-9-]{36}\/contests\?preselect=1$/

/** A save: the only POST the activity form sends, to its server function. */
export const isSaveRequest = (request: Request) => request.method() === 'POST' && request.url().includes('_serverFn')

export const saveButton = (page: Page) => page.getByRole('button', { name: 'Log activity', exact: true })

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
  const input = page.getByRole('combobox', { name: /^Tags \(/ })
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
  await page.getByLabel(/^Description/).fill(description)
}

/** The open activity form still holds the draft, unsaved. */
export async function expectDraftKept(page: Page, amount: string, description: string) {
  await expect(amountInput(page)).toHaveValue(amount)
  await expect(page.getByLabel(/^Description/)).toHaveValue(description)
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

async function session(browser: Browser, account: Account) {
  const context = await browser.newContext({ baseURL: webv3 })
  try {
    await signIn(await context.newPage(), account)
    return await context.cookies()
  } finally {
    await context.close()
  }
}

async function signedIn(browser: Browser, cookies: Cookie[]) {
  if (!api || !cookies.length) return false
  const context = await browser.newContext()
  try {
    await context.addCookies(cookies)
    const response = await context.request.get(`${api}/identity/session`)
    return response.ok() && Boolean((await response.json()).account)
  } finally {
    await context.close()
  }
}

const sessions = join(tmpdir(), 'tadoku-webv3-e2e-sessions')

/**
 * The account's session on this host, kept on disk so later workers and runs reuse it while the API still accepts it.
 * ponytail: no lock, so workers that start together may each sign in once; add one if the suite runs in parallel.
 */
async function storedSession(browser: Browser, account: Account) {
  if (!webv3 || !account.email || !account.password) return []
  const file = join(sessions, `${createHash('sha256').update(`${webv3}\n${account.email}`).digest('hex').slice(0, 16)}.json`)
  const stored = await readFile(file, 'utf8').then(text => JSON.parse(text) as Cookie[], () => [])
  if (await signedIn(browser, stored)) return stored
  const cookies = await session(browser, account)
  await mkdir(sessions, { recursive: true, mode: 0o700 })
  await writeFile(`${file}.tmp`, JSON.stringify(cookies), { mode: 0o600 })
  await rename(`${file}.tmp`, file)
  return cookies
}

/** Registers an account in the selected tenant; registration starts its session without spending a login. */
async function register(browser: Browser): Promise<Learner> {
  const context = await browser.newContext({ baseURL: webv3 })
  const account = { email: `webv3-${randomUUID()}@example.com`, password: `Pw-${randomUUID()}`, name: `Learner ${randomUUID().slice(0, 8)}` }
  const origin = new URL(webv3!)
  origin.hostname = origin.hostname.replace(/(^|\.)next\./, '$1account.')
  const response = await context.request.post(`${api}/identity/registration`, {
    headers: { Origin: origin.origin },
    data: { email: account.email, display_name: account.name, password: account.password },
  })
  expect(response.status(), 'registration in the selected tenant').toBe(201)
  return { ...account, id: await accountId(context), context }
}

const credentialed = Boolean(webv3 && api && reader.email)
let loginAccount: Promise<Required<Account>> | undefined

/**
 * Accounts for journeys, within the identity API's budgets: ten logins per account and ten registrations per tenant
 * and client IP, each in 15 minutes.
 * - `readerContext` and `adminContext` hold Dev Reader's and Dev Admin's stored sessions. Never log them out.
 * - `learner` is a new account per test, signed in, for journeys whose writes must not carry over to the next run.
 * - `loginLearner` is one new account per worker for journeys that sign in, out or switch accounts through authv2.
 */
export const test = base.extend<{
  readerContext: BrowserContext
  adminContext: BrowserContext
  learner: Learner
  loginLearner: Required<Account>
}, { readerCookies: Cookie[]; adminCookies: Cookie[] }>({
  // Set here rather than with a module-level test.use(), which Playwright attributes to whichever spec loads this file
  // first and which then breaks the fixture digest of the other specs sharing its worker.
  trace: ['off', { scope: 'worker' }],
  video: ['off', { scope: 'worker' }],
  ignoreHTTPSErrors: false,
  readerCookies: [async ({ browser }, provide) => provide(await storedSession(browser, reader)), { scope: 'worker' }],
  adminCookies: [async ({ browser }, provide) => provide(await storedSession(browser, admin)), { scope: 'worker' }],
  // Cached per worker process; a test-scoped fixture can skip where the environment has no credentials.
  loginLearner: async ({ browser }, provide) => {
    base.skip(!credentialed, 'Set the credentialed development environment')
    loginAccount ??= register(browser).then(async ({ email, password, name, context }) => {
      await context.close()
      return { email, password, name }
    })
    await provide(await loginAccount)
  },
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
  learner: async ({ browser }, provide) => {
    base.skip(!credentialed, 'Set the credentialed development environment')
    const learner = await register(browser)
    await provide(learner)
    await learner.context.close()
  },
})

export { expect }
