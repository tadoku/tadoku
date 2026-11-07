import stylex from '@stylexjs/unplugin'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  // Under Vitest the plugin's dev-server CSS endpoint would keep the server open; tests need only its transform.
  plugins: [{ ...stylex.vite({ styleResolution: 'application-order', useCSSLayers: true }), ...(process.env.VITEST ? { configureServer: undefined } : {}) }, react()],
  resolve: {
    dedupe: ['react', 'react-dom', '@tanstack/react-form'],
  },
  build: {
    sourcemap: true,
  },
  server: {
    host: process.env.HOST || '0.0.0.0',
    port: process.env.PORT ? Number(process.env.PORT) : 5174,
    allowedHosts: true,
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    css: true,
  },
})
