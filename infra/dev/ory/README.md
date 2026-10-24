# Shared Ory test fixtures

This directory provides shared fixtures, not deployment configuration.

`internal/testketo` loads `namespaces.keto.ts` through Bazel runfiles for real
authorization tests.

The active **development-only** provider manifests live in
`k8s/dev/base/` (see [Development base](../../../docs/docs/operations/development-base.md)). Shared identity/role seeding
uses `make dev-seed`; branch migration/seeding uses the
[dev-cli tasks](../../../docs/docs/develop/development-data.md).
