---
title: Identity gateway configuration
description: Callback bearer injection, safe Oathkeeper logging, private frontend session reads and activation gates.
---

# Identity gateway configuration

Read this when changing Oathkeeper's identity callback or the legacy frontends'
private session route. Both repositories must retain the same callback contract.

Activate these manifests only after the API callback, session and logout runtime
and public identity routes are deployed and verified. Production activation
requires separate authorization. Configuration tests do not prove the full v1
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

## Source delivery and production activation

A source merge automatically releases the API and worker to production through
`build-bazel.yaml`, Image Updater and Argo CD. It also publishes development
frontend images; development reconciliation switches the shared callback and
private SSR routes. Treat this as a real production API/worker release.
Production Oathkeeper is managed in the separate GitOps repository. Legacy
webv2, admin and auth publish `prod` only through a separately authorized
`workflow_dispatch` with `publish_production: true`; production tracks those pins.
Keep their existing production images until the identity stack is ready.

The selected order permits compatible source delivery before production gateway
activation. It grants no production gateway, public-route, authv2, SMTP,
branch-tooling adoption, frontend publication, flag, import, deletion or rollback
permission. New production branches and pull-request deploys built from these
sources are **unsupported** until the paired production identity stack is ready:
provisioning enables v2 in test tenants, while the existing production gateway
still authenticates Kratos sessions. Native Reader results can be empty and new
frontends cannot read sessions. Do not create such branches or add `deploy`.
[Production branches](../develop/production-branches.md) describes this hold.

### Source delivery gates

1. Before source delivery, capture these read-only preflight results:
   - Resolve the actual production flag repository's `main` SHA from the Flipt
     Git configuration. Inspect its decision definitions: canonical production
     must decide `identity-v2` explicitly `false`. A missing canonical flag
     selects v2, and errors and timeouts make selection unavailable.
     Git evidence alone does not prove the running evaluator's state.
   - Require zero production branch Deployments, no preview routes beyond the
     catchall, and zero open pull requests labelled `deploy`. Inventory the API,
     frontend, gateway, routing and branch-marker namespaces; do not create a
     deployment or apply a label to test this condition.
   - Capture production API readiness at 2/2, worker readiness and image digests,
     all three legacy frontend production pins, gateway health, and its current
     Kratos `whoami` callback with the session-cookie `only` filter. Retain an
     anonymous known-profile 200 body hash and known-unknown UUID 404 baseline;
     use the same URLs after release. Send no production session or credential.
   - Confirm v1 compatibility with those deployed routes and pins: profile and
     account age retain the existing per-request Kratos reads; the admin list
     performs one bounded provider walk per request with compatible ordering,
     fuzzy matching, totals and paging. The caller change adds no API environment
     requirement. Kratos public URL, cookie domain and trusted-proxy configuration
     remain readiness gates for the new legacy session path before gateway
     activation; the current production gateway does not route that path.
   - Retain accepted runtime/callback/security evidence, unchanged token-bearing
     golden fixtures, deployed clean migration 52 proof in both environments,
     source review and current-head CI. Prove composition with actual current
     main, preserving branch-tooling queries and fixtures; normal compare-and-swap
     merge gates still apply. Record the impending shared development gateway
     switch on the coordination board. Image-pin-only changes do not require
     rebuilding equivalent accepted code.
2. Deliver source through the normal source merge process. Record the **actual
   merge commit**, release workflow run and artifact provenance, API/worker
   digests, and development reconciliation. Do not substitute the PR head for
   the merge commit. Then complete these post-merge gates in order:
   - **Production read-only:** require the new API digest at 2/2, a healthy worker
     traced to the release artifact, unchanged known-profile 200 body hash and
     unknown-profile 404, and unchanged gateway, frontend pins and flag
     definitions. A missing or mismatched result stops progression.
   - **Development provenance:** require healthy API/gateway revisions containing
     the merge and live webv2, admin and auth images built from the actual merge
     commit. Capture build revisions, digests, workflow runs and artifacts before
     any browser smoke; an old image or successful client hydration is not SSR
     proof.
   - **Isolated v2 routing:** use an owned isolated-database branch, native branch
     accounts and JSON login. Require cookie-bearing non-identity GETs proving
     the account/role and Reader result, plus a schema-invalid POST through a
     `cookie_session` route returning the actual branch API's 400. Gateway 401,
     wrong method or wrong backend fails the gate. Invalid input must prevent
     business mutation. The base host must remain guest even with the branch
     cookie and spoofed selection header. Prove physical database placement
     separately; pod configuration, lifecycle markers and backend headers alone
     leave catalog/row placement unverified. Use the runtime-role connection in
     [Jobs operations](../tadoku-api/jobs-operations.md#run-jobs-on-your-development-branch)
     through a loopback port-forward inside a read-only transaction; never exec
     into shared pods, create Jobs, write rows or repair metadata. After
     owner-scoped down, retain and report the marked database; verification
     grants no deletion.
   - **Canonical v1 browser smoke:** reuse the same retained Reader/Admin sessions
     on canonical development hosts; new v2 branches reject those v1 sessions.
     Verify signed-in server-rendered document state with those private storage
     states, reload, Reader navigation/profile and admin denial, and Admin
     list/search/empty/clear and paging where controls exist. Record API-only
     paging separately. Check session retry, deep-link query/fragment return,
     full cookies and app-origin CORS. Only after all smoke preconditions and
     observations pass, publicly log out exactly that pair and require old
     `whoami` cookies to return 401 and protected access to be denied. No new
     shared login, replacement session, admin revocation or role/flag change.
   - Publish sanitized evidence to the source PR and coordination board, retaining
     exact commands, setup, tested revisions, results and unresolved boundaries.
     Physical isolation and live smoke are required; neither is waived by local
     composition evidence.

### Separately held production activation

1. Deploy and verify account session/logout/callback runtime and public identity
   routes under their own gates. Confirm callback credential compatibility,
   banned-account session/logout admission, trusted selection, Kratos public URL,
   cookie domain and proxy configuration. Public routing/package readiness must
   satisfy the account runtime's live gate. Production public-route and authv2
   preparations and branch-tooling adoption stay independently held. Preserve any
   landed email rule and token-header mutator; otherwise the email owner refreshes
   and revalidates later. SMTP/worker preparation retains its own gates.
2. Refresh the existing callback GitOps branch onto actual upstream main,
   preserving public anonymous identity and private session entries alongside
   any landed email rule. Retain final head/base SHAs, Helm render and image
   digests. Rerun the real-image callback matrix, production identity-proxy proof
   and applicable production email-route cases: token-only; token with unrelated,
   expired, revoked and valid cookies; reader-cookie precedence; administrator
   cookie; guest; forged JWT; invalid-token no-fallback; other-route rejection
   and tenant scope. Obtain focused independent review and current-head checks
   plus actual composed-callback proof. Only that final reviewed revision may
   receive separate exact-candidate activation authorization (the D4 gate);
   an earlier preparation head is not an activation candidate.
   Its authorized merge automatically syncs, prunes and self-heals through the
   `tadoku-oathkeeper` Application; no manual sync or apply is implied.
3. Verify the reconciled revision, healthy API/Oathkeeper, guest and authenticated
   private/public `GET /identity/session` and the credential-protected callback.
   Require callback GET with its correct bearer, cookie-bearing mutations,
   banned session/logout access, unrelated cookies, canonical/branch full-cookie
   coexistence, trusted tenant and selected physical database. Verify logs
   without exposing the credential. **Stop if any result is missing or mismatched.**
4. Publish webv2, admin and auth production images only with the owner's normal
   separate publication authorization after all prerequisites above. Record
   selected source SHA and image digests; verify anonymous pages, deep-link login
   with query/fragment return, authenticated navigation, Reader/Admin permissions,
   session retry behavior, logout and protected-page denial through production
   routes using authorized test sessions. **Do not create a production session
   without authorization.**

Before each step, record the exact authorized revisions, deployed image digests,
route/configuration revision, check results and remaining authority. A prior
configuration-only stand-in result is not actual API callback or live-routing
proof. Missing production session authority leaves that check unverified and
blocks claiming completion.

On failure, stop progression and further publications, retain sanitized evidence
and report the failing route/revision and user impact. Reverting source,
republishing an older image or changing live configuration requires explicit
permission unless minimum containment is urgently needed to prevent active harm.
A prepared inverse patch or runbook grants no rollback authority.
