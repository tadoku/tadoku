import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { expect, type Browser, type BrowserContextOptions, type Page, type TestInfo } from '@playwright/test'

export const readerEmail = process.env.AUTHV2_READER_EMAIL || 'reader@tadoku.app'
export const adminEmail = process.env.AUTHV2_ADMIN_EMAIL || 'dev@tadoku.app'
export const adminPassword = () => process.env.AUTHV2_ADMIN_PASSWORD || 'tadoku'

// The recovery journey replaces the branch Reader password; repeat runs read it back from this private file.
const stateFile = process.env.AUTHV2_READER_STATE_FILE

export function readerPassword() {
  if (stateFile && existsSync(stateFile)) return readFileSync(stateFile, 'utf8')
  return process.env.AUTHV2_READER_PASSWORD || 'tadoku'
}

export function saveReaderPassword(password: string) {
  if (stateFile) writeFileSync(stateFile, password, { mode: 0o600 })
}

export function origins(testInfo: TestInfo) {
  const authv2 = new URL(testInfo.project.use.baseURL as string).origin
  return { authv2, main: authv2.replace(/(\/\/|\.)account\./, '$1') }
}

export async function newPage(browser: Browser, testInfo: TestInfo, options: BrowserContextOptions = {}) {
  const context = await browser.newContext({ baseURL: testInfo.project.use.baseURL, ...options })
  return context.newPage()
}

export async function logIn(page: Page, email: string, password: string, path = '/login') {
  await page.goto(path)
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Log in', exact: true }).click()
}

export async function expectSettings(page: Page) {
  await expect(page.getByRole('heading', { level: 1, name: 'Account settings' })).toBeVisible()
}

export async function expectContained(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
}

/**
 * React attaches its props to every hydrated element, so this works for any build of the screens. Settings at rest
 * has no input, only links.
 */
export async function waitForHydration(page: Page) {
  await page.waitForFunction(() => {
    const control = document.querySelector('main :is(input:not([type="hidden"]), a[href])')
    return control !== null && Object.keys(control).some(key => key.startsWith('__reactProps'))
  })
}

/** A Settings row, the region its heading labels: "Display name", "Email" or "Password". */
export function settingRow(page: Page, name: string) {
  return page.getByRole('region', { name, exact: true })
}

/** The stored value a collapsed Settings row shows. */
export function settingValue(page: Page, name: string) {
  return settingRow(page, name).locator('.settings-row__value')
}

/** Opens a collapsed Settings row's editor through its Change link, with or without JavaScript. */
export async function openSetting(page: Page, name: string) {
  const row = settingRow(page, name)
  await row.getByRole('link', { name: `Change ${name.toLowerCase()}`, exact: true }).click()
  await expect(row.getByRole('link', { name: 'Cancel', exact: true })).toBeVisible()
}

export async function holdScripts(page: Page) {
  let release!: () => void
  const gate = new Promise<void>(resolve => { release = resolve })
  let held = 0
  await page.route('**/*', async route => {
    if (route.request().resourceType() === 'script') {
      held++
      await gate
    }
    await route.continue()
  })
  return {
    held: () => held,
    release: async () => {
      release()
      await waitForHydration(page)
      await page.unroute('**/*')
    },
  }
}

/** Resolves with the first non-GET request the page sends after `action` starts. */
export async function firstWrite(page: Page, action: () => Promise<void>) {
  const request = page.waitForRequest(candidate => candidate.method() !== 'GET')
  await action()
  return request
}
