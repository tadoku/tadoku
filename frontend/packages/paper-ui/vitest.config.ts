import { defineConfig } from 'vitest/config'
import { resolve } from 'node:path'

export default defineConfig({
  resolve: {
    alias: [
      { find: /^paper-ui$/, replacement: resolve('src/index.ts') },
      { find: /^paper-ui\/icons$/, replacement: resolve('src/icons/index.ts') },
    ],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    css: true,
  },
})
