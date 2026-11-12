import { randomUUID } from 'node:crypto'
import type { APIRequestContext, Page } from '@playwright/test'
import {
  contestLogs,
  createContest,
  createLog,
  expectBreakdowns,
  expectSummary,
  filterQuery,
  fitsViewport,
  profilePath,
  profileSummary,
  register,
  type Contest,
  type Filters,
  type NewLog,
} from './contest-profile'
import { accountId, activityEntries, admin, api, expect, hydrated, logLinks, reader, statValue, test as base, webv3 } from './logging-fixtures'

interface Seeded { contest: Contest; userId: string; total: number }

/**
 * Dev Reader's profile in a contest Dev Admin creates once per worker: two languages, both activities, overlapping tags
 * and enough listening logs for a second page, unfiltered and filtered by listening.
 */
const seeded: NewLog[] = [
  { language: 'jpn', activity: 2, minutes: 50, tags: ['audiobook'] },
  { language: 'jpn', activity: 2, minutes: 20, tags: ['podcast'] },
  { language: 'jpn', activity: 1, pages: 30, tags: ['manga'] },
  { language: 'kor', activity: 1, pages: 10, tags: ['audiobook', 'manga'] },
  ...Array.from({ length: 26 }, (): NewLog => ({ language: 'jpn', activity: 2, minutes: 1, tags: [] })),
]

const test = base.extend<{ profile: Seeded }, { seededProfile: Seeded | undefined }>({
  seededProfile: [async ({ browser, readerCookies, adminCookies }, provide) => {
    if (!webv3 || !api || !readerCookies.length || !adminCookies.length) return provide(undefined)
    const [readerContext, adminContext] = await Promise.all([browser.newContext({ baseURL: webv3 }), browser.newContext({ baseURL: webv3 })])
    await readerContext.addCookies(readerCookies)
    await adminContext.addCookies(adminCookies)
    const contest = await createContest(adminContext, 30, `Webv3 profile filters ${randomUUID().slice(0, 8)}`)
    const registration = await register(readerContext, contest.id, ['jpn', 'kor'])
    for (const [index, log] of seeded.entries()) await createLog(readerContext, registration, log, `webv3-filters-${index}`)
    const userId = await accountId(readerContext)
    await Promise.all([readerContext.close(), adminContext.close()])
    await provide({ contest, userId, total: seeded.length })
  }, { scope: 'worker' }],
  profile: async ({ seededProfile }, provide) => {
    base.skip(!reader.email || !admin.email || !seededProfile, 'Set WEBV3_BASE_URL, TADOKU_API_URL and both development fixture identities')
    await provide(seededProfile!)
  },
})

const path = ({ contest, userId }: Seeded, page = 1, filters?: Filters) => {
  const query = filterQuery(filters).toString()
  return `${profilePath(contest.id, userId, page)}${query ? `?${query}` : ''}`
}
const filtersButton = (page: Page) => page.getByRole('button', { name: /^Filters/ }).filter({ visible: true })
const logCount = (page: Page) => page.getByRole('group', { name: 'Applied filters' }).getByText(/ logs? of \d+$/)

/** Ticks values in the desktop Filters menu, which stays open between choices, and closes it. */
async function choose(page: Page, values: string[]) {
  await filtersButton(page).click()
  for (const value of values) await page.getByRole('menuitemcheckbox', { name: new RegExp(`^${value}`) }).click()
  await page.keyboard.press('Escape')
}

/**
 * Every figure on the page follows the filters and equals the API's filtered summary and log list: the summary, the
 * applied filters' count, the listed logs and the activity and tag lists. The summary's totals are the log list's sums.
 */
async function expectFiltered(page: Page, request: APIRequestContext, profile: Seeded, filters: Filters) {
  const [summary, logs] = await Promise.all([
    profileSummary(request, profile.contest.id, profile.userId, filters),
    contestLogs(request, profile.contest.id, 1, profile.userId, filters),
  ])
  expect(summary.log_count).toBe(logs.total_size)
  if (logs.total_size <= 25) expect(summary.score).toBeCloseTo(logs.logs.reduce((sum, log) => sum + log.score, 0), 1)
  await expectSummary(page, summary)
  await expect(statValue(page, 'Score')).toContainText(' of ')
  await expect(logCount(page)).toHaveText(`${logs.total_size} ${logs.total_size === 1 ? 'log' : 'logs'} of ${profile.total}`)
  await expect.poll(() => logLinks(page).evaluateAll(links => links.map(link => link.getAttribute('href'))))
    .toEqual(logs.logs.map(log => `/logs/${log.id}`))
  await expectBreakdowns(page, summary)
  return summary
}

/** The unfiltered page: the API's whole summary, its place in the Score figure and no applied filters. */
async function expectUnfiltered(page: Page, request: APIRequestContext, profile: Seeded) {
  await expectSummary(page, await profileSummary(request, profile.contest.id, profile.userId))
  await expect(statValue(page, 'Score')).toContainText(' place')
  await expect(page.getByRole('group', { name: 'Applied filters' })).toHaveCount(0)
  await expect(activityEntries(page)).toHaveCount(25)
}

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('choosing an activity in Filters narrows the address, summary, logs and lists to the API\'s filtered answers', async ({ page, request, profile }, testInfo) => {
  await page.goto(path(profile))
  await expectUnfiltered(page, request, profile)
  await expect(page.getByRole('region', { name: 'Score by language' })).toBeVisible()
  await hydrated(page)

  await choose(page, ['Reading'])
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId)}\\?activity=reading$`))
  const summary = await expectFiltered(page, request, profile, { activity: [1] })
  // The activity list ignores its own filter, so listening stays listed beside reading.
  expect(summary.activities.map(activity => activity.id).sort()).toEqual([1, 2])
  await page.screenshot({ path: testInfo.outputPath('filtered-activity.png'), fullPage: true })
})

test('two tags match either, and an activity narrows them further', async ({ page, request, profile }) => {
  await page.goto(path(profile))
  await expectUnfiltered(page, request, profile)
  await hydrated(page)

  await choose(page, ['audiobook', 'manga'])
  await expect(page).toHaveURL(/\?tag=audiobook&tag=manga$/)
  const either = await expectFiltered(page, request, profile, { tag: ['audiobook', 'manga'] })
  expect(either.log_count).toBe(3)

  await choose(page, ['Reading'])
  await expect(page).toHaveURL(/\?activity=reading&tag=audiobook&tag=manga$/)
  const both = await expectFiltered(page, request, profile, { activity: [1], tag: ['audiobook', 'manga'] })
  expect(both.log_count).toBe(2)
})

test('removing the last chip and Clear all restore the unfiltered page and return focus to Filters', async ({ page, request, profile }) => {
  await page.goto(path(profile))
  await expectUnfiltered(page, request, profile)
  await hydrated(page)

  await choose(page, ['Listening'])
  await expectFiltered(page, request, profile, { activity: [2] })
  await page.getByRole('button', { name: 'Remove Listening' }).click()
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId)}$`))
  await expectUnfiltered(page, request, profile)
  await expect(filtersButton(page)).toBeFocused()

  await choose(page, ['Listening', 'podcast'])
  await expectFiltered(page, request, profile, { activity: [2], tag: ['podcast'] })
  await page.getByRole('group', { name: 'Applied filters' }).getByRole('button', { name: 'Clear all' }).click()
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId)}$`))
  await expectUnfiltered(page, request, profile)
  await expect(filtersButton(page)).toBeFocused()
})

test('a filtered address opens the filtered page, and nothing matching offers Clear filters', async ({ page, request, profile }) => {
  const filters = { language: ['kor'], tag: ['manga'] }
  const response = await page.goto(path(profile, 1, filters))
  expect(response?.status()).toBe(200)
  await expectFiltered(page, request, profile, filters)
  await expect(page.getByRole('button', { name: 'Remove Korean' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Remove manga' })).toBeVisible()
  await hydrated(page)
  await filtersButton(page).click()
  await expect(page.getByRole('menuitemcheckbox', { name: /^Korean/ })).toBeChecked()
  await expect(page.getByRole('menuitemcheckbox', { name: /^manga/ })).toBeChecked()
  await page.keyboard.press('Escape')

  await page.goto(path(profile, 1, { language: ['kor'], tag: ['podcast'] }))
  await expect(page.getByRole('heading', { name: 'No logs match these filters' })).toBeVisible()
  await page.getByRole('link', { name: 'Clear filters' }).click()
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId)}$`))
  await expectUnfiltered(page, request, profile)
})

test('a filter change on page 2 returns to page 1, and page links keep the filters', async ({ page, request, profile }) => {
  await page.goto(path(profile, 2))
  await expect(statValue(page, 'Logs')).toHaveText(String(profile.total))
  await expect(activityEntries(page)).toHaveCount(profile.total - 25)
  await hydrated(page)

  await choose(page, ['Listening'])
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId, 1)}\\?activity=listening$`))
  await expectFiltered(page, request, profile, { activity: [2] })
  const next = page.getByRole('navigation', { name: 'Update pages', exact: true }).getByRole('link', { name: 'Next page' })
  await expect(next).toHaveAttribute('href', `${profilePath(profile.contest.id, profile.userId, 2)}?activity=listening`)
  await next.click()
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId, 2)}\\?activity=listening$`))
  const second = await contestLogs(request, profile.contest.id, 2, profile.userId, { activity: [2] })
  await expect.poll(() => logLinks(page).evaluateAll(links => links.map(link => link.getAttribute('href'))))
    .toEqual(second.logs.map(log => `/logs/${log.id}`))
})

test('invalid filter values in the address are dropped before the API reads them', async ({ page, request, profile }) => {
  const long = 'x'.repeat(51)
  const query = `activity=flying&activity=reading&tag=%20&tag=${long}&tag=%20manga%20&language=far-too-long-code`
  const response = await page.goto(`${profilePath(profile.contest.id, profile.userId)}?${query}`)
  expect(response?.status()).toBe(200)
  await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId)}\\?activity=reading&tag=manga$`))
  await expectFiltered(page, request, profile, { activity: [1], tag: ['manga'] })
})

test.describe('without JavaScript', () => {
  test.use({ javaScriptEnabled: false })

  test('the Filters form submits and filters the page', async ({ page, request, profile }, testInfo) => {
    await page.goto(path(profile, 2))
    await expect(statValue(page, 'Logs')).toHaveText(String(profile.total))
    await page.getByText('Filters', { exact: true }).click()
    const form = page.getByRole('form', { name: 'Filters' })
    await form.getByRole('group', { name: 'Activity' }).getByRole('checkbox', { name: /^Reading/ }).check()
    await form.getByRole('group', { name: 'Tag' }).getByRole('checkbox', { name: /^manga/ }).check()
    await page.screenshot({ path: testInfo.outputPath('no-js-form.png'), fullPage: true })
    await form.getByRole('button', { name: 'Apply' }).click()
    await expect(page).toHaveURL(new RegExp(`${profilePath(profile.contest.id, profile.userId, 1)}\\?activity=reading&tag=manga$`))
    await expectFiltered(page, request, profile, { activity: [1], tag: ['manga'] })
  })
})

test('at 390px Filters opens a full-screen sheet whose button names the result, and the page fits', async ({ page, request, profile }, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto(path(profile))
  await expectUnfiltered(page, request, profile)
  await hydrated(page)
  await page.evaluate(() => document.fonts.ready)
  expect(await fitsViewport(page), 'the unfiltered page at 390px').toBe(true)
  const lists = page.getByText('Score by language, activity and tag')
  await expect(lists).toBeVisible()
  await expect(page.getByRole('region', { name: 'Score by tag' })).toBeHidden()
  await lists.click()
  await expect(page.getByRole('region', { name: 'Score by tag' })).toBeVisible()

  await filtersButton(page).click()
  const sheet = page.getByRole('dialog', { name: 'Filters' })
  await expect(sheet).toBeVisible()
  await page.waitForFunction(() => document.getAnimations().every(animation => animation.playState !== 'running'))
  expect(await sheet.boundingBox()).toMatchObject({ x: 0, y: 0, width: 390, height: 844 })
  await sheet.getByRole('checkbox', { name: /^Reading/ }).check()
  const { log_count: count } = await profileSummary(request, profile.contest.id, profile.userId, { activity: [1] })
  await page.screenshot({ path: testInfo.outputPath('sheet-390.png') })
  await sheet.getByRole('button', { name: `Show ${count} logs` }).click()
  await expect(sheet).toHaveCount(0)
  await expect(page).toHaveURL(/\?activity=reading$/)
  await expect(statValue(page, 'Logs')).toHaveText(String(count))
  expect(await fitsViewport(page), 'the filtered page at 390px').toBe(true)
  await page.screenshot({ path: testInfo.outputPath('filtered-390.png'), fullPage: true })
})

