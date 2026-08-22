import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      '@app': fileURLToPath(new URL('./app', import.meta.url)),
      '@pages': fileURLToPath(new URL('./pages', import.meta.url)),
    },
  },
})
