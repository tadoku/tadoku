import { join } from 'node:path'
import { faultyWebv3, type Fault } from './fault-server'
import { api, expect, hydrated, reader, test, webv3 } from './logging-fixtures'
import { apiJson } from './standings'

// A new account can't log on earlier days, so a local server stands in its profile and an earlier year's activity and
// contests.
let server: Awaited<ReturnType<typeof faultyWebv3>>
const year = new Date().getUTCFullYear()

test.describe('getting going across years', () => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  test.describe.configure({ timeout: 90_000 })

  test.beforeAll(async () => {
    test.setTimeout(150_000)
    server = await faultyWebv3(join(test.info().project.testDir, '..'), api!)
  })
  test.afterAll(async () => server?.close())

  test('a learner who cleared Getting going in an earlier year sees their map, and the browser remembers it', async ({ learner }, testInfo) => {
    const { context, id, name } = learner
    const earlier = year - 3
    // Browsers keep a Secure cookie from plain-HTTP `localhost` as they would over HTTPS, but not from an IP address.
    const origin = server.url.replace('127.0.0.1', 'localhost')
    await context.addCookies((await context.cookies()).filter(cookie => cookie.name.startsWith('tadoku_session'))
      .map(({ name, value }) => ({ name, value, url: origin })))
    const page = await context.newPage()
    const home = async () => (await page.goto(`${origin}/`))!.text()
    const remembered = async () => (await context.cookies(origin)).find(cookie => cookie.name === `home_cleared_${id}`)

    // Nothing logged in any year: the server renders the checklist and remembers nothing.
    server.faults.splice(0)
    expect(await home()).toContain('>Getting going<')
    expect(await remembered()).toBeUndefined()

    // Three years ago, with nothing logged since. Each stand-in answer counts its own reads.
    const profile: Fault = {
      path: new RegExp(`^/immersion/users/${id}/profile(\\?|$)`),
      kind: 'json',
      body: { id, display_name: name, created_at: `${earlier}-03-01T12:00:00Z` },
    }
    const activityPath = new RegExp(`^/immersion/users/${id}/activity/${earlier}(\\?|$)`)
    const activity = (days: number[]): Fault => ({
      path: activityPath,
      kind: 'json',
      body: { total_updates: days.length, scores: days.map(day => ({ date: `${earlier}-03-${String(day).padStart(2, '0')}`, score: 4 })) },
    })
    // A real registration's shape, from a participant in the current round, moved to the learner and given dates.
    const latest = await apiJson<{ id: string }>(context.request, '/immersion/contests/latest-official')
    const { entries: [entry] } = await apiJson<{ entries: { user_id: string }[] }>(
      context.request,
      `/immersion/contests/${latest.id}/leaderboard?page_size=1&page=0`,
    )
    const { registrations: [template] } = await apiJson<{ registrations: { contest: object }[] }>(
      context.request,
      `/immersion/users/${entry.user_id}/contest-registrations/${year}?page_size=1&page=0`,
    )
    const joined = (...periods: [start: string, end: string][]): Fault => ({
      path: new RegExp(`^/immersion/users/${id}/contest-registrations/${earlier}(\\?|$)`),
      kind: 'json',
      body: {
        registrations: periods.map(([start, end]) => ({
          ...template,
          user_id: id,
          user_display_name: name,
          contest: { ...template.contest, contest_start: start, contest_end: end },
        })),
        total_size: periods.length,
        next_page_token: '',
      },
    })
    const stand = (...faults: Fault[]) => server.faults.splice(0, server.faults.length, ...faults)

    // A failed read of the earlier year clears nothing and remembers nothing.
    const failed: Fault = { path: activityPath, kind: 'status', status: 500 }
    stand(profile, failed)
    expect(await home()).toContain('>Getting going<')
    expect(failed.hits).toBeGreaterThan(0)
    expect(await remembered()).toBeUndefined()

    // Two active days in a row count only within the dates of a contest the learner joined.
    for (const contests of [
      [[`${earlier}-04-01`, `${earlier}-04-30`]],
      // Each day in a different contest.
      [[`${earlier}-03-01`, `${earlier}-03-03`], [`${earlier}-03-04`, `${earlier}-03-31`]],
    ] as [string, string][][]) {
      stand(profile, activity([3, 4]), joined(...contests))
      expect(await home()).toContain('>Getting going<')
      expect(await remembered()).toBeUndefined()
    }

    const history = activity([3, 4])
    stand(profile, history, joined([`${earlier}-03-01`, `${earlier}-03-31`]))
    expect(await home()).toContain('>Past 30 days<')
    expect(history.hits, 'the earlier year was read once').toBe(1)
    const cookie = await remembered()
    expect(cookie).toMatchObject({ value: '1', sameSite: 'Lax', secure: true, httpOnly: true, path: '/' })
    expect(cookie!.expires * 1000 - Date.now(), 'a long-lived cookie').toBeGreaterThan(300 * 86_400_000)

    // Later loads render the map from the cookie without reading the history again, nor once the page has hydrated.
    expect(await home()).toContain('>Past 30 days<')
    await hydrated(page)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Past 30 days')
    expect(history.hits, 'no further history read').toBe(1)
    await page.screenshot({ path: testInfo.outputPath('getting-going-earlier-year.png'), fullPage: true })

    // Without the cookie, the server reads the history again: five active days within 30 clear it without a contest.
    await context.clearCookies({ name: `home_cleared_${id}` })
    const fiveDays = activity([3, 5, 9, 12, 20])
    stand(profile, fiveDays)
    expect(await home()).toContain('>Past 30 days<')
    expect(fiveDays.hits).toBe(1)
    expect(await remembered()).toBeDefined()
  })
})
