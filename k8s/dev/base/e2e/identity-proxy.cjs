// Focused disposable proxy proof; optional third argument is the GitOps Oathkeeper values file.
const path = require('node:path')
const Y = require('../../../../frontend/node_modules/yaml')
const { createHarness } = require('./harness.cjs')
const { proveIdentityProxy } = require('./oathkeeper.cjs')
const harness = createHarness(path.resolve(process.argv[2]))
try {
  harness.report.revision = harness.run('git', ['rev-parse', 'HEAD'])
  harness.report.worktree = harness.run('git', ['status', '--short'])
  const rendered = harness.run('kubectl', ['--context', 'homelab-talos-dev', 'kustomize', 'k8s/dev/base'])
  harness.createNetwork()
  proveIdentityProxy(harness, Y.parseAllDocuments(rendered).map(d => d.toJSON()), process.argv[3])
  harness.report.result = 'passed'
} catch (error) {
  harness.report.result = 'failed'
  harness.report.error = error.message
  process.exitCode = 1
} finally { harness.cleanup() }
