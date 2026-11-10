import type { APIRequestContext } from '@playwright/test'
import { api, expect, reader, test, webv3 } from './logging-fixtures'
import { apiJson } from './standings'

interface Route {
  path: string
  title: string
  /** The page's own path without a page number, when its page links take Webv2's `<base>/<page>` shape. */
  paged?: string
  /** The page number the path selects, when it is not the first. */
  page?: number
}

async function redirect(request: APIRequestContext, path: string) {
  const response = await request.get(`${webv3}${path}`, { maxRedirects: 0 })
  return { status: response.status(), location: new URL(response.headers().location, webv3).href }
}

test('every Webv2 route shape, bookmarked with or without page numbers, opens the same page on Webv3', async ({ browser, request }, testInfo) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  test.setTimeout(180_000)
  const year = new Date().getUTCFullYear()
  const today = new Date().toISOString().slice(0, 10)
  const contest = await apiJson<{ id: string; title: string }>(request, '/immersion/contests/latest-official')
  const [{ entries: [participant] }, { logs: [log] }, { posts: [post] }, cmsPage] = await Promise.all([
    apiJson<{ entries: { user_id: string; user_display_name: string }[] }>(request, `/immersion/contests/${contest.id}/leaderboard?page_size=1&page=0`),
    apiJson<{ logs: { id: string }[] }>(request, `/immersion/contests/${contest.id}/logs?page_size=1&page=0`),
    apiJson<{ posts: { slug: string; title: string }[] }>(request, '/content/posts/tadoku?page_size=1&page=0'),
    // The tenant hook seeds this CMS page.
    apiJson<{ title: string }>(request, '/content/pages/tadoku/dev-welcome'),
  ])
  expect(participant && log && post, 'the contest has a participant and a log, and the blog a post').toBeTruthy()
  const c = `/contests/${contest.id}`
  const name = participant.user_display_name
  const profile = `/profile/${participant.user_id}`

  const routes: Route[] = [
    { path: '/', title: 'Tadoku' },
    { path: '/contests/official', title: 'Official contests · Tadoku', paged: '/contests/official' },
    { path: '/contests/official/2', title: 'Official contests · Tadoku', paged: '/contests/official', page: 2 },
    { path: '/contests/user-contests', title: 'Community contests · Tadoku', paged: '/contests/user-contests' },
    { path: '/contests/user-contests/2', title: 'Community contests · Tadoku', paged: '/contests/user-contests', page: 2 },
    { path: '/contests/my-contests', title: 'Contests you organize · Tadoku' },
    { path: '/contests/my-contests/2', title: 'Contests you organize · Tadoku' },
    { path: '/contests/new', title: 'Create new contest · Tadoku' },
    { path: `${c}/leaderboard`, title: `${contest.title} · Tadoku`, paged: `${c}/leaderboard` },
    { path: `${c}/leaderboard/2`, title: `${contest.title} · Tadoku`, paged: `${c}/leaderboard`, page: 2 },
    { path: `${c}/updates`, title: `Activity · ${contest.title} · Tadoku`, paged: `${c}/updates` },
    { path: `${c}/updates/2`, title: `Activity · ${contest.title} · Tadoku`, paged: `${c}/updates`, page: 2 },
    { path: `${c}/profile/${participant.user_id}`, title: `${name} · ${contest.title} · Tadoku` },
    { path: `${c}/profile/${participant.user_id}/2`, title: `${name} · ${contest.title} · Tadoku` },
    { path: '/leaderboard/latest', title: 'Latest leaderboard · Tadoku', paged: '/leaderboard/latest' },
    { path: '/leaderboard/latest/2', title: 'Latest leaderboard · Tadoku', paged: '/leaderboard/latest', page: 2 },
    { path: `/leaderboard/yearly/${year}`, title: `${year} leaderboard · Tadoku`, paged: `/leaderboard/yearly/${year}` },
    { path: `/leaderboard/yearly/${year}/2`, title: `${year} leaderboard · Tadoku`, paged: `/leaderboard/yearly/${year}`, page: 2 },
    // Webv2 reads a segment's leading integer.
    { path: `/leaderboard/yearly/${year}x/2abc`, title: `${year} leaderboard · Tadoku`, paged: `/leaderboard/yearly/${year}`, page: 2 },
    { path: '/leaderboard/all-time', title: 'All time leaderboard · Tadoku', paged: '/leaderboard/all-time' },
    { path: '/leaderboard/all-time/2', title: 'All time leaderboard · Tadoku', paged: '/leaderboard/all-time', page: 2 },
    { path: '/logs/new', title: 'Log activity · Tadoku' },
    { path: '/logs/new?amount=5', title: 'Log activity · Tadoku' },
    { path: `/logs/${log.id}`, title: 'Activity · Tadoku' },
    { path: `/logs/${log.id}/edit`, title: 'Edit log · Tadoku' },
    { path: `/logs/${log.id}/contests`, title: 'Submit to contests · Tadoku' },
    { path: `/logs/${log.id}/contests?preselect=1`, title: 'Submit to contests · Tadoku' },
    { path: '/pages/manual', title: 'Guide · Tadoku' },
    { path: '/pages/dev-welcome', title: `${cmsPage.title} · Tadoku` },
    { path: '/blog', title: 'Blog · page 1 · Tadoku', paged: '/blog' },
    { path: '/blog/2', title: 'Blog · page 2 · Tadoku', paged: '/blog', page: 2 },
    { path: `/blog/posts/${post.slug}`, title: `${post.title} · Blog · Tadoku` },
    { path: `${profile}/statistics`, title: `Profile statistics · ${name} · Tadoku` },
    { path: `${profile}/statistics/${year}`, title: `Profile statistics · ${name} · Tadoku` },
    { path: `${profile}/updates`, title: `Profile updates · ${name} · Tadoku`, paged: `${profile}/updates` },
    { path: `${profile}/updates/2`, title: `Profile updates · ${name} · Tadoku`, paged: `${profile}/updates`, page: 2 },
    { path: `${profile}/updates/1?date=${today}`, title: `Profile updates · ${name} · Tadoku` },
  ]

  // Bookmarks open without JavaScript, so the server's answer is what each one shows.
  const context = await browser.newContext({ baseURL: webv3, javaScriptEnabled: false })
  const page = await context.newPage()
  for (const route of routes) {
    const response = await page.goto(route.path)
    expect(response?.status(), route.path).toBe(200)
    expect(await page.title(), route.path).toBe(route.title)
    await expect(page.getByRole('heading', { level: 1 }), route.path).toHaveCount(1)
    if (!route.paged) continue

    // Page links keep Webv2's shape, and a later page links back to the first.
    const pageLinks = page.getByRole('navigation', { name: /pages/ }).getByRole('link')
    for (const href of await pageLinks.evaluateAll(links => links.map(link => link.getAttribute('href')))) {
      expect(href, `${route.path} links ${href}`).toMatch(new RegExp(`^${route.paged.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/\\d+$`))
    }
    if (route.page) {
      const first = page.getByRole('main').locator(`a[href="${route.paged}/1"], a[href="${route.paged}"]`)
      await expect(first.first(), `${route.path} is a later page, which links back to the first`).toBeAttached()
    }
  }
  await page.screenshot({ path: testInfo.outputPath('last-route.png'), fullPage: true })

  // Webv2's permanent redirect and the sign-in redirects of session-only pages.
  expect(await redirect(request, '/contests')).toEqual({ status: 308, location: `${webv3}/contests/official` })
  for (const path of [`${c}/registration`, `${c}/scoring`]) {
    const { status, location } = await redirect(request, path)
    expect(status, path).toBe(307)
    expect(new URL(location).searchParams.get('return_to'), path).toBe(`${webv3}${path}`)
  }
  await context.close()
})

const trailingSlashPaths = async (request: APIRequestContext) => {
  const { id } = await apiJson<{ id: string }>(request, '/immersion/contests/latest-official')
  return ['/blog/', '/contests/official/?ref=bookmark', `/contests/${id}/leaderboard/2/`]
}

test('a bookmark with a trailing slash redirects to the page without it', async ({ request }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  for (const path of await trailingSlashPaths(request)) {
    const { location } = await redirect(request, path)
    expect(location, path).toBe(`${webv3}${path.replace(/\/(\?|$)/, '$1')}`)
  }
})

test('a trailing-slash redirect is permanent, as in Webv2', async ({ request }) => {
  test.skip(!webv3 || !api || !reader.email, 'Set development URLs and fixture credentials')
  const paths = await trailingSlashPaths(request)
  const statuses = await Promise.all(paths.map(async path => (await redirect(request, path)).status))
  expect(statuses).toEqual(paths.map(() => 308))
})
