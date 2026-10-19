import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { defineConfig } from '@playwright/test'

const artifacts = process.env.PLAYWRIGHT_OUTPUT_DIR || join(tmpdir(), 'tadoku-authv2-e2e')

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: { timeout: 1_000 },
  workers: 1,
  fullyParallel: false,
  reporter: [['list'], ['html', { outputFolder: join(artifacts, 'report'), open: 'never' }]],
  outputDir: join(artifacts, 'results'),
  use: {
    baseURL: process.env.AUTHV2_BASE_URL || 'http://127.0.0.1:5184',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
})
