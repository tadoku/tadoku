import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { defineConfig } from '@playwright/test'

const artifacts = process.env.PLAYWRIGHT_OUTPUT_DIR || join(tmpdir(), 'tadoku-webv3-e2e')

export default defineConfig({
  testDir: './e2e',
  workers: 1,
  fullyParallel: false,
  reporter: [
    ['list'],
    ['html', { outputFolder: join(artifacts, 'report'), open: 'never' }],
    ['json', { outputFile: join(artifacts, 'results.json') }],
  ],
  outputDir: join(artifacts, 'results'),
  use: {
    baseURL: process.env.WEBV3_BASE_URL || 'http://127.0.0.1:5183',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    // A branch's Vite server sends a page as about 200 module requests, and the development ingress closes an HTTP/2
    // connection after 1,000 requests, failing whichever are in flight and leaving that page unable to hydrate.
    launchOptions: { args: ['--disable-http2'] },
  },
})
