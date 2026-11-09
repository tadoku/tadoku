import { expect, reader, saveButton, test, webv3 } from './logging-fixtures'

/**
 * The journey `.github/workflows/webv3-production-canary.yaml` runs against production. The account may hold real data,
 * so it loads pages only: any write the pages send is refused and fails the run.
 */
test('a signed-in member loads home, a contest, their profile and the log form without writing', async ({ readerContext }) => {
  test.skip(!webv3 || !reader.email, 'Set WEBV3_BASE_URL and the canary account')
  test.setTimeout(60_000)
  const writes: string[] = []
  await readerContext.route('**/*', (route, request) => {
    if (['GET', 'HEAD', 'OPTIONS'].includes(request.method())) return route.continue()
    writes.push(`${request.method()} ${new URL(request.url()).pathname}`)
    return route.abort()
  })
  const page = await readerContext.newPage()
  const origin = new URL(webv3!).origin
  const visited: string[] = []
  page.on('framenavigated', frame => {
    if (frame === page.mainFrame()) visited.push(frame.url())
  })
  const banner = page.getByRole('banner')

  await page.goto('/')
  await expect(banner.getByRole('link', { name: 'Log activity', exact: true })).toBeVisible()
  await expect(page.getByRole('heading', { level: 2, name: 'Your contests', exact: true })).toBeVisible()
  // The account menu stays disabled until hydration, so a page that fails to hydrate stops here.
  const menu = banner.locator('[aria-haspopup="menu"]')
  await expect(menu).toBeEnabled()

  await banner.getByRole('link', { name: 'Contests', exact: true }).click()
  await expect(page).toHaveURL(/\/contests\/official\/1$/)
  const contest = page.locator('.contest-row').first().getByRole('link').first()
  const title = (await contest.textContent())!
  await contest.click()
  await expect(page).toHaveURL(/\/contests\/[a-f0-9-]{36}\/leaderboard\/1$/)
  await expect(page.getByRole('heading', { level: 1, name: title })).toBeVisible()

  const name = (await menu.getAttribute('aria-label'))!
  await menu.click()
  await page.getByRole('menuitem', { name: 'Profile', exact: true }).click()
  await expect(page).toHaveURL(/\/profile\/[a-f0-9-]{36}\/statistics$/)
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible()

  await banner.getByRole('link', { name: 'Log activity', exact: true }).click()
  await expect(page).toHaveURL(/\/logs\/new$/)
  await expect(saveButton(page)).toBeEnabled({ timeout: 15_000 })

  expect(writes).toEqual([])
  expect(visited.filter(url => new URL(url).origin !== origin)).toEqual([])
})
