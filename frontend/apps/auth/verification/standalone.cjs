// Proves runtime endpoint selection in actual built Next standalone servers; API responses are fixtures.
const fs = require('node:fs')
const path = require('node:path')
const http = require('node:http')
const { spawn, execFileSync } = require('node:child_process')
const root = path.resolve(__dirname, '../../..')
const evidence = path.resolve(process.argv[2])
fs.mkdirSync(evidence, { recursive: true })
const report = { revision: execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(), cases: [], boundary: 'Built SSR runtime with loopback fixture, no real identity authentication or gateway' }
const cookie = 'ory_kratos_session=canonical-fixture; tadoku_session_proof-0123abcd=branch-fixture; unrelated=fixture'
const guard = path.join(evidence, 'deny-external.cjs')
fs.writeFileSync(guard, `const original=global.fetch;global.fetch=(url,...args)=>{if(new URL(String(url)).hostname!=='127.0.0.1')return Promise.reject(new Error('External network blocked by standalone verification'));return original(url,...args)};`)
const request = (port, headers = {}, route = '/') => new Promise((resolve, reject) => {
  const req = http.get({ hostname: '127.0.0.1', port, path: route, headers, timeout: 5000 }, res => {
    let body = ''
    res.on('data', chunk => { body += chunk })
    res.on('end', () => resolve({ status: res.statusCode, headers: res.headers, body }))
  })
  req.on('timeout', () => req.destroy(new Error('SSR request timed out')))
  req.on('error', reject)
})
async function main() {
  const observations = []
  let app
  const api = http.createServer((req, res) => {
    observations.push({ app, path: req.url, cookieExact: req.headers.cookie === cookie, branch: req.headers['x-dev-branch'] })
    res.setHeader('Content-Type', 'application/json')
    res.end(JSON.stringify({ account_app: app === 'auth' ? 'v2' : 'v1', account: null }))
  })
  await new Promise(resolve => api.listen(0, '127.0.0.1', resolve))
  try {
    for (app of ['webv2', 'admin', 'auth']) {
      const server = path.join(root, 'apps', app, '.next/standalone/apps', app, 'server.js')
      const log = fs.openSync(path.join(evidence, `${app}-server.log`), 'w')
      const child = spawn(process.execPath, [server], {
        cwd: path.dirname(server),
        env: { ...process.env, PORT: '43276', HOSTNAME: '127.0.0.1', NEXT_SERVER_API_ENDPOINT: `http://127.0.0.1:${api.address().port}/api/internal`, NODE_OPTIONS: `--require ${guard}` },
        stdio: ['ignore', log, log],
      })
      try {
        let response
        const headers = { Cookie: cookie, 'X-Dev-Branch': 'proof-0123abcd', Host: app === 'webv2' ? 'proof-0123abcd.preview.tadoku.app' : `proof-0123abcd.${app === 'auth' ? 'account' : 'admin'}.preview.tadoku.app` }
        for (let i = 0; i < 100; i++) {
          try { response = await request(43276, headers, app === 'auth' ? '/login?return_to=https%3A%2F%2Fproof-0123abcd.preview.tadoku.app%2Flogs%2Fnew' : app === 'admin' ? '/users' : '/'); break } catch (error) {
            if (child.exitCode !== null || i === 99) throw error
            await new Promise(resolve => setTimeout(resolve, 50))
          }
        }
        fs.writeFileSync(path.join(evidence, `${app}.html`), response.body)
        const seen = observations.filter(row => row.app === app)
        const passed = seen.length > 0 && seen.every(row => row.path === '/api/internal/identity/session' && row.cookieExact && row.branch === 'proof-0123abcd') && !response.body.includes('Session temporarily unavailable') && (app !== 'auth' || response.status === 307)
        report.cases.push({ app, status: response.status, unavailable: response.body.includes('Session temporarily unavailable'), observedPrivateRequests: seen, passed })
      } finally {
        child.kill('SIGTERM')
        await new Promise(resolve => { if (child.exitCode !== null) resolve(); else child.once('exit', resolve) })
        fs.closeSync(log)
      }
    }
    report.result = report.cases.every(row => row.passed) ? 'passed' : 'failed'
    if (report.result !== 'passed') process.exitCode = 1
  } catch (error) { report.result = 'failed'; report.error = error.message; process.exitCode = 1 }
  finally {
    await new Promise(resolve => api.close(resolve))
    fs.writeFileSync(path.join(evidence, 'report.json'), JSON.stringify(report, null, 2))
    console.log(report.result, path.join(evidence, 'report.json'))
  }
}
main()
