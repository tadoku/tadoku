# dev-cli configuration

This directory holds Tadoku's dev-cli setup for the `homelab-dev` development
cluster: `config.yaml` (the real, non-secret configuration), the overlay
workload manifests (`webv2.yaml`, `auth.yaml`, `admin.yaml`, `tadoku-api.yaml`),
the branch task manifests (`database.yaml`, `migrate.yaml`, `seed.yaml`,
`seed.sql`) and the Bazel metadata that builds them.

- [Development environment](../docs/docs/develop/environment.md): install dev-cli,
  run, open, seed and clean up a branch.
- [Development base](../docs/docs/operations/development-base.md): the Argo CD
  base these overlays route around.
- [Acceptance gates](acceptance.md): live checks for routing and sync changes.
