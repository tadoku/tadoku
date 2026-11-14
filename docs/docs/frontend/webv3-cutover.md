---
title: Webv3 cutover runbook
description: How tadoku.app moves from Webv2 to Webv3 after the flagged release, what to change in production before and during the host switch, how to verify it and how to roll it back.
sidebar_position: 3
---

# Webv3 cutover runbook

Read this before you prepare, run or roll back the switch of `tadoku.app` from
Webv2 to Webv3, or when a Webv3 alert fires after it.

The production manifests live in the maintainer's GitOps repository; every step
there is a reviewed pull request that Argo CD syncs. Never change production by
hand.

## Phases

1. **Flagged release.** Webv3 serves `next.tadoku.app`, and accounts in the
   `release-webv3-access` segment are sent there from Webv2, as
   [Webv3 flagged release](webv3-flagged-release.md) describes. This is the step
   before the host switch: grow the segment until the accounts on it use Webv3
   without problems. Keep the `release-webv3` flag through and after the switch;
   it is the fastest way to send accounts back to Webv2 if a rollback restores
   Webv2.
2. **Before the switch.** Make Webv3 able to carry all traffic.
3. **Host switch.** Route `tadoku.app` to Webv3 in one change.
4. **Afterwards.** Keep Webv2 running until the rollback window closes, then
   retire it and redirect `next.tadoku.app`.

## Before the switch

- **Replicas and disruption budget.** Webv3 runs two replicas spread across
  nodes, with a PodDisruptionBudget of `minAvailable: 1` and a 5-second
  `preStop` delay, so a node drain, an eviction or a rollout never leaves it
  without a ready pod. Keep all three while Webv3 serves `tadoku.app`.
- **Environment.** Production already sets `TADOKU_WEB_URL=https://tadoku.app`
  and `TADOKU_API_URL=https://tadoku.app/api/internal`, so `tadoku.app` is
  Webv3's web host and becomes indexable once it serves Webv3. Nothing else
  changes; [Webv3 production container](webv3-container.md) lists the variables.
- **Health checks.** Webv2 answers `/api/health`; Webv3 answers `/healthz`.
  After the switch, `https://tadoku.app/api/health` reaches Webv3 and answers
  404. Point every monitor or probe of that path at `/healthz` in the same change.
  The homelab blackbox probe reads `https://tadoku.app/` and needs no change.
- **Canary.** `.github/workflows/webv3-production-canary.yaml` signs in on the
  host Webv3 serves for everyone: `WEBV3_BASE_URL` is `https://next.tadoku.app`
  until `tadoku.app` reaches Webv3, and `https://tadoku.app` after. The host
  moves in GitOps pull requests and the base URL in a pull request here, so the
  two never land together, and a run against the host Webv2 serves goes red
  without any Webv3 fault. Change the base URL only where the host switch and
  the rollback below say, so it always names a host Webv3 serves;
  `next.tadoku.app` serves Webv3 throughout.
- **Request IDs.** The tadoku HTTPS listeners and the internal Gateway must keep
  an incoming `X-Request-Id`, or Webv3's and Tadoku API's log lines will not
  join on `correlation_id`; see [Webv3 server logs](webv3-server-logs.md).

## Host switch

Two GitOps pull requests, with a check between them. Both HTTPRoutes are
attached to `tadoku.app` in between; for the same path match, Envoy Gateway
routes to the route created first, which is `frontend-webv2`, so traffic stays
on Webv2 until its route is removed.

1. **Admit and attach Webv3.**
   - Add `tdk-prod-frontend-webv3` to the `https-apex` listener's
     `allowedRoutes` namespace list on Gateway `tadoku`. Without it the listener
     rejects Webv3's route, and removing Webv2's would leave `tadoku.app`
     without a page route.
   - Add `tadoku.app` and the `https-apex` listener to the `frontend-webv3`
     HTTPRoute. Keep the route's name: the `TadokuWebv3ServerErrorBurst` alert
     reads the edge access logs of the `frontend-webv3` HTTPRoute, so
     `tadoku.app` must stay on it. The Oathkeeper HTTPRoute keeps serving
     `/api/internal`, a longer path prefix.
   - **No edge `X-Robots-Tag` header.** The `frontend-webv3` HTTPRoute must
     not set `X-Robots-Tag`: HTTPRoute filters cannot match a hostname, so with
     `tadoku.app` on the same route every page would be noindex. Webv3 itself sends
     `X-Robots-Tag: noindex` from every host that is not its web host, as
     [headers, indexing and sharing](webv3-indexing.md) describes, so
     `next.tadoku.app` stays out of search results without it. The route that
     serves `tadoku.app` must add no `X-Robots-Tag`.
2. **Check admission.** After Argo CD syncs, both conditions must be `True` for
   the `https-apex` parent before you go on:

   ```sh
   kubectl -n tdk-prod-frontend-webv3 get httproute frontend-webv3 \
     -o jsonpath='{range .status.parents[?(@.parentRef.sectionName=="https-apex")].conditions[*]}{.type}={.status}{"\n"}{end}'
   ```

   It prints `Accepted=True` and `ResolvedRefs=True`. Anything else, or no
   output, means the route is not admitted on `https-apex`; fix that first.
3. **Remove Webv2's route.** Delete the `frontend-webv2` HTTPRoute, which serves
   only `tadoku.app`, and keep the `www-redirect` route beside it. `tadoku.app`
   now reaches Webv3.
4. **Point the canary at `tadoku.app`.** Only after step 3 has synced and
   `curl -sI https://tadoku.app/` shows Webv3's headers (see
   [Verify](#verify)), change `WEBV3_BASE_URL` in the canary workflow, and the
   host its Notify message names, to `https://tadoku.app` in a pull request
   here. Until it merges, the canary keeps checking Webv3 on `next.tadoku.app`.

Leave the `www.tadoku.app` redirect and the account and admin hosts as they are.

## Verify

After Argo CD syncs:

- `curl -sI https://tadoku.app/` answers 200 with Webv3's headers,
  `X-Content-Type-Options: nosniff` and a `Content-Security-Policy-Report-Only`,
  which Webv2 does not send, and no `X-Robots-Tag`.
- `https://tadoku.app/robots.txt` allows crawlers, a page links its canonical
  URL on `https://tadoku.app`, and `https://tadoku.app/healthz` answers 200.
- `next.tadoku.app` answers `X-Robots-Tag: noindex` and its `robots.txt` answers
  `Disallow: /`.
- An `/api/internal` read and signing in and out through the account host work.
- Run the canary's signed-in journey against `https://tadoku.app` with a
  `workflow_dispatch` once step 4 has merged, and watch the edge 5xx ratio of the
  `frontend-webv3` HTTPRoute and Webv3's `ssr_error` and `upstream_failure` lines
  in Loki for the next rollout.

## Rollback

- **Host switch.** Mirror the switch, again in two pull requests with a check
  between them, after moving the canary. A restored `frontend-webv2` HTTPRoute
  is created after `frontend-webv3`, so Webv3 keeps the traffic until it leaves
  `https-apex`.
  1. Point the canary's `WEBV3_BASE_URL`, and the host its Notify message names,
     back at `https://next.tadoku.app` in a pull request here, and merge it
     before the GitOps changes. `next.tadoku.app` keeps serving Webv3, so the
     canary never reaches Webv2. A canary run that started or queued before the
     merge keeps the old URL, so wait until no `webv3-production-canary` run
     from an earlier commit is queued or in progress (cancel it if it must not
     finish) before you go on.
  2. Restore the `frontend-webv2` HTTPRoute. Its namespace is still on the
     listener's list.
  3. Check that it is admitted for the `https-apex` parent:

     ```sh
     kubectl -n tdk-prod-frontend-webv2 get httproute frontend-webv2 \
       -o jsonpath='{range .status.parents[?(@.parentRef.sectionName=="https-apex")].conditions[*]}{.type}={.status}{"\n"}{end}'
     ```

     It must print `Accepted=True` and `ResolvedRefs=True`.
  4. Remove `tadoku.app` and the `https-apex` listener from the `frontend-webv3`
     HTTPRoute, and restore its `X-Robots-Tag` filter for `next.tadoku.app`.
     `tadoku.app` returns to Webv2. Webv3's namespace can then leave the
     listener's list.

  This needs Webv2 still deployed at its last image, so do not retire it while a
  rollback is possible. If Webv3 itself is at fault, also empty the
  `release-webv3-access` segment, or flagged accounts are sent back to it.
- **A broken release.** While Webv2 is still deployed, nothing rolls Webv3 back
  automatically; rolling back is the owner's decision. Once Webv2 is retired, a
  Webv3 5xx burst or a red canary rolls Webv3 back to the previous image
  automatically. The homelab's Tadoku web alerts runbook gives the rollback for
  each phase: the `release-webv3-access` segment before the host switch, the
  previous image after it.

## When a Webv3 alert fires

The Webv3 alerts open an SRE incident that investigates and reports to the
owner with its evidence. While Webv2 is still deployed, none of them rolls
Webv3 back; after that, see [A broken release](#rollback).

- **Check Tadoku API's health first.** An API outage makes Webv3's resource pages
  answer 503 ([status codes](webv3-status-codes.md)), and those 503s count toward
  `TadokuWebv3ServerErrorBurst`, so an API failure can fire it for a healthy
  Webv3 release. Rolling Webv3 back does not fix the API.
- **A red canary.** Check that its `WEBV3_BASE_URL` names the host Webv3 serves
  for everyone; a run against the host Webv2 serves fails without a Webv3 fault.
- **A gateway 504 with no Webv3 log line.** Nothing on the production path times
  out within Webv3's 8-second upstream deadline, so a 504 with no
  `upstream_failure` or `ssr_error` line means Webv3 did not answer the edge, not
  that the deadline was skipped. Read that request's edge access log: `UT` means
  the edge gave up on Webv3, so read Webv3's stderr, where framework errors go,
  around that time; `UC` means an established connection to the pod was
  terminated, so check restarts, OOM kills and rollouts at that time; `UF` means
  the edge could not establish a connection to the pod, so read the line's
  `upstream_transport_failure_reason` and `response_code_details`, such as a
  refused connection or a TLS error, and the pod's readiness and endpoints; a 504
  on the API route means the API timed out,
  which Webv3 logs as an `http_status` 504. A single one is noise; take a pattern
  to the API owners with the request ID.

## After the rollback window

- Retire Webv2's Deployment and HTTPRoute, and the `/api/health` entry of the
  production config contract. Removing Webv2's service directory from the
  GitOps repository ends the rollout: from then on a Webv3 5xx burst or a red
  canary rolls Webv3 back automatically.
- Redirect `next.tadoku.app` with a route of its own: move `next.tadoku.app`
  and the `https-next` listener off the `frontend-webv3` HTTPRoute onto a new
  HTTPRoute for `next.tadoku.app` only, whose rule answers a `301` to
  `tadoku.app` as the `www-redirect` route does. Keep `frontend-webv3`, under
  that name, as the `tadoku.app` backend route; a redirect on it would send
  `tadoku.app` to itself. The identity API keeps the `next.` host on its app
  allowlist, so sign-ins that return there keep working until then.
