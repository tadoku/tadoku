import type { Page } from '@playwright/test'
import { admin, api, expect, joinedContest, reader, statValue, test, webv3 } from './logging-fixtures'
import { apiJson, score } from './standings'

const points = (value: number) => `${score.format(value)} ${score.format(value) === '1.0' ? 'point' : 'points'}`
const days = (value: number) => `${value} ${value === 1 ? 'day' : 'days'}`
const dayLabel = (date: string) => new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeZone: 'UTC' }).format(new Date(`${date}T00:00:00Z`))
const step = (page: Page, name: string) => page.getByRole('main').getByRole('listitem').filter({ hasText: name })

test('a signed-in learner’s home shows their past 30 days, their board row and newest logs as the API reports them', async ({ page, readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const { request } = readerContext
  const year = new Date().getUTCFullYear()
  const today = new Date().toISOString().slice(0, 10)
  const { account: { id: userId } } = await apiJson<{ account: { id: string } }>(request, '/identity/session')

  // A log in the current round today, so the learner has a place, today's points and a newest log.
  const latest = await apiJson<{ id: string; title: string }>(request, '/immersion/contests/latest-official')
  let registration = await request.get(`${api}/immersion/contests/${latest.id}/registration`)
  if (registration.status() === 404) {
    const joined = await request.post(`${api}/immersion/contests/${latest.id}/registration`, { data: { language_codes: ['jpn'] } })
    expect(joined.status(), 'registration').toBe(200)
    registration = await request.get(`${api}/immersion/contests/${latest.id}/registration`)
  }
  const { id: registrationId } = await registration.json() as { id: string }
  const logged = await request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount: 2, unit_key: 'pages', tags: [], description: `home-values-${Date.now()}`, registration_ids: [registrationId] },
  })
  expect(logged.ok(), `log answered ${logged.status()}`).toBeTruthy()

  const [entry, activity, recent] = await Promise.all([
    apiJson<{ rank: number; score: number }>(request, `/immersion/contests/${latest.id}/leaderboard/${userId}`),
    apiJson<{ scores: { date: string; score: number }[] }>(request, `/immersion/users/${userId}/activity/${year}`),
    apiJson<{ logs: { id: string }[] }>(request, `/immersion/users/${userId}/logs?page_size=5&page=0&include_deleted=false`),
  ])
  // The 30 UTC days ending today. A streak counts back from today, or from yesterday while today is still empty.
  const byDate = new Map(activity.scores.map(day => [day.date, day.score]))
  const daysAgo = (count: number) => new Date(Date.parse(`${today}T00:00:00Z`) - count * 86_400_000).toISOString().slice(0, 10)
  const window = Array.from({ length: 30 }, (_, index) => byDate.get(daysAgo(29 - index)) ?? 0)
  const total = window.reduce((sum, each) => sum + each, 0)
  const active = window.filter(each => each > 0).length
  let streak = 0
  for (let back = byDate.get(today) ? 0 : 1; (byDate.get(daysAgo(back)) ?? 0) > 0; back++) streak++

  // The values are rendered on the server, before any script runs.
  const html = await (await request.get(`${webv3}/`)).text()
  expect(html).toContain(`>${score.format(total)}<`)

  // Dev Reader is in a contest and has logged on two days in a row, so they see their activity map.
  const loggedDays = [...byDate].filter(([, value]) => value > 0).map(([date]) => Date.parse(date)).sort()
  expect(loggedDays.some((date, index) => index > 0 && date - loggedDays[index - 1] === 86_400_000), 'seeded consecutive days').toBe(true)
  await page.goto('/')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Past 30 days')
  // The unit is its own element beside the number.
  await expect(page.getByRole('main').getByText(points(total).replace(' ', ''), { exact: true })).toBeVisible()
  await expect(statValue(page, 'Today')).toHaveText(points(byDate.get(today) ?? 0).replace(' ', ''))
  await expect(statValue(page, 'Streak')).toHaveText(days(streak).replace(' ', ''))
  await expect(statValue(page, 'Active days')).toHaveText(`${active}of 30`)
  const map = page.getByRole('group', { name: 'Daily activity for the past 13 weeks', exact: true })
  await expect(map.locator(`a[href$="${today}"]`)).toHaveAccessibleName(`${points(byDate.get(today)!)} on ${dayLabel(today)}`)
  // The map replaces the year calendar, which the profile keeps.
  await expect(page.getByRole('heading', { name: 'Daily activity', exact: true })).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Your profile', exact: true })).toHaveAttribute('href', `/profile/${userId}/statistics`)

  // The board, not the column beside it, carries the learner's place.
  const you = page.getByRole('region', { name: latest.title, exact: true }).locator('tr[aria-current="true"]')
  await expect(you.getByRole('cell').first()).toHaveText(String(entry.rank))
  await expect(you.getByRole('cell').last()).toHaveText(score.format(entry.score))

  const newest = page.getByRole('region', { name: 'Recent activity', exact: true }).getByRole('listitem').getByRole('link')
    .and(page.locator('[href^="/logs/"]'))
  await expect(newest).toHaveCount(recent.logs.length)
  for (const [index, log] of recent.logs.entries()) await expect(newest.nth(index)).toHaveAttribute('href', `/logs/${log.id}`)
  await page.screenshot({ path: testInfo.outputPath('home-values.png'), fullPage: true })
})

test('a new learner keeps the getting-going checklist until every step is done', async ({ learner, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !admin.email, 'Set development URLs and the Dev Admin fixture account')
  const { context } = learner
  const page = await context.newPage()

  await page.goto('/')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Getting going')
  for (const name of ['Log your first activity', 'Log something today', 'Join a contest', 'Log two days in a row']) {
    await expect(step(page, name)).toContainText('Not done')
  }
  await page.screenshot({ path: testInfo.outputPath('getting-going-new.png'), fullPage: true })

  // One log ticks two steps; two of four keeps the checklist.
  const logged = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount: 1, unit_key: 'pages', tags: [], description: 'Getting going', registration_ids: [] },
  })
  expect(logged.ok(), `log answered ${logged.status()}`).toBeTruthy()
  await page.reload()
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Getting going')
  await expect(step(page, 'Log your first activity')).toContainText('Done')
  await expect(step(page, 'Log something today')).toContainText('Done')
  await expect(step(page, 'Join a contest')).toContainText('Not done')
  await page.screenshot({ path: testInfo.outputPath('getting-going-logged.png'), fullPage: true })

  // A log and a contest are three of four steps, which still isn't enough: a new account can't log two days in a row
  // in one run, so Dev Reader's test covers the activity map.
  await joinedContest(adminContext, context)
  await page.reload()
  await expect(step(page, 'Join a contest')).toContainText('Done')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Getting going')
  await expect(step(page, 'Log two days in a row')).toContainText('Not done')
  await expect(page.getByRole('group', { name: 'Daily activity for the past 13 weeks', exact: true })).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('getting-going-three-steps.png'), fullPage: true })
})
