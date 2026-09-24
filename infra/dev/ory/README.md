# Shared Ory test fixtures

This directory provides shared fixtures, not deployment configuration.

Keep `identity.default.schema.json`, `namespaces.keto.ts` and `BUILD.bazel`:
the native API's `internal/testkratos` and `internal/testketo` helpers load them
through Bazel runfiles for real authentication/authorization tests.

The active **development-only** provider manifests live in
[`k8s/dev/base`](../../../k8s/dev/base/README.md). Shared identity/role seeding
uses `make dev-seed`; branch migration/seeding uses the
[DevCLI tasks](../../../.dev/README.md).
