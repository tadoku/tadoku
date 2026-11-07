import type { APIRequestContext, BrowserContext, Page } from '@playwright/test'
import { admin, api, expect, test, webv3 } from './logging-fixtures'

interface Post { slug: string; title: string; content: string; published_at: string }
interface Posts { posts: Post[]; total_size: number }

const pageSize = 10
const seededPage = 'dev-welcome'
// Stays in the admin's branch tenant and is reused by later runs; it keeps one element of each kind DOMPurify removes.
const hostilePage = {
  slug: 'webv3-e2e-sanitized',
  title: 'Webv3 sanitized page',
  html: '<p id="kept">Kept <strong>bold</strong> and <a href="/blog/1">a link</a>.</p>'
    + '<img src="data:," alt="" onerror="window.__xss=1"><script>window.__xss=2</script>'
    + '<a id="unsafe" href="javascript:window.__xss=3">unsafe</a>',
}

async function apiJson<T>(request: APIRequestContext, path: string): Promise<T> {
  const response = await request.get(`${api}${path}`)
  expect(response.ok(), `${path} answered ${response.status()}`).toBeTruthy()
  return response.json() as Promise<T>
}

const posts = (request: APIRequestContext, page = 1) =>
  apiJson<Posts>(request, `/content/posts/tadoku?page_size=${pageSize}&page=${page - 1}`)

/** The text a reader sees in HTML, parsed inertly in the page without scripts or styles. */
const visibleText = (page: Page, html: string) => page.evaluate(source => {
  const template = document.createElement('template')
  template.innerHTML = source
  template.content.querySelectorAll('script, style').forEach(element => element.remove())
  return (template.content.textContent ?? '').replace(/\s+/g, ' ').trim()
}, html)

const prose = (page: Page) => page.locator('.content-prose')

async function expectSanitized(page: Page) {
  await expect(prose(page).locator('script, [onerror], a[href^="javascript:"]')).toHaveCount(0)
  expect(await page.evaluate(() => (window as { __xss?: number }).__xss)).toBeUndefined()
}

async function expectContained(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
}

/** One page of the blog: page 1 leads with the newest post, the rest are listed with their titles. */
async function expectBlogPage(page: Page, data: Posts, number: number) {
  const listed = number === 1 ? data.posts.slice(1) : data.posts
  if (number === 1) await expect(page.locator('.blog-lead h2')).toHaveText(data.posts[0].title)
  else await expect(page.locator('.blog-lead')).toHaveCount(0)
  await expect(page.locator('.blog-posts a')).toHaveText(listed.map(post => post.title))
  await expect(page.getByRole('navigation', { name: 'Blog pages' })).toHaveCount(data.total_size > pageSize ? 1 : 0)
}

async function hydrated(page: Page) {
  await expect(page.locator('.paper-navbar__mobile-trigger')).toBeEnabled()
}

/** Gives the admin's branch tenant more than one page of published posts, a draft post and page, and the hostile page. */
async function ensureContent(context: BrowserContext) {
  const created = (response: { status(): number }) => [201, 409].includes(response.status())
  const { total_size: total } = await posts(context.request)
  const run = Date.now()
  for (let index = total; index <= pageSize; index++) {
    const response = await context.request.post(`${api}/content/posts/tadoku`, {
      data: {
        slug: `webv3-e2e-${run}-${index}`,
        title: `Webv3 blog check ${index + 1}`,
        content: `## Check ${index + 1}\n\nCreated by the **Webv3** content journey.`,
        published_at: new Date(Date.now() - 60_000).toISOString(),
      },
    })
    expect(response.status(), 'post creation').toBe(201)
  }
  const draftPost = await context.request.post(`${api}/content/posts/tadoku`, {
    data: { slug: 'webv3-e2e-draft-post', title: 'Webv3 draft post', content: 'Not published.' },
  })
  const draftPage = await context.request.post(`${api}/content/pages/tadoku`, {
    data: { slug: 'webv3-e2e-draft-page', title: 'Webv3 draft page', html: '<p>Not published.</p>' },
  })
  const hostile = await context.request.post(`${api}/content/pages/tadoku`, {
    data: { ...hostilePage, published_at: new Date(Date.now() - 60_000).toISOString() },
  })
  expect([draftPost, draftPage, hostile].map(created)).toEqual([true, true, true])
}

test.beforeEach(() => {
  test.skip(!webv3 || !api, 'Set WEBV3_BASE_URL and TADOKU_API_URL')
})

test('the manual and its navigation are server-rendered and fit a phone', async ({ browser }, testInfo) => {
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()
  const response = await page.goto('/pages/manual')
  expect(response?.status()).toBe(200)

  await expect(page).toHaveTitle('Manual · Tadoku')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Manual')
  await expect(page.getByRole('heading', { level: 2, name: 'Signing up for a contest' })).toBeVisible()
  for (const activity of ['Reading', 'Writing', 'Listening', 'Speaking', 'Study']) {
    await expect(page.getByRole('region', { name: activity }).getByRole('table')).toBeVisible()
  }
  await expect(page.getByRole('region', { name: 'Reading' }).getByRole('row', { name: /^Japanese/ })).toHaveText(/Japanese\s*1\s*1\.6\s*0\.2\s*0\.05\s*0\.0025/)

  const header = page.getByRole('banner')
  const main = header.getByRole('navigation', { name: 'Main navigation' })
  // Webv2's main navigation plus the guide.
  for (const [name, href] of [
    ['Home', '/'],
    ['Leaderboard', '/leaderboard/latest/1'],
    ['Contests', '/contests/official/1'],
    ['Blog', '/blog/1'],
    ['Guide', '/pages/manual'],
  ]) {
    await expect(main.getByRole('link', { name, exact: true }).first()).toHaveAttribute('href', href)
  }
  const footer = page.getByRole('navigation', { name: 'Footer' })
  const community = page.getByRole('contentinfo').getByRole('list', { name: 'Community' })
  // Webv2's footer destinations, in its order, with its community links beside them.
  await expect(footer.getByRole('link')).toHaveText([
    'Home', 'Leaderboard', 'Contests', 'Blog', 'About Tadoku', 'Guide', 'Rules', 'Questions & help', 'Privacy', 'Contact',
  ])
  await expect(community.getByRole('link')).toHaveCount(3)
  for (const [name, href] of [
    ['Home', '/'],
    ['Leaderboard', '/leaderboard/latest/1'],
    ['Contests', '/contests/official/1'],
    ['Blog', '/blog/1'],
    ['About Tadoku', '/pages/about'],
    ['Guide', '/pages/manual'],
    ['Rules', '/pages/rules'],
    ['Questions & help', '/pages/faq'],
    ['Privacy', '/pages/privacy'],
    ['Contact', '/pages/contact'],
  ]) {
    await expect(footer.getByRole('link', { name, exact: true })).toHaveAttribute('href', href)
  }
  for (const [name, href] of [
    ['Discord', 'https://discord.gg/AsC9vZs2Ex'],
    ['GitHub', 'https://github.com/tadoku'],
    ['Twitter', 'https://twitter.com/tadoku_app'],
  ]) {
    await expect(community.getByRole('link', { name, exact: true })).toHaveAttribute('href', href)
  }
  await page.screenshot({ path: testInfo.outputPath('manual-1280.png'), fullPage: true })

  await page.setViewportSize({ width: 390, height: 900 })
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('manual-390.png'), fullPage: true })
  await context.close()
})

test('a CMS page shows the API content after sanitizing it, and unknown or missing slugs are not found', async ({ page, request }, testInfo) => {
  const cms = await apiJson<{ title: string; html: string }>(request, `/content/pages/tadoku/${seededPage}`)
  const response = await page.goto(`/pages/${seededPage}`)
  expect(response?.status()).toBe(200)
  await expect(page).toHaveTitle(`${cms.title} · Tadoku`)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(cms.title)
  expect((await prose(page).textContent())!.replace(/\s+/g, ' ').trim()).toBe(await visibleText(page, cms.html))
  await expectSanitized(page)
  await hydrated(page)
  await expectSanitized(page)
  await page.screenshot({ path: testInfo.outputPath('cms-page-1280.png'), fullPage: true })

  await page.setViewportSize({ width: 390, height: 900 })
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('cms-page-390.png'), fullPage: true })

  for (const path of [`/pages/webv3-e2e-missing-${Date.now()}`, `/blog/posts/webv3-e2e-missing-${Date.now()}`]) {
    const missing = await page.goto(path)
    expect(missing?.status(), path).toBe(404)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Page not found')
  }
})

test('the blog lists the API posts, pages through them and opens a post', async ({ page, request }, testInfo) => {
  const first = await posts(request)
  await page.goto('/blog')
  await expect(page).toHaveTitle('Blog · page 1 · Tadoku')
  if (first.total_size === 0) {
    await expect(page.getByRole('main').getByRole('heading', { level: 2 })).toHaveText('No blog posts yet')
    return
  }
  await expectBlogPage(page, first, 1)
  await page.screenshot({ path: testInfo.outputPath('blog-1280.png'), fullPage: true })
  await hydrated(page)

  let opened = first.posts[0]
  if (first.total_size > pageSize) {
    await page.getByRole('navigation', { name: 'Blog pages' }).getByRole('link', { name: 'Next page' }).click()
    await expect(page).toHaveURL(/\/blog\/2$/)
    const second = await posts(request, 2)
    await expectBlogPage(page, second, 2)
    opened = second.posts[0]
    await page.locator('.blog-posts a', { hasText: opened.title }).first().click()
  } else {
    await page.locator('.blog-lead h2 a').click()
  }

  await expect(page).toHaveURL(new RegExp(`/blog/posts/${opened.slug}$`))
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(opened.title)
  await expect(page).toHaveTitle(`${opened.title} · Blog · Tadoku`)
  await expect(prose(page)).not.toBeEmpty()
  await expectSanitized(page)
  await page.screenshot({ path: testInfo.outputPath('blog-post-1280.png'), fullPage: true })
  await page.setViewportSize({ width: 390, height: 900 })
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('blog-post-390.png'), fullPage: true })

  await page.getByRole('link', { name: 'Blog', exact: true }).first().click()
  await expect(page).toHaveURL(/\/blog\/1$/)
  await expectContained(page)
  await page.screenshot({ path: testInfo.outputPath('blog-390.png'), fullPage: true })

  const past = Math.ceil(first.total_size / pageSize) + 1
  await page.goto(`/blog/${past}`)
  await expect(page.getByRole('main').getByRole('heading', { level: 2 })).toHaveText(`There is no page ${past}`)
  await expect(page.locator('.blog-posts, .blog-lead')).toHaveCount(0)
})

test('published content an administrator adds appears, sanitized and paginated, while drafts stay hidden', async ({ adminContext, browser }, testInfo) => {
  test.skip(!admin.email || !admin.password, 'Set E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD')
  await ensureContent(adminContext)
  const context = await browser.newContext()
  const page = await context.newPage()
  const first = await posts(context.request)
  expect(first.total_size).toBeGreaterThan(pageSize)

  await page.goto('/blog/1')
  await expectBlogPage(page, first, 1)
  await page.goto('/blog/2')
  await expectBlogPage(page, await posts(context.request, 2), 2)
  await expect(page.getByText('Webv3 draft post')).toHaveCount(0)

  for (const draft of ['/blog/posts/webv3-e2e-draft-post', '/pages/webv3-e2e-draft-page']) {
    expect((await page.goto(draft))?.status(), draft).toBe(404)
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Page not found')
  }

  expect((await page.goto(`/pages/${hostilePage.slug}`))?.status()).toBe(200)
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(hostilePage.title)
  await expect(prose(page).locator('#kept')).toHaveText('Kept bold and a link.')
  await expect(prose(page).locator('#kept a')).toHaveAttribute('href', '/blog/1')
  await hydrated(page)
  await expectSanitized(page)
  await page.screenshot({ path: testInfo.outputPath('cms-sanitized.png'), fullPage: true })
  await context.close()
})

test('active announcements show above pages until this browser dismisses them', async ({ page, request }, testInfo) => {
  test.skip(!api, 'Set TADOKU_API_URL')
  const { announcements } = await apiJson<{ announcements: { id: string; href?: string | null }[] }>(
    request,
    '/content/announcements/tadoku/active',
  )
  test.skip(announcements.length === 0, 'The selected tenant has no active announcement')

  await page.goto('/blog/1')
  const banners = page.locator('main .paper-flash')
  await expect(banners).toHaveCount(announcements.length)
  const first = banners.first()
  if (announcements[0].href) await expect(first.getByRole('link', { name: 'Learn more' })).toHaveAttribute('href', announcements[0].href)
  await page.screenshot({ path: testInfo.outputPath('announcement.png') })

  await first.getByRole('button', { name: 'Dismiss announcement' }).click()
  await expect(banners).toHaveCount(announcements.length - 1)
  expect(JSON.parse((await page.evaluate(() => localStorage.getItem('dismissed_announcements')))!)).toEqual([announcements[0].id])
  await page.reload()
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
  await expect(banners).toHaveCount(announcements.length - 1)
})
