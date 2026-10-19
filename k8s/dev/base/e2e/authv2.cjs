const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const Y = require('../../../../frontend/node_modules/yaml')

function checkAuthv2({ root, report }, docs) {
  const find = (kind, name) => docs.find(doc => doc.kind === kind && doc.metadata.name === name)
  const deployment = find('Deployment', 'frontend-authv2')
  assert.ok(deployment, 'authv2 must have an always-on base Deployment')
  const auth = find('Deployment', 'frontend-auth')
  assert.equal(deployment.metadata.namespace, auth.metadata.namespace, 'both account apps must reuse the existing auth namespace')
  const namespace = deployment.metadata.namespace
  const identities = docs.map(doc => `${doc.kind}/${doc.metadata.namespace || ''}/${doc.metadata.name}`)
  assert.equal(new Set(identities).size, identities.length, 'rendered resources must not collide')
  const service = find('Service', 'frontend-authv2')
  const matches = (selector, labels) => Object.entries(selector).every(([key, value]) => labels[key] === value)
  for (const app of [deployment, auth]) {
    const selected = docs.filter(doc => doc.kind === 'Deployment' && doc.metadata.namespace === namespace && matches(find('Service', app.metadata.name).spec.selector, doc.spec.template.metadata.labels))
    assert.deepEqual(selected.map(doc => doc.metadata.name), [app.metadata.name], 'each account Service must select only its own workload')
  }
  const pod = deployment.spec.template.spec
  const container = pod.containers[0]
  assert.equal(pod.automountServiceAccountToken, false)
  assert.equal(pod.securityContext.runAsUser, 1000)
  assert.equal(container.securityContext.readOnlyRootFilesystem, true)
  assert.equal(container.command, undefined)
  for (const probe of ['readinessProbe', 'livenessProbe']) assert.equal(container[probe].httpGet.path, '/healthz')
  const runtime = find('ConfigMap', 'tadoku-cli-authv2')
  assert.equal(runtime.metadata.namespace, namespace)
  for (const ref of container.envFrom) assert.ok(docs.some(doc => doc.kind === 'ConfigMap' && doc.metadata.name === ref.configMapRef.name && doc.metadata.namespace === namespace))
  const overlay = Y.parse(fs.readFileSync(path.join(root, '.dev/authv2.yaml'), 'utf8'))
  const overlayContainer = overlay.spec.containers[0]
  assert.equal(overlayContainer.envFrom, undefined, 'the overlay must not depend on unpublished runtime ConfigMaps')
  const overlayEnv = Object.fromEntries(overlayContainer.env.map(({ name, value }) => [name, value]))
  for (const name of ['TADOKU_WEB_URL', 'TADOKU_API_URL', 'NODE_EXTRA_CA_CERTS']) assert.equal(overlayEnv[name], runtime.data[name])
  for (const volume of [...pod.volumes, ...overlay.spec.volumes]) {
    const ca = docs.find(doc => doc.kind === 'ConfigMap' && doc.metadata.namespace === namespace && doc.metadata.name === volume.configMap.name)
    assert.equal(ca?.data['ca.crt'], fs.readFileSync(path.join(root, 'k8s/dev/base/lab-root-ca.crt'), 'utf8'), 'both runtimes must resolve the existing public CA')
  }
  assert.equal(runtime.data.TADOKU_WEB_URL, 'https://tadoku.dev.lab')
  assert.equal(runtime.data.TADOKU_API_URL, 'https://tadoku.dev.lab/api/internal')
  const route = find('HTTPRoute', 'tadoku-cli-authv2-base')
  assert.deepEqual(route.spec.hostnames, ['authv2.tadoku.dev.lab'])
  assert.ok(route.spec.rules[0].filters.some(f => f.requestHeaderModifier?.remove.includes('x-dev-branch')))
  assert.equal(route.spec.rules[0].backendRefs[0].name, 'tadoku-cli-authv2')
  assert.equal(service.metadata.namespace, namespace)
  assert.equal(find('Backend', 'tadoku-cli-authv2').spec.endpoints[0].fqdn.hostname, `${service.metadata.name}.${service.metadata.namespace}.svc.cluster.local`)
  assert.deepEqual(find('Ingress', 'tadoku-authv2').spec.tls[0].hosts, ['authv2.tadoku.dev.lab'])
  assert.ok(find('HTTPRoute', 'tadoku-cli-proxy-base').spec.hostnames.includes('authv2.tadoku.dev.lab'))
  const config = Y.parse(fs.readFileSync(path.join(root, '.dev/config.yaml'), 'utf8'))
  assert.ok(config.publicHosts.includes('authv2.tadoku.dev.lab'))
  assert.ok(config.namespaces.includes(deployment.metadata.namespace))
  assert.equal(new Set(config.namespaces).size, config.namespaces.length)
  for (const name of config.namespaces) assert.ok(docs.some(doc => doc.kind === 'Namespace' && doc.metadata.name === name), 'every configured namespace must be rendered by the base')
  const rule = Y.parse(fs.readFileSync(path.join(root, 'k8s/dev/base/oathkeeper/rules.yaml'), 'utf8')).find(r => r.id === 'tadoku:identity:api')
  assert.ok(rule, 'identity requests must pass through Oathkeeper')
  assert.deepEqual(rule.authenticators.map(a => a.handler), ['cookie_session', 'anonymous'])
  assert.equal(rule.mutators[0].handler, 'id_token')
  report.authv2 = { host: route.spec.hostnames[0], image: container.image, health: '/healthz', baseBranchHeader: 'removed' }
}

module.exports = { checkAuthv2 }
