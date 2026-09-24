# Shared Ory test fixtures

This directory provides shared fixtures, not deployment configuration.

Keep `identity.default.schema.json`, `namespaces.keto.ts` and `BUILD.bazel`:
Tadoku API's `internal/testkratos` and `internal/testketo` helpers load them
through Bazel runfiles for real authentication/authorization tests.

The active **development-only** provider manifests live in
`k8s/dev/base/` (see [Development base](../../../docs/docs/operations/development-base.md)). Shared identity/role seeding
uses `make dev-seed`; branch migration/seeding uses the
[dev-cli tasks](../../../docs/docs/develop/environment.md).
