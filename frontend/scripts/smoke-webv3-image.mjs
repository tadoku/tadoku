import assert from 'node:assert/strict'
import { spawn, execFile, execFileSync } from 'node:child_process'
import { createServer, get } from 'node:http'
import { mkdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { promisify } from 'node:util'

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
const learner = { id: '01000000-0000-0000-0000-000000000000', display_name: 'Fixture learner', created_at: '2024-03-01T08:00:00Z' }
// The learner's profile year: one Japanese reading day in whichever year is requested.
function profileYear(kind, year) {
  const score = 12.5
  switch (kind) {
    case 'activity': return { total_updates: 3, scores: [{ date: `${year}-01-02`, score }] }
    case 'scores': return { overall_score: score, scores: [{ language_code: 'jpn', language_name: 'Japanese', score }] }
    case 'activity-split': return { activities: [{ activity_id: 1, activity_name: 'Reading', score }] }
    case 'contest-registrations': return {
      registrations: [{ id: '22222222-2222-4222-8222-222222222222', contest_id: contest.id, user_id: learner.id,
        user_display_name: learner.display_name, languages: [], contest }],
      total_size: 1, next_page_token: '',
    }
  }
}
// The learner's logs, newest first: 51 fill two pages of 50; the third is tagged NSFW, which hides its description.
const learnerLogs = Array.from({ length: 51 }, (_, index) => ({
  id: `33333333-3333-4333-8333-${String(index).padStart(12, '0')}`, user_id: learner.id,
  user_display_name: learner.display_name, description: `Fixture reading ${51 - index}`,
  language: { code: 'jpn', name: 'Japanese' }, activity: { id: 1, name: 'Reading' }, unit_name: 'Page',
  tags: index === 2 ? ['nsfw'] : [], amount: 4, modifier: 1, score: 4,
  created_at: new Date(Date.parse('2024-03-01T08:00:00Z') + (51 - index) * 3_600_000).toISOString(), deleted: false,
}))
// Twelve published posts, newest first, so the blog has two pages; the CMS page holds markup DOMPurify must remove.
const posts = Array.from({ length: 12 }, (_, index) => ({
  id: `00000000-0000-4000-8000-${String(index + 1).padStart(12, '0')}`, slug: `fixture-post-${12 - index}`,
  title: `Fixture post ${12 - index}`, content: `## Fixture\n\nPost **${12 - index}** from the container fixture.`,
  published_at: `2026-01-${String(24 - index).padStart(2, '0')}T08:00:00Z`,
}))
const cmsPage = {
  id: '00000000-0000-4000-8000-000000000401', slug: 'dev-welcome', title: 'Fixture welcome',
  html: '<p>Fixture page body.</p><img src="data:," alt="" onerror="window.__xss=1"><script>window.__xss=2</script>',
}
const announcement = {
  id: '00000000-0000-4000-8000-000000000501', namespace: 'tadoku', title: 'Fixture notice',
  content: 'The **container fixture** announcement.', style: 'info', href: '/pages/dev-welcome',
  starts_at: '2020-01-01T00:00:00Z', ends_at: '2099-12-31T00:00:00Z',
}
const writes = []
// Every fixture answer waits this long, so a check can hold a page request open.
let fixtureDelayMs = 0
const fixture = createServer((request, response) => setTimeout(() => answer(request, response), fixtureDelayMs))
function answer(request, response) {
  if (request.method !== 'GET') {
    writes.push(`${request.method} ${request.url}`)
    response.writeHead(405).end()
    return
  }
  const url = new URL(request.url, 'http://fixture')
  const path = url.pathname
  let data
  if (path.endsWith('/latest-official') || path === `/immersion/contests/${contest.id}`) data = contest
  else if (path.endsWith('/summary')) data = { participant_count: 1, language_count: 1, total_score: 13319.6455 }
  // Imported accounts keep their source IDs without a UUID version check, so standings can hold a version-0 user ID.
  else if (path.endsWith('/leaderboard') || /^\/immersion\/leaderboard\/(global|yearly\/\d+)$/.test(path)) data = {
    // One participant: later pages are empty.
    entries: Number(url.searchParams.get('page') ?? 0) > 0 ? [] : [{ rank: 1, user_id: '01000000-0000-0000-0000-000000000000', user_display_name: 'Fixture learner', score: 13319.6455, is_tie: false }],
    total_size: 1, next_page_token: '',
  }
  // The fixture contest is official, so community lists are empty.
  else if (path === `/content/pages/tadoku/${cmsPage.slug}`) data = cmsPage
  else if (path === '/content/announcements/tadoku/active') data = { announcements: [announcement] }
  else if (path === '/content/posts/tadoku') {
    const size = Number(url.searchParams.get('page_size') ?? 10)
    const page = Number(url.searchParams.get('page') ?? 0)
    data = { posts: posts.slice(page * size, (page + 1) * size), total_size: posts.length, next_page_token: '' }
  }
  else if (path.startsWith('/content/posts/tadoku/')) data = posts.find(post => path.endsWith(`/${post.slug}`))
  else if (path === '/immersion/contests') data = url.searchParams.get('official') === 'false'
    ? { contests: [], total_size: 0, next_page_token: '' }
    : { contests: [contest], total_size: 1, next_page_token: '' }
  else if (path === `/immersion/users/${learner.id}/profile`) data = learner
  else if (path === `/immersion/users/${learner.id}/logs`) {
    const size = Number(url.searchParams.get('page_size') ?? 50)
    const start = Number(url.searchParams.get('page') ?? 0) * size
    data = { logs: learnerLogs.slice(start, start + size), total_size: learnerLogs.length, next_page_token: '' }
  }
  else if (path.startsWith('/immersion/logs/')) data = learnerLogs.find(log => path === `/immersion/logs/${log.id}`)
  // The contest's updates are the learner's logs, two pages of 50; the learner is its only participant.
  else if (path === `/immersion/contests/${contest.id}/logs`) {
    const size = Number(url.searchParams.get('page_size') ?? 50)
    const start = Number(url.searchParams.get('page') ?? 0) * size
    const logs = [learner.id, null].includes(url.searchParams.get('user_id')) ? learnerLogs : []
    data = { logs: logs.slice(start, start + size), total_size: logs.length, next_page_token: '' }
  }
  else if (path === `/immersion/contests/${contest.id}/profile/${learner.id}/scores`) data = {
    overall_score: 13319.6455, scores: [{ language_code: 'jpn', language_name: 'Japanese', score: 13319.6455 }],
    registration: { id: '22222222-2222-4222-8222-222222222222', contest_id: contest.id, user_id: learner.id,
      user_display_name: learner.display_name, languages: [{ code: 'jpn', name: 'Japanese' }], contest },
  }
  else if (path === `/immersion/contests/${contest.id}/profile/${learner.id}/activity`) data = {
    rows: [{ date: '2024-03-03', language_code: 'jpn', score: 13319.6455 }],
  }
  else {
    const match = path.match(new RegExp(`^/immersion/users/${learner.id}/([a-z-]+)/(\\d{4})$`))
    if (match) data = profileYear(match[1], match[2])
  }
  if (!data) { response.writeHead(404).end(); return }
  response.writeHead(200, { 'Content-Type': 'application/json' }).end(JSON.stringify(data))
}
await new Promise(resolve => fixture.listen(0, '0.0.0.0', resolve))
const apiPort = fixture.address().port
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
    fixtureDelayMs = 6_500
    const page = fetch(drainBase, { signal: AbortSignal.timeout(20_000) })
      .then(async response => [response.status, await response.text()], () => [])
    await new Promise(resolve => setTimeout(resolve, 1_000))
    await run('docker', ['stop', '-t', '20', draining])
    const [status, drainedHtml] = await page
    assert.equal(status, 200, 'docker stop dropped an in-flight page request')
    assert(drainedHtml.includes(contest.title), 'The drained page lost its API data')
  } finally {
    fixtureDelayMs = 0
    docker('rm', '-f', draining)
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
        TADOKU_API_URL: `http://127.0.0.1:${apiPort}`, PLAYWRIGHT_OUTPUT_DIR: artifacts },
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
