## Standalone migrations

Every migration ships in its own pull request; see
[Database and migrations](../../../docs/docs/tadoku-api/database.md#migrations).

## Scheduled official contests

Run the following in the database where `pg_cron` is installed, which is separate
from the Tadoku application database. These jobs are excluded from application
migrations because their setup requires different credentials. They call the
historical `data.create_contest_round` function retained in migration `0004`.

```sql
insert into cron.job (schedule, command, nodename, nodeport, database, username)
values
  -- Round 1: Dec 21 @ 01:00 for next year
  (
    '0 1 21 12 *',
    $$select data.create_contest_round(1, extract(year from current_date)::int + 1);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 2: Feb 21 @ 01:00 for this year
  (
    '0 1 21 2 *',
    $$select data.create_contest_round(2, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 3: Apr 21 @ 01:00
  (
    '0 1 21 4 *',
    $$select data.create_contest_round(3, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 4: Jun 21 @ 01:00
  (
    '0 1 21 6 *',
    $$select data.create_contest_round(4, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 5: Aug 21 @ 01:00
  (
    '0 1 21 8 *',
    $$select data.create_contest_round(5, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 6: Oct 21 @ 01:00
  (
    '0 1 21 10 *',
    $$select data.create_contest_round(6, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  );
```
