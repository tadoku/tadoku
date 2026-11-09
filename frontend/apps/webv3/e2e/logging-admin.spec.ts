import type { BrowserContext, Page } from '@playwright/test'
import { admin, amountInput, api, branch, createLog, expect, hydrated, openForm, reader, sessionCookie, test, webv3 } from './logging-fixtures'

type Submission = { registration_id: string }

const saveChanges = (page: Page) => page.getByRole('button', { name: 'Save', exact: true })
const saveSubmissions = (page: Page) => page.getByRole('button', { name: 'Save submissions', exact: true })

/** Dev Reader saves a Japanese reading log and submits it through the API to every ongoing registration it fits. */
async function createSubmittedLog(page: Page, readerContext: BrowserContext, description: string) {
  await openForm(page)
  await page.getByLabel('Language', { exact: true }).selectOption('jpn')
  await page.getByLabel('Activity', { exact: true }).selectOption({ label: 'Reading' })
  await amountInput(page).fill('7')
  await page.getByLabel(/^Description/).fill(description)
  const id = await createLog(page)

  const ongoing = await readerContext.request.get(`${api}/immersion/contests/ongoing-registrations`)
  const { registrations } = await ongoing.json() as {
    registrations: { id: string; languages: { code: string }[]; contest: { allowed_activities: { name: string }[] } }[]
  }
  const fitting = registrations.filter(registration =>
    registration.languages.some(language => language.code === 'jpn')
    && registration.contest.allowed_activities.some(activity => activity.name === 'Reading'))
  expect(fitting.length, 'Dev Reader has two ongoing registrations for Japanese reading').toBeGreaterThanOrEqual(2)

  const submitted = await readerContext.request.put(`${api}/immersion/logs/${id}/contest-registrations`, {
    data: { registration_ids: fitting.map(registration => registration.id) },
  })
  expect(submitted.status()).toBe(200)
  return { id, registrationIds: fitting.map(registration => registration.id) }
}

/** The log as its owner reads it from the API in the branch tenant. */
async function ownerRecord(readerContext: BrowserContext, id: string) {
  const response = await readerContext.request.get(`${api}/immersion/logs/${id}`)
  expect(response.status()).toBe(200)
  expect(response.headers()['x-dev-selected']).toBe(branch)
  return await response.json() as { user_id: string; amount: number; description?: string; registrations?: Submission[] }
}

test.beforeEach(() => {
  test.skip(!webv3 || !api || !reader.email || !admin.email, 'Set development URLs and both fixture accounts')
})

test('an administrator edits another learner\'s log and changes one contest submission, keeping the others', async ({
  page,
  readerContext,
  adminContext,
}, testInfo) => {
  const { id, registrationIds } = await createSubmittedLog(page, readerContext, `admin-edit-${Date.now()}`)
  const before = await ownerRecord(readerContext, id)
  expect(before.registrations?.map(submission => submission.registration_id).sort()).toEqual([...registrationIds].sort())

  const moderator = await adminContext.newPage()
  await moderator.goto(`/logs/${id}`)
  await hydrated(moderator)
  await expect(moderator.getByRole('button', { name: 'Delete log', exact: true })).toBeEnabled()
  await expect(moderator.getByRole('link', { name: 'Edit log', exact: true })).toBeVisible()
  await moderator.screenshot({ path: testInfo.outputPath('admin-detail.png'), fullPage: true })

  await moderator.getByRole('link', { name: 'Edit log', exact: true }).click()
  await expect(moderator).toHaveURL(`${webv3}/logs/${id}/edit`)
  await expect(saveChanges(moderator)).toBeEnabled({ timeout: 15_000 })
  await amountInput(moderator).fill('9')
  await saveChanges(moderator).click()
  await expect(moderator).toHaveURL(`${webv3}/logs/${id}`)
  const edited = await ownerRecord(readerContext, id)
  expect(edited).toMatchObject({ user_id: before.user_id, amount: 9 })
  expect(edited.registrations).toHaveLength(registrationIds.length)

  await moderator.getByRole('link', { name: 'Manage contest submissions', exact: true }).click()
  await expect(moderator).toHaveURL(`${webv3}/logs/${id}/contests`)
  await expect(saveSubmissions(moderator)).toBeEnabled({ timeout: 15_000 })
  // The page offers the owner's registrations with the owner's submissions checked, not the administrator's own.
  const contests = moderator.getByRole('checkbox', { disabled: false })
  await expect(contests).toHaveCount(registrationIds.length)
  for (const contest of await contests.all()) await expect(contest).toBeChecked()
  // The summary speaks about the learner's log, not the administrator's.
  await expect(moderator.getByRole('region', { name: 'This log', exact: true })).toBeVisible()
  await moderator.screenshot({ path: testInfo.outputPath('admin-contests.png'), fullPage: true })

  await contests.last().uncheck()
  await saveSubmissions(moderator).click()
  await expect(moderator).toHaveURL(`${webv3}/logs/${id}`)
  const changed = await ownerRecord(readerContext, id)
  expect(changed.registrations).toHaveLength(registrationIds.length - 1)
  for (const submission of changed.registrations!) expect(registrationIds).toContain(submission.registration_id)

  await hydrated(moderator)
  await moderator.getByRole('button', { name: 'Delete log', exact: true }).click()
  await moderator.getByRole('dialog').getByRole('button', { name: 'Delete log', exact: true }).click()
  await expect(moderator).toHaveURL(`${webv3}/profile/${before.user_id}/statistics`)
  expect((await readerContext.request.get(`${api}/immersion/logs/${id}`)).status()).toBe(404)
})

test('an administrator\'s edit of another learner\'s log is refused after the session switches account', async ({
  page,
  readerContext,
  adminContext,
}) => {
  const description = `admin-switch-${Date.now()}`
  const { id } = await createSubmittedLog(page, readerContext, description)

  const moderator = await adminContext.newPage()
  await moderator.goto(`/logs/${id}/edit`)
  await expect(saveChanges(moderator)).toBeEnabled({ timeout: 15_000 })
  await moderator.getByLabel(/^Description/).fill(`${description}-changed`)

  // Another tab signs in as Dev Reader: the administrator's draft must not be saved under that session.
  await adminContext.addCookies([await sessionCookie(readerContext)])
  await saveChanges(moderator).click()
  await expect(moderator.getByRole('alert')).toContainText('different account. Nothing was saved')
  await expect(moderator).toHaveURL(`${webv3}/logs/${id}/edit`)
  expect((await ownerRecord(readerContext, id)).description).toBe(description)
})
