import { randomUUID } from 'node:crypto'
import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { accountId, admin, api, expect, reader, test, webv3 } from './logging-fixtures'
import { apiJson, score } from './standings'

interface Contest { id: string; title: string }
interface Log { id: string; description?: string; tags: string[]; language: { name: string }; activity: { name: string }; user_display_name?: string }
interface Logs { logs: Log[]; total_size: number }
interface Profile { scores: { language_code: string; score: number }[]; registration: { user_display_name: string; languages: { code: string; name: string }[] } }

const day = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
const updatesPath = (id: string, page = 1) => `/contests/${id}/updates/${page}`
const profilePath = (id: string, userId: string, page = 1) => `/contests/${id}/profile/${userId}/${page}`
const fitsViewport = (page: Page) => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)
const latest = (request: APIRequestContext) => apiJson<Contest>(request, '/immersion/contests/latest-official')

/** One page of a contest's logs, counted from 1, as the API lists them. */
function contestLogs(request: APIRequestContext, contestId: string, page = 1, userId?: string) {
  const query = new URLSearchParams({ page_size: '50', page: String(page - 1), include_deleted: 'false', ...(userId ? { user_id: userId } : {}) })
  return apiJson<Logs>(request, `/immersion/contests/${contestId}/logs?${query}`)
}

const entries = (page: Page) => page.locator('.profile-entries > li')
const removeButtons = (page: Page) => page.getByRole('button', { name: 'Remove from contest' })

/** The page lists exactly the API's logs in order, each linking to its native detail page and naming its learner. */
async function expectLogs(page: Page, expected: Logs, participants: boolean) {
  const titles = page.locator('.profile-entries > li a.profile-entry-title')
  await expect(titles).toHaveCount(expected.logs.length)
  expect(await titles.evaluateAll(links => links.map(link => link.getAttribute('href'))))
    .toEqual(expected.logs.map(log => `/logs/${log.id}`))
  if (participants) {
    await expect(page.locator('.profile-entry-meta')).toHaveText(expected.logs.map(log => new RegExp(`^${escape(log.user_display_name ?? 'Unknown user')} · `)))
  }
  // The range summary appears only beside page links.
  if (expected.total_size > 50) {
    await expect(page.locator('.profile-entries-footer > p')).toHaveText(new RegExp(`of ${score.format(expected.total_size)} updates$`))
  } else {
    await expect(page.locator('.profile-entries-footer')).toHaveCount(0)
  }
}

const escape = (text: string) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

/** The page's overall score and each registered language's score agree with the API's contest profile. */
async function expectScores(page: Page, profile: Profile) {
  const stats = page.locator('.profile-stats')
  const total = profile.scores.reduce((sum, item) => sum + item.score, 0)
  const values = [score.format(total), ...profile.registration.languages.map(language =>
    score.format(profile.scores.find(item => item.language_code === language.code)?.score ?? 0))]
  await expect(stats.locator('.paper-stat-grid__value')).toHaveText(values)
}

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('the server-rendered updates feed lists the API\'s contest logs, pages through them and opens a log', async ({ browser, request }, testInfo) => {
  const contest = await latest(request)
  const first = await contestLogs(request, contest.id)
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()

  const response = await page.goto(updatesPath(contest.id))
  expect(response?.status()).toBe(200)
  await expect(page).toHaveTitle(`Updates · ${contest.title} · Tadoku`)
  await expect(page.getByRole('heading', { level: 1, name: contest.title })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Updates', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(page.getByRole('link', { name: 'Standings', exact: true })).toHaveAttribute('href', `/contests/${contest.id}/leaderboard/1`)
  await expectLogs(page, first, true)
  await expect(removeButtons(page)).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('updates-ssr.png'), fullPage: true })
  await context.close()

  const live = await browser.newPage()
  await live.goto(`/contests/${contest.id}/leaderboard/1`)
  await live.getByRole('link', { name: 'Updates', exact: true }).click()
  await expect(live).toHaveURL(new RegExp(`${updatesPath(contest.id)}$`))
  if (first.total_size > 50) {
    await live.getByRole('navigation', { name: 'Update pages', exact: true }).getByRole('link', { name: 'Next page' }).click()
    await expect(live).toHaveURL(new RegExp(`${updatesPath(contest.id, 2)}$`))
    await expectLogs(live, await contestLogs(request, contest.id, 2), true)
  }
  if (first.total_size > 0) {
    await live.goto(updatesPath(contest.id, Math.ceil(first.total_size / 50) + 1))
    await expect(live.getByRole('heading', { name: 'This page is past the end of the updates' })).toBeVisible()
    await live.goto(updatesPath(contest.id))
    const entry = first.logs[0]
    await live.locator(`a.profile-entry-title[href="/logs/${entry.id}"]`).click()
    await expect(live).toHaveURL(new RegExp(`/logs/${entry.id}$`))
    await expect(live.getByRole('heading', { level: 1 })).toBeVisible()
    await expect(live.getByRole('link', { name: 'Edit log' })).toHaveCount(0)
  }
  await live.close()
})

test('a standings participant opens their contest profile, with the API\'s scores and logs', async ({ page, request }, testInfo) => {
  const contest = await latest(request)
  const standings = await apiJson<{ entries: { user_id: string; user_display_name: string }[] }>(
    request,
    `/immersion/contests/${contest.id}/leaderboard?page_size=50&page=0`,
  )
  test.skip(standings.entries.length === 0, 'The latest official contest has no participants')
  const participant = standings.entries[0]

  await page.goto(`/contests/${contest.id}/leaderboard/1`)
  await page.locator('.standings-table').getByRole('link', { name: participant.user_display_name, exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`${profilePath(contest.id, participant.user_id)}$`))
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(participant.user_display_name)
  await expect(page).toHaveTitle(`${participant.user_display_name} · ${contest.title} · Tadoku`)
  await expect(page.getByRole('link', { name: 'View profile' })).toHaveAttribute('href', `/profile/${participant.user_id}/statistics`)
  await expectScores(page, await apiJson<Profile>(request, `/immersion/contests/${contest.id}/profile/${participant.user_id}/scores`))
  const logs = await contestLogs(request, contest.id, 1, participant.user_id)
  await expectLogs(page, logs, false)
  await page.screenshot({ path: testInfo.outputPath('contest-profile.png'), fullPage: true })

  if (logs.total_size > 50) {
    await page.getByRole('navigation', { name: 'Update pages', exact: true }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(new RegExp(`${profilePath(contest.id, participant.user_id, 2)}$`))
    await expectLogs(page, await contestLogs(request, contest.id, 2, participant.user_id), false)
  }
})

test('unknown and malformed contests and participants are not found', async ({ page, request }) => {
  const contest = await latest(request)
  for (const path of [updatesPath(randomUUID()), '/contests/not-a-contest/updates/1']) {
    expect((await page.goto(path))?.status(), path).toBe(404)
    await expect(page.getByRole('heading', { level: 1, name: 'Contest not found' })).toBeVisible()
  }
  for (const path of [profilePath(contest.id, randomUUID()), profilePath(randomUUID(), randomUUID()), `/contests/${contest.id}/profile/not-a-user/1`]) {
    expect((await page.goto(path))?.status(), path).toBe(404)
    await expect(page.getByRole('heading', { level: 1, name: 'Participant not found' })).toBeVisible()
  }
})

test('the updates feed and a contest profile fit at 390px', async ({ page, request }, testInfo) => {
  const contest = await latest(request)
  const standings = await apiJson<{ entries: { user_id: string }[] }>(request, `/immersion/contests/${contest.id}/leaderboard?page_size=1&page=0`)
  const paths = [['updates', updatesPath(contest.id)]]
  if (standings.entries[0]) paths.push(['profile', profilePath(contest.id, standings.entries[0].user_id)])

  await page.setViewportSize({ width: 390, height: 900 })
  for (const [name, path] of paths) {
    await page.goto(path)
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
    await page.evaluate(() => document.fonts.ready)
    expect(await fitsViewport(page), `${path} at 390px`).toBe(true)
    await page.screenshot({ path: testInfo.outputPath(`${name}-390.png`), fullPage: true })
  }
})

/** A contest created through the API as the context's account, open today. */
async function createContest(context: BrowserContext): Promise<Contest> {
  const title = `Webv3 moderation check ${randomUUID().slice(0, 8)}`
  const response = await context.request.post(`${api}/immersion/contests`, {
    data: {
      title,
      description: 'Created by the Webv3 contest moderation journey.',
      contest_start: day(),
      contest_end: day(30),
      registration_end: day(14),
      private: false,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return { ...(await response.json()), title }
}

/** Registers the context's account for the contest and submits one log to it through the API. */
async function submitLog(context: BrowserContext, contestId: string, description: string) {
  const joined = await context.request.post(`${api}/immersion/contests/${contestId}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.status()).toBe(200)
  const registration = await apiJson<{ id: string }>(context.request, `/immersion/contests/${contestId}/registration`)
  const created = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 2, duration_seconds: 600, tags: [], description, registration_ids: [registration.id] },
  })
  expect(created.status()).toBe(200)
  return (await created.json()) as Log & { score: number }
}

test('an administrator removes a log from a contest with a reason; a participant sees no such action', async ({ readerContext, adminContext }, testInfo) => {
  test.skip(!reader.email || !admin.email, 'Set both development fixture identities')
  const contest = await createContest(adminContext)
  const readerId = await accountId(readerContext)
  const log = await submitLog(readerContext, contest.id, `webv3-moderation-${randomUUID().slice(0, 8)}`)

  // The participant's contest profile shows the submitted log's score.
  const own = await readerContext.newPage()
  await own.goto(profilePath(contest.id, readerId))
  await expectScores(own, await apiJson<Profile>(readerContext.request, `/immersion/contests/${contest.id}/profile/${readerId}/scores`))
  await expect(own.locator('.profile-stats')).toContainText(score.format(log.score))
  await expectLogs(own, await contestLogs(readerContext.request, contest.id, 1, readerId), false)
  await expect(removeButtons(own)).toHaveCount(0)
  await own.goto(updatesPath(contest.id))
  await expectLogs(own, await contestLogs(readerContext.request, contest.id), true)
  await expect(removeButtons(own)).toHaveCount(0)
  await own.close()

  const page = await adminContext.newPage()
  await page.goto(updatesPath(contest.id))
  await expect(entries(page)).toHaveCount(1)
  await expect(removeButtons(page)).toHaveCount(1)
  await removeButtons(page).click()
  const dialog = page.getByRole('dialog', { name: 'Remove log from contest' })
  await dialog.getByRole('button', { name: 'Yes, remove it' }).click()
  await expect(dialog.getByText('Give a reason for removing this log.')).toBeVisible()
  await dialog.getByRole('button', { name: 'Cancel' }).click()
  await expect(dialog).toHaveCount(0)
  expect((await contestLogs(adminContext.request, contest.id)).total_size).toBe(1)

  await removeButtons(page).click()
  await dialog.getByLabel('Reason').fill('Duplicate entry from the Webv3 moderation journey')
  await expect(dialog).toBeVisible()
  await page.waitForFunction(() => document.getAnimations().every(animation => animation.playState !== 'running'))
  await page.screenshot({ path: testInfo.outputPath('moderation-dialog.png') })
  await dialog.getByRole('button', { name: 'Yes, remove it' }).click()
  await expect(page.getByText('Log removed from contest.')).toBeVisible()
  await expect(dialog).toHaveCount(0)
  await expect(page.getByRole('heading', { name: 'No updates submitted yet' })).toBeVisible()
  expect((await contestLogs(adminContext.request, contest.id)).total_size).toBe(0)
  await page.screenshot({ path: testInfo.outputPath('moderation-removed.png'), fullPage: true })

  // The log itself remains in the learner's personal record.
  expect((await readerContext.request.get(`${api}/immersion/logs/${log.id}`)).status()).toBe(200)
  await page.goto(profilePath(contest.id, readerId))
  await expect(page.getByRole('heading', { name: 'No updates submitted yet' })).toBeVisible()
  await page.close()
})

test('a community contest\'s organizer can remove a log without being an administrator', async ({ readerContext, adminContext }) => {
  test.skip(!reader.email || !admin.email, 'Set both development fixture identities')
  const contest = await createContest(readerContext)
  const log = await submitLog(adminContext, contest.id, `webv3-organizer-${randomUUID().slice(0, 8)}`)

  const page = await readerContext.newPage()
  await page.goto(updatesPath(contest.id))
  await expect(page.locator(`a.profile-entry-title[href="/logs/${log.id}"]`)).toBeVisible()
  await removeButtons(page).click()
  const dialog = page.getByRole('dialog', { name: 'Remove log from contest' })
  await dialog.getByLabel('Reason').fill('Organizer moderation check')
  await dialog.getByRole('button', { name: 'Yes, remove it' }).click()
  await expect(page.getByText('Log removed from contest.')).toBeVisible()
  expect((await contestLogs(readerContext.request, contest.id)).total_size).toBe(0)
  await page.close()
})
