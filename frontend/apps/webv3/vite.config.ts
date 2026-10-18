import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import react from '@vitejs/plugin-react'
import { nitro } from 'nitro/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [tanstackStart({ serverFns: { disableCsrfMiddlewareWarning: true } }), nitro({ serverDir: './server', preset: 'node-server' }), react()],
  resolve: {
    dedupe: ['react', 'react-dom', '@tanstack/react-query'],
  },
  ssr: {
    external: ['react', 'react-dom', 'react-dom/server', 'react/jsx-runtime'],
    noExternal: ['@tanstack/react-query', '@tanstack/react-router-with-query'],
  },
  server: {
    host: process.env.HOST || '0.0.0.0',
    port: process.env.PORT ? Number(process.env.PORT) : 5183,
    strictPort: true,
    allowedHosts: ['localhost', '.code.apps.lab', 'next.tadoku.dev.lab', '.next.tadoku.dev.lab'],
  },
})
