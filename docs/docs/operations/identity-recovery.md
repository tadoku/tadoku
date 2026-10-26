---
title: Identity incident recovery
description: How to respond to a production incident in sign-in, sessions or account data, who authorizes each step, which identity data must stay and how to verify database recovery without assuming backup coverage.
---

# Identity incident recovery

Read this when a production incident in sign-in, sessions or account data might
call for reverting identity changes or restoring identity data.

**Fix forward by default.** The canonical tenant `tadoku/prod` runs on native
accounts, sessions and recovery tokens in the Tadoku API database, with the
`tadoku_session` cookie; see
[Identity accounts and sessions](../tadoku-api/identity.md). Treat any restore
of that data as a stateful change that needs an explicit, authorized
disposition.

## Triggers and authority

Revert only for a real burst of `5xx` answers from identity routes or the
session callback. These are not triggers on their own: a failed sign-in, which
can come from a credential or client problem, `401` answers to guests, a gap in
sparse logs, Argo CD revision drift, or a single failed request.

The maintainer is the incident owner and authorizes each reversal scope by
name. This page grants no authority. A prepared revert, an earlier approval for
a different scope or a failed check is not permission to run it. Without that
authorization, apply only the minimum containment needed to stop active harm,
such as data loss or a security exposure, and report it at once.

Use simple health gates when deciding: the site answers `200`, identity routes
answer without `5xx`, and a real sign-in works. Do not add waiting periods or
statistical thresholds to an authorized decision.

## Data that must stay

Keep native accounts, sessions, recovery tokens, retained email and access
tokens unless the incident owner authorizes a specific disposition. Never copy
password hashes, email addresses or tokens off the cluster, and never reset or
delete accounts or their history to repair an incident. No step on this page
restores, resets or deletes a database.

## Database recovery

Production runs on PlanetScale Postgres. This repository does not record the
production database, branch, backup schedule, retention or latest recoverable
point, and none of them is verified here. Read them from the provider account
before relying on any recovery.

PlanetScale documents [backups and restores](https://planetscale.com/docs/postgres/backups),
[point-in-time recovery](https://planetscale.com/docs/postgres/backups/point-in-time-recovery),
[role credentials](https://planetscale.com/docs/postgres/connecting/roles) and
[pricing](https://planetscale.com/docs/postgres/pricing). A restore creates a
new, separately billed branch with its own cluster and resets role passwords.
Custom and manual backups can also cost money.

1. **Identify the target.** From provider account metadata, record the database
   and branch that the production Tadoku API uses. Secret names do not identify
   them.
2. **List the tables in scope.** Native identity: `accounts`, `sessions`,
   `recovery_tokens`, `emails` and `access_tokens`, from migration 52. Related:
   `users`, `tenants` and `tenant_overrides`.
3. **Read coverage.** Record the backup schedule, retention, oldest backup and
   latest recoverable point. Point-in-time recovery needs a base backup plus
   write-ahead log covering the commits to recover.
4. **Pass the cost gate.** Get the incident owner's explicit approval for the
   cost before creating a restore branch or a manual backup.
5. **Restore in isolation.** Restore to a new branch, never over production.
   Creating credentials for it is part of the authorization; never reuse the
   production runtime credential.
6. **Check integrity with aggregates only.** Compare row counts per tenant and
   the latest `created_at` or `updated_at` per table with the expected recovery
   point. Never select email addresses, hashes or tokens.
7. **Decide the disposition.** A restored branch is evidence for
   reconciliation. Pointing production at it, keeping it or removing it are
   separate decisions for the incident owner.
