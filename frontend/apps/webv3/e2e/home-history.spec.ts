import { join } from 'node:path'
import { faultyWebv3, type Fault } from './fault-server'
import { api, expect, hydrated, reader, test, webv3 } from './logging-fixtures'

// A new account can't log on earlier days, so a local server stands in the profile and an earlier year's activity.
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

    // Five active days within 30, three years ago, with nothing logged since.
    const profile: Fault = {
      path: new RegExp(`^/immersion/users/${id}/profile(\\?|$)`),
      kind: 'json',
      body: { id, display_name: name, created_at: `${earlier}-03-01T12:00:00Z` },
    }
    const history: Fault = {
      path: new RegExp(`^/immersion/users/${id}/activity/${earlier}(\\?|$)`),
      kind: 'json',
      body: { total_updates: 5, scores: [3, 5, 9, 12, 20].map(day => ({ date: `${earlier}-03-${String(day).padStart(2, '0')}`, score: 4 })) },
    }
    server.faults.push(profile, history)
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

    // Without the cookie, the server reads the history again and sets it.
    await context.clearCookies({ name: `home_cleared_${id}` })
    expect(await home()).toContain('>Past 30 days<')
    expect(history.hits).toBe(2)
  })
})
