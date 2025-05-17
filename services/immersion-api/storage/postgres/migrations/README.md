## Scheduling official contests

The following should be ran in the database in which `pg_cron` is installed, which is not the database Tadoku is using.
This is excluded from the migrations because different credentials are required and typically only needs to be performed once.

```sql
insert into cron.job (schedule, command, nodename, nodeport, database, username)
values
  -- Round 1: Dec 21 @ 01:00 for next year
  (
    '0 1 21 12 *',
    $$select create_contest_round(1, extract(year from current_date)::int + 1);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 2: Feb 21 @ 01:00 for this year
  (
    '0 1 21 2 *',
    $$select create_contest_round(2, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 3: Apr 21 @ 01:00
  (
    '0 1 21 4 *',
    $$select create_contest_round(3, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 4: Jun 21 @ 01:00
  (
    '0 1 21 6 *',
    $$select create_contest_round(4, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  ),

  -- Round 5: Aug 21 @ 01:00
  (
    '0 1 21 8 *',
    $$select create_contest_round(5, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_staging_immersion',
    'postgres'
  ),

  -- Round 6: Oct 21 @ 01:00
  (
    '0 1 21 10 *',
    $$select create_contest_round(6, extract(year from current_date)::int);$$,
    '/run/postgresql',
    5432,
    'tadoku_prod_immersion',
    'postgres'
  );
```
