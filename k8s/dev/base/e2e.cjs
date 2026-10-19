// Real-image migration and frontend startup proof; no cluster or mocks.
// Run: node k8s/dev/base/e2e.cjs /tmp/tadoku-base-evidence
const fs = require('node:fs')
const path = require('node:path')
const Y = require('../../../frontend/node_modules/yaml')
const { createHarness } = require('./e2e/harness.cjs')
const { checkConfiguration, checkWorkerMetadata } = require('./e2e/configuration.cjs')
const { proveMigrations } = require('./e2e/migrations.cjs')
const { checkAuthv2 } = require('./e2e/authv2.cjs')
const { proveFrontends } = require('./e2e/frontends.cjs')
const { proveIdentityProxy } = require('./e2e/oathkeeper.cjs')
const evidence = path.resolve(process.argv[2] || `/tmp/tadoku-base-evidence-${process.pid}`)
const harness = createHarness(evidence)
const { containers, prefix, report, run } = harness

try {
  report.revision = run('git', ['rev-parse', 'HEAD'])
  report.worktree = run('git', ['status', '--short'])
  const rendered = run('kubectl', ['--context', 'homelab-talos-dev', 'kustomize', 'k8s/dev/base'])
  fs.writeFileSync(path.join(evidence, 'rendered.yaml'), rendered)
  report.renderedSHA256 = require('node:crypto').createHash('sha256').update(rendered).digest('hex')
  const docs = Y.parseAllDocuments(rendered).map(d => d.toJSON())
  checkConfiguration(harness, docs)
  checkAuthv2(harness, docs)
  const db = proveMigrations(harness, docs)
  proveIdentityProxy(harness, docs)
  checkWorkerMetadata(harness)
  proveFrontends(harness, docs)
  const removedFixture = containers.find(name => name !== db)
  const fixtureOwner = run('docker', ['inspect', removedFixture, '--format', '{{index .Config.Labels "tadoku.dev/test"}}'])
  if (fixtureOwner !== prefix) throw new Error('Refusing removal of cleanup control fixture')
  run('docker', ['rm', removedFixture])
  report.cleanupControl = { removedFixture, expected: 'already absent owned fixture is clean' }
  report.result = 'passed'
} catch (error) {
  report.result = 'failed'
  report.error = error.message
  process.exitCode = 1
} finally {
  harness.cleanup()
}
