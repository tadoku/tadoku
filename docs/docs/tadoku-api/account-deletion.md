---
title: Account deletion
description: How a deletion request locks an account and the worker's account.delete.v1 job removes its data, what a deletion keeps, and why the two transactions' lock orders are safe.
---

# Account deletion

Read this when you change how an account deletion is filed or executed, or
need to know what a deletion removes and keeps.

A deletion has two transactions: the request locks the account at once, and the
worker deletes its data later. Operators file requests with
`tadoku-worker delete-account`; see
[Account deletion](../operations/account-deletion.md).

The filing operation, `accountdeletion.Filing.File` in
`services/tadoku-api/app/accountdeletion/`, runs inside its caller's
transaction:

1. It refuses an administrator with `admin_account`, and a banned user who asks
   through self-service with `banned_account`, both `409`, before it writes
   anything. An operator request may delete a banned user.
2. It locks the `users` row and sets `deletion_locked_at`. A user whose deletion
   is already filed or done returns `account_deletion_in_progress`.
3. It records `request_account_deletion` with the actor, the reason, the
   `target_user_id` and a `source` of `operator` or `self_service`.
4. It enqueues `account.delete.v1` with the user ID.

From that commit, login answers `wrong_password`, session lookup treats the
cookie as a guest, recovery and settings find no account, and log, contest and
registration writes return `409` `account_deletion_in_progress`. A write that
committed before the lock is visible to the worker and is deleted with the rest.

The worker's `account.delete.v1` handler makes every change in one
transaction. It locks the `users` row first and changes nothing for a user who
is already deleted, so a duplicate or replayed job is harmless. A job for a user
without a deletion lock or without a `users` row fails without retry. Any other
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

Neither audit event stores the email address or display name.

The request and the worker take row locks in opposite orders. A settings write
locks the account row and then the `users` row; the worker locks the `users`
row and then deletes the account row. This is safe because every account row
lock in `services/tadoku-api/sql/identity/identity.sql` either excludes
deletion-locked users or skips locked rows, so once the deletion lock commits
no transaction holds an account row lock while it waits for that `users` row.
