import { api, expect, reader, test, webv3 } from './logging-fixtures'
import { apiJson, score } from './standings'

const ordinal = new Intl.PluralRules('en-US', { type: 'ordinal' })
const suffix: Record<string, string> = { one: 'st', two: 'nd', few: 'rd', other: 'th' }
const place = (rank: number) => `${rank}${suffix[ordinal.select(rank)]}`
const points = (value: number) => `${score.format(value)} ${score.format(value) === '1.0' ? 'point' : 'points'}`
const count = new Intl.NumberFormat('en-US', { maximumFractionDigits: 1 })

test('a signed-in learner’s home shows their place, points, year totals and newest logs as the API reports them', async ({ page, readerContext }, testInfo) => {
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

  const [entry, summary, scores, activity, recent] = await Promise.all([
    apiJson<{ rank: number; is_tie: boolean; score: number }>(request, `/immersion/contests/${latest.id}/leaderboard/${userId}`),
    apiJson<{ participant_count: number }>(request, `/immersion/contests/${latest.id}/summary`),
    apiJson<{ scores: { score: number }[] }>(request, `/immersion/users/${userId}/scores/${year}`),
    apiJson<{ scores: { date: string; score: number }[] }>(request, `/immersion/users/${userId}/activity/${year}`),
    apiJson<{ logs: { id: string }[] }>(request, `/immersion/users/${userId}/logs?page_size=5&page=0&include_deleted=false`),
  ])
  const heading = `${entry.is_tie ? 'Tied ' : ''}${place(entry.rank)} place`
  const yearTotal = scores.scores.reduce((sum, each) => sum + each.score, 0)
  const todayTotal = activity.scores.find(day => day.date === today)?.score ?? 0

  // The values are rendered on the server, before any script runs.
  const html = await (await request.get(`${webv3}/`)).text()
  expect(html).toContain(`>${heading}<`)

  await page.goto('/')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(heading)
  const lead = `${points(entry.score)} in ${latest.title}. Out of ${count.format(summary.participant_count)} participants.`
  await expect(page.getByText(lead, { exact: true })).toBeVisible()

  const totals = page.getByRole('region', { name: `Your ${year}`, exact: true })
  const stat = (label: string) => totals.locator('dl > div').filter({ has: page.getByRole('term').filter({ hasText: label }) })
    .getByRole('definition')
  await expect(stat('Overall score')).toHaveText(score.format(yearTotal))
  // The unit is its own element beside the number.
  await expect(stat('Today')).toHaveText(points(todayTotal).replace(' ', ''))

  const newest = page.getByRole('region', { name: 'Recent activity', exact: true }).getByRole('listitem').getByRole('link')
    .and(page.locator('[href^="/logs/"]'))
  await expect(newest).toHaveCount(recent.logs.length)
  for (const [index, log] of recent.logs.entries()) await expect(newest.nth(index)).toHaveAttribute('href', `/logs/${log.id}`)
  await page.screenshot({ path: testInfo.outputPath('home-values.png'), fullPage: true })
})
