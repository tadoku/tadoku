---
title: HTTP email gateway fixture
description: Verify cookie and bearer routing through checksum-pinned Oathkeeper, local identity accounts and the real Tadoku API router, including rendered production gateway rules.
sidebar_position: 6.5
---

# HTTP email gateway fixture

Read this when changing or running the HTTP E2E email gateway fixture or checking
rendered Oathkeeper rules against disposable local services. General fixture rules
are in [HTTP end-to-end tests](./http-e2e.md).

`TestEmailGatewayCookieAndBearer` runs the checked-in development rules with
checksum-pinned Oathkeeper v26.2.0, local PostgreSQL and the production
router on the restricted role. It seeds disposable native admin and reader
accounts with `bcrypt.MinCost` hashes, then logs in through the real gateway
with the unchanged three-second client. Cookie checks call the API's session
callback using the deployment's additional-header environment override with a
synthetic bearer. JWT signing and verification remain real.

Assertions cover email list/detail, reader-cookie token precedence, invalid
tokens despite admin cookies, unrelated/forged/expired/revoked/mixed cookies,
guest access outside email, protected-route denial, wrong-deployment `421`,
revoked tokens, public identity access, logout and credential absence from logs.
Native session expiry is fixture-only SQL; no shared provider is used.

To check production routing, render the paired GitOps checkout's Oathkeeper
chart (version 0.60.1) into a scratch file and pass its absolute path through
`TADOKU_TEST_EMAIL_GATEWAY_RENDER`. The render must include the public
`tadoku.app` identity rule, preview identity rules, email rules and this callback
configuration. These public identity rules
are a prerequisite for gateway login; a render without them cannot run this
mode. Rendering does not activate those rules in any deployment. The same test
reads the rendered ConfigMaps and uses checksum-pinned Oathkeeper v25.4.0. Adaptations
are limited to loopback upstream/callback addresses, listener ports, ephemeral
signing keys and the synthetic callback bearer. Matching, authenticator options,
mutators and logging remain from the checked-in configuration or render.
The second router used for the wrong-deployment case is test-only fixture
instrumentation; no test route is added to production.
It checks forwarded HTTPS on `tadoku.app` and a preview host, plus HTTP denial.
Both binaries are fetched and checksum-verified by Bazel on cold machines.

```sh
helm template oathkeeper oathkeeper --repo https://k8s.ory.sh/helm/charts \
  --version 0.60.1 --namespace tdk-prod-oathkeeper \
  --values "$GITOPS_CHECKOUT/platform/oathkeeper/values.yaml" > "$SCRATCH/oathkeeper.yaml"
TADOKU_TEST_EMAIL_GATEWAY_RENDER="$SCRATCH/oathkeeper.yaml" \
  bazel test //services/tadoku-api/e2e:e2e_test \
  --test_filter=TestEmailGatewayCookieAndBearer --cache_test_results=no --test_output=all
```

Use the disposable PostgreSQL and Valkey settings above. This local production
configuration check deploys nothing and does not replace public branch-host
verification after the rules are released.
