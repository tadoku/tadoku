---
title: Tenant lifecycle command
description: The tadoku-tenant command that provisions, tears down and overrides test tenants, its database role, settings, image and idempotency.
sidebar_position: 9.2
---

# Tenant lifecycle command

Read this when you change tenant provisioning, teardown or worker overrides, or
run `tadoku-tenant` yourself.

`services/tadoku-api/cmd/tadoku-tenant` provides provisioning, teardown and
worker overrides. Its tenant argument is a full parsed key, such
as `tadoku/branch-0123abcd`; the id must end in a hyphen and eight lowercase
hexadecimal digits. Bare route ids, `tadoku/prod` and production registry rows
are refused before provider writes.

```sh
bazel run //services/tadoku-api/cmd/tadoku-tenant -- provision \
  --tenant tadoku/branch-0123abcd --flipt-features k8s/dev/base/flipt/features.yaml \
  --tester 11111111-1111-4111-8111-111111111111
bazel run //services/tadoku-api/cmd/tadoku-tenant -- override set \
  --tenant tadoku/branch-0123abcd --component tadoku-worker
bazel run //services/tadoku-api/cmd/tadoku-tenant -- override clear \
  --tenant tadoku/branch-0123abcd --component tadoku-worker
bazel run //services/tadoku-api/cmd/tadoku-tenant -- teardown \
  --tenant tadoku/branch-0123abcd
```

All commands refuse credentials that neither own the tenant registry nor
belong to the `tadoku_tenant_lifecycle` role. That role needs only schema usage,
`select`, `insert` and `delete` on `tenants` and `tenant_overrides`, and
`execute` on `copy_production_platform_scoring(text)` and
`copy_production_published_pages(text)`; registry policies limit
its writes to test rows, and deleting a row cascades through the owner's
foreign keys. Production branch Jobs use it, and development uses the owner.
Credentials come from `TENANT_POSTGRES_HOST`,
`TENANT_POSTGRES_PORT` (default 5432), `TENANT_POSTGRES_DATABASE`,
`TENANT_POSTGRES_USER`, `TENANT_POSTGRES_PASSWORD` and `TENANT_POSTGRES_SSLMODE`.
`TENANT_POSTGRES_URL` is rejected. Provision and teardown additionally require
`TENANT_KETO_READ_URL`, `TENANT_KETO_WRITE_URL`, `TENANT_VALKEY_URL` and
`TENANT_FLIPT_MANAGEMENT_URL`; `TENANT_FLIPT_ENVIRONMENT` defaults to `test` and
must differ from `production`. The Keto and Flipt management URLs must be
absolute HTTP(S) without credentials, query or fragment; trailing slashes are
removed. Use the direct provisioning-authorized Flipt endpoint, rather than the
API's restricted service-token route. Override commands need only PostgreSQL.
The only registered override component is `jobqueue.WorkerComponent`, currently
`tadoku-worker`.

`//services/tadoku-api/cmd/tadoku-tenant:image` packages the command for
branch provisioning outside development: a distroless base with the CA bundle
`verify-full` connections need, `/tadoku-tenant` as its entrypoint and the
repository's flag definitions at `/flipt/features.yaml`. Its `branch_push`
target sets no repository or tags; the caller supplies both. The development
`//.dev:tenant_push` image has no CA bundle.

Provision commits an idempotent test registry row under a transaction-local
advisory lock. In the same transaction it copies `tadoku/prod`'s active
platform scoring rule set and its rules into the tenant, with new ids and a
`platform_scoring_config` row pointing at the copy, so logs can be scored.
It then copies `tadoku/prod`'s published, undeleted CMS pages, each with only
the content revision it displays, keeping namespaces, slugs, titles and
timestamps under new ids. Drafts, scheduled and deleted pages, earlier
revisions and posts are not copied. Each copy is skipped when the tenant
already has platform scoring config or any page, so reruns keep the tenant's
own changes; contest rule sets and user data are never copied. It then grants
the canonical Keto parent and repeated `--tester` UUIDs, then creates only
missing Flipt resources from the strict boolean seed format.
Keto grants are checked by their complete tuple before creating them because
the provider accepts duplicate relationship writes.
Reruns preserve existing grants and flag values. Teardown deletes the tenant's
Keto relationships, cache keys and Flipt namespace before deleting the registry
row and its cascading data. Failure leaves the row for a retry; cleanup still
runs when a prior attempt already removed that row. Commands sharing a tenant
must be serialized by the caller's route lease. Provider HTTP calls have a
30-second timeout; the command has a five-minute deadline and observes shutdown
signals. Secrets are supplied through environment variables and never printed.
