// This file doesn't work with ES6 modules, so we need to disable the linter for it
/* eslint-disable @typescript-eslint/no-var-requires */
const express = require('express')
const next = require('next')
const proxy = require('http-proxy-middleware')
const publicRuntimeConfig = require('./config')

const port = parseInt(process.env.PORT, 10) || 3000
const dev = process.env.NODE_ENV !== 'production'
const config = { publicRuntimeConfig }
const app = next({ dev, config })
const handle = app.getRequestHandler()

app.prepare().then(() => {
  const server = express()

  server.use(
    '/api',
    proxy.createProxyMiddleware({
      target: process.env.API_ROOT,
      changeOrigin: true,
      pathRewrite: { '^/api': '' },
    }),
  )

  server.all('*', (req, res) => {
    return handle(req, res)
  })

  server.listen(port, err => {
    if (err) {
      throw err
    }

    console.log(`> Ready on http://localhost:${port}`)
  })
})
