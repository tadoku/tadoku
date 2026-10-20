import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  testMatch: 'identity-resilience.spec.ts',
  workers: 1,
  retries: 0,
  timeout: 60_000,
  outputDir: process.env.IDENTITY_UI_EVIDENCE,
  reporter: [['list'], ['json', { outputFile: `${process.env.IDENTITY_UI_EVIDENCE}/results.json` }]],
  use: {
    trace: 'retain-on-failure',
    launchOptions: {
      executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH,
      args: process.env.PLAYWRIGHT_LAUNCH_ARGS?.split(' '),
    },
  },
})
