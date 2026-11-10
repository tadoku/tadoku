import { expect, test } from './logging-fixtures'

const assets: [string, RegExp][] = [
  ['/favicon.svg', /^image\/svg\+xml/],
  ['/favicon.ico', /^image\/(x-icon|vnd\.microsoft\.icon)/],
  ['/favicon.png', /^image\/png/],
  ['/apple-touch-icon.png', /^image\/png/],
  ['/site.webmanifest', /^application\/(manifest\+)?json/],
]

test('pages offer the logomark favicon, its fallbacks and a manifest', async ({ page, request }) => {
  await page.goto('/pages/manual')
  await expect(page.locator('link[rel="icon"][type="image/svg+xml"]')).toHaveAttribute('href', '/favicon.svg')
  await expect(page.locator('link[rel="icon"][type="image/png"]')).toHaveAttribute('href', '/favicon.png')
  await expect(page.locator('link[rel="apple-touch-icon"]')).toHaveAttribute('href', '/apple-touch-icon.png')
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', '/site.webmanifest')
  await expect(page.locator('meta[name="theme-color"]')).toHaveCount(0)

  for (const [path, type] of assets) {
    const response = await request.get(path)
    expect(response.status(), path).toBe(200)
    expect(response.headers()['content-type'], path).toMatch(type)
  }
  const manifest = await (await request.get('/site.webmanifest')).json() as { name: string; icons: { src: string }[] }
  expect(manifest.name).toBe('Tadoku')
  expect(manifest.icons.map(icon => icon.src)).toEqual(expect.arrayContaining(['/favicon.svg', '/apple-touch-icon.png']))
})

test('the favicon draws the logomark dark on a light appearance and light on a dark one', async ({ page, request }, testInfo) => {
  // Fetched as an image, as a browser fetches an icon; the development server answers an HTML navigation itself.
  const svg = await (await request.get('/favicon.svg', { headers: { Accept: 'image/svg+xml' } })).text()
  for (const [scheme, fill] of [['light', 'rgb(33, 26, 29)'], ['dark', 'rgb(255, 254, 250)']] as const) {
    await page.emulateMedia({ colorScheme: scheme })
    await page.setContent(`<div style="width:128px;padding:16px;background:${scheme === 'dark' ? '#1c1c1c' : '#f4f4f4'}">${svg}</div>`)
    await expect(page.locator('path[mask]')).toHaveCSS('fill', fill)
    await page.locator('div').screenshot({ path: testInfo.outputPath(`favicon-${scheme}.png`) })
  }
})
