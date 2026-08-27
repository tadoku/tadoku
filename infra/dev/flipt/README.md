# Flipt development

The local Flipt deployment is deliberately separate from production. It has a
single `local` environment backed only by Flipt's in-memory storage. It has no
Git remote or credential, persistent volume, database, or cache, and its pod is
not allowed general internet egress.

Local flag state is disposable. Restarting or replacing the Flipt pod may
return it to an empty state. Developers create the flags they need through the
Oathkeeper-protected local Flipt UI; the repository does not seed or promote
local flag definitions. Tadoku applications must remain usable with their
compiled safe defaults when that state is empty or unavailable.

Prometheus-format metrics are available to the cluster monitoring workload at
`/metrics`. The Service is cluster-internal, and NetworkPolicies accept Flipt
HTTP traffic only from the Oathkeeper and monitoring namespaces. Flipt egress
is limited to cluster DNS; in particular, no GitHub egress is allowed.

## Operator access

The local Flipt operator UI uses the dedicated host configured as `hosts.flipt`
in the ignored `tilt_config.json`. The host resolves to Oathkeeper, which
requires a Kratos session and sends only the Kratos subject to authz-api's
fixed Tadoku-admin check. There is no anonymous authenticator on this rule.
Authz-api returns 200 for an allowed administrator, 403 for a denial, and 503
when Keto is unavailable; Oathkeeper's `remote_json` authorizer requires the
200 response to allow the request.

Tilt initializes the callback credential automatically. It generates a random
token when the active cluster does not already have one, then creates matching
`dev-oathkeeper-authz` Secrets for Oathkeeper and authz-api. No credential needs
to be added to `tilt_config.json` before running `tilt up`.

## Application access

Backend application SDKs use the internal base URL
`http://oathkeeper-proxy.default:4455/flipt`. They first exchange their
projected Kubernetes service-account token at
`/token-exchange/flipt-evaluation/immersion-api`, then send the returned
Oathkeeper-signed JWT as a bearer token. The evaluation rule accepts only
`GET` on the `default` namespace snapshot endpoint and strips `/flipt` before
forwarding to Flipt. Management paths and the streaming endpoint are not
included.

## Production feature-state contract

Production Flipt uses the private `tadoku/feature-flags` repository as its only
state store and commits ordinary changes directly to `main`. The branch allows
direct pushes without a pull-request requirement; branch protection is not
part of the accepted production contract.

The repository is initialized with a normal non-Flipt commit. Do not add sample
features, environment branches, pull-request templates, lifecycle manifests,
or bootstrap flag state. Creating the first production feature remains an
explicit operator action in Flipt.
