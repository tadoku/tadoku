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

/**
 * Waits until a Webv3 page has hydrated, when Paper's navbar replaces its link to the menu with a button, which wide
 * viewports hide with CSS. Leaving a page before then cancels its module requests.
 */
export const hydrated = (page: Page) =>
  expect(page.getByRole('button', { name: /^(Open|Close) menu$/, includeHidden: true })).toBeAttached({ timeout: 15_000 })

/** Submits authv2's login form once the page has hydrated, when React has attached its props to the Email field. */
export async function authenticate(page: Page, account = reader) {
  const email = page.getByLabel('Email')
  await expect.poll(() => email.evaluate(input => Object.keys(input).some(key => key.startsWith('__reactProps'))), { timeout: 15_000 })
    .toBe(true)
  await email.fill(account.email!)
  await page.getByLabel('Password').fill(account.password!)
  await page.getByRole('button', { name: 'Log in', exact: true }).click()
}

/** Opens the activity form anonymously, logs in through authv2 and waits for the form on its return. */
export async function signIn(page: Page, account = reader) {
  await page.goto('/logs/new')
  await hydrated(page)
  await page.getByRole('link', { name: 'Log in', exact: true }).first().click()
  await authenticate(page, account)
  await page.waitForURL(`${webv3}/logs/new`)
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
}

/** Logs out from the account menu, as Webv2's user dropdown does, and waits for the homepage it returns to. */
export async function logOut(page: Page, { name } = reader) {
  await expect(page.getByRole('button', { name, exact: true })).toBeEnabled()
  await page.getByRole('button', { name, exact: true }).click()
  await page.getByRole('menuitem', { name: 'Log out', exact: true }).click()
  await page.waitForURL(`${webv3}/`)
  await hydrated(page)
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

/** The links to `href`, by their destination, because a link's visible name often is the data the test reads. */
export const linkTo = (page: Page, href: string) => page.getByRole('link').and(page.locator(`[href="${href}"]`))

/** Each listed contest's row, in the list every contest collection labels. */
export const contestRows = (page: Page) => page.getByRole('list', { name: /contests/i }).getByRole('listitem')

/** Each log's row in a page of activity. */
export const activityEntries = (page: Page) => page.getByRole('list', { name: 'Activity', exact: true }).getByRole('listitem')

/** Each listed log's link to its detail page, beside any link to its learner. */
export const logLinks = (page: Page) => activityEntries(page).getByRole('link').and(page.locator('[href^="/logs/"]'))

/** The value of a page's statistic with this label. */
export const statValue = (page: Page, label: string) =>
  page.locator('dl > div').filter({ has: page.getByRole('term').filter({ hasText: label }) }).getByRole('definition')

/** Every statistic's value on the page. */
export const statValues = (page: Page) => page.getByRole('main').getByRole('definition')

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

/**
 * A new open contest created by `organizer` and joined in Japanese by `participant`, so a test need not rely on a
 * seeded round or its dates. `organizer` must be able to create contests, which rules out freshly registered accounts.
 */
export async function joinedContest(organizer: BrowserContext, participant: BrowserContext) {
  const day = (offset: number) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
  const created = await organizer.request.post(`${api}/immersion/contests`, {
    data: {
      title: `Webv3 joined contest ${randomUUID().slice(0, 8)}`,
      description: 'Created by a Webv3 journey.',
      contest_start: day(0),
      contest_end: day(30),
      registration_end: day(14),
      private: false,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(created.ok(), `contest creation answered ${created.status()}`).toBeTruthy()
  const { id } = await created.json() as { id: string }
  const joined = await participant.request.post(`${api}/immersion/contests/${id}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.status(), 'registration').toBe(200)
  const registration = await (await participant.request.get(`${api}/immersion/contests/${id}/registration`)).json() as { id: string }
  return { id, registrationId: registration.id }
}

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

/** The file that holds `key`'s state for this host, shared by every worker and run on the machine. */
const stateFile = (key: string) => join(sessions, `${createHash('sha256').update(`${webv3}\n${key}`).digest('hex').slice(0, 16)}.json`)

const readState = <T>(file: string) => readFile(file, 'utf8').then(text => JSON.parse(text) as T, () => undefined)

async function writeState(file: string, state: unknown) {
  await mkdir(sessions, { recursive: true, mode: 0o700 })
  // ponytail: no lock, so two runs that start together on one machine can each do the work and the last rename wins;
  // add a cross-process lock if the suite ever runs with several workers.
  const temporary = `${file}.${process.pid}.${randomUUID()}.tmp`
  await writeFile(temporary, JSON.stringify(state), { mode: 0o600 })
  await rename(temporary, file)
}

/** The account's session on this host, kept on disk so later workers and runs reuse it while the API still accepts it. */
async function storedSession(browser: Browser, account: Account) {
  if (!webv3 || !account.email || !account.password) return []
  const file = stateFile(account.email)
  const stored = await readState<Cookie[]>(file) ?? []
  if (await signedIn(browser, stored)) return stored

  const cookies = await session(browser, account)
  await writeState(file, cookies)
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
const loginWindow = 15 * 60_000
/** Below the API's ten, so a test that signs in twice or a run cut short cannot reach the refusal. */
const loginsPerAccount = 8

interface StoredLearner extends Omit<Learner, 'context'> { cookies: Cookie[]; logins: number[] }

/**
 * The host's reusable learner, kept on disk so a worker restarted by a failed test registers nothing. Each login
 * spends from the account's budget; a spent account, or one whose session the API no longer accepts, is replaced.
 */
async function storedLearner(browser: Browser, spendLogin: boolean) {
  const file = stateFile('learner')
  const now = Date.now()
  const stored = await readState<StoredLearner>(file)
  const logins = stored?.logins.filter(at => now - at < loginWindow) ?? []
  let learner = stored && { ...stored, logins }
  if (!learner || (spendLogin && logins.length >= loginsPerAccount) || !await signedIn(browser, learner.cookies)) {
    const { context, ...account } = await register(browser)
    learner = { ...account, cookies: await context.cookies(), logins: [] }
    await context.close()
  }
  if (spendLogin) learner.logins.push(now)
  await writeState(file, learner)
  return learner
}

/**
 * Accounts for journeys, within the identity API's budgets: ten logins per account and ten registrations per tenant
 * and client IP, each in 15 minutes.
 * - `readerContext` and `adminContext` hold Dev Reader's and Dev Admin's stored sessions. Never log them out.
 * - `learner` is a new account per test, signed in, for journeys that need an untouched account.
 * - `returningLearner` is a stored, signed-in account reused across tests and runs, for journeys that only need a young
 *   account and whose writes are safe to repeat.
 * - `loginLearner` is that stored account's credentials, for journeys that sign in, out or switch accounts through
 *   authv2. Each test that takes it may sign in once.
 */
export const test = base.extend<{
  readerContext: BrowserContext
  adminContext: BrowserContext
  learner: Learner
  returningLearner: Learner
  loginLearner: Required<Account>
}, { readerCookies: Cookie[]; adminCookies: Cookie[] }>({
  // Set here rather than with a module-level test.use(), which Playwright attributes to whichever spec loads this file
  // first and which then breaks the fixture digest of the other specs sharing its worker.
  trace: ['off', { scope: 'worker' }],
  video: ['off', { scope: 'worker' }],
  ignoreHTTPSErrors: false,
  readerCookies: [async ({ browser }, provide) => provide(await storedSession(browser, reader)), { scope: 'worker' }],
  adminCookies: [async ({ browser }, provide) => provide(await storedSession(browser, admin)), { scope: 'worker' }],
  loginLearner: async ({ browser }, provide) => {
    base.skip(!credentialed, 'Set the credentialed development environment')
    const { email, password, name } = await storedLearner(browser, true)
    await provide({ email, password, name })
  },
  returningLearner: async ({ browser }, provide) => {
    base.skip(!credentialed, 'Set the credentialed development environment')
    const { cookies, id, email, password, name } = await storedLearner(browser, false)
    const context = await browser.newContext({ baseURL: webv3 })
    await context.addCookies(cookies)
    await provide({ id, email, password, name, context })
    await context.close()
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
