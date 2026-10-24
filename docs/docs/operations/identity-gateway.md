---
title: Identity gateway configuration
description: Callback bearer injection, safe Oathkeeper logging, private frontend session reads, activation gates and the closed session-check latency gate.
---

# Identity gateway configuration

Read this when changing Oathkeeper's identity callback or the frontends' private
session route, or when comparing session-check latency. The development base and the production GitOps repository must
retain the same callback contract. Configuration tests do not prove the full
login, browser, banned-account or branch-isolation journey.

Oathkeeper forwards the complete Cookie and X-Dev-Branch headers through the
branch-aware API hop. Its cookie-session checks always use GET, including when
the original request is a mutation. The internal identity session entry uses
anonymous authentication so the API can admit banned accounts to session and
logout operations without passing through the shared ban gate.

The callback retains its required bearer credential. Kubernetes expands
`OATHKEEPER_AUTHZ_TOKEN` from its existing Secret into
`AUTHENTICATORS_COOKIE_SESSION_CONFIG_ADDITIONAL_HEADERS`; Oathkeeper does not
expand templates in `additional_headers`. Keep Oathkeeper at `log.level: warn`
and `leak_sensitive_values: false`: info and debug startup configuration logs
expose the expanded credential. Do not raise that log level while this override
is configured. The callback verification entry point rejects unsafe effective
config and explicit logging environment overrides, including unresolved sources.
Errors and denied-request warnings remain available.

Run `node k8s/dev/base/e2e/identity-session.cjs <evidence-directory>` to exercise
the pinned development image with a disposable callback stand-in and dummy
credential. Pass a Tadoku GitOps checkout as a second argument to include its
pinned production image. This proves configuration and logging behavior;
actual API sessions, ban admission, trusted routing and live tenant isolation
require the branch journey.

## Production

Production Oathkeeper is managed in the separate GitOps repository. Its
cookie-session authenticator calls the API's session callback, so the
`tadoku_session` cookie authenticates application routes. Public
`/api/internal/identity/*` rules use the anonymous authenticator, so the
identity API owns session reads, logout and CSRF.

A source merge automatically releases the API and worker to production through
`build-bazel.yaml`, Image Updater and Argo CD. Treat it as a real production
release. It also publishes development frontend images; development
reconciliation switches the shared callback and private SSR routes. Legacy
webv2 and admin publish `prod` only through a separately authorized
`workflow_dispatch` with `publish_production: true`. See
[Production branches](../develop/production-branches.md) for branch support.

Change production gateway rules only through a reviewed GitOps pull request,
merged through the normal compare-and-swap process and reconciled by Argo CD.
After reconciliation, verify:

- guest and authenticated `GET /identity/session` through public and private
  routes;
- the callback with its correct bearer, and `401` without it;
- cookie-bearing mutations, and session and logout access for banned accounts;
- unrelated cookies, canonical and branch cookies together, and the trusted
  tenant.

Verify logs without exposing the credential. Authenticated checks need an
authorized production session; without one, report them as unverified.
Reverting a change requires the incident owner's authorization for each scope;
see [Identity incident recovery](./identity-recovery.md#triggers-and-authority).

## Session-check latency

Production session checks were observed passively on 22 October 2026. The
figures are a reference for later comparisons, not a latency target.

| Observation | Timed interval | Successful samples | p50 | p95 |
| --- | --- | --- | --- | --- |
| Production, 177-second window | API session handler | 19 | 0.15 ms | 6.16 ms |
| Production, 600-second window | API session handler | 79 | 0.32 ms | 12.54 ms |

Existing request logs from both production API Pods were read without
generating requests. The figures cover successful
`GET /identity/internal/v1/session` records, using the `latency` field that
`services/tadoku-api/transport/http/observability.go` measures around the
handler, converted from nanoseconds, with nearest-rank quantiles. It covers the
handler through writing the response and excludes Oathkeeper, the network, the
browser and password sign-in. The two windows ran different API images. In the
19-sample window the p95 is the maximum; the 600-second window's maximum was
37.45 ms. The endpoint logs its tenant as `unknown`, so the scope is the
production API deployment, not a proven `tadoku/prod`-only cohort.

Do not derive a capacity or statistical claim from these samples. To compare a
later change, measure the old and new behavior at the same boundary with the
same method.
