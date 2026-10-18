// Runs commands with per-step evidence logs and removes only this run's labelled Docker fixtures.
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')
const root = path.resolve(__dirname, '../../../..')

function createHarness(evidence) {
  fs.mkdirSync(evidence, { recursive: true })
  const prefix = `tdk-base-proof-${process.pid}-${Date.now().toString(36)}`
  const containers = []
  let networkCreated = false
  let sequence = 0
  const report = { command: process.argv.join(' '), revision: '', images: {}, steps: [], unverified: ['Argo hook sequencing and failure gating', 'live ingress, login, overlays and Image Updater write-back'] }

  function run(cmd, args, input, expectFailure = false) {
    const result = spawnSync(cmd, args, { cwd: root, input, encoding: 'utf8', timeout: 300000, maxBuffer: 16 * 1024 * 1024 })
    const log = `${String(++sequence).padStart(3, '0')}-${cmd}.log`
    fs.writeFileSync(path.join(evidence, log), (result.stdout || '') + (result.stderr || ''))
    report.steps.push({ command: [cmd, ...args], exit: result.status, log, expectedFailure: expectFailure })
    if (result.error || result.status === null || (expectFailure ? result.status === 0 : result.status !== 0)) {
      throw new Error(`Unexpected result for ${cmd}; inspect ${log}: ${result.error || result.status}`)
    }
    return result.stdout.trim()
  }

  function createNetwork() {
    run('docker', ['network', 'create', '--label', `tadoku.dev/test=${prefix}`, prefix])
    networkCreated = true
  }

  function cleanup() {
    for (const name of containers.reverse()) {
      try {
        const fixture = run('docker', ['ps', '-a', '--no-trunc', '--filter', `name=^/${name}$`, '--format', '{{.ID}}|{{.Label "tadoku.dev/test"}}'])
        if (!fixture) {
          if (name === report.cleanupControl?.removedFixture) report.cleanupControl.result = 'passed'
          continue
        }
        const [id, owner] = fixture.split('|')
        if (owner !== prefix || !/^[0-9a-f]{64}$/.test(id)) throw new Error(`Refusing cleanup of ${name}`)
        run('docker', ['rm', '-f', '-v', id])
      } catch (error) { report.cleanupError = error.message; process.exitCode = 1 }
    }
    if (networkCreated) {
      try { run('docker', ['network', 'rm', prefix]) }
      catch (error) { report.cleanupError = error.message; process.exitCode = 1 }
    }
    fs.writeFileSync(path.join(evidence, 'report.json'), JSON.stringify(report, null, 2))
    console.log(`${report.result}: ${path.join(evidence, 'report.json')}`)
    if (report.error) console.error(report.error)
  }

  return { root, evidence, prefix, containers, report, run, createNetwork, cleanup }
}

module.exports = { createHarness }
