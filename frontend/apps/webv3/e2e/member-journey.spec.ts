import { admin, api, createLog, expect, fillDraft, joinedContest, saveButton, test, webv3 } from './logging-fixtures'
import { standingRows } from './standings'

test('a member walks home, contest, standings, profile, logging and the saved log without leaving Webv3', async ({ learner, adminContext }, testInfo) => {
  test.skip(!webv3 || !api || !admin.email, 'Set development URLs and the Dev Admin fixture account')
  test.setTimeout(90_000)
  // A new account's home lists the contest it joined; Dev Reader's lists only its oldest five unfinished contests.
  // Joined and logged to in setup, so the journey doesn't rely on a seeded round or its dates.
  const { context, name: member, id: memberId } = learner
  const { id: contestId, title: contestTitle, registrationId } = await joinedContest(adminContext, context)
  const logged = await context.request.post(`${api}/immersion/logs`, {
    data: { language_code: 'jpn', activity_id: 1, amount: 1, unit_key: 'pages', tags: [], description: 'Member journey', registration_ids: [registrationId] },
  })
  expect(logged.ok(), `log answered ${logged.status()}`).toBeTruthy()
  const page = await context.newPage()
  const origin = new URL(webv3!).origin
  const visited: string[] = []
  page.on('framenavigated', frame => {
    if (frame === page.mainFrame()) visited.push(frame.url())
  })
  const shot = (name: string) => page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true })

  await page.goto('/')
  await expect(page.getByRole('button', { name: member, exact: true })).toBeVisible()
  // A member's home carries their year, newest updates and contests under the hero.
  for (const heading of [`Your ${new Date().getUTCFullYear()}`, 'Recent activity', 'Your contests']) {
    await expect(page.getByRole('heading', { level: 2, name: heading, exact: true })).toBeVisible()
  }
  // Having joined a contest this year, the home page links to it and to everything they joined.
  await expect(page.getByRole('link', { name: 'All joined', exact: true })).toHaveAttribute('href', '/contests/joined/1')
  await shot('1-home')

  // The appearance switch is the account menu's; the choice survives a reload.
  const theme = await page.locator('html').getAttribute('data-theme')
  const other = theme === 'dark' ? 'light' : 'dark'
  await page.getByRole('button', { name: member, exact: true }).click()
  await page.getByRole('menuitem', { name: `Use ${other} theme` }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', other)
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', other)
  await page.getByRole('button', { name: member, exact: true }).click()
  await page.getByRole('menuitem', { name: `Use ${theme} theme` }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', theme!)

  await page.getByRole('region', { name: 'Your contests', exact: true }).getByRole('link', { name: contestTitle, exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contestId}/profile/${memberId}/1$`))
  await page.getByRole('main').getByRole('link', { name: contestTitle, exact: true }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contestId}/leaderboard/1$`))
  const standings = standingRows(page)
  await expect(standings.first()).toBeVisible()
  await shot('2-contest-standings')

  // A standings row opens its participant from anywhere on the row, here its score.
  const participant = standings.first().getByRole('link')
  const name = (await participant.textContent())!
  const score = await standings.first().getByRole('cell').last().boundingBox()
  await page.mouse.click(score!.x + score!.width / 2, score!.y + score!.height / 2)
  await expect(page).toHaveURL(/\/contests\/[a-f0-9-]{36}\/profile\/[a-f0-9-]{36}\/1$/)
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible()
  await shot('3-participant')

  await page.getByRole('link', { name: 'View profile', exact: true }).click()
  await expect(page).toHaveURL(/\/profile\/[a-f0-9-]{36}\/statistics$/)
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible()
  await shot('4-profile')

  await page.getByRole('banner').getByRole('link', { name: 'Log activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/new$/)
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })
  const description = `member-journey-${Date.now()}`
  await fillDraft(page, '3', description)
  await createLog(page)
  await expect(page.getByText(description)).toBeVisible()
  await shot('5-log-detail')

  expect(visited.length).toBeGreaterThanOrEqual(6)
  expect(visited.filter(url => new URL(url).origin !== origin)).toEqual([])
})
