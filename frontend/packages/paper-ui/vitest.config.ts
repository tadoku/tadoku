import stylex from '@stylexjs/unplugin'
import { defineConfig } from 'vitest/config'
import { resolve } from 'node:path'

export default defineConfig({
  // The plugin's dev-server CSS endpoint would keep Vitest's server open, so tests use only its transform. Its own
  // resolver reads catalogue examples' `paper-ui/tokens.stylex` from source, so tests don't need a build.
  plugins: [{
    ...stylex.vite({
      aliases: { 'paper-ui/tokens.stylex': [resolve('src/tokens/tokens.stylex.ts')] },
      devMode: 'css-only',
      styleResolution: 'application-order',
    }),
    configureServer: undefined,
  }],
  resolve: {
    alias: [
      { find: /^paper-ui$/, replacement: resolve('src/index.ts') },
      { find: /^paper-ui\/icons$/, replacement: resolve('src/icons/index.ts') },
      { find: /^paper-ui\/tokens\.stylex$/, replacement: resolve('src/tokens/tokens.stylex.ts') },
    ],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    css: true,
  },
})
