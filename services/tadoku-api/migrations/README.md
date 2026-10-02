## Standalone migrations

Every migration ships in its own pull request; see
[Database and migrations](../../../docs/docs/tadoku-api/database.md#migrations).

## Scheduled official contests

Production runs `pg_cron` 1.6 in the `tadoku` application database. These jobs
are configured separately from application migrations because their setup requires
administrative credentials. They call the historical `data.create_contest_round`
function retained in migration `0004`.
Both statements in each command run in one implicit transaction, so the tenant
setting lasts through contest creation and resets afterward.

Use the direct PostgreSQL port 5432 for administrative and migration work; runtime
connections use PgBouncer on port 6432. Production cron executes as `postgres`,
with background workers and GMT schedules. Keep each existing job’s schedule,
database, username and active state when updating its command.

`create_official_contest` catches insert errors and returns null. A successful cron
run therefore does not prove that a round exists: verify the expected official
contest row and its `tenant = 'tadoku/prod'`. Manual round creation must also set
the transaction-local tenant before calling the function.

```sql
insert into cron.job (schedule, command, nodename, nodeport, database, username)
values
  -- Round 1: Dec 21 @ 01:00 for next year
  (
    '0 1 21 12 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(1, extract(year from current_date)::int + 1);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  ),

  -- Round 2: Feb 21 @ 01:00 for this year
  (
    '0 1 21 2 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(2, extract(year from current_date)::int);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  ),

  -- Round 3: Apr 21 @ 01:00
  (
    '0 1 21 4 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(3, extract(year from current_date)::int);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  ),

  -- Round 4: Jun 21 @ 01:00
  (
    '0 1 21 6 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(4, extract(year from current_date)::int);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  ),

  -- Round 5: Aug 21 @ 01:00
  (
    '0 1 21 8 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(5, extract(year from current_date)::int);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  ),

  -- Round 6: Oct 21 @ 01:00
  (
    '0 1 21 10 *',
    $$select set_config('tadoku.tenant', 'tadoku/prod', true); select data.create_contest_round(6, extract(year from current_date)::int);$$,
    'localhost',
    5432,
    'tadoku',
    'postgres'
  );
```
