---
title: Webv3 checks and browser tests
description: The Webv3 build checks, its Playwright browser suite, the deployed branch tenant credentialed runs use, and the shared fixture accounts and login budget every spec relies on.
sidebar_position: 3
---

# Webv3 checks and browser tests

Read this when you run Webv3's checks or browser suite, or write a spec that
needs an account.

From `frontend/`, with Node 22 and the development server running:

```sh
pnpm --filter webv3 typecheck
pnpm --filter webv3 lint
pnpm --filter webv3 build
pnpm check:paper-boundaries
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm --filter webv3 test:e2e
```

The browser suite lives in `frontend/apps/webv3/e2e/`, one spec per feature
area; each feature page's checks section says what its spec covers. A spec
reads `WEBV3_BASE_URL` for the server under test and `TADOKU_API_URL` to compare
pages with the API, and skips what its environment cannot run rather than
failing. A skip names a missing environment variable, never missing tenant data:
a test that needs data creates it in setup. Specs that write create their own
data in the selected tenant, never assume it is empty and never depend on an
earlier run, another spec's data or a seeded round's dates. Specs select elements
by role, label or text; a CSS class is styling, not a contract. Browser tests
cover what the pages show for each API answer;
the API's golden cases in `services/tadoku-api/e2e/testdata/` own the rules
behind those answers, such as contest creation's account age and yearly quota.
The yearly quota counts every contest an account has created, deleted ones
included, so only Dev Admin, who is exempt, creates contests on every run.
The suite needs Chromium installed through Playwright. `PLAYWRIGHT_OUTPUT_DIR`
selects where reports and screenshots go, defaulting to a temporary directory
outside the source tree.

Chromium runs with HTTP/2 disabled. A branch's Vite server sends each page as
about 200 module requests, and the development ingress closes an HTTP/2
connection after 1,000 requests, failing whichever modules are in flight; that
page then never hydrates. A wider timeout cannot help it.

## Credentialed runs and fixture accounts

Tests that need an account run against a development-cluster deployment of the
pull request's branch with its own tenant: a dev-cli branch, whose `tenant` hook
seeds Dev Reader and Dev Admin into `tadoku/<route>` (see
[Branch tenants](../develop/development-data.md#branch-tenants)). Pull-request
deploys create no fixture accounts and no administrator; to run the credentialed
suite there, Dev Reader and Dev Admin must first exist in the branch tenant and
their credentials be supplied through the variables below. Run it from
`frontend/` with the branch's hosts:

| Variable | Value |
| --- | --- |
| `WEBV3_BASE_URL` | `https://<route>.next.tadoku.dev.lab` |
| `TADOKU_API_URL` | `https://<route>.tadoku.dev.lab/api/internal` |
| `E2E_READER_EMAIL`, `E2E_READER_PASSWORD` | Dev Reader in the branch tenant |
| `E2E_SECOND_EMAIL`, `E2E_SECOND_PASSWORD` | Dev Admin in the branch tenant |

Without the Dev Reader credentials, tests that need an account skip; the
container smoke runs that way.

The `tenant` hook also seeds sixty participants, with high scores, into the
branch tenant's Dev Tadoku Round, so a `learner` who joins it and logs a little
ranks below the first standings page. Pull-request deploys do not have them.

Specs get accounts only from the fixtures in
`frontend/apps/webv3/e2e/logging-fixtures.ts`, never by signing in or
registering themselves:

- `readerContext` and `adminContext` hold Dev Reader's and Dev Admin's shared
  sessions. Use them when signing in is incidental and the test's writes are
  safe to repeat. Never log them out, and restore any role change before the
  test ends.
- `learner` is a newly registered, signed-in account for one test, with its ID
  and credentials. Use it only when the test needs an untouched account, such as
  one with no logs or no registrations: each use spends a registration. The API
  refuses contest creation to accounts younger than a month, so a learner cannot
  create contests.
- `returningLearner` has the same shape but is one stored account reused across
  tests and runs. Use it when the test needs a young, signed-in account and its
  writes are safe to repeat.
- `loginLearner` is that stored account's credentials. Use it with `signIn`,
  `authenticate` and `logOut` when the test is about signing in, signing out or
  switching accounts through authv2. Each test that takes it may sign in once.

The identity API admits ten logins per account and ten registrations per tenant
and client IP, each in 15 minutes. The fixtures keep the Dev Reader and Dev
Admin sessions and the stored learner in `tadoku-webv3-e2e-sessions` under the
system temporary directory, keyed by host, and later workers and runs reuse them
while `GET /identity/session` still accepts them, so a failed test's restarted
worker spends nothing. The stored learner also records its logins; after eight
in 15 minutes the fixtures register a replacement. Switch to Dev Admin with
`adminCookies` rather than typing its password, because every login with a shared
fixture account spends from a budget no replacement can restore. A complete run
must stay repeatable straight away on the same tenant: a test that adds a
registration or a login with a fixture account on every run must keep two
consecutive runs within both budgets. Delete that directory to force new
sign-ins.

Two contests a run cannot create come from the `tenant` hook's seed, and
`frontend/apps/webv3/e2e/seeded.ts` derives their IDs from the branch tenant as
the seed does: Dev Finished Round, which ended before the run and holds a Dev
Reader log, and Dev Scoring Round, which has started and scores with its own
override rule set. A spec may change the seeded log, since the finished round
keeps its saved score. Development clusters run Tadoku API's scoring engine in
shadow mode, so saved scores follow the legacy rules and only the score preview
applies rule sets; `scoring-outcomes.spec.ts` reads the mode from the logging
configuration and expects the saved scores that mode produces.

A page's own upstream reads happen on the server, out of the browser's reach,
and the deployed branch's server cannot reach the machine running the suite.
`frontend/apps/webv3/e2e/fault-server.ts` therefore starts this checkout's
development server with its API behind a local proxy to `TADOKU_API_URL` that
answers chosen requests with a status, a dropped connection, a malformed body
or a delay; `outages.spec.ts` uses it for partial outages and recovery. Fault
the API this way only to fail a read: a test never fakes a successful write.

`frontend/apps/webv3/e2e/release-journey.spec.ts` is the signed-in journey a
release depends on: a newly registered learner logs in through authv2, logs,
submits to a contest, finds the log in their profile and the standings, edits
and deletes it, and logs out.

`frontend/apps/webv3/e2e/mutation-freshness.spec.ts` checks that a page visited
before a save, an administrator's edit, a registration or a moderation removal
shows the change when reached again through the router, with no document load.
Its assertions don't wait for the server: the API serves every board as of the
writes committed before the read, so only the browser's cache can be stale.

`pnpm --filter webv3 build` creates a local Node build in `.output`. This is a build
check; it does not publish or deploy the application.
