import assert from 'node:assert/strict'
import { spawn, execFile, execFileSync } from 'node:child_process'
import { get } from 'node:http'
import { mkdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { promisify } from 'node:util'
import { apiPort, contest, fixture, fixtureOptions, requestIds, writes } from './smoke-webv3-fixture.mjs'

const image = process.argv[2]
assert(image, 'Usage: node scripts/smoke-webv3-image.mjs <image>')
const artifacts = resolve(process.env.PLAYWRIGHT_OUTPUT_DIR || '/tmp/tadoku-webv3-image')
await mkdir(artifacts, { recursive: true })
const docker = (...args) => execFileSync('docker', args, { encoding: 'utf8' }).trim()
const run = promisify(execFile)
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
    const response = await new Promise((resolve, reject) => {
      const request = get(`${base}/pages/manual`, {
        headers: { Host: host, 'X-Forwarded-Host': 'attacker.invalid', 'X-Dev-Branch': 'attacker' },
        timeout: 15_000,
      }, response => {
        let html = ''
        response.setEncoding('utf8')
        response.on('data', chunk => { html += chunk })
        response.on('end', () => resolve({ status: response.statusCode, headers: response.headers, html }))
        response.on('error', reject)
      })
      request.on('timeout', () => request.destroy(new Error(`Guide SSR timed out for ${host}`)))
      request.on('error', reject)
    })
    assert.equal(response.status, 200, `Guide SSR failed for ${host}`)
    assert.equal(response.headers['cache-control'], 'private, no-store')
    assert(response.headers.vary.toLowerCase().includes('cookie'))
    const auth = origin.includes('.preview.tadoku.app')
      ? origin.replace('.preview.tadoku.app', '.account.preview.tadoku.app')
      : origin === 'https://tadoku.app' || origin === 'https://tadoku.dev.lab'
        ? origin.replace('https://', 'https://account.')
        : origin.replace('.tadoku.dev.lab', '.account.tadoku.dev.lab')
    const login = `${auth}/login?return_to=${encodeURIComponent(`https://${host}/pages/manual`)}`
    const html = response.html
    assert(html.includes(`href="${login}"`), `Wrong login return for ${host}`)
    await writeFile(`${artifacts}/links-${host}.html`, html)
    const register = `${auth}/register?return_to=${encodeURIComponent(`https://${host}/pages/manual`)}`
    assert(html.includes(`href="${register}"`), `Wrong sign-up return for ${host}`)
    assert(html.includes(`href="/pages/manual"`), `Wrong guide link for ${host}`)
    for (const [, href] of html.matchAll(/href="(https:\/\/[^" ]+)"/g)) {
      const url = new URL(href)
      if (url.hostname.endsWith('.tadoku.app') || url.hostname === 'tadoku.app' || url.hostname.endsWith('.dev.lab')) {
        assert([origin, auth].includes(url.origin), `Cross-request product URL leaked on ${host}`)
      }
    }
    return { host, origin, outcome: 'passed' }
  }))
}

let container
let developmentContainer
const runtimeLinks = []
const mutationProtection = []
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
  for (const path of ['/pages/%FF', '/pages/%E0%A4', '/pages/%']) {
    const answer = await fetch(base + path, { signal: AbortSignal.timeout(5_000) }).catch(() => assert.fail(`${path} hung`))
    assert([400, 404].includes(answer.status), `${path} answered ${answer.status}`)
  }
  const bootLog = execFileSync('sh', ['-c', `docker logs ${container} 2>&1`], { encoding: 'utf8' })
  assert(!bootLog.includes('uncaughtException'), 'A malformed path threw')
  assert.equal(bootLog.match(/"event":"startup"/g)?.length, 1, 'Boot wrote no single startup line')
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
  // A directory is not a CA file, so it must stop the container as a bad URL does.
  for (const [name, value] of [['TADOKU_API_URL', 'ftp://tadoku.app'], ['NODE_EXTRA_CA_CERTS', '/app']]) {
    const misconfigured = docker('run', '-d', '--label', 'tadoku.check=webv3-config', '-e', `${name}=${value}`, image)
    try {
      const exit = await run('docker', ['wait', misconfigured], { timeout: 15_000 })
        .catch(() => assert.fail(`A bad ${name} did not stop the container`))
      assert.notEqual(exit.stdout.trim(), '0', `A bad ${name} exited successfully`)
      const logs = await run('docker', ['logs', misconfigured])
      assert.match(logs.stdout + logs.stderr, new RegExp(`cannot start: .*${name}`), 'The startup error does not name the variable')
    } finally {
      docker('rm', '-f', misconfigured)
    }
  }
  // A page still waiting on the API past srvx's default 5-second shutdown timeout finishes during docker stop.
  const draining = docker('run', '-d', '--label', 'tadoku.check=webv3-drain', '--add-host', 'host.docker.internal:host-gateway',
    '-p', '127.0.0.1::3000', '-e', `TADOKU_API_URL=http://host.docker.internal:${apiPort}`, image)
  try {
    const drainBase = `http://127.0.0.1:${JSON.parse(docker('inspect', draining))[0].NetworkSettings.Ports['3000/tcp'][0].HostPort}`
    await waitForHealth(drainBase)
    fixtureOptions.delayMs = 6_500
    const page = fetch(drainBase, { signal: AbortSignal.timeout(20_000) })
      .then(async response => [response.status, await response.text()], () => [])
    await new Promise(resolve => setTimeout(resolve, 1_000))
    await run('docker', ['stop', '-t', '20', draining])
    const [status, drainedHtml] = await page
    assert.equal(status, 200, 'docker stop dropped an in-flight page request')
    assert(drainedHtml.includes(contest.title), 'The drained page lost its API data')
  } finally {
    fixtureOptions.delayMs = 0
    docker('rm', '-f', draining)
  }
  // Upstream calls carry the request's ID, so webv3's failure lines and Tadoku API's lines join on correlation_id.
  // An unsafe ingress ID is replaced, never forwarded.
  for (const ingress of ['smoke-forwarded', 'not safe!']) {
    requestIds.length = 0
    fixtureOptions.status = 500
    try {
      await fetch(base, { headers: { 'X-Request-ID': ingress }, signal: AbortSignal.timeout(15_000) })
    } finally {
      fixtureOptions.status = 0
    }
    const [id] = requestIds
    assert(typeof id === 'string' && /^[A-Za-z0-9_-]{1,128}$/.test(id) && requestIds.every(each => each === id), `Upstream calls carried ${requestIds}`)
    assert.equal(id === ingress, ingress === 'smoke-forwarded', `Ingress ID ${ingress} was forwarded as ${id}`)
    const lines = docker('logs', container).split('\n').filter(line => line.includes(`"correlation_id":"${id}"`))
    assert(lines.some(line => JSON.parse(line).event === 'upstream_failure'), `No failure line carries ${id}`)
  }
  const refused = await fetch(`${base}/logout`, { method: 'POST', headers: { Origin: 'https://evil.tadoku.app', 'Sec-Fetch-Site': 'same-site' }, redirect: 'manual' })
  assert.equal(refused.status, 403)
  // The synthetic API has no identity routes, so a session cookie meets an unavailable identity service.
  const outage = await fetch(`${base}/pages/manual`, { headers: { Cookie: 'tadoku_session=fixture' }, signal: AbortSignal.timeout(15_000) })
  const outageHtml = await outage.text()
  assert.equal(outage.status, 200)
  assert(outageHtml.includes('Account service unavailable'), 'Identity outage did not render the unavailable account state')
  assert(!outageHtml.includes('>Log in<'), 'Identity outage rendered a guest')
  const createPath = docker('exec', container, 'node', '--input-type=module', '-e', String.raw`
    import { readdirSync, readFileSync } from 'node:fs';
    const source = readdirSync('/app/server/_ssr').filter(name => name.endsWith('.mjs'))
      .map(name => readFileSync('/app/server/_ssr/' + name, 'utf8')).join('\n');
    const id = source.match(/id:\s*"([a-f0-9]+)",\s*name:\s*"createLog"/)[1];
    console.log('/_serverFn/' + id);
  `)
  for (const headers of [{ Origin: 'https://evil.tadoku.app', 'Sec-Fetch-Site': 'same-site' }, {}]) {
    const response = await fetch(base + createPath, {
      method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded', ...headers },
      body: new URLSearchParams({}), redirect: 'manual',
    })
    assert.equal(response.status, 403, 'Built createLog did not reject an unsafe request')
    mutationProtection.push({ path: createPath, origin: headers.Origin || 'absent', status: response.status })
  }
  const html = await (await fetch(base, { signal: AbortSignal.timeout(15_000) })).text()
  assert(html.includes(contest.title), 'Runtime API override or SSR failed')
  assert(html.includes('https://account.tadoku.app/login'), 'Production account links missing')
  const assets = [...html.matchAll(/(?:src|href)="([^" ]+\.(?:js|css)(?:\?[^" ]*)?)"/g)]
  assert(assets.length > 0, 'SSR has no built assets')
  for (const [, asset] of assets) assert((await fetch(new URL(asset, base), { signal: AbortSignal.timeout(5_000) })).ok, `Asset missing: ${asset}`)
  const exitCode = await new Promise((resolve, reject) => {
    const child = spawn('pnpm', ['--filter', 'webv3', 'test:e2e'], {
      stdio: 'inherit', env: { ...process.env, WEBV3_BASE_URL: base,
        TADOKU_API_URL: `http://127.0.0.1:${apiPort}`, WEBV3_WEB_URL: 'https://tadoku.app', PLAYWRIGHT_OUTPUT_DIR: artifacts },
    })
    child.once('error', reject)
    child.once('exit', resolve)
  })
  assert.equal(exitCode, 0, 'Browser journeys failed')
  const newLog = await fetch(`${base}/logs/new`)
  assert.equal(newLog.status, 200)
  assert((await newLog.text()).includes('Sign in to log activity'), 'Anonymous native logging prompt missing')
  assert.deepEqual(writes, [], 'Application attempted upstream writes')
  await new Promise(resolve => fixture.close(resolve))
  const health = await fetch(`${base}/healthz`, { signal: AbortSignal.timeout(2_000) })
  assert.equal(health.status, 200, 'API outage broke health')
  assert.equal(await health.text(), 'ok\n')
  // Without the API, a contest page's main read fails: 503, the error page, one record per failed read and one for the page.
  const apiOutage = await fetch(`${base}/contests/${contest.id}/leaderboard/1`, { headers: { 'X-Request-ID': 'smoke-outage' } })
  assert.equal(apiOutage.status, 503, 'An API outage did not answer 503')
  assert((await apiOutage.text()).includes('This page couldn’t load'), 'An API outage did not render the error page')
  const records = docker('logs', container).split('\n').filter(line => line.includes('"correlation_id":"smoke-outage"')).map(line => JSON.parse(line))
  const failedReads = records.filter(record => record.event === 'upstream_failure').map(record => record.operation)
  assert(failedReads.length > 0, 'An API outage logged no failed read')
  assert.equal(new Set(failedReads).size, failedReads.length, 'An API outage logged a failed read twice')
  assert.deepEqual(records.filter(record => record.event === 'ssr_error').map(record => record.status), [503], 'An API outage did not log exactly one page failure')
  // A logout the API cannot confirm renders its error page in the app shell.
  const logout = await fetch(`${base}/logout`, { method: 'POST', redirect: 'manual',
    headers: { Origin: base.replace('http:', 'https:'), Cookie: 'tadoku_session=fixture', Accept: 'text/html' } })
  const logoutHtml = await logout.text()
  assert.equal(logout.status, 503)
  assert(logoutHtml.includes('Log out is unavailable') && logoutHtml.includes('Skip to content'), 'Failed logout lost the app shell')
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
    fixture: `Synthetic local read-only contest and content API (${process.env.WEBV3_LANGUAGE_FIXTURE === 'array' ? 'array' : 'null'} language lists); real API/auth/deployment unverified`,
    command: `node scripts/smoke-webv3-image.mjs ${image}`, writes, runtimeLinks, mutationProtection }, null, 2))
}
