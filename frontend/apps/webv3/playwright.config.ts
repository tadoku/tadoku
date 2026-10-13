import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { defineConfig } from '@playwright/test'

const artifacts = process.env.PLAYWRIGHT_OUTPUT_DIR || join(tmpdir(), 'tadoku-webv3-e2e')

export default defineConfig({
  testDir: './e2e',
  workers: 1,
  fullyParallel: false,
  reporter: [['list'], ['html', { outputFolder: join(artifacts, 'report'), open: 'never' }]],
  outputDir: join(artifacts, 'results'),
  use: {
    baseURL: process.env.WEBV3_BASE_URL || 'http://127.0.0.1:5183',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
})
