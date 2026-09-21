// Development deployment adapter for the existing Next standalone GHCR images.
// server.js freezes build-time URLs; retain its build config but override only
// the development runtime endpoints. No frontend rebuild or source mutation.
const { readFileSync } = require('node:fs')
const { createRequire } = require('node:module')
const path = require('node:path')

const dir = process.env.NEXT_APP_DIR
if (!dir || !dir.startsWith('/app/apps/')) throw new Error('NEXT_APP_DIR is required')
process.chdir(dir)
process.env.NODE_ENV = 'production'
const config = JSON.parse(readFileSync(path.join(dir, '.next/required-server-files.json'))).config
config.publicRuntimeConfig = {
  ...config.publicRuntimeConfig,
  kratosPublicEndpoint: 'https://account.tadoku.dev.lab/kratos',
  kratosInternalEndpoint: 'http://kratos-public.tdk-dev-kratos',
  authUiUrl: 'https://account.tadoku.dev.lab',
  homeUrl: 'https://tadoku.dev.lab',
  adminUrl: 'https://admin.tadoku.dev.lab',
  apiEndpoint: 'https://tadoku.dev.lab/api/internal',
  cookieDomain: '.tadoku.dev.lab',
  cookieSecure: true,
}
config.serverRuntimeConfig = {
  ...config.serverRuntimeConfig,
  apiEndpoint: 'https://tadoku.dev.lab/api/internal',
}
process.env.__NEXT_PRIVATE_STANDALONE_CONFIG = JSON.stringify(config)
const appRequire = createRequire(path.join(dir, 'server.js'))
appRequire('next')
appRequire('next/dist/server/lib/start-server').startServer({
  dir, isDev: false, config, hostname: '0.0.0.0', port: 3000,
  allowRetry: false, useWorkers: true,
}).catch(error => { console.error(error); process.exit(1) })
