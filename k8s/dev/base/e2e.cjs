// Real-image migration and frontend startup proof; no cluster or mocks.
// Run: node k8s/dev/base/e2e.cjs /tmp/tadoku-base-evidence
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')
const Y = require('../../../frontend/node_modules/yaml')
const root = path.resolve(__dirname, '../../..')
const evidence = path.resolve(process.argv[2] || `/tmp/tadoku-base-evidence-${process.pid}`)
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

try {
  report.revision = run('git', ['rev-parse', 'HEAD'])
  report.worktree = run('git', ['status', '--short'])
  const rendered = run('kubectl', ['--context', 'homelab-talos-dev', 'kustomize', 'k8s/dev/base'])
  fs.writeFileSync(path.join(evidence, 'rendered.yaml'), rendered)
  report.renderedSHA256 = require('node:crypto').createHash('sha256').update(rendered).digest('hex')
  const docs = Y.parseAllDocuments(rendered).map(d => d.toJSON())
  const api = docs.find(d => d.kind === 'Deployment' && d.metadata.namespace === 'tdk-dev-tadoku-api' && d.metadata.name === 'tadoku-api')
  if (!api) throw new Error('Missing base Tadoku API Deployment')
  if (docs.some(d => d.kind === 'Deployment' && d.metadata.name === 'immersion-api')) throw new Error('Legacy Immersion Deployment remains in dev base')
  const envValue = (container, name) => container.env.find(e => e.name === name)?.value
  const worker = docs.find(d => d.kind === 'Deployment' && d.metadata.namespace === 'tdk-dev-tadoku-api' && d.metadata.name === 'tadoku-worker')
  if (!worker || worker.spec.replicas !== 1 || worker.spec.strategy?.type !== 'Recreate') throw new Error('Base worker must be a single Recreate Deployment')
  const workerPod = worker.spec.template.spec
  const workerContainer = workerPod.containers[0]
  const apiContainer = api.spec.template.spec.containers[0]
  const branchApi = Y.parse(fs.readFileSync(path.join(root, '.dev/tadoku-api.yaml'), 'utf8')).spec.containers[0]
  const branchWorker = Y.parse(fs.readFileSync(path.join(root, '.dev/tadoku-worker.yaml'), 'utf8')).spec.containers[0]
  if (envValue(workerContainer, 'WORKER_POSTGRES_DATABASE') !== 'tadoku' || envValue(workerContainer, 'WORKER_LEADERBOARD_CACHE_PREFIX') !== undefined || envValue(apiContainer, 'API_LEADERBOARD_CACHE_PREFIX') !== undefined) throw new Error('Base worker must use the base database and both base processes must use unprefixed cache')
  if (envValue(branchWorker, 'WORKER_POSTGRES_DATABASE') !== 'tadoku-${DEV_ROUTE}' || envValue(branchWorker, 'WORKER_LEADERBOARD_CACHE_PREFIX') !== 'dev:${DEV_ROUTE}:') throw new Error('Branch worker database/cache isolation is not configured')
  if (envValue(branchApi, 'API_POSTGRES_DATABASE') !== envValue(branchWorker, 'WORKER_POSTGRES_DATABASE') || envValue(branchApi, 'API_LEADERBOARD_CACHE_PREFIX') !== envValue(branchWorker, 'WORKER_LEADERBOARD_CACHE_PREFIX')) throw new Error('Branch API and worker must use the same database/cache namespace')
  if (workerPod.automountServiceAccountToken !== false || workerContainer.image === apiContainer.image || !workerContainer.image.startsWith('ghcr.io/tadoku/tadoku/tadoku-worker:latest')) throw new Error('Worker image or private Pod configuration is invalid')
  if (docs.some(d => d.kind === 'Service' && d.spec?.selector?.app === 'tadoku-worker') || docs.some(d => d.kind === 'HTTPRoute' && JSON.stringify(d.spec).includes('tadoku-worker'))) throw new Error('Worker must have no Service or public route')
  const resources = workerContainer.resources
  if (!resources?.requests?.cpu || !resources.requests.memory || !resources.limits?.cpu || !resources.limits.memory) throw new Error('Worker must declare CPU and memory requests and limits')
  report.leaderboardWorkers = { asyncJobs: 'tadoku-worker', baseDatabase: 'tadoku', branchDatabase: envValue(branchWorker, 'WORKER_POSTGRES_DATABASE'), branchPrefix: envValue(branchWorker, 'WORKER_LEADERBOARD_CACHE_PREFIX'), image: workerContainer.image }
  const jobs = docs.filter(d => d.kind === 'Job')
  const frontends = docs.filter(d => d.kind === 'Deployment' && d.metadata.name.startsWith('frontend-'))
  const roles = { 'tadoku-api-migrate': 'tadoku_owner', 'kratos-migrate': 'kratos', 'keto-migrate': 'keto' }
  const database = { tadoku_owner: 'tadoku', kratos: 'kratos', keto: 'keto' }
  const postgres = docs.find(d => d.kind === 'postgresql' && d.metadata.name === 'tadoku-dev-db')
  const ownership = jobs.find(j => j.metadata.name === 'tadoku-api-ownership')
  const grants = jobs.find(j => j.metadata.name === 'tadoku-api-grants')
  for (const image of ['postgres:17', ...jobs.map(j => j.spec.template.spec.containers[0].image)]) {
    // Use an existing resolved artifact; avoid an unnecessary large base pull
    // on the shared T3 disk. Missing images still use the normal registry path.
    const cached = spawnSync('docker', ['image', 'inspect', image], { stdio: 'ignore', timeout: 30000 })
    if (cached.status !== 0) run('docker', ['pull', image])
    report.images[image] = JSON.parse(run('docker', ['image', 'inspect', image, '--format', '{{json .RepoDigests}}']))
  }
  run('docker', ['network', 'create', '--label', `tadoku.dev/test=${prefix}`, prefix])
  networkCreated = true
  const db = `${prefix}-db`
  run('docker', ['create', '--name', db, '--network', prefix, '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '512m', '--tmpfs', '/var/lib/postgresql/data:rw', '-e', 'POSTGRES_PASSWORD=disposable-test-only', 'postgres:17'])
  containers.push(db)
  run('docker', ['start', db])
  run('docker', ['exec', db, 'sh', '-ec', 'for i in $(seq 1 60); do pg_isready -h 127.0.0.1 -U postgres && exit 0; sleep 1; done; exit 1'])
  const roleSetup = Object.values(roles)
    .map(r => `create role ${r} login password 'disposable-test-only';\ncreate database ${database[r]} owner ${postgres.spec.databases[database[r]]};`)
    .join('\n')
  run('docker', ['exec', '-i', db, 'psql', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1'], `create role tadoku login password 'disposable-test-only';\n` + roleSetup)

  function migrate(job, suffix, fail = false, target = 'tadoku', selectedContainer) {
    const spec = job.spec.template.spec
    const container = selectedContainer || spec.containers[0]
    const role = roles[job.metadata.name] || (
      job === ownership || job === grants ? 'postgres'
        : job.metadata.name === 'migrate' ? 'tadoku_owner'
        : undefined
    )
    if (!role) throw new Error(`Unexpected migration Job: ${job.metadata.name}`)
    const name = `${prefix}-${job.metadata.name}-${suffix}`
    const isPsql = container.env.some(e => e.name === 'PGUSER')
    const userSecret = container.env.find(e => e.name === 'POSTGRES_USER' || e.name === 'PGUSER')?.valueFrom.secretKeyRef.name
    const migrationRole = userSecret?.startsWith('tadoku-owner.') ? 'tadoku_owner' : 'tadoku'
    const env = isPsql
      ? { PGHOST: db, PGUSER: role === 'postgres' ? role : migrationRole, PGPASSWORD: 'disposable-test-only', PGDATABASE: target, PGSSLMODE: 'disable' }
      : role === 'tadoku_owner'
        ? { POSTGRES_HOST: db, POSTGRES_USER: migrationRole, POSTGRES_PASSWORD: 'disposable-test-only', POSTGRES_DATABASE: fail ? 'missing_fixture_database' : target, POSTGRES_SSLMODE: 'disable' }
        : { DSN: `postgres://${role}@${db}:5432/${role}?sslmode=disable`, PGPASSWORD: 'disposable-test-only' }
    const args = container.args.map(arg => arg.replaceAll('$$', '$'))
    run('docker', [
      'create', '--name', name, '--network', prefix, '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '256m',
      ...Object.entries(env).flatMap(([k,v]) => ['-e', `${k}=${v}`]),
      ...(container.command ? ['--entrypoint', container.command[0]] : []),
      container.image,
      ...(container.command?.slice(1) || []),
      ...args,
    ])
    containers.push(name)
    // docker cp works with the T3 DinD sidecar; host bind mounts do not.
    for (const volume of spec.volumes || []) {
      const config = docs.find(d => d.kind === 'ConfigMap' && d.metadata.name === volume.configMap.name && d.metadata.namespace === job.metadata.namespace)
      const mount = container.volumeMounts.find(m => m.name === volume.name)
      if (!config || !mount) throw new Error('Missing migration configuration')
      const dir = path.join(evidence, name)
      fs.mkdirSync(dir)
      for (const [file, content] of Object.entries(config.data)) fs.writeFileSync(path.join(dir, file), content)
      // /etc exists in Ory images; copy the directory itself as /etc/config.
      run('docker', ['cp', dir, `${name}:${mount.mountPath}`])
    }
    run('docker', ['start', '-a', name], undefined, fail)
    const exit = run('docker', ['inspect', name, '--format', '{{.State.ExitCode}}'])
    if (fail ? exit === '0' : exit !== '0') throw new Error(`Migration ${name} exited ${exit}`)
  }
  const syncWave = job => Number(job.metadata.annotations['argocd.argoproj.io/sync-wave'])
  const orderedJobs = [...jobs].sort((a, b) => syncWave(a) - syncWave(b))
  for (const job of orderedJobs) migrate(job, 'fresh')
  for (const job of orderedJobs) migrate(job, 'noop')
  const application = jobs.find(j => j.metadata.name === 'tadoku-api-migrate')
  migrate(application, 'failure', true)
  migrate(application, 'recovery')
  for (const role of Object.values(roles)) {
    const count = Number(run('docker', ['exec', db, 'psql', '-U', 'postgres', '-d', database[role], '-Atc', "select count(*) from information_schema.tables where table_schema='public'"]))
    if (count < 1) throw new Error(`${role} schema is empty`)
  }

  function assertRuntime(target) {
    const owners = run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-d', target, '-Atc', "select distinct tableowner from pg_tables where schemaname='public'"])
    if (owners !== 'tadoku_owner') throw new Error(`Public tables must be owned only by tadoku_owner; found ${owners}`)
    run('docker', ['exec', '-i', db, 'psql', '-X', '-U', 'postgres', '-d', target, '-v', 'ON_ERROR_STOP=1'], `
      do $$ begin
        if not exists(select from pg_roles where rolname='tadoku' and not rolbypassrls and not rolsuper)
          or pg_has_role('tadoku', 'tadoku_owner', 'member') then
          raise exception 'runtime must not bypass RLS or inherit ownership';
        end if;
        if exists(select from pg_tables where schemaname='public' and tablename<>'schema_migrations'
          and not has_table_privilege('tadoku', format('%I.%I', schemaname, tablename), 'select,insert,update,delete')) then
          raise exception 'missing runtime DML';
        end if;
      end $$;
      set role tadoku_owner;
      create table owner_split_proof(id bigserial primary key, value text not null);
    `)
    run('docker', ['exec', '-i', db, 'psql', '-X', '-U', 'tadoku', '-d', target, '-v', 'ON_ERROR_STOP=1'], `
      insert into owner_split_proof(value) values('runtime');
      update owner_split_proof set value='updated';
      select * from owner_split_proof;
      delete from owner_split_proof;
      select * from schema_migrations;
    `)
    const forbidden = [
      'create table forbidden_owner_split_proof()',
      'insert into schema_migrations select * from schema_migrations where false',
      'update schema_migrations set dirty=dirty where false',
      'delete from schema_migrations where false',
      'truncate schema_migrations',
    ]
    for (const sql of forbidden) {
      const output = run('docker', ['exec', db, 'psql', '-X', '-U', 'tadoku', '-d', target, '-v', 'ON_ERROR_STOP=1', '-c', sql], undefined, true)
      const log = fs.readFileSync(path.join(evidence, report.steps.at(-1).log), 'utf8')
      if (!log.includes('permission denied')) throw new Error(`Expected permission denial for ${sql}: ${output}`)
    }
    run('docker', ['exec', db, 'psql', '-X', '-U', 'tadoku_owner', '-d', target, '-c', 'drop table owner_split_proof'])
  }
  assertRuntime('tadoku')
  if (!ownership || !grants) throw new Error('Both pre-migration ownership and post-migration grants hooks are required')
  const branchConfig = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === 'tdk-dev-data' && d.metadata.name === 'tadoku-api-ownership')
  const branchConfigDir = path.join(evidence, 'branch-ownership')
  fs.mkdirSync(branchConfigDir)
  for (const [file, content] of Object.entries(branchConfig.data)) fs.writeFileSync(path.join(branchConfigDir, file), content)
  run('docker', ['cp', branchConfigDir, `${db}:/config`])
  for (const branch of ['owner-split-existing-0123abcd', 'owner-split-fresh-0123abcd']) {
    if (branch.includes('existing')) run('docker', ['exec', '-i', db, 'psql', '-X', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1'], `
      create database "tadoku-${branch}" owner tadoku;
      comment on database "tadoku-${branch}" is 'dev-cli branch database route=${branch}';
      \\connect "tadoku-${branch}"
      set role tadoku;
      create table legacy_owner_split_proof(id bigint);
    `)
    migrate(ownership, `${branch}-existing`)
    const provision = Y.parse(fs.readFileSync(path.join(root, '.dev/database.yaml'), 'utf8')).spec.template.spec.containers[0]
    function prepareBranch() {
      run('docker', [
        'exec', '-i', '-e', `DEV_ROUTE=${branch}`, '-e', 'PGUSER=postgres', '-e', 'PGDATABASE=postgres', db,
        ...provision.command,
        ...provision.args.map(arg => arg.replaceAll('$$', '$')),
      ])
    }
    prepareBranch()
    const branchMigrate = Y.parse(fs.readFileSync(path.join(root, '.dev/migrate.yaml'), 'utf8'))
    const migrationInit = branchMigrate.spec.template.spec.initContainers?.find(c => c.name === 'migrate')
    if (!migrationInit) throw new Error('Branch migration must complete before the runtime-grant post-step')
    migrationInit.image = application.spec.template.spec.containers[0].image
    migrate(branchMigrate, `${branch}-init`, false, `tadoku-${branch}`, migrationInit)
    migrate(branchMigrate, `${branch}-grants`, false, `tadoku-${branch}`)
    assertRuntime(`tadoku-${branch}`)
    prepareBranch()
  }
  report.ownerSplit = { fresh: 'passed', existing: 'passed', repeated: 'passed', branch: 'passed', runtime: 'tadoku', owner: 'tadoku_owner' }

  run('bazel', ['build', '//services/tadoku-api:dev', '//services/tadoku-api:worker_dev'])
  const apiMetadata = JSON.parse(fs.readFileSync(path.join(root, 'bazel-bin/services/tadoku-api/dev.dev.json')))
  const workerMetadata = JSON.parse(fs.readFileSync(path.join(root, 'bazel-bin/services/tadoku-api/worker_dev.dev.json')))
  if (apiMetadata.selectionGroup !== 'tadoku-api' || workerMetadata.selectionGroup !== 'tadoku-api' || workerMetadata.kind !== 'worker') throw new Error('Branch API and worker are not selected as one group')
  if (workerMetadata.imageTarget !== '//services/tadoku-api/cmd/tadoku-worker:cli_image' || workerMetadata.pushTarget !== '//services/tadoku-api/cmd/tadoku-worker:cli_push' || workerMetadata.workloadTemplate !== '.dev/tadoku-worker.yaml') throw new Error('Worker is not an independent DevCLI image/deployable')
  if (['publicPath', 'internalHost', 'publicProxy', 'baseService', 'servicePort'].some(key => key in workerMetadata)) throw new Error('Worker metadata must not request a route or Service')

  for (const deployment of frontends) {
    const spec = deployment.spec.template.spec
    const container = spec.containers[0]
    const cached = spawnSync('docker', ['image', 'inspect', container.image], { stdio: 'ignore', timeout: 30000 })
    if (cached.status !== 0) run('docker', ['pull', container.image])
    report.images[container.image] = JSON.parse(run('docker', ['image', 'inspect', container.image, '--format', '{{json .RepoDigests}}']))
    const name = `${prefix}-${deployment.metadata.name}`
    const volume = spec.volumes.find(v => v.name === 'startup')
    const config = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === deployment.metadata.namespace && d.metadata.name === volume.configMap.name)
    const dir = path.join(evidence, deployment.metadata.name)
    fs.mkdirSync(dir)
    for (const [file, content] of Object.entries(config.data)) fs.writeFileSync(path.join(dir, file), content)
    // No external network: this checks the published server/asset startup, not
    // login or API acceptance, and cannot accidentally contact production.
    run('docker', ['create', '--name', name, '--network', 'none', '--label', `tadoku.dev/test=${prefix}`, '--cpus', '0.5', '--memory', '512m', ...container.env.flatMap(e => ['-e', `${e.name}=${e.value}`]), '--entrypoint', container.command[0], container.image, ...container.command.slice(1)])
    containers.push(name)
    run('docker', ['cp', dir, `${name}:/etc/dev`])
    run('docker', ['start', name])
    try {
      run('docker', ['exec', name, 'node', '-e', String.raw`
        (async () => {
          let response;
          for (let i = 0; i < 60; i++) {
            try {
              response = await fetch('http://127.0.0.1:3000/', {signal: AbortSignal.timeout(5000)});
              if (response.ok) break;
            } catch {}
            await new Promise(resolve => setTimeout(resolve, 1000));
          }
          if (!response?.ok) throw new Error('Frontend did not become ready');
          const html = await response.text();
          const match = html.match(/<script id="__NEXT_DATA__"[^>]*>(.*?)<\/script>/s);
          if (!match) throw new Error('No Next runtime data in rendered page');
          const runtime = JSON.parse(match[1]).runtimeConfig;
          if (runtime?.homeUrl !== 'https://tadoku.dev.lab' || runtime?.kratosPublicEndpoint !== 'https://account.tadoku.dev.lab/kratos' || runtime?.apiEndpoint !== 'https://tadoku.dev.lab/api/internal' || runtime?.cookieDomain !== '.tadoku.dev.lab' || runtime?.cookieSecure !== true) throw new Error('Development runtime URLs/cookies were not applied');
          const assets = [...html.matchAll(/<script[^>]*src="([^\"]+)"/g)].map(m => m[1]).filter(p => p.startsWith('/_next/'));
          if (!assets.length) throw new Error('No compiled Next assets');
          for (const asset of assets) {
            const result = await fetch('http://127.0.0.1:3000' + asset);
            if (!result.ok) throw new Error('Missing compiled asset: ' + asset);
            await result.arrayBuffer();
          }
          console.log(JSON.stringify({runtime, assets: assets.length}));
        })().catch(error => { console.error(error); process.exit(1) });
      `])
    } finally { run('docker', ['logs', name]) }
    run('docker', ['rm', '-f', name])
    containers.pop()
  }
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
  for (const name of containers.reverse()) {
    try {
      const fixture = run('docker', ['ps', '-a', '--no-trunc', '--filter', `name=^/${name}$`, '--format', '{{.ID}}|{{.Label "tadoku.dev/test"}}'])
      if (!fixture) {
        if (name === report.cleanupControl?.removedFixture) report.cleanupControl.result = 'passed'
        continue
      }
      const [id, owner] = fixture.split('|')
      if (owner !== prefix || !/^[0-9a-f]{64}$/.test(id)) throw new Error(`Refusing cleanup of ${name}`)
      run('docker', ['rm', '-f', id])
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
