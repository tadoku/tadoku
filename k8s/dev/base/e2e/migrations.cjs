// Real-image migrations, owner and runtime grant separation, and branch database provisioning
// against a disposable Postgres container.
const fs = require('node:fs')
const path = require('node:path')
const { spawnSync } = require('node:child_process')
const Y = require('../../../../frontend/node_modules/yaml')

function proveMigrations({ root, evidence, prefix, containers, report, run, createNetwork }, docs) {
  const jobs = docs.filter(d => d.kind === 'Job')
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
  createNetwork()
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
  run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-d', 'tadoku', '-v', 'ON_ERROR_STOP=1', '-c', 'grant create on schema public to tadoku'])
  migrate(ownership, 'invalid', true)
  const invalidOwnership = report.steps.find(step => step.command.includes(`${prefix}-tadoku-api-ownership-invalid`) && step.command.includes('start'))
  if (!fs.readFileSync(path.join(evidence, invalidOwnership.log), 'utf8').includes('database ownership or runtime grant validation failed')) throw new Error('Invalid ownership or runtime grants must fail the ownership hook')
  run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-d', 'tadoku', '-v', 'ON_ERROR_STOP=1', '-c', 'revoke create on schema public from tadoku'])
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
  const provision = Y.parse(fs.readFileSync(path.join(root, '.dev/database.yaml'), 'utf8')).spec.template.spec.containers[0]
  function prepareBranch(branch, expectFailure = false) {
    run('docker', [
      'exec', '-i', '-e', `DEV_ROUTE=${branch}`, '-e', 'PGUSER=postgres', '-e', 'PGDATABASE=postgres', db,
      ...provision.command,
      ...provision.args.map(arg => arg.replaceAll('$$', '$')),
    ], undefined, expectFailure)
  }
  for (const invalidRoute of ['invalid/route', `${'a'.repeat(56)}-0123abcd`]) {
    prepareBranch(invalidRoute, true)
    const refusal = fs.readFileSync(path.join(evidence, report.steps.at(-1).log), 'utf8')
    if (!refusal.includes('refusing invalid or overlong branch database name')) throw new Error('Invalid names must fail before database creation')
    const created = run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-Atc', `select exists(select from pg_database where datname = 'tadoku-${invalidRoute}')`])
    if (created !== 'f') throw new Error('Invalid branch provisioning created a database')
  }
  const unmarkedRoute = 'owner-split-unmarked-0123abcd'
  run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1', '-c', `create database "tadoku-${unmarkedRoute}" owner tadoku_owner`])
  prepareBranch(unmarkedRoute, true)
  const unmarkedOwner = run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-Atc', `select pg_get_userbyid(datdba) || '|' || coalesce(shobj_description(oid, 'pg_database'), '') from pg_database where datname = 'tadoku-${unmarkedRoute}'`])
  if (unmarkedOwner !== 'tadoku_owner|') throw new Error('Refused provisioning changed an unmarked database')
  for (const branch of ['owner-split-existing-0123abcd', 'owner-split-fresh-0123abcd']) {
    if (branch.includes('existing')) run('docker', ['exec', '-i', db, 'psql', '-X', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1'], `
      create database "tadoku-${branch}" owner tadoku;
      comment on database "tadoku-${branch}" is 'dev-cli branch database route=${branch}';
      \\connect "tadoku-${branch}"
      set role tadoku;
      create table legacy_owner_split_proof(id bigint);
    `)
    if (branch.includes('existing')) {
      prepareBranch(branch, true)
      const refusal = fs.readFileSync(path.join(evidence, report.steps.at(-1).log), 'utf8')
      if (!refusal.includes('phase 4.6')) throw new Error('Legacy-owned retained databases must fail closed and point to approved phase 4.6 removal')
      const unchangedOwner = run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-Atc', `select pg_get_userbyid(datdba) from pg_database where datname = 'tadoku-${branch}'`])
      if (unchangedOwner !== 'tadoku') throw new Error('Refused branch provisioning must preserve legacy database ownership')
    }
    migrate(ownership, `${branch}-existing`)
    prepareBranch(branch)
    const branchMigrate = Y.parse(fs.readFileSync(path.join(root, '.dev/migrate.yaml'), 'utf8'))
    const migrationInit = branchMigrate.spec.template.spec.initContainers?.find(c => c.name === 'migrate')
    if (!migrationInit) throw new Error('Branch migration must complete before the runtime-grant post-step')
    migrationInit.image = application.spec.template.spec.containers[0].image
    migrate(branchMigrate, `${branch}-init`, false, `tadoku-${branch}`, migrationInit)
    migrate(branchMigrate, `${branch}-grants`, false, `tadoku-${branch}`)
    assertRuntime(`tadoku-${branch}`)
    if (branch.includes('fresh')) {
      run('docker', ['exec', db, 'psql', '-X', '-U', 'postgres', '-d', `tadoku-${branch}`, '-v', 'ON_ERROR_STOP=1', '-c', 'revoke select on schema_migrations from tadoku'])
      migrate(branchMigrate, `${branch}-invalid-grants`, true, `tadoku-${branch}`)
      const invalidGrants = report.steps.find(step => step.command.includes(`${prefix}-migrate-${branch}-invalid-grants`) && step.command.includes('start'))
      if (!fs.readFileSync(path.join(evidence, invalidGrants.log), 'utf8').includes('runtime migration-table grants are invalid')) throw new Error('Invalid runtime grants must fail the branch migration task')
    }
    prepareBranch(branch)
  }
  report.ownerSplit = { fresh: 'passed', existing: 'passed', repeated: 'passed', branch: 'passed', invalidNames: 'refused', unmarkedDatabase: 'refused unchanged', legacyOwner: 'refused unchanged', invalidMigrationGrants: 'refused', runtime: 'tadoku', owner: 'tadoku_owner' }
  return db
}

module.exports = { proveMigrations }
