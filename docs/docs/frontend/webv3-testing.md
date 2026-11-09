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
failing. Specs that write create their own data in the selected tenant and never
assume it is empty. Browser tests cover what the pages show for each API answer;
the API's golden cases in `services/tadoku-api/e2e/testdata/` own the rules
behind those answers, such as contest creation's account age and yearly quota.
The suite needs Chromium installed through Playwright. `PLAYWRIGHT_OUTPUT_DIR`
selects where reports and screenshots go, defaulting to a temporary directory
outside the source tree.

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
  and credentials. Use it when the test needs an untouched account or leaves
  state a rerun would trip over. The API refuses contest creation to accounts
  younger than a month, so a learner cannot create contests.
- `loginLearner` is one newly registered account per worker, given as
  credentials. Use it with `signIn`, `authenticate` and `logOut` when the test
  is about signing in, signing out or switching accounts.

The identity API admits ten logins per account and ten registrations per tenant
and client IP, each in 15 minutes. The fixtures store the Dev Reader and Dev
Admin sessions in `tadoku-webv3-e2e-sessions` under the system temporary
directory, keyed by host and email, and later workers and runs reuse them while
`GET /identity/session` still accepts them. A complete run must stay
repeatable straight away on the same tenant: a test that adds a login with a
fixture account or a registration on every run must keep two consecutive runs
within both budgets. Delete that directory to force new sign-ins.

`frontend/apps/webv3/e2e/mutation-freshness.spec.ts` checks that a page visited
before a save, an administrator's edit, a registration or a moderation removal
shows the change when reached again through the router, with no document load.
Its assertions don't wait for the server: the API serves every board as of the
writes committed before the read, so only the browser's cache can be stale.

`pnpm --filter webv3 build` creates a local Node build in `.output`. This is a build
check; it does not publish or deploy the application.
