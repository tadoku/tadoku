---
title: Account deletion
description: How an administrator files an account deletion through the admin endpoint, how filing locks the account and the worker's account.delete.v1 job removes its data, what a deletion keeps, the account-then-user lock order, and how to follow and replay a deletion.
---

# Account deletion

Read this when you file, follow or replay an account deletion, change how one
is filed or executed, or need to know what a deletion removes and keeps.

A deletion has two transactions: filing locks the account at once, and the
worker deletes its data later. Administrators file with
`POST /profile/users/{id}/account-deletion`; see
[File a deletion](#file-a-deletion).

## Filing

The admin operation, `RequestAccountDeletion` in
`services/tadoku-api/app/api/account_deletion.go`, requires an administrator,
records the signed-in administrator as the actor and passes the request to
`FileAccountDeletion`. Filing checks no actor and runs in its own transaction:

1. It locks the account row, then the `users` row, and sets
   `deletion_locked_at`. An unknown or deleted user returns `404`, and a user
   whose deletion is already filed returns `409`
   `account_deletion_in_progress`.
2. An administrator's request must name the account's email address. It is
   normalized as login normalizes it and compared with the address of the
   locked account row, so a concurrent address change cannot slip past it. A
   mismatch returns `400` `confirmation_mismatch`.
3. It refuses an administrator with `409` `admin_account`, and a banned user
   who asks through self-service with `409` `banned_account`. An administrator
   may delete a banned user. A refusal rolls the transaction back, so nothing
   is written. Roles are read only while filing holds the `users` row, which
   role changes also lock, so a ban or grant that commits first is seen.
4. It records `request_account_deletion` with the actor, the reason, the
   `target_user_id` and a `source` of `admin` or `self_service`.
5. It enqueues `account.delete.v1` with the user ID.

From that commit, login answers `wrong_password`, session lookup treats the
cookie as a guest, recovery and settings find no account, and log, contest and
registration writes and role changes for the user return `409`
`account_deletion_in_progress`. Those writes lock the `users` row, so a write
that committed before the lock is visible to the worker and is deleted with the
rest, and one still in flight waits for the deletion and is refused.

## Execution

The worker's `account.delete.v1` handler makes every change in one
transaction. It locks the account row and then the `users` row, and changes
nothing for a user who is already deleted, so a duplicate or replayed job is
harmless. A job for a user without a deletion lock or without a `users` row
fails without retry. Any other
failure rolls the whole deletion back and leaves the account locked until the
job is retried or replayed. A finished contest is one whose end date is before
the current UTC date, the rule that closes log editing.

| Data | Effect |
| --- | --- |
| `accounts`, `sessions`, `recovery_tokens` | Deleted. The address is free; a later registration creates an unrelated user. |
| `users` | Kept as a tombstone: `display_name` becomes `Deleted participant` and `deleted_at` is set. The ID and `created_at` stay. |
| Logs attached to a finished contest | Kept with their contest scores, without description or tags, and frozen. |
| Every other log, including soft-deleted logs, with its contest logs and tags | Deleted. |
| Registrations | Deleted in unfinished contests, kept in finished ones. |
| Contests the user owns | Kept; the stored organizer name becomes `Deleted organizer`. |
| `permission_relationships` with the user as subject | Deleted. |
| Leaderboards | One invalidation job per unfinished contest that loses a log or registration and per year with official logs. |
| Contest event feeds | `log.withdrawn` for each log leaving an unfinished contest and `log.updated` for each kept log whose description or tags were cleared. |
| `moderation_audit_log` | One `complete_account_deletion` event, with the deleted user as actor and the deleted and kept log and registration counts and the owned contest count. |
| `emails` | Unchanged; the retention pass removes earlier messages after three days in production and fourteen elsewhere. The worker adds an `account.deleted` notice to the account's address, unless that address cannot receive mail. |

Neither audit event stores the email address or display name. The
administrator's reason is free text in the request event's description.

Every contest event of a deletion carries the time the worker started, but feed
readers see them only when the transaction commits, which takes seconds on a
large account.

## Lock order

Every transaction that locks both an account row and its `users` row takes the
account row first: the settings writes, login and recovery (whose session and
token inserts lock the `users` row through their foreign keys), filing and the
worker. Filing therefore waits for a settings write that already holds the
account row, and the worker waits for any of them instead of deadlocking.

- Filing's account lock skips a user whose deletion was already filed when the
  statement started, and filing reads the committed deletion state before it
  waits for the `users` row. Filing again answers
  `account_deletion_in_progress` at once, even while the worker runs.
- A settings write queued behind filing's account lock gets the row after
  filing commits, rechecks its session and fails before it touches the `users`
  row.
- A login or recovery request queued the same way can still store a session or
  send a recovery email. The session never authenticates, because session lookup
  excludes deletion-locked users, and the worker deletes it and the recovery
  token.

## File a deletion

Start only after you have confirmed that the request comes from the account's
owner. Administrators delete:

- a banned account;
- an administrator's account, after removing the administrator role, because
  filing refuses administrators;
- the account of a user who cannot sign in.

A deletion cannot be undone. Only a database backup holds the earlier rows; see
[Database recovery](../operations/identity-recovery.md#database-recovery).

The admin UI cannot file a deletion yet. Until it can, send the request with
your administrator session. In production, the session cookie is
`tadoku_session`:

```sh
curl -sS -X POST "https://tadoku.app/api/internal/profile/users/<ACCOUNT_ID>/account-deletion" \
  --cookie "tadoku_session=<YOUR_SESSION>" \
  -H 'Content-Type: application/json' \
  --data '{"email":"<ACCOUNT_EMAIL>","reason":"<HOW_THE_OWNER_ASKED_AND_HOW_YOU_CONFIRMED_IT>"}'
```

- The account ID is `accounts.user_id` and `users.id`, the `id` that
  `GET /profile/users` lists. Imported accounts keep their former Kratos
  identity ID.
- `email` confirms the target: type the account's address as that list shows
  it. Letter case and surrounding whitespace do not matter.
- `reason` is required, at most 1,000 characters after trimming, and stays in
  the audit log for good. Describe the request and how you confirmed it; do not include the
  user's email address or display name.
- `202`, with no body, means the account is locked and its `account.delete.v1`
  job is queued; [Follow pending deletions](#follow-pending-deletions) finds the
  job by account ID. From then on the account cannot sign in, its sessions and
  recovery links stop working, and its writes return `409`
  `account_deletion_in_progress`.
- A refusal changes nothing: `400` `confirmation_mismatch` for another
  address, `404` for an unknown or deleted user, and `409` with
  `admin_account` or `account_deletion_in_progress`.

## Follow pending deletions

The worker usually finishes within seconds. A pending deletion has
`deletion_locked_at` set and `deleted_at` empty. Query with `psql` as the
non-owner runtime role and the `search_path` that the Tadoku API uses. Row-level
security returns no rows until a transaction sets the tenant; see
[Tenancy](./database.md#tenancy). Never use the owner role: it bypasses
row-level security.

```sql
\set ON_ERROR_STOP on
\set tenant 'tadoku/prod'
\set account_id '<ACCOUNT_ID>'

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

The request event names the administrator as the actor, with `source` `admin`
and the reason. Once the job completes, the completion event holds the counts
of deleted and kept logs and registrations and of owned contests, the user's
display name is `Deleted participant` and `deleted_at` is set. The admin user
list reads native accounts, so the account disappears from it at once.

## Replay a failed deletion

A deletion that uses up its attempts leaves the account locked, its address
taken and its job `failed`, and nothing is deleted. Read the job's
`last_error` and the worker's `job failed` log line for the same job ID, repair
the cause, then replay it with the
[replay command](./jobs-operations.md#inspect-and-replay-retained-work):

```sh
tadoku-worker replay --id <FAILED_JOB_ID> --actor <YOUR_NAME> --reason '<repaired cause>'
```

Follow the new job with the queries above. A job that failed with
`invalid_payload` named a user without a deletion lock or without a `users`
row; it changed nothing and a replay fails the same way. Replaying a job for an
account that is already deleted completes without changes.

## Data outside the deletion

- **Feature access.** Flipt allowlist segments declared in
  `k8s/dev/base/flipt/features.yaml` may still name the account ID. The ID
  stays on the tombstone `users` row and can no longer sign in.
- **Legacy identity data.** Production accounts were imported from Ory Kratos.
  Production Kratos is offline, and its database and Secrets are retained, so
  they still hold the identity's pre-import email address and password hash.
  Never start Kratos or use its self-service or admin APIs for a deletion.
  Record the account ID with the request's evidence so that the separately
  authorized disposition of retained legacy data can include it.
- **Database backups** keep earlier copies until the provider's retention
  removes them; see
  [Database recovery](../operations/identity-recovery.md#database-recovery).
