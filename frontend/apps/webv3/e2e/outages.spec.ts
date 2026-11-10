import { join } from 'node:path'
import type { Page } from '@playwright/test'
import { faultyWebv3, type Fault } from './fault-server'
import { api, expect, hydrated, reader, statValue, test, webv3 } from './logging-fixtures'
import { apiJson, score as scoreFormat } from './standings'

// One local server for the file: the outages replace API answers the deployed branch's server would read.
let server: Awaited<ReturnType<typeof faultyWebv3>>
const year = new Date().getUTCFullYear()
const breakdownFailed = 'Could not retrieve this breakdown.'

/** Applies `faults` alone, so a test starts from a healthy API and ends with whichever faults it left. */
const inject = (...faults: Fault[]) => server.faults.splice(0, server.faults.length, ...faults)

test.describe('partial outages', () => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  test.describe.configure({ timeout: 90_000 })

  test.beforeAll(async () => {
    test.setTimeout(150_000)
    server = await faultyWebv3(join(test.info().project.testDir, '..'), api!)
  })
  test.afterAll(async () => server?.close())
  test.beforeEach(() => { inject() })

  async function participant(page: Page) {
    const contest = await apiJson<{ id: string; title: string }>(page.request, '/immersion/contests/latest-official')
    const { entries: [entry] } = await apiJson<{ entries: { user_id: string; user_display_name: string; score: number }[] }>(
      page.request,
      `/immersion/contests/${contest.id}/leaderboard?page_size=1&page=0`,
    )
    return { contest, userId: entry.user_id, name: entry.user_display_name, score: entry.score }
  }

  test('a failed and a malformed profile section are marked while the rest of the year shows, until the API recovers', async ({ page }, testInfo) => {
    const { userId, name } = await participant(page)
    const split: Fault = { path: new RegExp(`^/immersion/users/${userId}/activity-split/${year}`), kind: 'status', status: 503 }
    const scores: Fault = { path: new RegExp(`^/immersion/users/${userId}/scores/${year}`), kind: 'malformed', body: '{"scores":"none"}' }
    inject(split, scores)

    const response = await page.goto(`${server.url}/profile/${userId}/statistics/${year}`)
    expect(response?.status()).toBe(200)
    expect(split.hits, 'the server-rendered page read the failing section').toBeGreaterThan(0)
    expect(scores.hits).toBeGreaterThan(0)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(name)
    await expect(page.getByRole('region', { name: 'Score by activity' })).toContainText(breakdownFailed)
    await expect(page.getByRole('region', { name: 'Score by language' })).toContainText(breakdownFailed)
    // The yearly activity still answers, so its totals and calendar show.
    await expect(page.getByRole('term').filter({ hasText: 'Logs' })).toBeVisible()
    await expect(statValue(page, 'Overall score')).toHaveText('Unavailable')
    await expect(page.getByRole('term').filter({ hasText: 'Languages' })).toHaveCount(0)
    await page.screenshot({ path: testInfo.outputPath('profile-sections-failed.png'), fullPage: true })

    inject()
    await page.reload()
    await expect(page.getByRole('region', { name: 'Score by activity' }).getByRole('listitem').first()).toBeVisible()
    await expect(page.getByRole('region', { name: 'Score by language' }).getByRole('listitem').first()).toBeVisible()
    await expect(page.getByText(breakdownFailed)).toHaveCount(0)
    await expect(statValue(page, 'Overall score')).not.toHaveText('Unavailable')
  })

  test('a section slower than the request deadline is marked and the page still answers in time', async ({ page }) => {
    const { userId, name } = await participant(page)
    const slow: Fault = { path: new RegExp(`^/immersion/users/${userId}/activity-split/${year}`), kind: 'slow', delayMs: 15_000 }
    inject(slow)

    const started = Date.now()
    const response = await page.goto(`${server.url}/profile/${userId}/statistics/${year}`)
    expect(response?.status()).toBe(200)
    // The server's deadline for a request's upstream calls is 8 seconds; the section's answer would take 15.
    expect(Date.now() - started).toBeLessThan(14_000)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(name)
    await expect(page.getByRole('region', { name: 'Score by activity' })).toContainText(breakdownFailed)
    await expect(page.getByRole('region', { name: 'Score by language' }).getByRole('listitem').first()).toBeVisible()
  })

  test('home keeps the board when the schedule cannot be read', async ({ page }, testInfo) => {
    const { contest, name, score } = await participant(page)
    const schedule: Fault = { path: /^\/immersion\/contests\?.*upcoming=true/, kind: 'reset' }
    inject(schedule)

    const response = await page.goto(`${server.url}/`)
    expect(response?.status()).toBe(200)
    expect(schedule.hits, 'the server-rendered page read the schedule').toBeGreaterThan(0)
    const board = page.getByRole('region', { name: new RegExp(contest.title) })
    // The board's table still lists the round's leader with the API's score.
    const leader = board.getByRole('table').getByRole('row').filter({ has: page.getByRole('link', { name, exact: true }) })
    await expect(leader.getByRole('cell').last()).toHaveText(scoreFormat.format(score))
    await expect(page.getByText('The board could not be loaded')).toHaveCount(0)
    // A running round covers the schedule, so an unreadable one is left out; otherwise its row says so.
    const next = page.getByRole('region', { name: 'Next official round' })
    if (await next.count()) await expect(next).toContainText('Temporarily unavailable')
    await page.screenshot({ path: testInfo.outputPath('home-schedule-failed.png'), fullPage: true })

    inject()
    await page.reload()
    const { contests: [upcoming] } = await apiJson<{ contests: { registration_end: string }[] }>(
      page.request,
      '/immersion/contests?upcoming=true&official=true&include_deleted=false&page_size=1',
    )
    if (upcoming) await expect(next).toContainText('Registration closes')
    else if (await next.count()) await expect(next).toContainText('Not announced yet')
    await expect(page.getByText('Temporarily unavailable')).toHaveCount(0)
  })

  test('a contest page whose main read fails or answers malformed is a 503 error page, and Try again recovers it', async ({ page }, testInfo) => {
    const { contest } = await participant(page)
    const main = new RegExp(`^/immersion/contests/${contest.id}(\\?|$)`)
    const errorPage = page.getByRole('heading', { name: 'This page couldn’t load' })

    const faults: Fault[] = [{ path: main, kind: 'reset' }, { path: main, kind: 'malformed', body: '{"id":42}' }]
    for (const fault of faults) {
      inject(fault)
      const response = await page.goto(`${server.url}/contests/${contest.id}/leaderboard/1`)
      expect(response?.status(), fault.kind).toBe(503)
      expect(fault.hits, fault.kind).toBeGreaterThan(0)
      await expect(errorPage, fault.kind).toBeVisible()
    }
    await page.screenshot({ path: testInfo.outputPath('contest-unavailable.png'), fullPage: true })

    inject()
    await hydrated(page)
    await page.getByRole('button', { name: 'Try again' }).click()
    await expect(page.getByRole('heading', { level: 1, name: contest.title })).toBeVisible()
    await expect(errorPage).toHaveCount(0)
  })
})
