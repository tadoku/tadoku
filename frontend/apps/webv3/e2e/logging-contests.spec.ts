import { score } from './standings'
import type { BrowserContext, Page } from '@playwright/test'
import {
  createLog,
  admin,
  amountInput,
  api,
  branch,
  expect,
  joinedContest,
  logDescription,
  openForm,
  reader,
  test,
  webv3,
} from './logging-fixtures'

type Submission = { registration_id: string; contest_id: string; title: string; score?: number }

const saveSubmissions = (page: Page) => page.getByRole('button', { name: 'Save submissions', exact: true })
const manageLink = (page: Page) => page.getByRole('link', { name: 'Manage contest submissions', exact: true })

/** Saves a reading log to the personal record only and returns its ID, leaving the page on its detail. */
async function createPersonalLog(page: Page, text: string) {
  await openForm(page)
  // The default follows the account's previously used languages; contests the test joined are in Japanese.
  await page.getByLabel('Language', { exact: true }).selectOption('jpn')
  await page.getByLabel('Activity', { exact: true }).selectOption({ label: 'Reading' })
  await amountInput(page).fill('7')
  await page.getByLabel(/^Description/).fill(text)
  return createLog(page)
}

async function submissions(context: BrowserContext, id: string) {
  const response = await context.request.get(`${api}/immersion/logs/${id}`)
  expect(response.status()).toBe(200)
  expect(response.headers()['x-dev-selected']).toBe(branch)
  return ((await response.json()).registrations ?? []) as Submission[]
}

test('an anonymous visitor is asked to sign in and sees no contest choices', async ({ page }) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto(`/logs/${crypto.randomUUID()}/contests`)
  await expect(page.getByRole('heading', { name: 'Sign in to submit to contests' })).toBeVisible()
  await expect(page.getByRole('checkbox')).toHaveCount(0)
  await expect(saveSubmissions(page)).toHaveCount(0)
})

test('owner submits a saved log to a contest, sees it after reload, then removes it', async ({ page, readerContext, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  await joinedContest(adminContext, readerContext)
  const id = await createPersonalLog(page, `contests-${Date.now()}`)
  expect(await submissions(readerContext, id)).toEqual([])
  await expect(page.getByText('Saved to your personal record only.')).toBeVisible()

  await manageLink(page).click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}/contests`)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Submit to contests')
  await expect(saveSubmissions(page)).toBeEnabled({ timeout: 15_000 })
  const eligible = page.getByRole('checkbox', { disabled: false })
  expect(await eligible.count()).toBeGreaterThan(0)
  for (const contest of await page.getByRole('checkbox').all()) await expect(contest).not.toBeChecked()

  // Webv2's `?preselect=1` starts a log not yet submitted anywhere with every eligible contest selected.
  await page.goto(`/logs/${id}/contests?preselect=1`)
  await expect(saveSubmissions(page)).toBeEnabled({ timeout: 15_000 })
  for (const contest of await eligible.all()) await expect(contest).toBeChecked()
  await expect(page.getByRole('link', { name: 'Skip', exact: true })).toHaveAttribute('href', `/logs/${id}`)
  await page.goto(`/logs/${id}/contests`)
  await expect(saveSubmissions(page)).toBeEnabled({ timeout: 15_000 })
  await expect(page.getByRole('link', { name: 'Cancel', exact: true })).toHaveAttribute('href', `/logs/${id}`)
  await page.screenshot({ path: testInfo.outputPath('contests-desktop.png'), fullPage: true })
  await page.setViewportSize({ width: 390, height: 844 })
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390)
  await page.screenshot({ path: testInfo.outputPath('contests-390.png'), fullPage: true })
  await page.setViewportSize({ width: 1280, height: 800 })

  const chosen = (await eligible.first().getAttribute('value'))!
  await eligible.first().check()
  await saveSubmissions(page).click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}`)

  const [submission] = await submissions(readerContext, id)
  expect(submission.registration_id).toBe(chosen)
  const { title } = submission
  expect(submission.score).toBeGreaterThan(0)
  await page.reload()
  const contests = page.getByRole('region', { name: 'Contests' })
  await expect(contests.getByRole('link', { name: title, exact: true })).toBeVisible()
  await expect(contests).toContainText(`${score.format(submission.score ?? 0)} points`)
  await page.screenshot({ path: testInfo.outputPath('detail-submitted.png'), fullPage: true })

  await page.goto(`/logs/${id}/contests`)
  await expect(saveSubmissions(page)).toBeEnabled({ timeout: 15_000 })
  const saved = page.getByRole('checkbox', { name: title })
  await expect(saved).toBeChecked()
  await saved.uncheck()
  await saveSubmissions(page).click()
  await expect(page).toHaveURL(`${webv3}/logs/${id}`)
  await expect(page.getByText('Saved to your personal record only.')).toBeVisible()
  expect(await submissions(readerContext, id)).toEqual([])
  await page.reload()
  await expect(page.getByText('Saved to your personal record only.')).toBeVisible()
})

test('another learner cannot manage the log\'s contests', async ({ page, readerContext, adminContext }) => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set both development fixture identities')
  const description = `contests-visitor-${Date.now()}`
  // Dev Admin owns the log and Dev Reader visits it, because administrators may manage any learner's log.
  const id = await createPersonalLog(await adminContext.newPage(), description)

  await page.goto(`${webv3}/logs/${id}`)
  await expect(logDescription(page)).toHaveText(description)
  await expect(manageLink(page)).toHaveCount(0)
  await page.goto(`${webv3}/logs/${id}/contests`)
  await expect(page.getByRole('heading', { name: 'This is someone else\'s record' })).toBeVisible()
  await expect(page.getByRole('checkbox')).toHaveCount(0)
  await expect(saveSubmissions(page)).toHaveCount(0)

  // The API refuses an anonymous change outright.
  const anonymous = await readerContext.browser()!.newContext()
  try {
    const response = await anonymous.request.put(`${api}/immersion/logs/${id}/contest-registrations`, {
      data: { registration_ids: [] },
    })
    expect(response.status()).toBe(401)
  } finally {
    await anonymous.close()
  }
  expect(await submissions(adminContext, id)).toEqual([])
})
