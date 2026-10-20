---
title: Import Kratos accounts
description: Configure, dry-run and validate the guarded Kratos-to-account import command, its transactional boundaries and the operational holds on shared environments.
---

# Import Kratos accounts

Read this when preparing or verifying the account import command.

The command imports a consistent Kratos PostgreSQL snapshot into the target
database's canonical `tadoku/prod` tenant. It preserves UUIDs, email spelling,
display names, bcrypt hashes, account state and creation timestamps. It never
creates sessions, contacts identity providers, changes feature flags or applies
schema migrations.

**Operational hold:** command development and disposable database tests do not
authorize a shared-development rehearsal or production import. Shared source
credential access, account freezes, the development seed switch, flag changes,
production cutover, default changes, deployment, rollback and Kratos removal
each remain held for separately authorized operational work. Do not run this
command against shared development or production as a verification step.

## Configuration and source schema

Build the standalone operator command from the repository root:

```sh
bazel build //services/tadoku-api/cmd/tadoku-identity-import
```

Supply individual database fields through a private, mode-0600 environment file
or a secret manager. Both prefixes require `HOST`, `DATABASE`, `USER`, `PASSWORD`
and `SSLMODE`; `PORT` defaults to 5432:

| Prefix | Connection |
| --- | --- |
| `IMPORT_SOURCE_POSTGRES_` | Read-only Kratos PostgreSQL database |
| `IMPORT_TARGET_POSTGRES_` | Tadoku PostgreSQL database with the independently applied identity schema |

The legacy `IMPORT_SOURCE_POSTGRES_URL` and `IMPORT_TARGET_POSTGRES_URL` variables
are rejected. Do not put connection URLs, credentials or password hashes in
command arguments, shell tracing, logs, reports or review comments. Source and
target must be different databases. The source role needs schema usage and
`select` on `networks`, `identities`, `identity_credentials` and
`identity_credential_types`. The target role needs schema usage and
`select`, `insert`, `update` and `delete` on `users` and `accounts`; the existing
runtime grants provide these privileges and permit the required table locks.
No superuser or row-security bypass is required.

The source follows the repository's pinned Kratos v26.2.0 PostgreSQL schema:
identity UUID, traits, state and timestamps come from `identities`; credentials
join through `identity_credentials.identity_credential_type_id` to
`identity_credential_types.name = 'password'`. The password is
`identity_credentials.config.hashed_password`, not an identity trait. See the
upstream [credential model](https://github.com/ory/kratos/blob/v26.2.0/identity/credentials.go)
and [password configuration](https://github.com/ory/kratos/blob/v26.2.0/identity/credentials_password.go).
The CLI tests create this schema with the actual pinned Kratos migrator.

The target uses the existing account and user schema. Existing matching users
and accounts retain all their metadata, including `updated_at`, and sessions
are untouched. Missing rows receive source timestamps. The user's timestamp
without time zone is explicitly interpreted as UTC. A different existing
display name, creation timestamp, email, hash or state is a conflict; deleted
or deletion-locked users are also conflicts. The command never overwrites,
resurrects or automatically repairs conflicting rows.

## Preflight and dry run

With configuration pointing only to disposable test databases, run:

```sh
bazel run //services/tadoku-api/cmd/tadoku-identity-import -- --dry-run
```

Dry run is also the default when no mode flag is supplied. It prints current
source counts, UUIDs and the names of fields or rows that would change. It never
prints email addresses, display names, hashes, tokens or connection details.
Database errors expose only the failing stage and SQLSTATE when available.

Before any target writes, preflight requires:

- Exactly one source network and a nonempty identity snapshot.
- Exactly one password credential for every identity, with matching network
  membership. Other credential types require separate review rather than
  silently discarding additional authentication factors.
- A complete 60-character `$2a$`, `$2b$` or `$2y$` bcrypt encoding with a valid
  cost and alphabet; no password migration hook.
- Valid emails with no duplicates after the same normalization used by local
  accounts, plus display names of 2–32 Unicode characters with visible content.
- Valid source UUIDs, states and timestamps.
- Matching existing target rows, no email owned by a different UUID, and no
  canonical account absent from the source snapshot. Unrelated user rows and
  other tenants are retained.

There is no hardcoded historical identity count. Review the current
`source_identities` count and all proposed changes. A conflict stops the command
with a nonzero exit code, UUID and field name. Investigate it without exposing
values; do not delete or rewrite data to force a passing import.

## Apply and validate on disposable databases

Pass the count observed in the dry run explicitly. For the two-identity test
fixture:

```sh
bazel run //services/tadoku-api/cmd/tadoku-identity-import -- --apply --expect-count 2
bazel run //services/tadoku-api/cmd/tadoku-identity-import -- --dry-run --expect-count 2
```

`--apply` requires a positive expected count and cannot be combined with
`--dry-run`. A changed count fails before target writes. A successful apply
prints `committed_accounts` for newly inserted accounts and `target_accounts`
for the validated total. A repeat dry run must print zero insertions and an
unchanged-account count equal to `source_identities`. A repeat apply succeeds
without changing existing rows or sessions.

For additional count validation, run this in the disposable target database
through an authenticated PostgreSQL client that does not echo credentials:

```sql
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true);
select count(*) as canonical_accounts
from accounts
where tenant = current_setting('tadoku.tenant');
select count(*) as missing_users
from accounts a
left join users u on (u.tenant, u.id) = (a.tenant, a.user_id)
where a.tenant = current_setting('tadoku.tenant') and u.id is null;
commit;
```

The account count must equal the source snapshot count and `missing_users` must
be zero. The permanent CLI tests additionally compare every imported field,
source and target snapshots, unchanged metadata and other-tenant data without
printing protected values.

## Transaction and retry boundaries

Source reads use one read-only, repeatable-read PostgreSQL transaction. Counts
and payloads therefore belong to the same snapshot. The source transaction
finishes before opening the target transaction; there is no distributed
transaction or automatic source freeze. An eventual separately approved shared
import must arrange source quiescence and post-import reconciliation. Equal
counts alone cannot detect a source credential change after the snapshot.

The target uses the ordinary transaction-local tenant executor. Both modes
acquire `share row exclusive` locks on `users` and `accounts` before target
preflight. These table locks block writers across tenants while held and
serialize concurrent import commands. Lock acquisition has a five-second
timeout; command execution has a five-minute deadline. All source validation
and target conflict checks finish before the first insert. Every insert and
the final count check share one target transaction; a failure rolls back the
entire batch. A commit transport failure can leave the outcome uncertain:
inspect a fresh dry run before retrying. Never infer failure from missing
success output alone.

## Repeatable CLI verification

Use the disposable PostgreSQL setup in [Testing](./testing.md). The loopback
fixture URL must satisfy its guard. Then run:

```sh
bazel test //services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import_test --test_output=all
bazel test //services/tadoku-api/cmd/tadoku-identity-import:tadoku-identity-import_test --@rules_go//go/config:race --test_output=all
```

Tests execute the built command as a subprocess with explicit fixture-only
configuration. Each scenario owns separate disposable source and target
databases; the source is initialized by the real Kratos migrator and the target
uses all current Tadoku migrations. No provider login, shared service, dev-cli
overlay or deployment is part of this verification. Retain the tested revision,
commands and Bazel test report outside source commits.
