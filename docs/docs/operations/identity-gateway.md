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
is configured. Errors and denied-request warnings remain available.

Run `node k8s/dev/base/e2e/identity-session.cjs <evidence-directory>` to exercise
the pinned development image with a disposable callback stand-in and dummy
credential. Pass a Tadoku GitOps checkout as a second argument to include its
pinned production image. This proves configuration and logging behavior;
actual API sessions, ban admission, trusted routing and live tenant isolation
require the branch journey.
