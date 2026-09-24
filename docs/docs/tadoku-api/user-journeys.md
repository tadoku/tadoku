---
title: User journeys
description: How Tadoku API user journeys chain real requests against one reset to prove the important user flows, and how steps, cast members, time and goldens work.
sidebar_position: 7
---

# User journeys

Read this when you add or change a step in `services/tadoku-api/e2e/user_journeys_test.go`
or its fixtures under `services/tadoku-api/e2e/testdata/journeys/`. The general
test principles are in [Testing](./testing.md), and single-request golden cases
are in [HTTP end-to-end tests](./http-e2e.md).

A user journey chains requests against one reset, so every state after its
first step comes from the API itself: write and read paths must agree without a
handwritten seed between them, business time can move between steps, and one
user can observe another's writes. Journeys keep the functionality users expect
working, so cover every important user journey in the application.

All journeys live in `e2e/user_journeys_test.go`, one explicit Go step table per
journey passed to `runJourney`. They run only against Tadoku API and keep
dependent steps in one scenario. Fixtures live under `e2e/testdata/journeys/`:

```text
e2e/testdata/journeys/
  cast.json                  # cast member -> bearer token
  relationships.json         # cast Keto tuples, seeded for every journey
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
  a request, verify or job step.
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
  verify steps for effects no endpoint exposes, such as soft deletes, outbox
  rows and audit entries; API-observable persistence belongs in the next
  request or a repository test.
- Verify queries end with `order by` and select only application-supplied
  columns; database-defaulted IDs and `now()` timestamps are not deterministic.
- A job step runs background work and has no fixture directory. It can run a
  worker's synchronous pass when that is the behavior under test, or start the
  worker's real polling loop, wait for its ready signal and for an event written
  after startup, and cancel and join the worker during cleanup.

### Reset and time

- The runner resets PostgreSQL and Keto once per journey: cleanup, then the
  optional shared `journeys/setup.sql` and the shared
  `journeys/relationships.json`, then the journey's own files. There are no
  per-step seeds and no resets between steps.
- The JWT clock stays at the fixture instant for the whole journey, while each
  step's `at` sets its business instant through `timex`.
- The runner stops at the first failing step. Unknown entries in a journey or
  step directory fail the journey.

### Cast

- Cast members are `guest`, `user`, `user2`, `admin` and `banned`, plus the
  implicit `none`. `admin` and `banned` hold their `app:tadoku` tuples through
  the shared `relationships.json`.
- To add a member, sign a token with the
  [fixture recipe](./http-e2e.md#signing-a-new-fixture-token), append its public key when you
  use a new signing key, and add the token to `cast.json` and any tuple to
  `relationships.json`.

### Regenerating journey goldens

The same `-update-goldens` command and source root as the golden-case tables
also cover journey `golden.http` and `verify.json` files. It never creates a
missing file, so add empty placeholders before recording a new journey, and
review the complete diff.
