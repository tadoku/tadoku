import type { APIRequestContext, BrowserContext } from '@playwright/test'
import { api, branch, expect, hydrated, test, webv3 } from './logging-fixtures'

/** Port-forward to the development Flipt, whose test environment holds this branch tenant's namespace. */
const flipt = process.env.E2E_FLIPT_URL
const webv2 = api ? new URL(api).origin : undefined
const account = webv3?.replace('.next.', '.account.')
const segment = 'release-webv3-access'
const resources = `${flipt}/api/v2/environments/test/namespaces/tadoku_${branch}/resources`
const page = '/contests/official/1?from=webv2&next=%2Flogs'

/** Adds the account to, or removes it from, the branch tenant's access segment. The flag stays disabled by default. */
async function setAccess(request: APIRequestContext, id: string, granted: boolean) {
  const current = await request.get(`${resources}/flipt.core.Segment/${segment}`)
  expect(current.status()).toBe(200)
  const { resource, revision } = await current.json()
  const [constraint] = resource.payload.constraints
  const members = (JSON.parse(constraint.value) as string[]).filter(member => member !== id)
  constraint.value = JSON.stringify(granted ? [...members, id].sort() : members)
  const updated = await request.put(resources, { data: { key: segment, revision, payload: resource.payload } })
  expect(updated.status()).toBe(200)
}

/** Waits for the API to see the change, which it picks up from its next Flipt snapshot. */
const decisionBecomes = (context: BrowserContext, expected: boolean) => expect.poll(async () => {
  const response = await context.request.get(`${api}/immersion/feature-flags`)
  return (await response.json()).decisions['release-webv3']
}, { timeout: 90_000, intervals: [2_000] }).toBe(expected)

const get = (context: BrowserContext, url: string) => context.request.get(url, { maxRedirects: 0 })

// Before any fixture runs, so a run without Flipt spends no registration on this test.
test.skip(!flipt || !webv2, 'Set E2E_FLIPT_URL, WEBV3_BASE_URL and TADOKU_API_URL for a branch tenant')

test('webv2 sends an account with release-webv3 to the same page on Webv3 and nothing else', async ({ browser, learner }) => {
  test.setTimeout(240_000)
  const { context, id } = learner
  const guest = await browser.newContext({ baseURL: webv3 })

  try {
    await test.step('serves webv2 to a guest and to an account without the flag', async () => {
      for (const visitor of [guest, context]) {
        const response = await get(visitor, `${webv2}${page}`)
        expect(response.status()).toBe(200)
        expect(response.headers().location).toBeUndefined()
      }
    })

    await setAccess(context.request, id, true)
    await decisionBecomes(context, true)

    await test.step('redirects a page request once to the same path and query, without caching', async () => {
      const response = await get(context, `${webv2}${page}`)
      expect(response.status()).toBe(302)
      expect(response.headers().location).toBe(`${webv3}${page}`)
      expect(response.headers()['cache-control']).toContain('no-store')

      const tab = await context.newPage()
      const document = await tab.goto(`${webv2}${page}`)
      expect(tab.url()).toBe(`${webv3}${page}`)
      expect(document!.request().redirectedFrom()?.url()).toBe(`${webv2}${page}`)
      expect(document!.request().redirectedFrom()?.redirectedFrom()).toBeNull()
      await hydrated(tab)
      await tab.close()
    })

    await test.step('never sends Webv3 back to webv2', async () => {
      const response = await get(context, `${webv3}${page}`)
      expect(response.status()).toBe(200)
      expect(response.headers().location).toBeUndefined()
    })

    await test.step('keeps Webv3 login and settings links returning to Webv3', async () => {
      const tab = await context.newPage()
      await tab.goto('/')
      await hydrated(tab)
      await tab.getByRole('banner').locator('[aria-haspopup="menu"]').click()
      await tab.getByRole('menuitem', { name: 'Account settings', exact: true }).click()
      await tab.waitForURL(`${account}/settings**`)
      expect(new URL(tab.url()).searchParams.get('return_to')).toBe(`${webv3}/`)
      await tab.close()

      const anonymous = await guest.newPage()
      await anonymous.goto('/')
      await hydrated(anonymous)
      const login = await anonymous.getByRole('link', { name: 'Log in', exact: true }).first().getAttribute('href')
      expect(decodeURIComponent(login!)).toContain(`return_to=${webv3}/`)
      await anonymous.close()
    })

    await test.step('leaves health, API, auth, asset and missing-page requests alone', async () => {
      const html = await (await get(guest, `${webv2}/`)).text()
      const asset = /\/_next\/static\/[^"']+\.(?:js|css)/.exec(html)![0]

      for (const [url, status] of [
        [`${webv2}/api/health`, 200],
        [`${api}/identity/session`, 200],
        [`${account}/login`, 200],
        [`${webv2}${asset}`, 200],
        [`${webv2}/favicon.png`, 200],
        [`${webv2}/favicon-missing.png`, 404],
      ] as const) {
        const response = await get(context, url)
        expect(response.status(), url).toBe(status)
        expect(response.headers().location, url).toBeUndefined()
      }
    })

    await setAccess(context.request, id, false)
    await decisionBecomes(context, false)

    await test.step('stops redirecting as soon as the flag is off', async () => {
      const response = await get(context, `${webv2}${page}`)
      expect(response.status()).toBe(200)
      expect(response.headers().location).toBeUndefined()
    })
  } finally {
    await setAccess(context.request, id, false)
    await guest.close()
  }
})
