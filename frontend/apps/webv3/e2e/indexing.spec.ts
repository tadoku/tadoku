import { get } from 'node:http'
import type { APIRequestContext, Page } from '@playwright/test'
import { accountId, admin, api, expect, test, webv3 } from './logging-fixtures'

const base = new URL(webv3 || 'http://127.0.0.1:5183')
/** The web host this deployment's pages belong on: `next.` dropped from a Webv3 host, or `WEBV3_WEB_URL`. */
const canonical = new URL(process.env.WEBV3_WEB_URL || base.origin.replace(/\/\/(.+\.)?next\./, '//$1'))
const onWebHost = base.host === canonical.host

async function apiJson<T>(request: APIRequestContext, path: string): Promise<T> {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), path).toBe(true)
  return response.json() as Promise<T>
}

const meta = (page: Page, key: string) => page.locator(`meta[name="${key}"], meta[property="${key}"]`).first()

/** A request to the deployment under test as if it were the web host, which only a plain-HTTP server can take. */
function asWebHost(path: string) {
  return new Promise<{ status?: number; robotsTag?: string | string[]; body: string }>((resolve, reject) => {
    get(new URL(path, base), { headers: { Host: canonical.host } }, response => {
      let body = ''
      response.setEncoding('utf8')
      response.on('data', chunk => { body += chunk })
      response.on('end', () => resolve({ status: response.statusCode, robotsTag: response.headers['x-robots-tag'], body }))
    }).on('error', reject)
  })
}

test('pages send security headers and a new CSP nonce, report-only, with every response', async ({ request }) => {
  const nonces = []
  for (let attempt = 0; attempt < 2; attempt++) {
    const response = await request.get('/pages/manual')
    const headers = response.headers()
    const html = await response.text()
    // A development ingress may replace it with its own, longer policy.
    expect(headers['strict-transport-security']).toMatch(/^max-age=\d{7,}/)
    expect(headers['x-content-type-options']).toBe('nosniff')
    expect(headers['referrer-policy']).toBe('strict-origin-when-cross-origin')
    expect(headers['content-security-policy']).toBe("frame-ancestors 'none'")

    const nonce = html.match(/<meta property="csp-nonce" content="([A-Za-z0-9+/=_-]{16,})"/)?.[1]
    expect(nonce, 'csp-nonce meta').toBeTruthy()
    expect(html, 'the theme bootstrap carries the nonce').toContain(`<script nonce="${nonce}">try{`)
    expect(headers['content-security-policy-report-only']).toContain(`script-src 'self' 'nonce-${nonce}'`)
    expect(headers['content-security-policy-report-only']).toContain("frame-ancestors 'none'")
    nonces.push(nonce)
  }
  expect(nonces[0]).not.toBe(nonces[1])
})

test('every page names its canonical URL on the web host, and other hosts are not indexed', async ({ page, request }) => {
  const response = await page.goto('/pages/manual?ref=share#units-reading')
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', `${canonical.origin}/pages/manual`)
  await expect(meta(page, 'og:url')).toHaveAttribute('content', `${canonical.origin}/pages/manual`)
  if (!onWebHost) expect(response?.headers()['x-robots-tag']).toContain('noindex')

  const robots = await request.get('/robots.txt')
  expect(robots.status()).toBe(200)
  expect(robots.headers()['content-type']).toContain('text/plain')
  if (!onWebHost) expect(await robots.text()).toMatch(/^User-agent: \*\nDisallow: \/\n$/)
})

test('the web host is indexable and lets crawlers in', async () => {
  test.skip(base.protocol !== 'http:' || onWebHost, 'Needs a plain-HTTP server that is not the web host, such as the image smoke')
  const page = await asWebHost('/pages/manual')
  expect(page.status).toBe(200)
  expect(page.robotsTag).toBeUndefined()
  const robots = await asWebHost('/robots.txt')
  expect(robots.body).toMatch(/^User-agent: \*\nDisallow:\n$/)
})

test('blog posts, CMS pages, contests and profiles describe themselves to search results and link previews', async ({ page, request }) => {
  test.skip(!api, 'Set TADOKU_API_URL')
  const contest = await apiJson<{ id: string }>(request, '/immersion/contests/latest-official')
  const { entries } = await apiJson<{ entries: { user_id: string }[] }>(request, `/immersion/contests/${contest.id}/leaderboard?page_size=1&page=0`)
  const { posts } = await apiJson<{ posts: { slug: string }[] }>(request, '/content/posts/tadoku?page_size=1&page=0')
  const paths = ['/pages/dev-welcome', `/contests/${contest.id}/leaderboard/1`]
  if (posts[0]) paths.push(`/blog/posts/${posts[0].slug}`)
  if (entries[0]) paths.push(`/profile/${entries[0].user_id}/statistics`)

  for (const path of paths) {
    await page.goto(path)
    await expect(meta(page, 'description'), path).toHaveAttribute('content', /\S/)
    await expect(meta(page, 'og:title'), path).toHaveAttribute('content', /\S/)
    const description = (await meta(page, 'description').getAttribute('content'))!
    const title = (await meta(page, 'og:title').getAttribute('content'))!
    expect(description.length, path).toBeLessThanOrEqual(160)
    await expect(meta(page, 'og:description'), path).toHaveAttribute('content', description)
    await expect(meta(page, 'og:url'), path).toHaveAttribute('content', `${canonical.origin}${path}`)
    await expect(meta(page, 'twitter:card'), path).toHaveAttribute('content', 'summary')
    await expect(meta(page, 'twitter:title'), path).toHaveAttribute('content', title)
    await expect(meta(page, 'twitter:description'), path).toHaveAttribute('content', description)
    await expect(meta(page, 'robots'), path).toHaveCount(0)
  }
})

test('private and action pages are not indexed', async ({ page }) => {
  for (const path of ['/logs/new', '/contests/new', '/contests/joined/1', '/contests/my-contests/1']) {
    await page.goto(path)
    await expect(meta(page, 'robots'), path).toHaveAttribute('content', 'noindex')
  }
})

test('a private contest’s standings, activity and participant profiles are not indexed', async ({ adminContext }) => {
  test.skip(!api || !admin.email, 'Set TADOKU_API_URL and the Dev Admin fixture account')
  const day = (offset: number) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10)
  const created = await adminContext.request.post(`${api}/immersion/contests`, {
    data: {
      title: `Webv3 private indexing check ${Date.now().toString(36)}`,
      description: 'Created by the Webv3 indexing journey.',
      contest_start: day(0),
      contest_end: day(30),
      registration_end: day(14),
      private: true,
      official: false,
      language_code_allow_list: null,
      activity_type_id_allow_list: [1],
    },
  })
  expect(created.ok(), `contest creation answered ${created.status()}`).toBe(true)
  const { id } = await created.json() as { id: string }
  const joined = await adminContext.request.post(`${api}/immersion/contests/${id}/registration`, { data: { language_codes: ['jpn'] } })
  expect(joined.ok(), `registration answered ${joined.status()}`).toBe(true)

  const page = await adminContext.newPage()
  for (const path of [`/contests/${id}/leaderboard/1`, `/contests/${id}/updates/1`, `/contests/${id}/profile/${await accountId(adminContext)}`]) {
    await page.goto(path)
    await expect(meta(page, 'robots'), path).toHaveAttribute('content', 'noindex')
  }
})
