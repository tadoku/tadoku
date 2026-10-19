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

## Production release order

A merge to the source repository's `main` runs `build-bazel.yaml`, publishes the
API production image, and lets Image Updater and Argo CD reconcile it. Treat an
API merge as a production release. The legacy webv2, admin and auth workflows
publish commit and `latest` images on main; they publish `prod` only through an
owner-approved `workflow_dispatch` with `publish_production: true`. Production
tracks `prod`. Publishing any of these apps from a revision containing identity
session reads requires the private session rule and working public identity
routes. Otherwise SSR receives an error and visitors see the retry screen.

Use this prerequisite order; ordinary source-merge permission does not authorize
an activation or a workflow dispatch:

1. Deploy and verify the account session, logout and callback runtime and the
   public identity routes in production. Public routing and package readiness
   must satisfy the account runtime's own live gate before its owner lands it;
   do not waive that gate to impose an accounts-first order. Confirm callback
   credential compatibility, banned-account session/logout admission, trusted
   branch selection and the required public Kratos URL/proxy configuration.
2. Land the production public-route and authv2 GitOps preparations
   (G #238/#239) and email-route preparation (C #242) under their own
   authorization. B #240 remains separately held for its own prerequisites.
   Refresh the callback GitOps branch (#241) onto the resulting main,
   preserving the public anonymous identity route, private session route and
   any landed email rule with its token-header mutator. If #241 lands before
   #242, the email-route owner must refresh and revalidate that later change.
   Retain the final head/base SHAs, Helm render and image digests; rerun the
   real-image callback matrix, production identity-proxy proof and production
   email-route cases under the callback configuration. Include token-only,
   token with unrelated/expired/valid cookies, administrator cookie, guest,
   forged-JWT, other-route rejection and tenant-scope cases. Get independent
   review of that final candidate. After current-head checks and actual
   composed-callback proof pass, obtain separate authorization for that exact
   refreshed production gateway revision. An earlier preparation head cannot
   be the activation candidate.
   Merge the approved GitOps revision only after that authorization: the
   `tadoku-oathkeeper` Application tracks main and automatically syncs, prunes and
   self-heals. No manual sync or apply is required or implied.
3. Verify the reconciled gateway revision, healthy API and Oathkeeper, private
   and public `GET /identity/session` for guests and authenticated users, and
   the actual credential-protected callback. Exercise cookie-bearing mutations,
   banned session/logout access, unrelated cookies and canonical/branch cookie
   coexistence. Confirm callback GET, trusted tenant and selected database;
   verify logs without exposing the credential. Stop if any result is missing
   or mismatched.
4. Only then is the caller-change source merge eligible. Its development-base
   gateway switch is automatic on Argo reconciliation; development webv2,
   admin and auth also switch automatically through their main image publication.
   Complete the separately approved pre-merge proof and post-merge dev smoke. Source review, current-head
   CI and branch-journey gates must also pass.
5. Publish webv2, admin and auth production images only with the owner's normal
   publication approval after all prerequisites above. Record the selected
   source SHA and resulting image digests; verify anonymous pages, deep-link
   login and return URL, authenticated navigation, reader/admin permissions,
   session retry behavior and logout/protected-page denial through production
   routes using approved test sessions. Do not create a production session
   without authorization.

Before each step, record the exact approved revisions, deployed image digests,
route/configuration revision, check results and remaining permissions. A prior
configuration-only stand-in result is not actual API callback or live-routing
proof. Missing production session authority leaves that check unverified and
blocks claiming completion.

On failure, stop progression and further publications, retain sanitized evidence
and report the failing route/revision and its user impact. Keeping the current
frontend production images is the limited option before publication; this
sequence assumes no production-publish freeze. Reverting source, republishing an
older image or changing live configuration requires new explicit authorization
unless minimum containment is urgently needed to prevent active harm. A
prepared inverse patch or revert procedure grants no rollback authority.
