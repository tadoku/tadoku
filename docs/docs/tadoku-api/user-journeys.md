---
title: User journeys
description: How Tadoku API user journeys chain real requests against one reset to prove the important user flows, and how steps, cast members, time and goldens work.
sidebar_position: 7
---

# User journeys

Read this when you add or change a journey step in `services/tadoku-api/e2e/`
or its fixtures under `services/tadoku-api/e2e/testdata/journeys/`. The general
test principles are in [Testing](./testing.md), and single-request golden cases
are in [HTTP end-to-end tests](./http-e2e.md).

A user journey chains requests against one reset, so every state after its
first step comes from the API itself: write and read paths must agree without a
handwritten seed between them, business time can move between steps, and one
user can observe another's writes. Journeys keep the functionality users expect
working, so cover every important user journey in the application.

Journeys are grouped by domain, one explicit Go step table per journey passed
to `runJourney`:

| File | Journeys |
| --- | --- |
| `services/tadoku-api/e2e/user_journeys_test.go` | Tenant isolation, languages, contests, registrations and scoring rule sets |
| `services/tadoku-api/e2e/log_journeys_test.go` | Logs and leaderboard jobs |
| `services/tadoku-api/e2e/access_journeys_test.go` | Roles, bans and the user list |
| `services/tadoku-api/e2e/content_journeys_test.go` | Announcements, pages and posts |

Journeys keep dependent steps in one scenario and run the production API router
and, where needed, the worker application in process. Fixtures live under
`e2e/testdata/journeys/`:

```text
e2e/testdata/journeys/
  cast.json                  # cast member -> bearer token
  relationships.json         # cast roles, seeded for every journey
  <Journey>/
    setup.sql                # optional, journey-specific
    relationships.json       # optional, journey-specific
    01_<step>/
      request.http           # no Authorization header
      golden.http
    02_<step>/
      verify.sql             # one ordered query over deterministic columns
      verify.json
```

### Steps

- Steps are numbered by their position in the table. Each step is exactly one of
  a request, verify, Valkey snapshot or job step.
- A request step names the cast member that sends `request.http`; the runner
  injects that member's token, and `none` sends no credentials. Do not embed
  tokens in journey `request.http` files.
- A request step's optional `others` map replays the same request as other
  members before the primary request and checks only their status. Add replays
  only where they add information: identities that must be rejected on
  mutating steps, `banned` included, and a second user only to observe limited
  visibility of a resource.
- Replays must not legitimately mutate state, so an identity that is supposed to
  succeed at a write gets its own step.
- A verify step runs `verify.sql`, aggregates the rows into one JSON array in
  query order and compares the indented result with `verify.json`. Reserve
  verify steps for effects no endpoint exposes, such as soft deletes, job
  rows and audit entries; API-observable persistence belongs in the next
  request or a repository test.
- Verify queries end with `order by` and select only application-supplied
  columns; database-defaulted IDs and `now()` timestamps are not deterministic.
- A `run_worker` job step has no fixture directory. It starts the worker against
  the journey's disposable database and Valkey fixture, waits until every job
  eligible for the base worker has completed, including jobs those jobs enqueue,
  and stops it before the next step.
  Tenants overridden for `tadoku-worker` remain pending. A `run_branch_worker`
  job step requires `forTenant` with a parsed non-production key and completes
  only that tenant's jobs.
- A Valkey snapshot step uses `valkey: "<name>"` and a directory containing only
  `verify.json`. It checks ownership of the suite's leased Valkey database,
  scans leaderboard keys, removes duplicate scan results and compares the
  sorted names with the golden. Snapshots include tenant prefixes, cache
  markers and invalidation generations; they exclude the lease key.

### Reset and time

- The runner resets PostgreSQL and Valkey once per journey: cleanup, then the
  optional shared `journeys/setup.sql`, the journey's own `setup.sql` and the
  account cast, then the shared and the journey's own `relationships.json`, as
  [Relationship files](./http-e2e.md#relationship-files) describes. There are
  no per-step seeds and no resets between steps. A journey may seed an external
  cache or install a fault, such as a failing trigger, after reset; it removes
  a fault in cleanup.
- The JWT clock stays at the fixture instant for the whole journey, while each
  step's `at` sets its business instant through `timex`.
- The runner stops at the first failing step. Unknown entries in a journey or
  step directory fail the journey.

### Cast

- Cast members are `guest`, `user`, `user2`, `admin` and `banned`, plus the
  implicit `none`. `admin` and `banned` hold their roles on `tadoku/prod`
  through the shared `relationships.json`. Tenant variants use the same
  subjects in signed tokens with a different tenant claim, and hold a role
  there only through a relationship for that tenant.
- To add a member, sign a token with the
  [fixture recipe](./http-e2e.md#signing-a-new-fixture-token), append its public key when you
  use a new signing key, and add the token to `cast.json` and any relationship
  to `relationships.json`.

### Tenant isolation

`TestTenantIsolationJourney` uses the restricted PostgreSQL application role
and real Valkey. Its requests create one log in each tenant, then check
separate leaderboards, cache keys, scoped worker outcomes and bans.

Flipt evaluation and management use the suite's tenant-aware protocol fixture.
This journey therefore verifies application routing and independent flag
state, not the Flipt server's storage or SDK snapshot behavior. Those boundaries
require separate real-provider verification.

### Regenerating journey goldens

The same `-update-goldens` command and source root as the golden-case tables
also cover journey `golden.http` and `verify.json` files. It never creates a
missing file, so add empty placeholders before recording a new journey, and
review the complete diff.
