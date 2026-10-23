// Disposable HTTP boundaries: a banned session callback and an echo API, never shared providers.
const assert = require('node:assert/strict')
const http = require('node:http')
const https = require('node:https')
const { createPublicKey, verify } = require('node:crypto')
const { readFileSync } = require('node:fs')

if (process.argv[2] !== 'check') {
  const calls = { session: 0, upstream: 0 }
  http.createServer((req, res) => {
    res.setHeader('Content-Type', 'application/json')
    if (req.url === '/metrics') return res.end(JSON.stringify(calls))
    if (req.url === '/sessions/whoami') {
      calls.session++
      res.writeHead(403).end(JSON.stringify({ error: 'synthetic banned session' }))
      return
    }
    calls.upstream++
    res.end(JSON.stringify({ method: req.method, path: req.url, cookie: req.headers.cookie, authorization: req.headers.authorization }))
  }).listen(8080, '0.0.0.0')
} else {
  check().catch(error => { console.error(error); process.exitCode = 1 })
}

async function check() {
  const proxy = process.argv[3]
  const production = process.argv[4] === 'production'
  const publicKey = createPublicKey({ key: JSON.parse(readFileSync('/etc/config/jwks.json')).keys[0], format: 'jwk' })
  for (let i = 0; ; i++) {
    try { if ((await fetch(`http://${proxy}:4456/health/ready`)).ok) break } catch {}
    assert.ok(i < 60, 'Oathkeeper must become ready before testing requests')
    await new Promise(resolve => setTimeout(resolve, 250))
  }
  const metrics = () => fetch('http://127.0.0.1:8080/metrics').then(res => res.json())
  const cookie = 'ory_kratos_session=synthetic-banned-user; tadoku_session_tadoku_proof_0123abcd=synthetic-identity-session; unrelated=preserved'
  const send = (host, path, method = 'GET', withCookie = true, internal = false) => new Promise((resolve, reject) => {
    const req = (production ? https : http).request({ hostname: proxy, port: 4455, path, method,
      ...(production ? { ca: readFileSync('/etc/config/tls.crt'), servername: proxy } : {}), headers: {
      Host: host, Accept: 'application/json', ...(withCookie ? { Cookie: cookie } : {}),
      'X-Dev-Branch': 'proof-0123abcd', ...(production && !internal ? { 'X-Forwarded-Proto': 'https' } : {}),
    } }, res => {
      let body = ''
      res.on('data', chunk => { body += chunk })
      res.on('end', () => resolve({ status: res.statusCode, body }))
    })
    req.setTimeout(5000, () => req.destroy(new Error('proxy request timed out')))
    req.on('error', reject)
    req.end()
  })
  const hosts = production
    ? ['tadoku.app', 'account.tadoku.app', 'authv2.tadoku.app', 'proof-0123abcd.preview.tadoku.app', 'proof-0123abcd.account.preview.tadoku.app', 'proof-0123abcd.authv2.preview.tadoku.app']
    : ['tadoku.dev.lab', 'account.tadoku.dev.lab', 'proof-0123abcd.account.tadoku.dev.lab', 'authv2.tadoku.dev.lab', 'proof-0123abcd.authv2.tadoku.dev.lab']
  const cases = []
  for (const host of hosts) {
    for (const [method, endpoint] of [['GET', 'session'], ['POST', 'logout']]) {
      const before = await metrics()
      const response = await send(host, `/api/internal/identity/${endpoint}`, method)
      const after = await metrics()
      assert.equal(response.status, 200, `${host} ${endpoint}: synthetic banned cookie must reach identity upstream; callback calls=${after.session - before.session}, upstream calls=${after.upstream - before.upstream}`)
      const upstream = JSON.parse(response.body)
      assert.equal(upstream.cookie, cookie, 'original Cookie must reach the identity API unchanged')
      assert.equal(upstream.path, `/identity/${endpoint}`)
      assert.equal(upstream.method, method)
      const [header, payload, signature] = upstream.authorization.replace(/^Bearer /, '').split('.')
      assert.ok(verify('RSA-SHA256', Buffer.from(`${header}.${payload}`), publicKey, Buffer.from(signature, 'base64url')), 'Oathkeeper must sign the forwarded JWT')
      const claims = JSON.parse(Buffer.from(payload, 'base64url'))
      assert.equal(claims.sub, 'guest')
      assert.equal(claims.tenant, 'tadoku/proof-0123abcd')
      assert.equal(claims.type, 'user')
      assert.equal((await metrics()).session, before.session, 'identity must not invoke the global session/ban callback')
      cases.push({ host, method, endpoint, status: response.status, cookie: 'byte-equal', subject: claims.sub, sessionChecks: 0 })
    }
  }
  const before = await metrics()
  const blocked = await send(hosts[0], '/api/internal/content/pages')
  assert.equal(blocked.status, 401, 'unrelated cookie-authenticated routes must retain the rejected-session gate')
  assert.equal((await metrics()).session, before.session + 1)
  assert.equal((await metrics()).upstream, before.upstream, 'blocked request must not reach the API')
  assert.equal((await send(hosts[0], '/api/internal/content/pages', 'GET', false)).status, 200)
  if (!production) {
    const host = 'oathkeeper-proxy.tdk-dev-oathkeeper:4455'
    assert.equal((await send(host, '/api/internal/immersion/feature-flags', 'GET', false, true)).status, 401, 'session-only route must still reject guests')
    assert.equal((await send(host, '/api/internal/immersion/feature-flags', 'GET', true, true)).status, 401, 'session-only route must still reject the banned fixture')
  }
  console.log(JSON.stringify({ environment: process.argv[4], cases, unrelated: { banned: 401, anonymous: 200, sessionOnly: production ? 'not tested over fixture TLS listener' : 'guest and banned both 401' }, fixture: 'synthetic HTTP session rejection and echo API; real Oathkeeper and signing; production rules use a verified disposable TLS listener; no real identity backend or providers' }))
}
