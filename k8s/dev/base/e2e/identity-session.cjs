// Disposable callback stand-in: proves pinned Oathkeeper configuration, not API authentication.
// node k8s/dev/base/e2e/identity-session.cjs <evidence-dir> [tadoku-argocd-checkout]
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')
const http = require('node:http')
const Y = require('../../../../frontend/node_modules/yaml')
const { createHarness } = require('./harness.cjs')
const harness = createHarness(path.resolve(process.argv[2]))
const { evidence, prefix, report, containers, run } = harness
const dummy = 'disposable-identity-callback-credential'
const fixtures = [
  { name: 'dev', image: 'oryd/oathkeeper:v26.2.0', config: Y.parse(fs.readFileSync('k8s/dev/base/oathkeeper/oathkeeper.yaml', 'utf8')), rules: Y.parse(fs.readFileSync('k8s/dev/base/oathkeeper/rules.yaml', 'utf8')), env: Y.parse(fs.readFileSync('k8s/dev/base/oathkeeper/deployment.yaml', 'utf8')).spec.template.spec.containers[0].env },
]
if (process.argv[3]) {
  const values = Y.parse(fs.readFileSync(path.join(process.argv[3], 'platform/oathkeeper/values.yaml'), 'utf8'))
  const rendered = Y.parseAllDocuments(harness.run('helm', ['template', 'oathkeeper', 'ory/oathkeeper', '--version', '0.60.1', '--namespace', 'tdk-prod-oathkeeper', '-f', path.join(process.argv[3], 'platform/oathkeeper/values.yaml')])).map(doc => doc.toJSON())
  const deployment = rendered.find(doc => doc.kind === 'Deployment')
  const configMap = rendered.find(doc => doc.kind === 'ConfigMap' && doc.data?.['config.yaml'])
  const container = deployment.spec.template.spec.containers[0]
  fixtures.push({ name: 'prod', image: container.image, config: Y.parse(configMap.data['config.yaml']), rules: Y.parse(values.oathkeeper.accessRules), env: container.env })
}
const cases = [
  { name: 'guest', subject: 'guest' },
  { name: 'unrelated', cookie: 'analytics=unrelated', subject: 'guest' },
  { name: 'canonical', cookie: 'ory_kratos_session=canonical', subject: 'canonical-user' },
  { name: 'branch-and-canonical', branch: 'proof-0123abcd', cookie: 'ory_kratos_session=canonical; tadoku_session_proof-0123abcd=branch', subject: 'branch-user' },
  { name: 'branch-with-canonical-only', branch: 'proof-0123abcd', cookie: 'ory_kratos_session=canonical', subject: 'guest' },
  { name: 'callback-failure', cookie: 'failure=yes', status: 401 },
]
function request(port, method, route, headers) {
  return new Promise((resolve, reject) => {
    const req = http.request({ hostname: '127.0.0.1', port, path: route, method, headers }, res => {
      let body = ''
      res.on('data', chunk => { body += chunk })
      res.on('end', () => resolve({ status: res.statusCode, body: JSON.parse(body) }))
    })
    req.on('error', reject)
    req.end()
  })
}
async function main() {
  try {
    report.revision = run('git', ['rev-parse', 'HEAD'])
    report.worktree = run('git', ['status', '--short'])
    report.unverified = ['Actual A callback/session/logout/ban runtime', 'live branch ingress, trusted routing and isolated database', 'shared provider sessions']
    harness.createNetwork()
    fs.writeFileSync(path.join(evidence, 'stub.cjs'), `
const http=require('node:http');
http.createServer((req,res)=>{
 const callback=req.url==='/identity/internal/v1/session';
 const auth=req.headers.authorization==='Bearer ${dummy}';
 const branch=req.headers['x-dev-branch']||''; const cookie=req.headers.cookie||'';
 if(callback){
  console.log(JSON.stringify({method:req.method,auth,cookie,branch}));
  if(!auth||req.method!=='GET'||cookie.includes('failure=yes')){res.statusCode=401;return res.end('{}')}
  const id=branch?(cookie.includes('tadoku_session_'+branch+'=branch')?'branch-user':'guest'):(cookie.includes('ory_kratos_session=canonical')?'canonical-user':'guest');
  res.setHeader('content-type','application/json');return res.end(JSON.stringify({identity:{id},marker:'extra-preserved'}));
 }
 res.setHeader('content-type','application/json');res.end(JSON.stringify({method:req.method,subject:req.headers['x-subject'],extra:req.headers['x-extra'],cookie:req.headers.cookie||'',branch:req.headers['x-dev-branch']||''}));
}).listen(8080);
`)
    const stub = `${prefix}-callback`
    containers.push(stub)
    run('docker', ['run', '-d', '--name', stub, '--label', `tadoku.dev/test=${prefix}`, '--network', prefix, '-v', `${evidence}:/proof:ro`, 'node:22-alpine', 'node', '/proof/stub.cjs'])
    report.matrix = []
    report.logChecks = []
    for (const fixture of fixtures) {
      const config = structuredClone(fixture.config)
      config.serve = { proxy: { port: 4455 }, api: { port: 4456 } }
      config.access_rules = { matching_strategy: 'glob', repositories: ['file:///proof/rules.yaml'] }
      config.authenticators.cookie_session.config.check_session_url = `http://${stub}:8080/identity/internal/v1/session`
      config.mutators = { header: { enabled: true, config: { headers: { 'X-Subject': '{{ print .Subject }}', 'X-Extra': '{{ .Extra | toJson }}' } } } }
      config.authorizers = { allow: { enabled: true } }
      const rules = [{ id: 'configuration-proof', upstream: { url: `http://${stub}:8080` }, match: { url: 'http://<**>/protected', methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] }, authenticators: [{ handler: 'cookie_session' }, { handler: 'anonymous' }], authorizer: { handler: 'allow' }, mutators: [{ handler: 'header' }] }]
      const internal = structuredClone(fixture.rules.find(rule => rule.id === 'tadoku:identity:session:internal'))
      internal.upstream.url = `http://${stub}:8080`
      internal.mutators = [{ handler: 'header' }]
      rules.push(internal)
      fs.writeFileSync(path.join(evidence, 'rules.yaml'), Y.stringify(rules))
      fs.writeFileSync(path.join(evidence, `${fixture.name}.yaml`), Y.stringify(config))
      const name = `${prefix}-${fixture.name}`
      if (process.env.IDENTITY_CALLBACK_CONTROL === 'literal') {
        config.authenticators.cookie_session.config.additional_headers = { Authorization: `Bearer ${dummy}` }
        fs.writeFileSync(path.join(evidence, `${fixture.name}.yaml`), Y.stringify(config))
      }
      const env = fixture.env.filter(e => e.name === 'AUTHENTICATORS_COOKIE_SESSION_CONFIG_ADDITIONAL_HEADERS').flatMap(e => ['-e', `${e.name}=${e.value.replaceAll('$(OATHKEEPER_AUTHZ_TOKEN)', dummy)}`])
      report.images[fixture.image] = run('docker', ['image', 'inspect', fixture.image, '--format', '{{json .RepoDigests}}'])
      containers.push(name)
      run('docker', ['run', '-d', '--name', name, '--label', `tadoku.dev/test=${prefix}`, '--network', prefix, '-p', '127.0.0.1::4455', '-p', '127.0.0.1::4456', '-v', `${evidence}:/proof:ro`, '-e', `OATHKEEPER_AUTHZ_TOKEN=${dummy}`, ...env, fixture.image, 'serve', '-c', `/proof/${fixture.name}.yaml`])
      const port = run('docker', ['port', name, '4455']).split(':').pop()
      const apiPort = run('docker', ['port', name, '4456']).split(':').pop()
      for (let i = 0; ; i++) {
        try { if ((await fetch(`http://127.0.0.1:${apiPort}/health/ready`)).ok) break } catch {}
        if (i === 100) { run('docker', ['logs', name]); throw new Error(`${fixture.name} startup failed`) }
        await new Promise(resolve => setTimeout(resolve, 100))
      }
      for (const method of ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']) for (const test of cases) {
        const headers = { Host: 'localhost', ...(test.cookie ? { Cookie: test.cookie } : {}), ...(test.branch ? { 'X-Dev-Branch': test.branch } : {}) }
        const res = await request(port, method, '/protected', headers)
        const body = res.body
        const expectedStatus = test.status || 200
        const passed = res.status === expectedStatus && (test.status || (body.method === method && body.subject === test.subject && body.cookie === (test.cookie || '') && body.branch === (test.branch || '') && (!test.cookie || JSON.parse(body.extra).marker === 'extra-preserved')))
        report.matrix.push({ image: fixture.image, method, case: test.name, expectedStatus, status: res.status, passed: !!passed })
      }
      const internalHost = new URL(internal.match.url).host
      for (const cookie of ['failure=yes', 'unrelated=yes', 'ory_kratos_session=canonical; tadoku_session_proof-0123abcd=branch']) {
        const res = await request(port, 'GET', '/api/internal/identity/session', { Host: internalHost, Cookie: cookie, 'X-Dev-Branch': 'proof-0123abcd' })
        report.matrix.push({ image: fixture.image, case: 'internal-session-anonymous', status: res.status, passed: res.status === 200 && res.body.subject === 'guest' && res.body.cookie === cookie && res.body.branch === 'proof-0123abcd' })
      }
      run('docker', ['logs', name])
      // docker writes logs on stderr too; scan the complete captured output.
      const logFile = report.steps.at(-1).log
      const complete = fs.readFileSync(path.join(evidence, logFile), 'utf8')
      const result = spawnSync('docker', ['logs', stub], { encoding: 'utf8' })
      const callbackLines = (result.stdout + result.stderr).trim().split('\n').filter(Boolean).map(line => JSON.parse(line))
      const check = { image: fixture.image, secretAbsent: !complete.includes(dummy), callbackCount: callbackLines.length, allCallbacksAuthenticated: callbackLines.length > 0 && callbackLines.every(line => line.auth), allCallbacksGET: callbackLines.every(line => line.method === 'GET') }
      report.logChecks.push(check)
    }
    report.result = report.matrix.every(row => row.passed) && report.logChecks.every(row => row.secretAbsent && row.allCallbacksAuthenticated && row.allCallbacksGET) ? 'passed' : 'failed'
    if (report.result !== 'passed') process.exitCode = 1
  } catch (error) { report.result = 'failed'; report.error = error.message; process.exitCode = 1 }
  finally { harness.cleanup() }
}
main()
