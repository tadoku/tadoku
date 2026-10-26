const path = require('node:path')

function proveTenantSeed({ root, report, run }, db) {
  const tenant = 'tadoku/fixture-seed-0123abcd'
  const admin = 'edda0741-cf0c-4b71-a540-472c9726b1b5'
  const reader = '0d141891-b6f8-4150-9d2f-6272d68b3db0'
  run('docker', ['cp', path.join(root, 'scripts/dev/seed'), `${db}:/`])
  run('docker', ['cp', path.join(root, '.dev/seed.sql'), `${db}:/seed.sql`])
  const psql = ['exec', '-i', db, 'psql', '-X', '-qAt', '-U', 'postgres', '-d', 'tadoku', '-v', 'ON_ERROR_STOP=1']
  run('docker', psql, `
    insert into tenants(key,kind) values ('${tenant}','test');
    insert into users(tenant,id,display_name,created_at) values
      ('${tenant}','${admin}','Changed branch name',now()-interval '2 months'),
      ('${tenant}','${reader}','Dev Reader',now()-interval '2 months');
    insert into accounts(tenant,user_id,email,email_normalized,password_hash) values
      ('${tenant}','${admin}','changed@example.com','changed@example.com','changed-branch-hash'),
      ('${tenant}','${reader}','reader@tadoku.app','reader@tadoku.app','reader-hash');
    insert into sessions(tenant,user_id,token_hash,authenticated_at,expires_at)
      values ('${tenant}','${admin}','owned-session',now(),now()+interval '1 day');
  `)
  const snapshot = () => run('docker', psql, `
    select jsonb_agg(to_jsonb(t) order by to_jsonb(t)::text) from users t where tenant='${tenant}';
    select jsonb_agg(to_jsonb(t) order by to_jsonb(t)::text) from accounts t where tenant='${tenant}';
    select jsonb_agg(to_jsonb(t) order by to_jsonb(t)::text) from sessions t where tenant='${tenant}';
  `)
  const before = snapshot()
  for (let attempt = 0; attempt < 2; attempt++) {
    run('docker', ['exec', db, 'psql', '-X', '-qAt', '-U', 'tadoku', '-d', 'tadoku', '-v', 'ON_ERROR_STOP=1',
      '-v', `tenant=${tenant}`, '-v', `admin_user_id=${admin}`, '-v', `reader_user_id=${reader}`, '-f', '/seed.sql'])
    if (snapshot() !== before) throw new Error('Branch seed overwrote native fixture profile/account/session data')
  }
  const populated = run('docker', psql, `
    select exists(select from contests where tenant='${tenant}')
      and exists(select from posts where tenant='${tenant}');
  `)
  if (populated !== 't') throw new Error('Native reseed did not continue through branch content/contest fixtures')
  run('docker', psql, `delete from tenants where key='${tenant}';`)
  report.tenantSeed = { repeats: 2, preserved: 'native users/accounts/sessions', populated: 'contests/posts' }
}

module.exports = { proveTenantSeed }
