import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  workers: 1,
  timeout: 120_000,
  expect: { timeout: 15_000 },
  use: {
    actionTimeout: 15_000,
    navigationTimeout: 60_000,
    baseURL: process.env.AUTH_URL,
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
  },
})
