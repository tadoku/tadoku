# dev-cli configuration

This directory holds Tadoku's dev-cli setup for the `homelab-talos-dev` development
cluster: `config.yaml` (the real, non-secret configuration), the overlay
workload manifests (`webv2.yaml`, `auth.yaml`, `admin.yaml`, `tadoku-api.yaml`,
`tadoku-worker.yaml`),
the tenant lifecycle task manifests (`tenant.yaml`, `tenant-teardown.yaml`,
`worker-override-set.yaml`, `worker-override-clear.yaml`), the fixture loader
(`tenant.sh`, `seed.sql`), and the Bazel metadata that builds them. The retained
`database.yaml` and `migrate.yaml` support explicitly configured isolated
database work; ordinary branches use a test tenant in the shared database.

Use dev-cli v0.7.0 or newer. `dev up --owner <owner>` automatically provisions
and seeds `tadoku/<route>` before startup, including frontend-only work. Select
the main hostname before admin or auth API writes. `dev down` and expired-route
cleanup stop overlays and remove that tenant's application and provider data.

- [Development environment](../docs/docs/develop/environment.md): install dev-cli,
  run, open, seed and clean up a branch.
- [Development base](../docs/docs/operations/development-base.md): the Argo CD
  base these overlays route around.
- [Acceptance gates](acceptance.md): live checks for routing and sync changes.
