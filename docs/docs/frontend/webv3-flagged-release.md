---
title: Webv3 flagged release
description: How the release-webv3 flag sends named accounts from Webv2 to Webv3 before the host switch, what never redirects, and how to grant or turn off access.
sidebar_position: 3
---

# Webv3 flagged release

Read this when you change the Webv2 redirect, grant an account Webv3 or turn
the release off.

Before the host switch, Webv3 is released to named accounts with the
`release-webv3` flag. While the flag is on for an account, Webv2 answers that
account's page requests with an uncached `302` to the same path and query on
Webv3. Everyone else keeps Webv2. `frontend/apps/webv2/pages/_app.tsx` makes the
decision on the server, so it applies to full page loads; a Webv2 tab already
open moves over on its next load. API routes, health checks and assets do not
pass through it, a page Webv2 does not have stays a Webv2 404, and the account
and API hosts never redirect. Webv3 never redirects to Webv2.

`WEBV3_URL` on Webv2 selects the target and is read at runtime. Without it, the
target is the `next.` sibling of Webv2's home URL. A branch host redirects to
the same branch's Webv3 host. [Webv2 configuration](webv2.md#configuration) lists it
with the other variables.

Access is the account IDs in the `release-webv3-access` segment, changed through
the Flipt operator interface as [Feature flags](../architecture/feature-flags.md)
describes:

- **To release to an account,** add its ID to the segment's `entityId`
  constraint. Keep the flag's `enabled` value off: it is the answer for every
  account outside the segment, so turning it on redirects everyone.
- **To turn the release off,** remove the ID, or empty the list to stop it for
  every account. Webv2 serves the account again once the API loads its next
  Flipt snapshot (`API_FLIPT_UPDATE_INTERVAL`). No deployment is needed.

`frontend/apps/webv3/e2e/webv2-release-redirect.spec.ts` covers this on a branch
tenant. Set `E2E_FLIPT_URL` to a port-forward of the development Flipt, along
with the [credentialed variables](webv3-testing.md#credentialed-runs-and-fixture-accounts).
It grants a new account access in that tenant's namespace only.
