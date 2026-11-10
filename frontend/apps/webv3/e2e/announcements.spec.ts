import { admin, api, expect, test } from './logging-fixtures'

test('an announcement links only to an HTTPS or same-site page, and its images load lazily', async ({ adminContext, page }, testInfo) => {
  test.skip(!api || !admin.email, 'Set TADOKU_API_URL and the Dev Admin fixture account')
  const run = Date.now().toString(36)
  const cases = [
    { content: `HTTPS link ${run}`, href: 'https://example.com/tadoku', learnMore: true },
    { content: `Same-site link ${run}`, href: '/pages/manual', learnMore: true },
    { content: `Plain HTTP link ${run}`, href: 'http://example.com/tadoku', learnMore: false },
    { content: `Image ${run} ![Fixture image](https://example.com/announcement.png)`, href: null, learnMore: false },
  ]
  const ids: string[] = []
  try {
    for (const { content, href } of cases) {
      const created = await adminContext.request.post(`${api}/content/announcements/tadoku`, {
        data: {
          title: content,
          content,
          style: 'info',
          href,
          starts_at: new Date(Date.now() - 60_000).toISOString(),
          ends_at: new Date(Date.now() + 3_600_000).toISOString(),
        },
      })
      expect(created.ok(), `announcement creation answered ${created.status()}`).toBe(true)
      ids.push((await created.json() as { id: string }).id)
    }

    await page.goto('/blog/1')
    for (const { content, href, learnMore } of cases) {
      const banner = page.locator('main .paper-flash').filter({ hasText: content.split(' ![')[0] })
      await expect(banner).toBeVisible()
      await expect(banner.getByRole('link', { name: 'Learn more' }), content).toHaveCount(learnMore ? 1 : 0)
      if (learnMore) await expect(banner.getByRole('link', { name: 'Learn more' })).toHaveAttribute('href', href!)
    }
    await expect(page.getByRole('img', { name: 'Fixture image' })).toHaveAttribute('loading', 'lazy')
    await page.screenshot({ path: testInfo.outputPath('announcements.png'), fullPage: true })
  } finally {
    for (const id of ids) await adminContext.request.delete(`${api}/content/announcements/tadoku/${id}`)
  }
})
