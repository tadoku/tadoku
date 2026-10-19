const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const Y = require('../../../../frontend/node_modules/yaml')

function checkAuthv2({ root, report }, docs) {
  const find = (kind, name) => docs.find(doc => doc.kind === kind && doc.metadata.name === name)
  const deployment = find('Deployment', 'frontend-authv2')
  assert.ok(deployment, 'authv2 must have an always-on base Deployment')
  assert.equal(deployment.metadata.namespace, 'tdk-dev-frontend-authv2')
  const pod = deployment.spec.template.spec
  const container = pod.containers[0]
  assert.equal(pod.automountServiceAccountToken, false)
  assert.equal(pod.securityContext.runAsUser, 1000)
  assert.equal(container.securityContext.readOnlyRootFilesystem, true)
  assert.equal(container.command, undefined)
  for (const probe of ['readinessProbe', 'livenessProbe']) assert.equal(container[probe].httpGet.path, '/healthz')
  const runtime = find('ConfigMap', 'tadoku-cli-authv2')
  assert.equal(runtime.data.TADOKU_WEB_URL, 'https://tadoku.dev.lab')
  assert.equal(runtime.data.TADOKU_API_URL, 'https://tadoku.dev.lab/api/internal')
  const route = find('HTTPRoute', 'tadoku-cli-authv2-base')
  assert.deepEqual(route.spec.hostnames, ['authv2.tadoku.dev.lab'])
  assert.ok(route.spec.rules[0].filters.some(f => f.requestHeaderModifier?.remove.includes('x-dev-branch')))
  assert.equal(route.spec.rules[0].backendRefs[0].name, 'tadoku-cli-authv2')
  assert.equal(find('Backend', 'tadoku-cli-authv2').spec.endpoints[0].fqdn.hostname, 'frontend-authv2.tdk-dev-frontend-authv2.svc.cluster.local')
  assert.deepEqual(find('Ingress', 'tadoku-authv2').spec.tls[0].hosts, ['authv2.tadoku.dev.lab'])
  assert.ok(find('HTTPRoute', 'tadoku-cli-proxy-base').spec.hostnames.includes('authv2.tadoku.dev.lab'))
  const config = Y.parse(fs.readFileSync(path.join(root, '.dev/config.yaml'), 'utf8'))
  assert.ok(config.publicHosts.includes('authv2.tadoku.dev.lab'))
  assert.ok(config.namespaces.includes(deployment.metadata.namespace))
  const rule = Y.parse(fs.readFileSync(path.join(root, 'k8s/dev/base/oathkeeper/rules.yaml'), 'utf8')).find(r => r.id === 'tadoku:identity:api')
  assert.ok(rule, 'identity requests must pass through Oathkeeper')
  assert.deepEqual(rule.authenticators.map(a => a.handler), ['cookie_session', 'anonymous'])
  assert.equal(rule.mutators[0].handler, 'id_token')
  report.authv2 = { host: route.spec.hostnames[0], image: container.image, health: '/healthz', baseBranchHeader: 'removed' }
}

module.exports = { checkAuthv2 }
