---
sidebar_position: 1
title: Account deletion
description: How to manually delete a Tadoku account from the Tadoku API database, the leaderboard cache, Keto and Kratos, keeping anonymised history for finished contests.
---

# Account deletion

Read this when a user has asked for their Tadoku account to be deleted.

Account deletion is manual. Run the steps below in order. You can safely rerun any step.

The deletion removes the account's logs, contest registrations, log tags, roles
and profile rows, except the history of finished contests. A contest is finished
once its end date has passed and it has not been deleted. The account's
registrations and logs in finished contests stay, so their final results and
leaderboards do not change. Those logs are anonymised: their descriptions and tags
are removed and `frozen_at` is set, which stops any further change to them.

The `users` row stays as an anonymised tombstone that keeps the ID and
`created_at`: contest queries join contest owners and participants to `users`,
and the Tadoku API rejects writes and profile synchronisation for users whose
`deletion_locked_at` or `deleted_at` is set.

## Prerequisites

- Write access to the Tadoku API database through `psql`. The statements use
  unqualified table names, like the migrations in
  `services/tadoku-api/migrations/`, so connect with the `search_path` that the
  Tadoku API uses.
- The independently deployed migration 0032 must be complete: the
  queue is the `jobs` table. Follow the [migration deployment gate](../tadoku-api/database.md#migrations)
  before using this runbook; it does not support the old table name.
- A running `tadoku-worker` release that registers both
  `leaderboard.invalidate_contest.v1` and
  `leaderboard.invalidate_official.v1`, connected to this exact database and
  its matching Valkey cache prefix. Confirm the deployed release's registration
  before proceeding; the embedded legacy consumer does not process these jobs.
- Access to the Kratos admin API and the Keto read and write APIs.
- The account ID. This is the Kratos identity ID, which is also `users.id`. To
  find it from an email address, run
  `curl "$KRATOS_ADMIN_URL/admin/identities?credentials_identifier=<EMAIL>"`.

Set the ID once for `psql` and once for the shell. The Tadoku API compares
contest dates with the current UTC date, so set the `psql` session to UTC too:

```sql
\set ON_ERROR_STOP on
\set account_id '<ACCOUNT_ID>'
set time zone 'UTC';
```

```bash
ACCOUNT_ID='<ACCOUNT_ID>'
```

## 1. Preflight

Run these read-only queries. First, list the data the account owns:

```sql
select
  (select count(*) from users where id = :'account_id') as users,
  (select deletion_locked_at from users where id = :'account_id') as deletion_locked_at,
  (select deleted_at from users where id = :'account_id') as deleted_at,
  (select count(*) from logs where user_id = :'account_id') as logs,
  (select count(*) from contest_registrations where user_id = :'account_id') as registrations,
  (select count(*) from contests where owner_user_id = :'account_id') as owned_contests,
  (select count(*) from user_roles where user_id = :'account_id') as user_roles,
  (select count(*) from profiles where user_id = :'account_id') as profiles,
  (select count(*) from account_deletion_requests where identity_id = :'account_id') as deletion_requests,
  (select count(*) from moderation_audit_log
    where user_id = :'account_id' or metadata->>'target_user_id' = :'account_id') as audit_entries;
```

If `users` is `0`, the account has no Tadoku API data. Skip to
[Keto relations](#5-keto-relations-and-feature-access).

Next, count the finished-contest history that the delete step keeps anonymised:

```sql
with finished_contests as (
  select id from contests
  where deleted_at is null
    and contest_end < current_date
)
select
  (select count(*) from contest_registrations
    where user_id = :'account_id'
      and contest_id in (select id from finished_contests)) as kept_registrations,
  (select count(*) from logs
    where user_id = :'account_id'
      and deleted_at is null
      and exists (
        select 1 from contest_logs
        where contest_logs.log_id = logs.id
          and contest_logs.contest_id in (select id from finished_contests)
      )) as kept_logs;
```

Every other log and registration of the account is deleted.

Finally, review the contests that the account owns:

```sql
select id, title, contest_start, contest_end, "private", official, deleted_at
from contests
where owner_user_id = :'account_id';
```

The delete step keeps owned contests and shows them with the organiser
"Deleted organizer". Contest titles and descriptions are not changed.

## 2. Lock the account

Commit the lock on its own before you delete anything. Once it is committed, the
Tadoku API rejects every log, contest and registration write for this account
with `409` and `{"error":"account_deletion_in_progress"}`.

```sql
update users
set deletion_locked_at = coalesce(deletion_locked_at, now())
where id = :'account_id';
```

## 3. Delete the data

Run this transaction in the same `psql` session. The statements must run in this
order: the job insert reads the rows that the deletes remove, and the log
statements treat every log still attached to a contest as finished-contest
history. The CTE inserts the existing v1 contracts and captures their exact IDs
in `deletion_job_ids`, including an empty array when no invalidation is needed.

This direct SQL is part of this manual maintenance transaction. Application
writes continue to use typed jobs returned by features and transactional
`jobqueue.Enqueue`; see [Jobs and worker](../tadoku-api/jobs.md). Do not change
these message names or payloads without the consumer-version preflight.

```sql
begin;

with affected_contests as (
  select contest_id from contest_registrations where user_id = :'account_id'
  union
  select contest_logs.contest_id
  from contest_logs
  join logs on logs.id = contest_logs.log_id
  where logs.user_id = :'account_id'
), queued as (
  insert into jobs (task_type, payload)
  select 'leaderboard.invalidate_contest.v1',
         jsonb_build_object('contest_id', contest_id)
  from affected_contests
  where contest_id not in (
    select id from contests where deleted_at is null and contest_end < current_date
  )
  union all
  select distinct 'leaderboard.invalidate_official.v1',
                  jsonb_build_object('year', year)
  from logs
  where user_id = :'account_id'
    and eligible_official_leaderboard
  returning id
)
select coalesce(array_agg(id order by id), '{}'::bigint[]) as deletion_job_ids
from queued
\gset

-- Keep contest logs only for live logs in finished contests.
delete from contest_logs
using logs
where logs.id = contest_logs.log_id
  and logs.user_id = :'account_id'
  and (
    logs.deleted_at is not null
    or contest_logs.contest_id not in (
      select id from contests where deleted_at is null and contest_end < current_date
    )
  );

delete from contest_registrations
where user_id = :'account_id'
  and contest_id not in (
    select id from contests where deleted_at is null and contest_end < current_date
  );

-- Logs still attached to a contest are finished-contest history.
update logs
set
  description = null,
  frozen_at = coalesce(frozen_at, now())
where user_id = :'account_id'
  and exists (select 1 from contest_logs where contest_logs.log_id = logs.id);

delete from log_tags where user_id = :'account_id';

delete from logs
where user_id = :'account_id'
  and not exists (select 1 from contest_logs where contest_logs.log_id = logs.id);

delete from user_roles where user_id = :'account_id';
delete from profiles where user_id = :'account_id';
delete from account_deletion_requests where identity_id = :'account_id';

update contests
set owner_user_display_name = 'Deleted organizer'
where owner_user_id = :'account_id';

update users
set
  display_name = 'Deleted participant',
  deleted_at = coalesce(deleted_at, now()),
  updated_at = now()
where id = :'account_id'
  and deletion_locked_at is not null;
```

Check the last statement's output. If it printed `UPDATE 1`, commit:

```sql
commit;
\echo :deletion_job_ids
```

Record that exact job-ID array with the operation's evidence after the commit
succeeds, and keep this `psql` session for step 4. If the session is lost, restore
it with `\set deletion_job_ids '{123,124}'` using the recorded IDs. Keep every
committed batch's IDs if you rerun the deletion; a later empty batch does not
prove that an earlier batch completed.

If it printed `UPDATE 0`, the lock is missing. Run `rollback;` and go back to
step 2.

Moderation audit entries stay in `moderation_audit_log`, both entries where the
account was the moderator and entries where it was the target.

## 4. Leaderboards

The separate `tadoku-worker` processes the rows inserted into `jobs` in step 3.
It invalidates the cached leaderboards of the contests that lose the account's
registrations or logs, and the yearly and global leaderboards, in Valkey. On the next
read, each leaderboard is rebuilt from the Tadoku API database. The rebuilt
leaderboards leave out the account, because its registrations and logs in those
contests are gone and official leaderboards skip users with `deleted_at` set.
Finished contests keep their scores, so step 3 does not invalidate them. Every
leaderboard reads display names from `users` on each request, so finished
contests show the account as `Deleted participant`. You do not need to change
the cache by hand.

Use the recorded `deletion_job_ids` from the committed transaction to inspect
these jobs and any linked replay attempts in the same `psql` session:

```sql
with recursive requested as (
  select unnest(:'deletion_job_ids'::bigint[]) as original_id
), attempts as (
  select requested.original_id, task.id, task.state
  from requested
  join jobs as task on task.id = requested.original_id
  union all
  select attempts.original_id, task.id, task.state
  from attempts
  join jobs as task on task.replay_of_id = attempts.id
)
select requested.original_id,
       coalesce(bool_or(attempts.state = 'completed'), false) as completed,
       array_agg(attempts.id order by attempts.id)
         filter (where attempts.state in ('pending', 'running')) as outstanding_ids,
       array_agg(attempts.id order by attempts.id)
         filter (where attempts.state = 'failed') as failed_ids,
       count(attempts.id) = 0 as missing
from requested
left join attempts using (original_id)
group by requested.original_id
order by requested.original_id;
```

A batch with no IDs needs no invalidation. Otherwise every original ID needs
`completed = true` with `missing = false`; pending or running IDs remain work
for the worker. For an incomplete effect with failed IDs, inspect the failure
and repair its cause before using the [worker replay command](../tadoku-api/jobs.md#inspect-and-replay-retained-work).
Replay creates a linked job and leaves the failed original unchanged. Re-run
this query to observe the linked completion; do not wait for the original
failed row to become completed. A missing row is not proof of success: consult
retained operation evidence, since successful jobs older than three UTC
calendar months can be removed. Successful replay rows expire too while their
failed originals remain. If recorded completion has aged out, use that retained
evidence rather than replaying solely because the query no longer finds a
completed descendant.

## 5. Keto relations and feature access

Tadoku stores roles as `app:tadoku#admins` and `app:tadoku#banned` tuples. Remove
both through the Keto write API. The Verify section checks that none are left.

```bash
for relation in admins banned; do
  curl -sS -X DELETE \
    "$KETO_WRITE_URL/admin/relation-tuples?namespace=app&object=tadoku&relation=$relation&subject_id=$ACCOUNT_ID"
done
```

Flipt stores named-user feature access by account ID. The flags that support it
are listed in `services/tadoku-api/features/featureflags/domain.go`. For each of
those flags, call `GET /immersion/admin/feature-flags/{flagKey}/users/{userId}`
as an administrator. If the account has access, send `DELETE` to the same path
to revoke it.

## 6. Kratos identity

Delete the identity through the Kratos admin API:

```bash
curl -fsS -X DELETE "$KRATOS_ADMIN_URL/admin/identities/$ACCOUNT_ID"
```

A `404` response means the identity is already gone.

## Verify

In the Tadoku API database:

```sql
select display_name, deletion_locked_at, deleted_at from users where id = :'account_id';

with finished_contests as (
  select id from contests
  where deleted_at is null
    and contest_end < current_date
)
select
  (select count(*) from contest_registrations where user_id = :'account_id') as kept_registrations,
  (select count(*) from logs where user_id = :'account_id') as kept_logs,
  (select count(*) from contest_registrations
    where user_id = :'account_id'
      and contest_id not in (select id from finished_contests)) as other_registrations,
  (select count(*) from contest_logs
    join logs on logs.id = contest_logs.log_id
    where logs.user_id = :'account_id'
      and contest_logs.contest_id not in (select id from finished_contests)) as other_contest_logs,
  (select count(*) from logs
    where user_id = :'account_id'
      and (
        deleted_at is not null
        or frozen_at is null
        or description is not null
        or not exists (select 1 from contest_logs where contest_logs.log_id = logs.id)
      )) as unscrubbed_logs,
  (select count(*) from log_tags where user_id = :'account_id') as log_tags,
  (select count(*) from user_roles where user_id = :'account_id') as user_roles,
  (select count(*) from profiles where user_id = :'account_id') as profiles,
  (select count(*) from account_deletion_requests where identity_id = :'account_id') as deletion_requests,
  (select count(*) from contests
    where owner_user_id = :'account_id'
      and owner_user_display_name <> 'Deleted organizer') as named_contests;
```

The `users` query returns one row, with `Deleted participant` as the display name
and both timestamps set. In the second query, `kept_registrations` and
`kept_logs` match the preflight, and every other count is `0`.

In Keto and Kratos:

```bash
# Expect "relation_tuples":[]
curl -fsS "$KETO_READ_URL/relation-tuples?namespace=app&subject_id=$ACCOUNT_ID"

# Expect 404
curl -s -o /dev/null -w '%{http_code}\n' "$KRATOS_ADMIN_URL/admin/identities/$ACCOUNT_ID"
```

The admin user list reads identities from an in-memory Kratos cache. The account
disappears from that list within five minutes.
