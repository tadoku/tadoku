import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  esbuild: { jsx: 'automatic' },
  resolve: {
    alias: {
      '@app': fileURLToPath(new URL('./app', import.meta.url)),
      '@pages': fileURLToPath(new URL('./pages', import.meta.url)),
    },
  },
  test: {
    exclude: ['**/.next/**', '**/node_modules/**'],
    setupFiles: ['./vitest.setup.ts'],
  },
})
