# Local Kubernetes development

The Tilt entrypoint is `k8s/dev/Tiltfile`.

## Flipt

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
