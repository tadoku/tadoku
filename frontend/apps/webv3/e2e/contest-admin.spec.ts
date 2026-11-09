import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { accountId, admin, api, expect, reader, test, webv3 } from './logging-fixtures'

interface Contest {
  id: string
  title: string
  description?: string
  contest_start: string
  contest_end: string
  registration_end: string
  private: boolean
  official: boolean
  language_code_allow_list: string[] | null
  activity_type_id_allow_list: number[]
}
interface RuleSet {
  id: string
  version: number
  status: 'draft' | 'published'
  active: boolean
  mode?: 'override' | 'replace'
  fallback_rule_set_id?: string
  rules: { priority: number; stackable: boolean; activity_id: number; tag?: string; score_source: string; rate: number }[]
}

const utcDay = (offset = 0) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
const unique = () => `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`
const createButton = (page: Page) => page.getByRole('button', { name: 'Create contest', exact: true })

async function hydrated(page: Page) {
  await expect(page.locator('.paper-navbar__mobile-trigger')).toBeEnabled()
}

async function fitsWidth(page: Page, name: string, path: string) {
  await page.setViewportSize({ width: 390, height: 900 })
  await hydrated(page)
  await page.evaluate(() => document.fonts.ready)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `${name} at 390px`).toBe(true)
  await page.screenshot({ path, fullPage: true })
  await page.setViewportSize({ width: 1280, height: 900 })
}

/** Creates a community contest through the API as the request's account. */
async function postContest(request: APIRequestContext, start: string, end: string, title = `Webv3 scoring check ${unique()}`) {
  const response = await request.post(`${api}/immersion/contests`, {
    data: {
      title,
      description: 'Created by the Webv3 contest administration journey.',
      contest_start: start,
      contest_end: end,
      registration_end: end,
      private: true,
      official: false,
      language_code_allow_list: [],
      activity_type_id_allow_list: [1, 2],
    },
  })
  expect(response.ok(), `contest creation answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<Contest>
}

async function ruleSets(request: APIRequestContext, path: string) {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), `${path} answered ${response.status()}`).toBeTruthy()
  return (await response.json() as { rule_sets: RuleSet[] }).rule_sets
}

/** Exhausts Dev Reader's yearly contest allowance in the selected tenant unless the API already refuses it. */
async function ensureCreationRefused(context: BrowserContext) {
  for (let attempt = 0; attempt < 13; attempt++) {
    const permission = await context.request.get(`${api}/immersion/contests/create-permissions`)
    if (permission.status() === 403) return
    expect(permission.status()).toBe(200)
    await postContest(context.request, utcDay(30), utcDay(40), `Webv3 allowance check ${unique()}`)
  }
  throw new Error('The API kept allowing Dev Reader to create contests')
}

test('anonymous visitors are asked to log in to create a contest and sent to log in for scoring', async ({ page }, testInfo) => {
  test.skip(!webv3, 'Set WEBV3_BASE_URL')
  await page.goto('/contests/new')
  await expect(page.getByRole('heading', { level: 1, name: 'Sign in to create a contest' })).toBeVisible()
  await expect(createButton(page)).toHaveCount(0)
  await hydrated(page)
  const login = new URL((await page.locator('main').getByRole('link', { name: 'Log in', exact: true }).getAttribute('href'))!)
  expect(login.pathname).toBe('/login')
  expect(login.searchParams.get('return_to')).toBe(page.url())
  await fitsWidth(page, '/contests/new anonymous', testInfo.outputPath('new-anonymous-390.png'))

  // As Webv2's protected page does, scoring sends an anonymous visitor straight to login and back.
  const path = '/contests/00000000-0000-4000-8000-000000000001/scoring?from=e2e'
  const response = await page.request.get(path, { maxRedirects: 0 })
  expect(response.status()).toBeGreaterThanOrEqual(300)
  expect(response.status()).toBeLessThan(400)
  const redirect = new URL(response.headers()['location'])
  expect(redirect.pathname).toBe('/login')
  const returnTo = new URL(redirect.searchParams.get('return_to')!)
  expect(returnTo.pathname + returnTo.search).toBe(path)
  expect(returnTo.host).toBe(new URL(webv3!).host)
})

test('an organizer gets each invalid field focused, then creates a contest and lands on its contest page', async ({ adminContext }, testInfo) => {
  test.skip(!api || !admin.email || !admin.password, 'Set TADOKU_API_URL, E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD')
  const page = await adminContext.newPage()
  const failures: string[] = []
  page.on('pageerror', error => failures.push(error.message))

  // Starting from My contests keeps that list cached in the browser for the return below.
  await page.goto('/contests/my-contests/1')
  await hydrated(page)
  await page.getByRole('link', { name: 'Create contest', exact: true }).click()
  await expect(page).toHaveURL(/\/contests\/new$/)
  await expect(page.getByRole('heading', { level: 1, name: 'Create new contest' })).toBeVisible()
  await expect(page).toHaveTitle('Create new contest · Tadoku')
  await hydrated(page)
  await page.screenshot({ path: testInfo.outputPath('new-form.png'), fullPage: true })

  const title = page.getByRole('textbox', { name: 'Title', exact: true })
  await createButton(page).click()
  await expect(title).toBeFocused()
  await expect(page.getByText('Use a title of at least 4 characters.')).toBeVisible()
  await expect(page.getByText('Choose a date.')).toHaveCount(3)

  const name = `Webv3 creation check ${unique()}`
  await title.fill(name)
  await page.getByLabel(/^Description/).fill('  Read anything you like.  ')
  await page.getByLabel('Start date').fill(utcDay(20))
  await page.getByLabel('End date').fill(utcDay(10))
  await page.getByLabel('Sign-up deadline').fill(utcDay(15))
  await createButton(page).click()
  await expect(page.getByLabel('End date')).toBeFocused()
  await expect(page.getByText('Choose an end date on or after the start date.')).toBeVisible()
  await page.getByLabel('End date').fill(utcDay(50))

  // Without activities the API refuses the contest.
  const activities = page.getByRole('combobox', { name: 'Allowed activities' })
  for (let chips = page.getByRole('button', { name: /^Remove / }); await chips.count();) await chips.first().click()
  await createButton(page).click()
  await expect(activities).toBeFocused()
  await expect(page.getByText('Choose at least one activity.')).toBeVisible()
  await activities.fill('Reading')
  await page.getByRole('option', { name: 'Reading', exact: true }).click()
  await page.keyboard.press('Escape')
  const languages = page.getByRole('combobox', { name: 'Allowed languages' })
  await languages.fill('Japanese')
  await page.getByRole('option', { name: 'Japanese', exact: true }).click()
  await page.keyboard.press('Escape')
  await page.getByLabel('Only those with the link can sign up').check()
  await fitsWidth(page, '/contests/new form', testInfo.outputPath('new-form-390.png'))

  await createButton(page).click()
  await page.waitForURL(url => /^\/contests\/[0-9a-f-]{36}\/leaderboard\/1$/.test(url.pathname))
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible()
  const id = new URL(page.url()).pathname.split('/')[2]
  await page.goBack()
  await expect(page).toHaveURL(/\/contests\/my-contests\/1$/)
  await expect(page.locator('.contest-row').first().getByRole('heading', { level: 2 })).toHaveText(name)
  const created = await (await adminContext.request.get(`${api}/immersion/contests/${id}`)).json() as Contest & {
    allowed_languages: { code: string }[] | null
    allowed_activities: { id: number }[]
  }
  expect(created).toMatchObject({
    title: name,
    description: 'Read anything you like.',
    contest_start: utcDay(20),
    contest_end: utcDay(50),
    registration_end: utcDay(15),
    private: true,
    official: false,
  })
  expect(created.allowed_languages?.map(language => language.code)).toEqual(['jpn'])
  expect(created.allowed_activities.map(activity => activity.id)).toEqual([1])
  expect(failures).toEqual([])
})

test('an account the API refuses sees why instead of the contest form', async ({ readerContext }, testInfo) => {
  test.skip(!api || !reader.email || !reader.password, 'Set TADOKU_API_URL, E2E_READER_EMAIL and E2E_READER_PASSWORD')
  await ensureCreationRefused(readerContext)
  const page = await readerContext.newPage()

  await page.goto('/contests/new')
  await expect(page.getByRole('heading', { level: 1, name: 'You can’t create a contest yet' })).toBeVisible()
  await expect(createButton(page)).toHaveCount(0)
  await page.goto('/contests/my-contests/1')
  await expect(page.getByRole('link', { name: 'Create contest', exact: true })).toHaveCount(0)
  await page.goto('/contests/new')
  await fitsWidth(page, '/contests/new refused', testInfo.outputPath('new-refused-390.png'))
})

test('an administrator drafts, publishes and activates a contest rule set', async ({ adminContext }, testInfo) => {
  test.skip(!api || !admin.email || !admin.password, 'Set TADOKU_API_URL, E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD')
  const contest = await postContest(adminContext.request, utcDay(30), utcDay(60))
  const platform = (await ruleSets(adminContext.request, '/immersion/scoring/rule-sets'))
    .filter(ruleSet => ruleSet.status === 'published')
  const page = await adminContext.newPage()

  await page.goto(`/contests/${contest.id}/leaderboard/1`)
  await hydrated(page)
  await expect(page.locator('.standings-header .paper-page-header__actions').getByRole('link')).toHaveText(['Join contest'])
  await page.getByRole('button', { name: 'Contest options' }).click()
  await expect(page.getByRole('menuitem')).toHaveText(['Scoring rules'])
  await page.screenshot({ path: testInfo.outputPath('contest-admin-menu.png') })
  await page.getByRole('menuitem', { name: 'Scoring rules' }).click()
  await expect(page).toHaveURL(new RegExp(`/contests/${contest.id}/scoring$`))
  await expect(page.getByRole('heading', { level: 1, name: 'Contest scoring' })).toBeVisible()
  await expect(page).toHaveTitle('Contest scoring · Tadoku')
  await expect(page.getByText(contest.title, { exact: true })).toBeVisible()
  await expect(page.getByText('No contest-specific rule sets yet; this contest scores with the platform defaults.')).toBeVisible()
  await expect(page.getByRole('link', { name: contest.title, exact: true }))
    .toHaveAttribute('href', `/contests/${contest.id}/leaderboard/1`)
  await hydrated(page)

  const pinned = page.getByLabel('Pinned platform version')
  const expectedFallback = platform.find(ruleSet => ruleSet.active) ?? platform[0]
  await expect(pinned).toHaveValue(expectedFallback?.id ?? '')
  if (!expectedFallback) await page.getByLabel('Contest behavior').selectOption('replace')

  const rules = page.locator('.scoring-rule')
  const rate = rules.nth(0).getByLabel('Rate')
  await rate.fill('')
  await page.getByRole('button', { name: 'Create draft' }).click()
  await expect(rate).toBeFocused()
  await expect(rules.nth(0).getByText('Enter a rate.')).toBeVisible()
  await rate.fill('2.5')

  await page.getByRole('button', { name: 'Add rule' }).click()
  await expect(rules).toHaveCount(2)
  await rules.nth(1).getByLabel('Activity').selectOption({ label: 'Listening' })
  await rules.nth(1).getByLabel('Score source').selectOption('duration_minutes')
  await rules.nth(1).getByLabel('Tag matcher').fill('  Webv3 Podcast ')
  await rules.nth(1).getByLabel('Rate').fill('0.5')
  await rules.nth(1).getByLabel('Stackable modifier').check()
  // Moving the modifier first, then back, keeps each rule's values with it.
  await rules.nth(1).getByRole('button', { name: 'Up' }).click()
  await expect(rules.nth(0).getByLabel('Tag matcher')).toHaveValue('  Webv3 Podcast ')
  await rules.nth(0).getByRole('button', { name: 'Down' }).click()
  await expect(rules.nth(1).getByLabel('Tag matcher')).toHaveValue('  Webv3 Podcast ')
  await page.screenshot({ path: testInfo.outputPath('scoring-form.png'), fullPage: true })

  await page.getByRole('button', { name: 'Create draft' }).click()
  await expect(page.getByRole('status').filter({ hasText: 'Scoring rule-set draft created.' })).toBeVisible()
  const version = page.locator('.scoring-version')
  await expect(version).toHaveCount(1)
  let [draft] = await ruleSets(adminContext.request, `/immersion/contests/${contest.id}/scoring/rule-sets`)
  expect(draft).toMatchObject({
    status: 'draft',
    active: false,
    mode: expectedFallback ? 'override' : 'replace',
    ...(expectedFallback ? { fallback_rule_set_id: expectedFallback.id } : {}),
  })
  expect(draft.rules.map(({ priority, stackable, activity_id, tag, score_source, rate }) =>
    ({ priority, stackable, activity_id, tag, score_source, rate }))).toEqual([
    { priority: 0, stackable: false, activity_id: 1, tag: undefined, score_source: 'amount', rate: 2.5 },
    { priority: 10, stackable: true, activity_id: 2, tag: 'webv3 podcast', score_source: 'duration_minutes', rate: 0.5 },
  ])
  await expect(version.getByRole('heading', { level: 3 })).toHaveText(`Version ${draft.version}`)
  await expect(version.locator('.paper-status-tag')).toHaveText(['Draft'])
  await expect(version).toContainText('2.5')
  await expect(version).toContainText('webv3 podcast')

  await version.getByRole('button', { name: 'Publish' }).click()
  const dialog = page.getByRole('dialog', { name: 'Publish this version?' })
  await dialog.getByRole('button', { name: 'Cancel' }).click()
  await expect(dialog).toHaveCount(0)
  await version.getByRole('button', { name: 'Publish' }).click()
  await page.getByRole('dialog', { name: 'Publish this version?' }).getByRole('button', { name: 'Publish' }).click()
  await expect(version.locator('.paper-status-tag')).toHaveText(['Published'])
  await version.getByRole('button', { name: 'Activate' }).click()
  await page.getByRole('dialog', { name: 'Activate this version?' }).getByRole('button', { name: 'Activate' }).click()
  await expect(version.locator('.paper-status-tag')).toHaveText(['Published', 'Active'])
  await expect(version.getByRole('button')).toHaveCount(0)
  ;[draft] = await ruleSets(adminContext.request, `/immersion/contests/${contest.id}/scoring/rule-sets`)
  expect(draft).toMatchObject({ status: 'published', active: true })

  await page.reload()
  await expect(version.locator('.paper-status-tag')).toHaveText(['Published', 'Active'])
  await fitsWidth(page, 'scoring', testInfo.outputPath('scoring-390.png'))
})

test('a started contest’s scoring is locked, and only administrators reach scoring', async ({ adminContext, readerContext }, testInfo) => {
  test.skip(
    !api || !admin.email || !admin.password || !reader.email || !reader.password,
    'Set TADOKU_API_URL and both fixture accounts',
  )
  // Administrators may create contests that already started.
  const started = await postContest(adminContext.request, utcDay(-2), utcDay(30))
  const page = await adminContext.newPage()
  await page.goto(`/contests/${started.id}/scoring`)
  await expect(page.getByText('Scoring is locked because this contest has started.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Create draft' })).toHaveCount(0)
  // With no contest rule sets, the lock notice itself names the platform defaults.
  await expect(page.getByRole('link', { name: 'platform defaults' })).toBeVisible()
  await fitsWidth(page, 'locked scoring', testInfo.outputPath('scoring-locked-390.png'))

  await page.goto('/contests/not-a-contest/scoring')
  await expect(page.getByRole('heading', { level: 1, name: 'Contest not found' })).toBeVisible()

  // The API lets organizers manage their own contest's scoring; Webv2's page, and so Webv3's, is for administrators.
  const readerId = await accountId(readerContext)
  const own = await (await readerContext.request.get(
    `${api}/immersion/contests?page_size=1&page=0&official=false&include_deleted=false&user_id=${readerId}`,
  )).json() as { contests: Contest[] }
  const visitor = await readerContext.newPage()
  await visitor.goto(`/contests/${started.id}/leaderboard/1`)
  await expect(visitor.getByRole('heading', { level: 1, name: started.title })).toBeVisible()
  await hydrated(visitor)
  await expect(visitor.getByRole('button', { name: 'Contest options' })).toHaveCount(0)
  for (const id of [started.id, own.contests[0]?.id].filter(Boolean)) {
    await visitor.goto(`/contests/${id}/scoring`)
    await expect(visitor.getByRole('heading', { level: 1, name: 'This feature is not yet available.' })).toBeVisible()
    await expect(visitor.getByRole('button', { name: 'Create draft' })).toHaveCount(0)
  }
})
