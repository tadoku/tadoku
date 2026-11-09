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

The browser suite in `frontend/apps/webv3/e2e/home.spec.ts` checks configured read-only
data in server-rendered HTML, hydration and mobile containment, local goal
validation/editing/persistence without network writes, unavailable browser
storage, theme persistence, mobile navigation and unknown routes.
`frontend/apps/webv3/e2e/leaderboard.spec.ts` compares each leaderboard's server-rendered
rows, scope and profile links with the API, follows page, year and period links,
checks a signed-in participant's row and contest action, and checks containment
at 390px and 1280px. It needs `TADOKU_API_URL` and the Dev Reader credentials
described under [native activity logging](webv3-logging.md), and uses the shared
Dev Reader session.
`frontend/apps/webv3/e2e/contests.spec.ts` covers the
[contest lists](webv3-contests.md), `frontend/apps/webv3/e2e/contest.spec.ts`
the [contest page and registration](webv3-contest.md) and
`frontend/apps/webv3/e2e/contest-activity.spec.ts` its updates, participant
profiles and moderation; their signed-in journeys use both fixture accounts and
create community contests in the selected tenant.
`frontend/apps/webv3/e2e/contest-admin.spec.ts` covers
[contest creation and scoring](webv3-contest-admin.md) and writes contests and
rule sets in the selected tenant.
These suites require reachable contest and standings data and Chromium installed through
Playwright. It preserves HTML reports and screenshots outside the source tree.
`frontend/apps/webv3/e2e/member-journey.spec.ts` signs in, walks from the
homepage through the current contest's standings, a participant's contest and
main profiles and the activity form to the saved log, and fails if any
navigation leaves the Webv3 host; it writes one log in the selected tenant.
`WEBV3_BASE_URL` selects the running server; `PLAYWRIGHT_OUTPUT_DIR` selects the
artifact directory, defaulting to a temporary directory.

## Credentialed runs and fixture accounts

Tests that need an account run against a deployed branch with its own tenant:
a dev-cli branch, whose `tenant` hook seeds Dev Reader and Dev Admin into
`tadoku/<route>` (see [Branch tenants](../develop/development-data.md#branch-tenants)).
Pull-request deploys create no fixture accounts and no administrator, so they
cannot run this suite. Run it from `frontend/` with the branch's hosts:

| Variable | Value |
| --- | --- |
| `WEBV3_BASE_URL` | `https://<route>.next.tadoku.dev.lab` |
| `TADOKU_API_URL` | `https://<route>.tadoku.dev.lab/api/internal` |
| `E2E_READER_EMAIL`, `E2E_READER_PASSWORD` | Dev Reader in the branch tenant |
| `E2E_SECOND_EMAIL`, `E2E_SECOND_PASSWORD` | Dev Admin in the branch tenant |

Without the Dev Reader credentials, tests that need an account skip; the
container smoke runs that way.

Specs get accounts only from the fixtures in
`frontend/apps/webv3/e2e/logging-fixtures.ts`, never by signing in or
registering themselves:

- `readerContext` and `adminContext` hold Dev Reader's and Dev Admin's shared
  sessions. Use them when signing in is incidental and the test's writes are
  safe to repeat. Never log them out, and restore any role change before the
  test ends.
- `learner` is a newly registered, signed-in account for one test, with its ID
  and credentials. Use it when the test needs an untouched account or leaves
  state a rerun would trip over, such as a used-up allowance. The API refuses
  contest creation to accounts younger than a month, so a learner cannot
  create contests.
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

`pnpm --filter webv3 build` creates a local Node build in `.output`. This is a build
check; it does not publish or deploy the application.
