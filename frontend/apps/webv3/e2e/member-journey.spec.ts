import { api, expect, fillDraft, reader, savedLogUrl, saveButton, test, webv3 } from './logging-fixtures'

test('a member walks home, contest, standings, profile, logging and the saved log without leaving Webv3', async ({ readerContext }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  test.setTimeout(90_000)
  const page = await readerContext.newPage()
  const origin = new URL(webv3!).origin
  const visited: string[] = []
  page.on('framenavigated', frame => {
    if (frame === page.mainFrame()) visited.push(frame.url())
  })
  const shot = (name: string) => page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true })

  await page.goto('/')
  await expect(page.getByRole('button', { name: 'Dev Reader', exact: true })).toBeVisible()
  await shot('1-home')

  await page.getByRole('link', { name: /^(Contest details|Full leaderboard|See final board)$/ }).first().click()
  await expect(page).toHaveURL(/\/contests\/[a-f0-9-]{36}\/leaderboard\/1$/)
  const standings = page.locator('.standings-table tbody tr')
  await expect(standings.first()).toBeVisible()
  await shot('2-contest-standings')

  const participant = standings.first().getByRole('link')
  const name = (await participant.textContent())!
  await participant.click()
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
  await saveButton(page).click()
  await expect(page).toHaveURL(savedLogUrl)
  await expect(page.getByText(description)).toBeVisible()
  await shot('5-log-detail')

  expect(visited.length).toBeGreaterThanOrEqual(6)
  expect(visited.filter(url => new URL(url).origin !== origin)).toEqual([])
})
