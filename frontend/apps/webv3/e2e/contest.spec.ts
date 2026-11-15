import { randomUUID } from 'node:crypto'
import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { submitLog } from './contest-profile'
import { admin, api, expect, hydrated, reader, statValue, test, webv3 } from './logging-fixtures'
import { apiJson, board, expectRows, expectTotal, standingRows } from './standings'

interface Contest {
  id: string
  title: string
  contest_start: string
  contest_end: string
  registration_end: string
}

const day = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
const contestPath = (id: string, page = 1) => `/contests/${id}/leaderboard/${page}`
const latest = (request: APIRequestContext) => apiJson<Contest>(request, '/immersion/contests/latest-official')
const fitsViewport = (page: Page) => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)

/** Creates a community contest in the selected tenant as Dev Admin, who may date it in the past. */
async function createContest(context: BrowserContext, dates: Omit<Contest, 'id' | 'title'>): Promise<Contest> {
  const title = `Webv3 registration check ${randomUUID().slice(0, 8)}`
  const response = await context.request.post(`${api}/immersion/contests`, {
    data: {
      title,
      description: 'Created by the Webv3 contest registration journey.',
      ...dates,
      private: false,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return { ...(await response.json()), title }
}

const openDates = () => ({ contest_start: day(), contest_end: day(30), registration_end: day(14) })

/** The reader's registered language names for a contest, or null when the API answers that they haven't joined. */
async function registeredLanguages(context: BrowserContext, contestId: string) {
  const response = await context.request.get(`${api}/immersion/contests/${contestId}/registration`)
  if (response.status() === 204) return null
  expect(response.status()).toBe(200)
  return ((await response.json()) as { languages: { name: string }[] }).languages.map(language => language.name)
}

async function chooseLanguage(page: Page, name: string) {
  await page.getByRole('combobox', { name: 'Languages' }).fill(name.slice(0, 4))
  await page.getByRole('option', { name, exact: true }).click()
}

const joinLinks = (page: Page) => page.getByRole('main').getByRole('link', { name: 'Join contest', exact: true })
const contestOptions = (page: Page) => page.getByRole('main').getByRole('button', { name: 'Contest options' })
const updateRegistration = (page: Page) => page.getByRole('main').getByRole('link', { name: 'Update registration', exact: true })

async function contestMenu(page: Page) {
  await contestOptions(page).click()
  return page.getByRole('menuitem')
}

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('the server-rendered contest page shows its details, API standings and an anonymous visitor\'s prompts', async ({ browser, request }, testInfo) => {
  const contest = await latest(request)
  const data = await board(request, `/immersion/contests/${contest.id}/leaderboard`)
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()

  const response = await page.goto(contestPath(contest.id))
  expect(response?.status()).toBe(200)
  await expect(page).toHaveTitle(`${contest.title} · Tadoku`)
  await expect(page.getByRole('heading', { level: 1, name: contest.title })).toBeVisible()
  await expect(page.getByRole('link', { name: 'Official contests' })).toHaveAttribute('href', '/contests/official/1')
  await expectRows(page, data.entries, userId => `/contests/${contest.id}/profile/${userId}/1`)
  await expectTotal(page, data)
  await expect(page.getByRole('navigation', { name: 'Contest views' }).getByRole('link', { name: 'Activity' })).toHaveAttribute('href', `/contests/${contest.id}/updates/1`)

  const today = day()
  const ended = today > contest.contest_end
  const open = !ended && today <= contest.registration_end
  await expect(joinLinks(page)).toHaveCount(open ? 1 : 0)
  if (open) await expect(joinLinks(page)).toHaveAttribute('href', `/contests/${contest.id}/registration`)
  await expect(contestOptions(page)).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('contest-anonymous.png'), fullPage: true })
  await context.close()
})

test('optional and past-the-end page segments follow the API', async ({ page, request }) => {
  const contest = await latest(request)
  const data = await board(request, `/immersion/contests/${contest.id}/leaderboard`)

  await page.goto(`/contests/${contest.id}/leaderboard`)
  await expectTotal(page, data)
  if (data.total_size > 50) {
    await page.getByRole('navigation', { name: 'Standings pages' }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(new RegExp(`${contestPath(contest.id, 2)}$`))
    await expectTotal(page, await board(request, `/immersion/contests/${contest.id}/leaderboard`, 2), 2)
  }

  await page.goto(contestPath(contest.id, Math.ceil(data.total_size / 50) + 1))
  if (data.total_size === 0) return
  await expect(page.getByRole('heading', { name: 'This page is past the end of the standings' })).toBeVisible()
  await page.getByRole('link', { name: 'Go to the first page' }).click()
  await expect(page).toHaveURL(new RegExp(`${contestPath(contest.id)}$`))
  await expect(standingRows(page)).toHaveCount(data.entries.length)
})

const countdown = /^\d+ days?, \d+ hr, \d+ min$/
const utcClock = /^\d{1,2} [A-Z][a-z]{2,3} \d{4}, \d{2}:\d{2}:\d{2} UTC$/
const facts = (page: Page) => page.getByRole('region', { name: 'About this contest' })
const fact = (page: Page, label: string) => statValue(page, label).and(facts(page).getByRole('definition'))
const clockTerms = (page: Page) => facts(page).getByRole('term').filter({ hasText: /^(Starting in|Ending in|Tadoku time)$/ })

test('the contest page counts its participants from the API and counts down to its start or end on the UTC clock', async ({ browser, readerContext, adminContext }, testInfo) => {
  test.skip(!reader.email || !admin.email, 'Set both development fixture identities')
  const live = await createContest(adminContext, openDates())
  await submitLog(readerContext, live.id, `webv3-countdown-${randomUUID().slice(0, 8)}`)
  const upcoming = await createContest(adminContext, { contest_start: day(2), contest_end: day(30), registration_end: day(14) })
  const ended = await createContest(adminContext, { contest_start: day(-30), contest_end: day(-2), registration_end: day(-14) })

  // The server renders the count and the countdown from the API, without JavaScript.
  const server = await browser.newContext({ javaScriptEnabled: false })
  const document = await server.newPage()
  await document.goto(contestPath(live.id))
  const participants = (await board(document.request, `/immersion/contests/${live.id}/leaderboard`)).total_size
  expect(participants).toBe(1)
  await expect(fact(document, 'Participants')).toHaveText('1')
  await expect(fact(document, 'Ending in')).toHaveText(countdown)
  await expect(fact(document, 'Tadoku time')).toHaveText(utcClock)
  await document.goto(contestPath(upcoming.id))
  await expect(fact(document, 'Starting in')).toHaveText(/^[12] days?, \d+ hr, \d+ min$/)
  await expect(fact(document, 'Participants')).toHaveText('0')
  await document.goto(contestPath(ended.id))
  await expect(fact(document, 'Dates')).toBeVisible()
  await expect(clockTerms(document)).toHaveCount(0)
  await server.close()

  // With the browser's clock a month ahead, the page still hydrates with the server's time and then ticks on its own,
  // until the contest's last UTC second passes.
  const end = Date.parse(`${live.contest_end}T00:00:00Z`) + 86_400_000
  const page = await browser.newPage()
  const failures: string[] = []
  page.on('pageerror', error => failures.push(error.message))
  // An organizer without an avatar answers 404, and the page shows their initials instead.
  page.on('console', message => {
    if (message.type() === 'error' && !message.location().url.endsWith('/avatar')) failures.push(message.text())
  })
  await page.clock.install({ time: end - 65_000 })
  await page.goto(contestPath(live.id))
  await hydrated(page)
  await expect(fact(page, 'Ending in')).toHaveText('0 days, 0 hr, 1 min')
  const clock = await fact(page, 'Tadoku time').textContent()
  await page.clock.fastForward(2_000)
  await expect(fact(page, 'Tadoku time')).not.toHaveText(clock!)
  await page.screenshot({ path: testInfo.outputPath('contest-countdown.png'), fullPage: true })
  await page.clock.fastForward(65_000)
  await expect(clockTerms(page)).toHaveCount(0)
  await expect(fact(page, 'Participants')).toHaveText('1')
  expect(failures).toEqual([])
  await page.close()
})

test('unknown and malformed contests are not found', async ({ page }) => {
  for (const path of [contestPath(randomUUID()), '/contests/not-a-contest/leaderboard/1', '/contests/not-a-contest/scoring']) {
    const response = await page.goto(path)
    expect(response?.status(), path).toBe(404)
    await expect(page.getByRole('heading', { level: 1, name: 'Contest not found' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Browse contests' })).toHaveAttribute('href', '/contests/official/1')
  }
})

test('anonymous visitors who open a registration are sent to log in and returned to it', async ({ request }) => {
  const contest = await latest(request)
  const registration = `/contests/${contest.id}/registration`

  // Read the redirect without following it: in the container check the login host is production's account site.
  const response = await request.get(registration, { maxRedirects: 0 })
  expect(response.status()).toBe(307)
  const login = new URL(response.headers().location)
  expect(login.pathname).toBe('/login')
  const returnTo = new URL(login.searchParams.get('return_to')!)
  expect(returnTo.host + returnTo.pathname).toBe(new URL(webv3!).host + registration)
})

test('a signed-in learner joins an open contest, sees it on the contest page and changes languages', async ({ page, readerContext, adminContext }, testInfo) => {
  test.skip(!reader.email || !admin.email, 'Set both development fixture identities')
  const contest = await createContest(adminContext, openDates())
  expect(await registeredLanguages(readerContext, contest.id)).toBeNull()

  await page.goto(contestPath(contest.id))
  await expect(page.getByRole('link', { name: 'Community contests' })).toHaveAttribute('href', '/contests/user-contests/1')
  await expect(joinLinks(page)).toHaveCount(1)
  await joinLinks(page).click()
  await expect(page.getByRole('heading', { level: 1, name: `Join ${contest.title}` })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Register' })).toBeEnabled()

  // An empty selection is refused before anything is sent.
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page.getByText('Select at least one language.')).toBeVisible()
  await chooseLanguage(page, 'Japanese')
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page).toHaveURL(new RegExp(`${contestPath(contest.id)}$`))
  await expect(joinLinks(page)).toHaveCount(0)
  // A wide header shows Update registration as a button and needs no menu.
  await expect(updateRegistration(page)).toBeVisible()
  await expect(contestOptions(page)).toHaveCount(0)
  expect(await registeredLanguages(readerContext, contest.id)).toEqual(['Japanese'])

  // On a phone the menu holds it, on the title's row below the back link, centred on the title's first line.
  await page.setViewportSize({ width: 390, height: 844 })
  await page.reload()
  await expect(joinLinks(page)).toHaveCount(0)
  await expect(updateRegistration(page)).toHaveCount(0)
  await expect(contestOptions(page)).toBeEnabled()
  const heading = page.getByRole('heading', { level: 1 })
  const line = await heading.evaluate(element => parseFloat(getComputedStyle(element).lineHeight))
  // Content above the row, such as announcements, arrives after hydration, so measure until the layout settles.
  await expect(async () => {
    const options = await contestOptions(page).boundingBox()
    const back = await page.getByRole('link', { name: 'Community contests' }).boundingBox()
    const title = await heading.boundingBox()
    expect(options!.y).toBeGreaterThanOrEqual(back!.y + back!.height)
    expect(Math.abs(options!.y + options!.height / 2 - (title!.y + line / 2))).toBeLessThan(4)
  }).toPass({ timeout: 5_000 })
  expect(await fitsViewport(page)).toBe(true)
  await expect(await contestMenu(page)).toHaveText(['Update registration'])
  await page.screenshot({ path: testInfo.outputPath('contest-registered-390.png'), fullPage: true })
  await page.getByRole('menuitem', { name: 'Update registration' }).click()
  await page.setViewportSize({ width: 1280, height: 720 })
  await expect(page.getByRole('heading', { level: 1, name: 'Update registration' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Update registration' })).toBeEnabled()
  await chooseLanguage(page, 'Spanish')
  await page.getByRole('button', { name: 'Remove Japanese' }).click()
  await expect(page.getByText('Removing Japanese will also remove any contest logs you’ve submitted in that language.')).toBeVisible()
  await page.screenshot({ path: testInfo.outputPath('registration-update.png'), fullPage: true })
  await page.getByRole('button', { name: 'Update registration' }).click()
  await expect(page).toHaveURL(new RegExp(`${contestPath(contest.id)}$`))
  expect(await registeredLanguages(readerContext, contest.id)).toEqual(['Spanish'])

  const unknown = await page.goto(`/contests/${randomUUID()}/registration`)
  expect(unknown?.status()).toBe(404)
  await expect(page.getByRole('heading', { level: 1, name: 'Contest not found' })).toBeVisible()
})

test('registration after the deadline is refused as in Webv2', async ({ page, readerContext, adminContext }, testInfo) => {
  test.skip(!reader.email || !admin.email, 'Set both development fixture identities')
  const closed = await createContest(adminContext, { contest_start: day(-7), contest_end: day(30), registration_end: day(-1) })

  await page.goto(contestPath(closed.id))
  await expect(page.getByRole('heading', { level: 1, name: closed.title })).toBeVisible()
  await expect(page.getByRole('main').getByText('Live now', { exact: true })).toBeVisible()
  await expect(joinLinks(page)).toHaveCount(0)
  await expect(page.getByText('Closed to new participants')).toBeVisible()

  await page.goto(`/contests/${closed.id}/registration`)
  await expect(page.getByText('Unfortunately, registrations for this contest have ended.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Register' })).toHaveCount(0)
  await page.screenshot({ path: testInfo.outputPath('registration-closed.png'), fullPage: true })

  // A form opened while registration was open is refused by the API once it has closed: here the save is redirected
  // to the closed contest inside the server function's payload.
  const open = await createContest(adminContext, openDates())
  await page.goto(`/contests/${open.id}/registration`)
  await expect(page.getByRole('button', { name: 'Register' })).toBeEnabled()
  await page.route('**/_serverFn/**', async route => {
    const body = route.request().postData()
    if (route.request().method() !== 'POST' || !body?.includes(open.id)) return route.continue()
    return route.continue({ postData: body.replaceAll(open.id, closed.id) })
  })
  await chooseLanguage(page, 'Japanese')
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page.getByText('Registration for this contest has closed.')).toBeVisible()
  await expect(page).toHaveURL(new RegExp(`/contests/${open.id}/registration$`))
  expect(await registeredLanguages(readerContext, closed.id)).toBeNull()
  expect(await registeredLanguages(readerContext, open.id)).toBeNull()
})

test('contest and registration pages stay within mobile and desktop viewports', async ({ request, readerContext }, testInfo) => {
  const contest = await latest(request)
  const pages = [['contest', contestPath(contest.id)]]
  // Registration needs an account; the context holds the shared session when the fixture credentials are set.
  if (reader.email && reader.password) pages.push(['registration', `/contests/${contest.id}/registration`])
  const page = await readerContext.newPage()

  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 900 })
    for (const [name, path] of pages) {
      await page.goto(path)
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
      await hydrated(page)
      await page.evaluate(() => document.fonts.ready)
      expect(await fitsViewport(page), `${path} at ${width}px`).toBe(true)
      await page.screenshot({ path: testInfo.outputPath(`${name}-${width}.png`), fullPage: true })
    }
  }
})
