const fs = require('node:fs')
const path = require('node:path')
const { generateKeyPairSync } = require('node:crypto')
const Y = require('../../../../frontend/node_modules/yaml')

function proveIdentityProxy({ root, evidence, prefix, containers, report, run }, docs, productionValues) {
  const base = docs.find(d => d.kind === 'ConfigMap' && d.data?.['oathkeeper.yaml'])
  const values = productionValues && Y.parse(fs.readFileSync(productionValues, 'utf8')).oathkeeper
  const config = values ? values.config : Y.parse(base.data['oathkeeper.yaml'])
  const rules = Y.parse(values ? values.accessRules : base.data['rules.yaml'])
    .filter(rule => ['tadoku:identity:api', 'tadoku:content:api', 'tadoku:immersion:feature-flags:internal'].includes(rule.id))
  const fixture = `${prefix}-identity-fixture`
  const proxy = `${prefix}-oathkeeper`
  const dir = path.join(evidence, 'identity-proxy')
  fs.mkdirSync(dir)
  const { privateKey } = generateKeyPairSync('rsa', { modulusLength: 2048 })
  const jwk = { ...privateKey.export({ format: 'jwk' }), kid: 'disposable-e2e', alg: 'RS256', use: 'sig' }
  fs.writeFileSync(path.join(dir, 'jwks.json'), JSON.stringify({ keys: [jwk] }))
  for (const rule of rules) rule.upstream.url = `http://${fixture}:8080`
  config.access_rules.repositories = ['file:///etc/config/rules.yaml']
  config.authenticators.cookie_session.config.check_session_url = `http://${fixture}:8080/sessions/whoami`
  config.mutators.id_token.config.jwks_url = 'file:///etc/config/jwks.json'
  if (values) {
    run('openssl', ['req', '-x509', '-newkey', 'rsa:2048', '-nodes', '-days', '1', '-subj', `/CN=${proxy}`, '-addext', `subjectAltName=DNS:${proxy}`, '-keyout', path.join(dir, 'tls.key'), '-out', path.join(dir, 'tls.crt')])
    fs.chmodSync(path.join(dir, 'tls.key'), 0o644)
    config.serve.proxy.tls = { key: { path: '/etc/config/tls.key' }, cert: { path: '/etc/config/tls.crt' } }
  }
  fs.writeFileSync(path.join(dir, 'rules.yaml'), Y.stringify(rules))
  fs.writeFileSync(path.join(dir, 'oathkeeper.yaml'), Y.stringify(config))
  fs.copyFileSync(path.join(root, 'k8s/dev/base/e2e/identity-proxy-fixture.cjs'), path.join(dir, 'fixture.cjs'))
  const image = docs.find(d => d.kind === 'Deployment' && d.metadata.name === 'oathkeeper').spec.template.spec.containers[0].image
  const fixtureImage = process.env.TADOKU_BASE_AUTHV2_IMAGE || docs.find(d => d.kind === 'Deployment' && d.metadata.name === 'frontend-authv2').spec.template.spec.containers[0].image
  run('docker', ['pull', image])
  report.images[image] = JSON.parse(run('docker', ['image', 'inspect', image, '--format', '{{json .RepoDigests}}']))
  for (const [name, selectedImage, command] of [
    [fixture, fixtureImage, ['--entrypoint', 'node']],
    [proxy, image, []],
  ]) {
    run('docker', ['create', '--name', name, '--network', prefix, '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '256m', ...command, selectedImage, ...(name === fixture ? ['/etc/config/fixture.cjs'] : ['serve', '--config', '/etc/config/oathkeeper.yaml'])])
    containers.push(name)
    run('docker', ['cp', dir, `${name}:/etc/config`])
    run('docker', ['start', name])
  }
  try {
    const result = run('docker', ['exec', fixture, 'node', '/etc/config/fixture.cjs', 'check', proxy, values ? 'production' : 'development'])
    report.identityProxy = JSON.parse(result)
  } finally {
    for (const name of [proxy, fixture]) run('docker', ['logs', name])
  }
}

module.exports = { proveIdentityProxy }
