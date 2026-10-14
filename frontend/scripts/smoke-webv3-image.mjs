import assert from 'node:assert/strict'
import { spawn, execFileSync } from 'node:child_process'
import { createServer } from 'node:http'
import { mkdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'

const image = process.argv[2]
assert(image, 'Usage: node scripts/smoke-webv3-image.mjs <image>')
const artifacts = resolve(process.env.PLAYWRIGHT_OUTPUT_DIR || '/tmp/tadoku-webv3-image')
await mkdir(artifacts, { recursive: true })
const contest = {
  id: '11111111-1111-4111-8111-111111111111', title: 'Container fixture contest',
  contest_start: '2020-01-01', contest_end: '2099-12-31', registration_end: '2099-12-31',
  private: false, official: true,
  allowed_languages: process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? [{ code: 'jpn', name: 'Japanese' }] : null,
  language_code_allow_list: process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? ['jpn'] : null,
  allowed_activities: [], activity_type_id_allow_list: [],
}
const writes = []
const fixture = createServer((request, response) => {
  if (request.method !== 'GET') {
    writes.push(`${request.method} ${request.url}`)
    response.writeHead(405).end()
    return
  }
  const path = new URL(request.url, 'http://fixture').pathname
  let data
  if (path.endsWith('/latest-official')) data = contest
  else if (path.endsWith('/summary')) data = { participant_count: 1, language_count: 1, total_score: 13319.6455 }
  else if (path.endsWith('/leaderboard')) data = {
    entries: [{ rank: 1, user_id: contest.id, user_display_name: 'Fixture learner', score: 13319.6455, is_tie: false }],
    total_size: 1, next_page_token: '',
  }
  else if (path === '/immersion/contests') data = { contests: [contest], total_size: 1, next_page_token: '' }
  if (!data) { response.writeHead(404).end(); return }
  response.writeHead(200, { 'Content-Type': 'application/json' }).end(JSON.stringify(data))
})
await new Promise(resolve => fixture.listen(0, '0.0.0.0', resolve))
const apiPort = fixture.address().port
const docker = (...args) => execFileSync('docker', args, { encoding: 'utf8' }).trim()
async function waitForHealth(base) {
  for (let attempt = 0; attempt < 60; attempt++) {
    try {
      if ((await fetch(`${base}/healthz`, { signal: AbortSignal.timeout(2_000) })).ok) return
    } catch {}
    await new Promise(resolve => setTimeout(resolve, 500))
  }
  assert.fail('Standalone container did not become healthy')
}

async function checkRuntimeLinks(base, cases) {
  return Promise.all([...cases, ...cases].map(async ([host, origin]) => {
    const response = await fetch(`${base}/plan`, {
      headers: { Host: host, 'X-Forwarded-Host': 'attacker.invalid', 'X-Dev-Branch': 'attacker' },
      signal: AbortSignal.timeout(15_000),
    })
    assert.equal(response.status, 200, `Goal SSR failed for ${host}`)
    const html = await response.text()
    assert(html.includes(`href="${origin}/logs/new"`), `Wrong activity link for ${host}`)
    assert(html.includes(`href="${origin}/pages/manual"`), `Wrong guide link for ${host}`)
    for (const [, href] of html.matchAll(/href="(https:\/\/[^" ]+)"/g)) {
      const url = new URL(href)
      if (url.hostname.endsWith('.tadoku.app') || url.hostname === 'tadoku.app' || url.hostname.endsWith('.dev.lab')) {
        assert.equal(url.origin, origin, `Cross-request product URL leaked on ${host}`)
      }
    }
    await writeFile(`${artifacts}/links-${host}.html`, html)
    return { host, origin, outcome: 'passed' }
  }))
}

let container
let developmentContainer
const runtimeLinks = []
let outcome = 'failed'
try {
  container = docker('run', '-d', '--read-only', '--cap-drop=ALL', '--security-opt=no-new-privileges',
    '--label', 'tadoku.check=webv3-image', '--add-host', 'host.docker.internal:host-gateway',
    '-p', '127.0.0.1::3000', '-e', `TADOKU_API_URL=http://host.docker.internal:${apiPort}`, image)
  const info = JSON.parse(docker('inspect', container))[0]
  assert.deepEqual(JSON.parse(docker('exec', container, 'node', '-e',
    'console.log(JSON.stringify([process.getuid(), process.getgid()]))')), [1000, 1000])
  const base = `http://127.0.0.1:${info.NetworkSettings.Ports['3000/tcp'][0].HostPort}`
  await waitForHealth(base)
  runtimeLinks.push(...await checkRuntimeLinks(base, [
    ['next.tadoku.app', 'https://tadoku.app'],
    ['alice.next.preview.tadoku.app', 'https://alice.preview.tadoku.app'],
    ['bob.next.preview.tadoku.app', 'https://bob.preview.tadoku.app'],
  ]))

  developmentContainer = docker('run', '-d', '--read-only', '--cap-drop=ALL', '--security-opt=no-new-privileges',
    '--label', 'tadoku.check=webv3-runtime-links', '-p', '127.0.0.1::3000',
    '-e', 'TADOKU_WEB_URL=https://tadoku.dev.lab', '-e', 'TADOKU_API_URL=https://tadoku.dev.lab/api/internal', image)
  const developmentInfo = JSON.parse(docker('inspect', developmentContainer))[0]
  const developmentBase = `http://127.0.0.1:${developmentInfo.NetworkSettings.Ports['3000/tcp'][0].HostPort}`
  await waitForHealth(developmentBase)
  runtimeLinks.push(...await checkRuntimeLinks(developmentBase, [
    ['next.tadoku.dev.lab', 'https://tadoku.dev.lab'],
    ['alice.next.tadoku.dev.lab', 'https://alice.tadoku.dev.lab'],
    ['bob.next.tadoku.dev.lab', 'https://bob.tadoku.dev.lab'],
  ]))
  const html = await (await fetch(base, { signal: AbortSignal.timeout(15_000) })).text()
  assert(html.includes(contest.title), 'Runtime API override or SSR failed')
  assert(html.includes('https://tadoku.app'), 'Production product links missing')
  const assets = [...html.matchAll(/(?:src|href)="([^" ]+\.(?:js|css)(?:\?[^" ]*)?)"/g)]
  assert(assets.length > 0, 'SSR has no built assets')
  for (const [, asset] of assets) assert((await fetch(new URL(asset, base), { signal: AbortSignal.timeout(5_000) })).ok, `Asset missing: ${asset}`)
  const exitCode = await new Promise((resolve, reject) => {
    const child = spawn('pnpm', ['--filter', 'webv3', 'test:e2e'], {
      stdio: 'inherit', env: { ...process.env, WEBV3_BASE_URL: base,
        TADOKU_API_URL: `http://127.0.0.1:${apiPort}`, PLAYWRIGHT_OUTPUT_DIR: artifacts },
    })
    child.once('error', reject)
    child.once('exit', resolve)
  })
  assert.equal(exitCode, 0, 'Browser journeys failed')
  assert.deepEqual(writes, [], 'Application attempted upstream writes')
  await new Promise(resolve => fixture.close(resolve))
  const health = await fetch(`${base}/healthz`, { signal: AbortSignal.timeout(2_000) })
  assert.equal(health.status, 200, 'API outage broke health')
  assert.equal(await health.text(), 'ok\n')
  outcome = 'passed'
} finally {
  if (developmentContainer) {
    await writeFile(`${artifacts}/development-container.log`, docker('logs', developmentContainer))
    docker('rm', '-f', developmentContainer)
  }
  if (container) {
    await writeFile(`${artifacts}/container.log`, docker('logs', container))
    docker('rm', '-f', container)
  }
  fixture.close()
  await writeFile(`${artifacts}/result.json`, JSON.stringify({ outcome, image, nodeVersion: process.version,
    revision: execFileSync('git', ['rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(),
    fixture: `Synthetic local read-only contest API (${process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? 'array' : 'null'} language lists); real API/auth/deployment unverified`,
    command: `node scripts/smoke-webv3-image.mjs ${image}`, writes, runtimeLinks }, null, 2))
}
