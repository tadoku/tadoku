---
sidebar_position: 1
title: Account deletion
description: When an operator deletes a Tadoku account, how to file the deletion with tadoku-worker delete-account, follow pending deletions and their audit events, replay a failed account.delete.v1 job and verify the result.
---

# Account deletion

Read this when a user has asked an operator to delete their Tadoku account.

An operator files the deletion with the worker's `delete-account` command. It
locks the account at once, records a `request_account_deletion` audit event
and enqueues an `account.delete.v1` job. The worker then deletes the data in
one transaction, records a `complete_account_deletion` event and sends an
`account.deleted` notice to the account's address.
[Account deletion](../tadoku-api/account-deletion.md) lists what a deletion
removes and keeps. This guide writes no SQL; every query below is
read-only.

## When an operator deletes

Start only after you have confirmed that the request comes from the account's
owner. Operators delete:

- a banned account;
- an administrator's account, after removing the administrator role, because
  the command refuses administrators;
- the account of a user who cannot sign in.

A deletion cannot be undone. Only a database backup holds the earlier rows; see
[Database recovery](./identity-recovery.md#database-recovery).

## Prerequisites

- The account ID, which is `accounts.user_id` and `users.id`. Imported accounts
  keep their former Kratos identity ID.
- Your own Tadoku user ID. The command records it as the request's actor.
- A shell in a running `tadoku-worker` container of a release that registers
  `account.delete.v1`. The command needs only the worker's `WORKER_POSTGRES_*`
  settings and, like `replay`, acts on `tadoku/prod` only.
- For the queries, `psql` as the non-owner runtime role. Row-level security
  returns no rows until a transaction sets the tenant; see
  [Tenancy](../tadoku-api/database.md#tenancy). Never use the owner role: it
  bypasses row-level security. Connect with the `search_path` that the Tadoku
  API uses.

Set the variables once for `psql`. The Tadoku API compares contest dates with
the current UTC date, so set the `psql` session to UTC too:

```sql
\set ON_ERROR_STOP on
\set tenant 'tadoku/prod'
set time zone 'UTC';
```

To find the account ID from an email address, look it up by the address
lowercased with surrounding whitespace removed, the form that login uses, then
set it:

```sql
begin transaction read only;
select set_config('tadoku.tenant', :'tenant', true);
select user_id, state from accounts where email_normalized = '<NORMALIZED_EMAIL>';
commit;

\set account_id '<ACCOUNT_ID>'
```

## 1. Preflight

Count the finished-contest history that the deletion keeps, and review the
contests the account owns. A contest is finished once its end date is before
the current UTC date.

```sql
begin transaction read only;
select set_config('tadoku.tenant', :'tenant', true);

with finished_contests as (
  select id from contests where contest_end < current_date
)
select
  (select count(*) from contest_registrations
    where user_id = :'account_id'
      and contest_id in (select id from finished_contests)) as kept_registrations,
  (select count(distinct logs.id) from logs
    join contest_logs on contest_logs.log_id = logs.id
    where logs.user_id = :'account_id'
      and logs.deleted_at is null
      and contest_logs.contest_id in (select id from finished_contests)) as kept_logs;

select id, title, contest_start, contest_end, "private", official, deleted_at
from contests
where owner_user_id = :'account_id';

commit;
```

Record both counts. Every other log and registration of the account is
deleted. Owned contests stay and show the organizer as "Deleted organizer";
their titles and descriptions do not change.

## 2. File the deletion

In the worker container:

```sh
tadoku-worker delete-account --user <ACCOUNT_ID> --actor <YOUR_USER_ID> \
  --reason 'Owner asked by email on 2026-10-26'
```

- On success it logs `account deletion requested` with `user_id` and the new
  `job_id`. Record the job ID.
- From then on the account cannot sign in, its sessions and recovery links stop
  working, and a log, contest or registration write that is already
  authenticated as the user returns `409` with
  `{"error":"account_deletion_in_progress"}`.
- It changes nothing and exits with an error for an administrator
  (`remove the administrator role before deleting this account`), for an
  account whose deletion is already filed or done
  (`account deletion in progress`) and for an unknown user ID
  (`local user not found`).

## 3. Follow pending deletions

The worker usually finishes within seconds. A pending deletion has
`deletion_locked_at` set and `deleted_at` empty. List every pending deletion
with its jobs, then the audit events of one account:

```sql
begin transaction read only;
select set_config('tadoku.tenant', :'tenant', true);

select users.id, users.deletion_locked_at, jobs.id as job_id, jobs.state,
       jobs.attempts, jobs.last_error, jobs.replay_of_id
from users
left join jobs
  on jobs.task_type = 'account.delete.v1'
  and jobs.payload->>'user_id' = users.id::text
where users.deletion_locked_at is not null
  and users.deleted_at is null
order by users.deletion_locked_at, jobs.id;

select user_id as actor_id, action, metadata, description, created_at
from moderation_audit_log
where action in ('request_account_deletion', 'complete_account_deletion')
  and metadata->>'target_user_id' = :'account_id'
order by created_at;

commit;
```

The request event names you as the actor, with `source` `operator` and your
reason. Once the job completes, the completion event holds the counts of
deleted and kept logs and registrations and of owned contests.

## 4. Replay a failed deletion

A deletion that uses up its attempts leaves the account locked, its address
taken and its job `failed`, and nothing is deleted. Read the job's
`last_error` and the worker's `job failed` log line for the same job ID, repair
the cause, then replay it with the
[replay command](../tadoku-api/jobs-operations.md#inspect-and-replay-retained-work):

```sh
tadoku-worker replay --id <FAILED_JOB_ID> --actor <YOUR_NAME> --reason '<repaired cause>'
```

Follow the new job with the queries in step 3. A job that failed with
`invalid_payload` named a user without a deletion lock or without a `users`
row; it changed nothing and a replay fails the same way. Replaying a job for an
account that is already deleted completes without changes.

## Verify

```sql
begin transaction read only;
select set_config('tadoku.tenant', :'tenant', true);

select display_name, deletion_locked_at, deleted_at from users where id = :'account_id';

with finished_contests as (
  select id from contests where contest_end < current_date
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
  (select count(*) from permission_relationships
    where subject_type = 'user' and subject_id = lower(:'account_id')) as permission_relationships,
  (select count(*) from accounts where user_id = :'account_id') as accounts,
  (select count(*) from sessions where user_id = :'account_id') as sessions,
  (select count(*) from recovery_tokens where user_id = :'account_id') as recovery_tokens,
  (select count(*) from contests
    where owner_user_id = :'account_id'
      and owner_user_display_name <> 'Deleted organizer') as named_contests;

commit;
```

The `users` query returns one row, with `Deleted participant` as the display
name and both timestamps set. In the second query, `kept_registrations` and
`kept_logs` match the preflight, and every other count is `0`. Step 3's audit
query returns exactly one request and one completion event. The
[email read API](../tadoku-api/email.md#read-retained-messages) shows one
`account.deleted` message to the account's address. The admin user list reads
native accounts on each request, so the account disappears from it at once.

The worker also invalidates the leaderboards of the unfinished contests that
lost the account's logs or registrations and the yearly leaderboards of its
official logs. Rebuilt leaderboards leave the account out, and finished
contests keep its scores under `Deleted participant`. No cache step is needed.

## Data a deletion keeps

- **Moderation audit entries** in `moderation_audit_log`, both where the
  account was the moderator and where it was the target.
- **Retained email** rows addressed to the account, including the
  `account.deleted` notice, until the worker's
  [email retention](../tadoku-api/email.md#retention-and-operations) removes
  them after three days in production. Deleting the recovery tokens invalidates
  every link in them.
- **Feature access.** Flipt allowlist segments declared in
  `k8s/dev/base/flipt/features.yaml` may still name the account ID. The ID
  stays on the tombstone `users` row and can no longer sign in.
- **Legacy identity data.** Production accounts were imported from Ory Kratos.
  Production Kratos is offline, and its database and Secrets are retained, so
  they still hold the identity's pre-import email address and password hash.
  This guide never starts Kratos or uses its self-service or admin APIs. Record
  the account ID with the operation's evidence so that the separately
  authorized disposition of retained legacy data can include it.
- **Database backups** keep earlier copies until the provider's retention
  removes them; see [Database recovery](./identity-recovery.md#database-recovery).
