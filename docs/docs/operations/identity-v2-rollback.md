---
title: Identity v2 rollback
description: When and how production identity could return to the legacy Kratos stack, what each reversal restores and loses, the dependencies to prove first, and how to verify database recovery without assuming backup coverage.
---

# Identity v2 rollback

Read this when a production incident in sign-in, sessions or account data might
call for reverting identity changes or restoring identity data.

**Fix forward by default.** The canonical tenant `tadoku/prod` runs on native accounts, and the legacy
Kratos stack is offline. Setting the `identity-v2` flag to `false` or routing
the account host back does **not** restore the old system on its own: with
Kratos offline, every user is signed out and nobody can sign in. Bring back
every legacy dependency first, and treat any return to Kratos as a stateful
migration that needs reconciliation.

## Current production identity

| Part | Current state |
| --- | --- |
| Account selection | `tadoku/prod` selects native accounts through `identity-v2`; see [Account-app selection](../tadoku-api/identity-selection.md) |
| Accounts and sessions | `accounts`, `sessions` and `recovery_tokens` in the Tadoku API database, with the `tadoku_session` cookie; see [Identity accounts and sessions](../tadoku-api/identity.md) |
| Account host | `account.tadoku.app` redirects to `authv2.tadoku.app`, keeping the path and query; `/account-recovery` maps to `/recovery` |
| Legacy account frontend | `auth` has no running replicas; its production route only redirects |
| Ory Kratos | Kratos and its courier have no running Pods. Its database, Secrets and manifests are retained |
| Gateway | Oathkeeper's cookie-session authenticator calls the API's session callback; no route forwards to Kratos |
| Roles | Keto, unchanged. Imported accounts kept their Kratos UUIDs, so role tuples apply to both stacks |
| Mail | The worker sends canonical mail; see [Retained email](../tadoku-api/email.md) |

Production manifests live in the separate GitOps repository; this page names
what a rollback has to reverse, not its manifests.

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

## What a reversal restores

| Reversal | Restores | Does not restore |
| --- | --- | --- |
| Revert the Kratos shutdown | Kratos Pods and courier, reading the retained database as it was at shutdown | Anything written since the switch. The courier can send mail still queued in that database |
| Revert the account-host move | The legacy account frontend on `account.tadoku.app` | Sign-in, unless Kratos and every caller route to it are live first |
| Set canonical `identity-v2` to explicit `false` | API selection of Kratos for every identity operation | Native sessions, which stop authenticating, so everyone is signed out. Native accounts created after the switch cannot sign in. Users who changed their password, email address or display name since the switch see the old values |

Never remove the flag as a rollback. A missing canonical flag uses the
registered safe default in `services/tadoku-api/internal/featureflags/registry.go`;
set an explicit `false` instead.

## Data is one-way

The account import copied Kratos into native accounts once. Nothing writes back
to Kratos: native registrations, password, email and display-name changes,
sessions and recovery since the switch exist only in the Tadoku API database.
The importer cannot rerun to repair a difference after native writes: its
preflight rejects canonical accounts missing from the source and accounts that
no longer match it; see [Import Kratos accounts](../tadoku-api/identity-import.md).

A stateful rollback therefore needs a reconciliation plan with an explicit,
authorized disposition for each class of change: accounts that exist only
natively, changed passwords, changed email addresses and deletions. Do not
promise users a seamless downgrade. Never copy password hashes, email addresses
or tokens off the cluster, and never reset or delete native accounts or their
history to make the two stores agree.

## Data that must stay

Keep all of the following unless the incident owner authorizes a specific
disposition:

- Native accounts, sessions, recovery tokens, retained email and access tokens.
- The retained Kratos database, its Secrets and its manifests.
- The import's recovery holders, which exist only in memory and cannot be
  copied out. Do not delete their Pods.

No step on this page restores, resets or deletes a database, and none cleans up
recovery Pods automatically.

## Return to Kratos

Prepare each step as a staged instruction and run it only under the incident
owner's authorization for that exact scope.

### Prove dependencies before selecting traffic

- [ ] Kratos Pods are Ready against the retained database, and its public and
      admin APIs answer inside the cluster.
- [ ] The API's `API_KRATOS_ADMIN_URL` and `API_KRATOS_PUBLIC_URL` point at the
      live Kratos services, and the API is ready; see
      [Configuration](../tadoku-api/configuration.md). An unset public URL
      fails closed for every Kratos cookie.
- [ ] The legacy account frontend serves with an image that runs Kratos flows,
      and webv2 and admin run images that read the selected provider's session.
- [ ] Gateway routes forward the legacy Kratos paths, and the self-service
      registration, settings and recovery policy is decided.
- [ ] The courier's handling of retained queued mail is decided.
- [ ] The reconciliation plan above is approved.

Stop if any item is missing. Partial readiness signs users out without giving
them a working sign-in.

### Revert in reverse order

1. For each production change, open an ordinary revert pull request of its
   exact merge commit on the GitOps repository's current `main`. Revert in
   reverse order of application: dependencies first, traffic selection last.
2. Check the composed result, merge through the normal compare-and-swap
   process and let Argo CD reconcile. If an Application syncs manually, repeat
   the selective sync recorded with the original change, with pruning
   disabled and no hooks. Do not apply resources by hand.
3. Read back each Application's revision, sync and health before the next
   revert.
4. Change traffic selection last: the account-host route, then the flag.

### Verify

- `https://tadoku.app` answers `200`.
- A guest `GET https://tadoku.app/api/internal/identity/session` answers `200`
  with the expected `account_app`; `503` means selection is unavailable.
- A request to `/api/internal/identity/internal/v1/session` without the callback
  bearer answers `401`, not `503`.
- API logs show no `5xx` burst on identity routes.
- An authorized real user signs in, keeps the session across a reload and
  signs out. Guest checks never prove sign-in; report sign-in as unverified
  until this passes.

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
   them. The legacy Kratos database is a separate binding with its own recovery.
2. **List the tables in scope.** Native identity: `accounts`, `sessions`,
   `recovery_tokens`, `emails` and `access_tokens`, from migration 52. Related:
   `users`, `profiles`, `user_roles`, `account_deletion_requests`, `tenants` and
   `tenant_overrides`.
3. **Read coverage.** Record the backup schedule, retention, oldest backup and
   latest recoverable point. Point-in-time recovery needs a base backup plus
   write-ahead log covering the commits to recover. A backup taken before the
   import contains no native accounts, so it alone does not recover anything
   written since.
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
